import type { SettingsDTO, SLAPolicy } from '@/types/dto';

/**
 * Repository interface for application settings
 */
export interface SettingsRepository {
  /**
   * Get application settings
   * @returns Current settings
   */
  getSettings(): Promise<SettingsDTO>;

  /**
   * Save application settings
   * @param settings Settings to save
   */
  saveSettings(settings: SettingsDTO): Promise<void>;

  /**
   * Refresh AI models for a provider
   * @param provider Provider name
   * @param apiKey API key
   */
  refreshAIModels(provider: string, apiKey: string): Promise<void>;

  /**
   * Set SLA policy
   * @param policy SLA policy configuration
   */
  setSLAPolicy(policy: SLAPolicy): Promise<void>;

  /**
   * Reset settings to defaults
   */
  resetToDefaults(): Promise<void>;

  /**
   * Export settings for backup
   * @returns Settings as JSON string
   */
  exportSettings(): Promise<string>;

  /**
   * Import settings from backup
   * @param settingsJson Settings as JSON string
   */
  importSettings(settingsJson: string): Promise<void>;

  /**
   * Validate settings configuration
   * @param settings Settings to validate
   * @returns Validation result with errors
   */
  validateSettings(settings: SettingsDTO): Promise<{
    isValid: boolean;
    errors: string[];
    warnings: string[];
  }>;
}