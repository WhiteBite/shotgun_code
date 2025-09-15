import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import type { SettingsDTO, SLAPolicy } from '@/types/dto';
import { 
  GetSettings,
  SaveSettings,
  RefreshAIModels,
  SetSLAPolicy
} from '../../../wailsjs/go/main/App';

/**
 * Settings API Adapter - Infrastructure implementation of SettingsRepository
 * This handles settings operations while conforming to Clean Architecture
 */
export class SettingsApiAdapter implements SettingsRepository {
  async getSettings(): Promise<SettingsDTO> {
    try {
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
      // Create default settings object
      const defaultSettings: SettingsDTO = {
        aiProvider: 'openai',
        apiKey: '',
        model: 'gpt-3.5-turbo',
        temperature: 0.7,
        maxTokens: 2048,
        contextWindow: 4096,
        streamingEnabled: false,
        autoSave: true,
        theme: 'system',
        language: 'en',
        gitIntegration: true,
        guardrailsEnabled: true,
        memoryLimit: 512,
        debugMode: false
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
      if (!settings.aiProvider) {
        errors.push('AI provider is required');
      }
      
      if (!settings.apiKey && settings.aiProvider !== 'local') {
        warnings.push('API key is missing for external AI provider');
      }
      
      if (!settings.model) {
        errors.push('Model is required');
      }
      
      // Validate temperature range
      if (settings.temperature !== undefined) {
        if (settings.temperature < 0 || settings.temperature > 2) {
          errors.push('Temperature must be between 0 and 2');
        }
      }
      
      // Validate token limits
      if (settings.maxTokens !== undefined) {
        if (settings.maxTokens < 1 || settings.maxTokens > 128000) {
          errors.push('Max tokens must be between 1 and 128000');
        }
      }
      
      if (settings.contextWindow !== undefined) {
        if (settings.contextWindow < 1024 || settings.contextWindow > 200000) {
          errors.push('Context window must be between 1024 and 200000');
        }
      }
      
      // Validate memory limit
      if (settings.memoryLimit !== undefined) {
        if (settings.memoryLimit < 128 || settings.memoryLimit > 8192) {
          warnings.push('Memory limit should be between 128MB and 8GB');
        }
      }
      
      // Validate theme
      if (settings.theme && !['light', 'dark', 'system'].includes(settings.theme)) {
        warnings.push('Unknown theme value, will use system default');
      }
      
      // Validate language
      if (settings.language && !['en', 'es', 'fr', 'de', 'zh', 'ja'].includes(settings.language)) {
        warnings.push('Unsupported language, will use English');
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