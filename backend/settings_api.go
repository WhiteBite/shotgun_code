package main

import (
	"encoding/json"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/settingsfs"
	"strings"
)

// GetSettings returns current application settings
func (a *App) GetSettings() (domain.SettingsDTO, error) {
	return a.settingsHandler.GetSettings()
}

// SaveSettings saves application settings
func (a *App) SaveSettings(settingsJson string) error {
	return a.settingsHandler.SaveSettings(settingsJson)
}

// RefreshAIModels refreshes available AI models for a provider
func (a *App) RefreshAIModels(provider, apiKey string) error {
	return a.settingsHandler.RefreshAIModels(provider, apiKey)
}

// GetRecentProjects returns the list of recently opened projects
func (a *App) GetRecentProjects() (string, error) {
	projects := a.settingsService.GetRecentProjects()
	result, err := json.Marshal(projects)
	if err != nil {
		return "", fmt.Errorf("failed to marshal recent projects: %w", err)
	}
	return string(result), nil
}

// AddRecentProject adds a project to the recent list and saves settings
func (a *App) AddRecentProject(path, name string) error {
	a.settingsService.AddRecentProject(path, name)
	return a.settingsService.Save()
}

// RemoveRecentProject removes a project from the recent list and saves settings
func (a *App) RemoveRecentProject(path string) error {
	a.settingsService.RemoveRecentProject(path)
	return a.settingsService.Save()
}

// GetCustomIgnoreRules returns custom ignore rules from settings
func (a *App) GetCustomIgnoreRules() (string, error) {
	dto, err := a.settingsService.GetSettingsDTO()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %w", err)
	}
	return dto.CustomIgnoreRules, nil
}

// UpdateCustomIgnoreRules updates custom ignore rules in settings
func (a *App) UpdateCustomIgnoreRules(rules string) error {
	dto, err := a.settingsService.GetSettingsDTO()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}
	dto.CustomIgnoreRules = rules
	return a.settingsService.SaveSettingsDTO(dto)
}

// TestIgnoreRules tests ignore rules against project files
func (a *App) TestIgnoreRules(projectPath string, rules string) ([]string, error) {
	files, err := a.projectService.ListFiles(projectPath, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var ignoredFiles []string
	for _, file := range files {
		if !file.IsDir && a.matchesIgnorePattern(file.RelPath, rules) {
			ignoredFiles = append(ignoredFiles, file.RelPath)
		}
	}

	return ignoredFiles, nil
}

// matchesIgnorePattern checks if a path matches any ignore pattern
func (a *App) matchesIgnorePattern(path string, rules string) bool {
	if rules == "" {
		return false
	}

	lines := strings.Split(rules, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(path, line) || strings.HasSuffix(path, line) {
			return true
		}
	}

	return false
}

// SetSLAPolicy sets SLA policy
func (a *App) SetSLAPolicy(policyJson string) error {
	var policy domain.SLAPolicy
	if err := json.Unmarshal([]byte(policyJson), &policy); err != nil {
		return fmt.Errorf("failed to parse SLA policy JSON: %w", err)
	}

	a.log.Info(fmt.Sprintf("SLA Policy set: %s", policy.Name))
	return nil
}

// GetSLAPolicy returns current SLA policy
func (a *App) GetSLAPolicy() (string, error) {
	defaultPolicy := domain.SLAPolicy{
		Name:        "standard",
		Description: "Standard SLA policy",
		MaxTokens:   10000,
		MaxFiles:    50,
		MaxTime:     300,
		MaxMemory:   1024 * 1024 * 100,
		MaxRetries:  3,
		Timeout:     30,
	}

	policyJson, err := json.Marshal(defaultPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SLA policy: %w", err)
	}

	return string(policyJson), nil
}

// ============================================
// Secure API Key Storage
// ============================================

// SaveAPIKey saves an API key securely using encrypted storage
func (a *App) SaveAPIKey(provider, apiKey string) error {
	storage, err := a.getSecureStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize secure storage: %w", err)
	}

	key := formatAPIKeyName(provider)
	if err := storage.SaveCredential(key, apiKey); err != nil {
		return fmt.Errorf("failed to save API key for %s: %w", provider, err)
	}

	a.log.Info(fmt.Sprintf("API key saved securely for provider: %s", provider))
	return nil
}

// HasAPIKey checks if an API key exists for the given provider
func (a *App) HasAPIKey(provider string) bool {
	storage, err := a.getSecureStorage()
	if err != nil {
		a.log.Warning(fmt.Sprintf("Failed to check API key: %v", err))
		return false
	}

	key := formatAPIKeyName(provider)
	return storage.HasCredential(key)
}

// DeleteAPIKey removes an API key for the given provider
func (a *App) DeleteAPIKey(provider string) error {
	storage, err := a.getSecureStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize secure storage: %w", err)
	}

	key := formatAPIKeyName(provider)
	if err := storage.DeleteCredential(key); err != nil {
		return fmt.Errorf("failed to delete API key for %s: %w", provider, err)
	}

	a.log.Info(fmt.Sprintf("API key deleted for provider: %s", provider))
	return nil
}

// GetAPIKeyStatus returns status of all API keys (without exposing actual keys)
func (a *App) GetAPIKeyStatus() (string, error) {
	storage, err := a.getSecureStorage()
	if err != nil {
		return "", fmt.Errorf("failed to initialize secure storage: %w", err)
	}

	providers := []string{"openai", "anthropic", "gemini", "openrouter", "ollama"}
	status := make(map[string]bool)

	for _, provider := range providers {
		key := formatAPIKeyName(provider)
		status[provider] = storage.HasCredential(key)
	}

	result, err := json.Marshal(status)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API key status: %w", err)
	}

	return string(result), nil
}

// LoadAPIKey loads an API key for internal use (not exposed to frontend)
func (a *App) LoadAPIKey(provider string) (string, error) {
	storage, err := a.getSecureStorage()
	if err != nil {
		return "", fmt.Errorf("failed to initialize secure storage: %w", err)
	}

	key := formatAPIKeyName(provider)
	apiKey, err := storage.LoadCredential(key)
	if err != nil {
		return "", fmt.Errorf("failed to load API key for %s: %w", provider, err)
	}

	return apiKey, nil
}

// formatAPIKeyName formats the credential key name for a provider
func formatAPIKeyName(provider string) string {
	return fmt.Sprintf("api_key_%s", strings.ToLower(provider))
}

// getSecureStorage returns a SecureStorage instance for API key management
func (a *App) getSecureStorage() (*settingsfs.SecureStorage, error) {
	storage, err := settingsfs.NewSecureStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create secure storage: %w", err)
	}
	return storage, nil
}
