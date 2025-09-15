import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import type { SettingsDTO, SLAPolicy } from '@/types/dto';
import { 
  GetSettings,
  SaveSettings,
  RefreshAIModels,
  SetSLAPolicy
} from '../../../wailsjs/go/main/App';
import { APP_CONFIG } from '@/config/app-config';

/**
 * Settings API Adapter - Infrastructure implementation of SettingsRepository
 * This handles settings operations while conforming to Clean Architecture
 */
export class SettingsApiAdapter implements SettingsRepository {
  async getSettings(): Promise<SettingsDTO> {
    try {
      // Check if Wails is available
      if (!window || !window['go'] || !window['go']['main'] || !window['go']['main']['App']) {
        throw new Error('Wails runtime not available - application may not be fully initialized');
      }
      
      return await GetSettings();
    } catch (error) {
      throw this.handleError(error, 'Failed to get settings');
    }
  }

  async saveSettings(settings: SettingsDTO): Promise<void> {
    try {
      await SaveSettings(JSON.stringify(settings));
    } catch (error) {
      throw this.handleError(error, 'Failed to save settings');
    }
  }

  async refreshAIModels(provider: string, apiKey: string): Promise<void> {
    try {
      await RefreshAIModels(provider, apiKey);
    } catch (error) {
      throw this.handleError(error, 'Failed to refresh AI models');
    }
  }

  async setSLAPolicy(policy: SLAPolicy): Promise<void> {
    try {
      const policyJson = JSON.stringify(policy);
      await SetSLAPolicy(policyJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to set SLA policy');
    }
  }

  async resetToDefaults(): Promise<void> {
    try {
      // Create default settings object using APP_CONFIG
      const defaultSettings: SettingsDTO = {
        selectedProvider: APP_CONFIG.ai.defaultSettings.provider,
        openAIAPIKey: '',
        geminiAPIKey: '',
        openRouterAPIKey: '',
        localAIAPIKey: '',
        openaiModel: APP_CONFIG.ai.defaultSettings.model,
        geminiModel: 'gemini-pro',
        temperature: APP_CONFIG.ai.defaultSettings.temperature,
        maxTokens: APP_CONFIG.ai.defaultSettings.maxTokens,
        maxContextSize: APP_CONFIG.context.streaming.CHUNK_SIZE,
        maxFilesInContext: 20,
        includeDependencies: false,
        includeTests: false,
        autoFormat: true,
        includeComments: true,
        useGitignore: true,
        useCustomIgnore: true,
        customIgnoreRules: ''
      };
      
      await this.saveSettings(defaultSettings);
    } catch (error) {
      throw this.handleError(error, 'Failed to reset settings to defaults');
    }
  }

  async exportSettings(): Promise<string> {
    try {
      const settings = await this.getSettings();
      return JSON.stringify(settings, null, 2);
    } catch (error) {
      throw this.handleError(error, 'Failed to export settings');
    }
  }

  async importSettings(settingsJson: string): Promise<void> {
    try {
      const settings: SettingsDTO = JSON.parse(settingsJson);
      
      // Validate the imported settings
      const validation = await this.validateSettings(settings);
      if (!validation.isValid) {
        throw new Error(`Invalid settings: ${validation.errors.join(', ')}`);
      }
      
      await this.saveSettings(settings);
    } catch (error) {
      if (error instanceof SyntaxError) {
        throw this.handleError(error, 'Failed to parse settings JSON');
      }
      throw this.handleError(error, 'Failed to import settings');
    }
  }

  async validateSettings(settings: SettingsDTO): Promise<{
    isValid: boolean;
    errors: string[];
    warnings: string[];
  }> {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    try {
      // Validate AI provider settings
      if (!settings.selectedProvider) {
        errors.push('AI provider is required');
      }
      
      const apiKey = settings.openAIAPIKey || settings.geminiAPIKey || settings.openRouterAPIKey || settings.localAIAPIKey;
      if (!apiKey && settings.selectedProvider !== 'localai') {
        warnings.push('API key is missing for external AI provider');
      }
      
      if (!settings.openaiModel && !settings.geminiModel) {
        errors.push('Model is required');
      }
      
      // Validate temperature range
      if (settings.temperature !== undefined) {
        if (settings.temperature < 0 || settings.temperature > 2) {
          errors.push('Temperature must be between 0 and 2');
        }
      }
      
      // Validate context limits
      if (settings.maxContextSize !== undefined) {
        if (settings.maxContextSize < 1024 || settings.maxContextSize > 200000) {
          errors.push('Context size must be between 1024 and 200000');
        }
      }
      
      return {
        isValid: errors.length === 0,
        errors,
        warnings
      };
    } catch (error) {
      return {
        isValid: false,
        errors: ['Failed to validate settings: ' + (error instanceof Error ? error.message : String(error))],
        warnings
      };
    }
  }

  // Helper methods for settings management
  async updatePartialSettings(partialSettings: Partial<SettingsDTO>): Promise<void> {
    try {
      const currentSettings = await this.getSettings();
      const updatedSettings = { ...currentSettings, ...partialSettings };
      
      // Validate before saving
      const validation = await this.validateSettings(updatedSettings);
      if (!validation.isValid) {
        throw new Error(`Invalid settings: ${validation.errors.join(', ')}`);
      }
      
      await this.saveSettings(updatedSettings);
    } catch (error) {
      throw this.handleError(error, 'Failed to update settings');
    }
  }

  async getSettingValue<K extends keyof SettingsDTO>(key: K): Promise<SettingsDTO[K]> {
    try {
      const settings = await this.getSettings();
      return settings[key];
    } catch (error) {
      throw this.handleError(error, `Failed to get setting value: ${String(key)}`);
    }
  }

  async setSettingValue<K extends keyof SettingsDTO>(key: K, value: SettingsDTO[K]): Promise<void> {
    try {
      await this.updatePartialSettings({ [key]: value } as Partial<SettingsDTO>);
    } catch (error) {
      throw this.handleError(error, `Failed to set setting value: ${String(key)}`);
    }
  }

  // Private helper methods
  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    
    // Check if this is a domain error from backend
    if (message.startsWith('domain_error:')) {
      try {
        const domainErrorJson = message.substring('domain_error:'.length);
        const domainError = JSON.parse(domainErrorJson);
        
        const structuredError = new Error(`${context}: ${domainError.message}`);
        (structuredError as any).code = domainError.code;
        (structuredError as any).recoverable = domainError.recoverable;
        (structuredError as any).context = domainError.context;
        (structuredError as any).cause = domainError.cause;
        
        return structuredError;
      } catch (parseErr) {
        console.error('Failed to parse domain error:', parseErr);
      }
    }
    
    return new Error(`${context}: ${message}`);
  }
}