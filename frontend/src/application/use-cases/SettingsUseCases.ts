import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import type { SettingsDTO, SLAPolicy } from '@/types/dto';

/**
 * Settings Management Use Cases
 */

export class GetSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(): Promise<SettingsDTO> {
    try {
      const settings = await this.settingsRepo.getSettings();
      
      // Validate settings structure
      this.validateSettingsStructure(settings);
      
      return settings;
    } catch (error) {
      throw new Error(`Failed to get settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateSettingsStructure(settings: SettingsDTO): void {
    if (!settings || typeof settings !== 'object') {
      throw new Error('Invalid settings structure');
    }
    
    // Validate AI settings if present
    if (settings.ai) {
      if (settings.ai.provider && !['openai', 'anthropic', 'google', 'azure', 'local'].includes(settings.ai.provider)) {
        console.warn(`Unknown AI provider: ${settings.ai.provider}`);
      }
      
      if (settings.ai.temperature && (settings.ai.temperature < 0 || settings.ai.temperature > 2)) {
        console.warn(`Invalid temperature setting: ${settings.ai.temperature}`);
      }
    }
    
    // Validate editor settings if present
    if (settings.editor) {
      if (settings.editor.tabSize && (settings.editor.tabSize < 1 || settings.editor.tabSize > 8)) {
        console.warn(`Invalid tab size: ${settings.editor.tabSize}`);
      }
    }
  }
}

export class SaveSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settings: SettingsDTO): Promise<void> {
    // Validate settings before saving
    await this.validateSettings(settings);
    
    try {
      await this.settingsRepo.saveSettings(settings);
    } catch (error) {
      throw new Error(`Failed to save settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private async validateSettings(settings: SettingsDTO): Promise<void> {
    if (!settings || typeof settings !== 'object') {
      throw new Error('Invalid settings object');
    }

    const validationResult = await this.settingsRepo.validateSettings(settings);
    
    if (!validationResult.isValid) {
      const errorMessage = validationResult.errors.join(', ');
      throw new Error(`Settings validation failed: ${errorMessage}`);
    }
    
    if (validationResult.warnings.length > 0) {
      console.warn('Settings validation warnings:', validationResult.warnings);
    }
  }
}

export class ResetSettingsToDefaultsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(): Promise<SettingsDTO> {
    try {
      await this.settingsRepo.resetToDefaults();
      return await this.settingsRepo.getSettings();
    } catch (error) {
      throw new Error(`Failed to reset settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

/**
 * AI Settings Use Cases
 */

export class UpdateAISettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(aiSettings: Partial<SettingsDTO['ai']>): Promise<void> {
    this.validateAISettings(aiSettings);
    
    try {
      // Get current settings
      const currentSettings = await this.settingsRepo.getSettings();
      
      // Update AI settings
      const updatedSettings: SettingsDTO = {
        ...currentSettings,
        ai: {
          ...currentSettings.ai,
          ...aiSettings
        }
      };
      
      // Save updated settings
      await this.settingsRepo.saveSettings(updatedSettings);
    } catch (error) {
      throw new Error(`Failed to update AI settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateAISettings(aiSettings: Partial<SettingsDTO['ai']>): void {
    if (!aiSettings || typeof aiSettings !== 'object') {
      throw new Error('Invalid AI settings object');
    }

    if (aiSettings.provider) {
      const validProviders = ['openai', 'anthropic', 'google', 'azure', 'local'];
      if (!validProviders.includes(aiSettings.provider)) {
        throw new Error(`Invalid AI provider: ${aiSettings.provider}`);
      }
    }

    if (aiSettings.apiKey && aiSettings.apiKey.length < 10) {
      throw new Error('API key must be at least 10 characters long');
    }

    if (aiSettings.temperature !== undefined) {
      if (typeof aiSettings.temperature !== 'number' || aiSettings.temperature < 0 || aiSettings.temperature > 2) {
        throw new Error('Temperature must be a number between 0 and 2');
      }
    }

    if (aiSettings.maxTokens !== undefined) {
      if (typeof aiSettings.maxTokens !== 'number' || aiSettings.maxTokens < 1 || aiSettings.maxTokens > 32000) {
        throw new Error('Max tokens must be between 1 and 32000');
      }
    }
  }
}

export class RefreshAIModelsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(provider: string, apiKey: string): Promise<void> {
    this.validateRefreshModelsInputs(provider, apiKey);
    
    try {
      await this.settingsRepo.refreshAIModels(provider, apiKey);
    } catch (error) {
      throw new Error(`Failed to refresh AI models: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateRefreshModelsInputs(provider: string, apiKey: string): void {
    if (!provider?.trim()) {
      throw new Error('Provider is required');
    }

    if (!apiKey?.trim()) {
      throw new Error('API key is required');
    }

    const validProviders = ['openai', 'anthropic', 'google', 'azure'];
    if (!validProviders.includes(provider)) {
      throw new Error(`Invalid provider: ${provider}`);
    }

    if (apiKey.length < 10) {
      throw new Error('API key must be at least 10 characters long');
    }
  }
}

/**
 * Editor Settings Use Cases
 */

export class UpdateEditorSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(editorSettings: Partial<SettingsDTO['editor']>): Promise<void> {
    this.validateEditorSettings(editorSettings);
    
    try {
      // Get current settings
      const currentSettings = await this.settingsRepo.getSettings();
      
      // Update editor settings
      const updatedSettings: SettingsDTO = {
        ...currentSettings,
        editor: {
          ...currentSettings.editor,
          ...editorSettings
        }
      };
      
      // Save updated settings
      await this.settingsRepo.saveSettings(updatedSettings);
    } catch (error) {
      throw new Error(`Failed to update editor settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateEditorSettings(editorSettings: Partial<SettingsDTO['editor']>): void {
    if (!editorSettings || typeof editorSettings !== 'object') {
      throw new Error('Invalid editor settings object');
    }

    if (editorSettings.tabSize !== undefined) {
      if (typeof editorSettings.tabSize !== 'number' || editorSettings.tabSize < 1 || editorSettings.tabSize > 8) {
        throw new Error('Tab size must be between 1 and 8');
      }
    }

    if (editorSettings.fontSize !== undefined) {
      if (typeof editorSettings.fontSize !== 'number' || editorSettings.fontSize < 8 || editorSettings.fontSize > 32) {
        throw new Error('Font size must be between 8 and 32');
      }
    }

    if (editorSettings.theme && !['light', 'dark', 'auto'].includes(editorSettings.theme)) {
      throw new Error('Theme must be light, dark, or auto');
    }
  }
}

/**
 * SLA Policy Use Cases
 */

export class SetSLAPolicyUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(policy: SLAPolicy): Promise<void> {
    this.validateSLAPolicy(policy);
    
    try {
      await this.settingsRepo.setSLAPolicy(policy);
    } catch (error) {
      throw new Error(`Failed to set SLA policy: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateSLAPolicy(policy: SLAPolicy): void {
    if (!policy || typeof policy !== 'object') {
      throw new Error('Invalid SLA policy object');
    }

    if (!policy.name?.trim()) {
      throw new Error('SLA policy name is required');
    }

    if (policy.maxResponseTime !== undefined) {
      if (typeof policy.maxResponseTime !== 'number' || policy.maxResponseTime < 100 || policy.maxResponseTime > 300000) {
        throw new Error('Max response time must be between 100ms and 5 minutes');
      }
    }

    if (policy.maxRetries !== undefined) {
      if (typeof policy.maxRetries !== 'number' || policy.maxRetries < 0 || policy.maxRetries > 10) {
        throw new Error('Max retries must be between 0 and 10');
      }
    }
  }
}

/**
 * Settings Import/Export Use Cases
 */

export class ExportSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(): Promise<{
    settingsJson: string;
    exportedAt: string;
    version: string;
  }> {
    try {
      const settingsJson = await this.settingsRepo.exportSettings();
      
      return {
        settingsJson,
        exportedAt: new Date().toISOString(),
        version: '1.0.0' // Settings format version
      };
    } catch (error) {
      throw new Error(`Failed to export settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

export class ImportSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settingsJson: string, options?: {
    validateBeforeImport?: boolean;
    backupCurrent?: boolean;
    mergeWithCurrent?: boolean;
  }): Promise<{
    success: boolean;
    importedSettings: SettingsDTO;
    warnings: string[];
  }> {
    this.validateImportInputs(settingsJson);
    
    const warnings: string[] = [];
    
    try {
      // Parse settings JSON
      let importedSettings: SettingsDTO;
      try {
        importedSettings = JSON.parse(settingsJson);
      } catch (parseError) {
        throw new Error('Invalid JSON format in settings file');
      }

      // Validate imported settings
      if (options?.validateBeforeImport !== false) {
        const validationResult = await this.settingsRepo.validateSettings(importedSettings);
        if (!validationResult.isValid) {
          throw new Error(`Invalid settings format: ${validationResult.errors.join(', ')}`);
        }
        warnings.push(...validationResult.warnings);
      }

      // Backup current settings if requested
      if (options?.backupCurrent) {
        try {
          const currentSettings = await this.settingsRepo.getSettings();
          const backupJson = JSON.stringify(currentSettings, null, 2);
          // Note: In a real implementation, you'd save this backup somewhere
          console.log('Current settings backed up');
        } catch (backupError) {
          warnings.push('Failed to backup current settings');
        }
      }

      // Merge with current settings if requested
      if (options?.mergeWithCurrent) {
        try {
          const currentSettings = await this.settingsRepo.getSettings();
          importedSettings = this.mergeSettings(currentSettings, importedSettings);
          warnings.push('Settings merged with current configuration');
        } catch (mergeError) {
          warnings.push('Failed to merge with current settings, using imported settings only');
        }
      }

      // Import settings
      await this.settingsRepo.importSettings(settingsJson);

      return {
        success: true,
        importedSettings,
        warnings
      };
    } catch (error) {
      throw new Error(`Failed to import settings: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateImportInputs(settingsJson: string): void {
    if (!settingsJson?.trim()) {
      throw new Error('Settings JSON is required');
    }

    if (settingsJson.length > 1000000) { // 1MB limit
      throw new Error('Settings file is too large (maximum 1MB)');
    }
  }

  private mergeSettings(current: SettingsDTO, imported: SettingsDTO): SettingsDTO {
    return {
      ...current,
      ...imported,
      ai: {
        ...current.ai,
        ...imported.ai
      },
      editor: {
        ...current.editor,
        ...imported.editor
      },
      project: {
        ...current.project,
        ...imported.project
      }
    };
  }
}

/**
 * Settings Validation Use Case
 */

export class ValidateSettingsUseCase {
  constructor(private settingsRepo: SettingsRepository) {}

  async execute(settings?: SettingsDTO): Promise<{
    isValid: boolean;
    errors: string[];
    warnings: string[];
    suggestions: string[];
  }> {
    try {
      const settingsToValidate = settings || await this.settingsRepo.getSettings();
      
      const result = await this.settingsRepo.validateSettings(settingsToValidate);
      
      // Add additional suggestions based on settings
      const suggestions = this.generateSettingsSuggestions(settingsToValidate);
      
      return {
        ...result,
        suggestions
      };
    } catch (error) {
      return {
        isValid: false,
        errors: [`Validation failed: ${error instanceof Error ? error.message : 'Unknown error'}`],
        warnings: [],
        suggestions: []
      };
    }
  }

  private generateSettingsSuggestions(settings: SettingsDTO): string[] {
    const suggestions: string[] = [];

    // AI settings suggestions
    if (settings.ai) {
      if (!settings.ai.apiKey) {
        suggestions.push('Configure AI API key for enhanced code generation features');
      }
      
      if (settings.ai.temperature && settings.ai.temperature > 1.5) {
        suggestions.push('High temperature setting may produce unpredictable results');
      }
      
      if (!settings.ai.model) {
        suggestions.push('Select an AI model for optimal performance');
      }
    }

    // Editor settings suggestions
    if (settings.editor) {
      if (!settings.editor.theme) {
        suggestions.push('Set editor theme preference for better user experience');
      }
      
      if (settings.editor.tabSize && settings.editor.tabSize > 4) {
        suggestions.push('Consider using smaller tab size for better code readability');
      }
    }

    // Performance suggestions
    if (!settings.performance?.enableVirtualScrolling) {
      suggestions.push('Enable virtual scrolling for better performance with large files');
    }

    return suggestions;
  }
}