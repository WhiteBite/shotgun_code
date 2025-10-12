package config

import (
	"fmt"
	"shotgun_code/domain"
	"sync"
)

// ModelFetcher function type for getting model lists for providers
type ModelFetcher func(apiKey string) ([]string, error)

// SettingsService manages application settings
type SettingsService struct {
	log                           domain.Logger
	bus                           domain.EventBus
	settingsRepo                  domain.SettingsRepository
	modelFetchers                 map[string]ModelFetcher
	onIgnoreRulesChangedCallbacks []func() error
	muCallbacks                   sync.RWMutex
}

// NewSettingsService creates a new settings service
func NewSettingsService(
	log domain.Logger,
	bus domain.EventBus,
	settingsRepo domain.SettingsRepository,
	modelFetchers map[string]ModelFetcher,
) (*SettingsService, error) {
	s := &SettingsService{
		log:           log,
		bus:           bus,
		settingsRepo:  settingsRepo,
		modelFetchers: modelFetchers,
	}
	return s, nil
}

// GetSettingsDTO gathers all settings into one DTO for frontend
func (s *SettingsService) GetSettingsDTO() (domain.SettingsDTO, error) {
	return s.settingsRepo.GetSettingsDTO()
}

// SaveSettingsDTO takes DTO from frontend and updates settings
func (s *SettingsService) SaveSettingsDTO(dto domain.SettingsDTO) error {
	s.settingsRepo.SetCustomIgnoreRules(dto.CustomIgnoreRules)
	s.settingsRepo.SetCustomPromptRules(dto.CustomPromptRules)
	s.settingsRepo.SetOpenAIKey(dto.OpenAIAPIKey)
	s.settingsRepo.SetGeminiKey(dto.GeminiAPIKey)
	s.settingsRepo.SetOpenRouterKey(dto.OpenRouterAPIKey)
	s.settingsRepo.SetLocalAIKey(dto.LocalAIAPIKey)
	s.settingsRepo.SetLocalAIHost(dto.LocalAIHost)
	s.settingsRepo.SetLocalAIModelName(dto.LocalAIModelName)
	s.settingsRepo.SetSelectedAIProvider(dto.SelectedProvider)
	s.settingsRepo.SetUseGitignore(dto.UseGitignore)
	s.settingsRepo.SetUseCustomIgnore(dto.UseCustomIgnore)

	for provider, model := range dto.SelectedModels {
		s.settingsRepo.SetSelectedModel(provider, model)
	}
	for provider, models := range dto.AvailableModels {
		s.settingsRepo.SetModels(provider, models)
	}

	if err := s.settingsRepo.Save(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	s.notifyIgnoreRulesChanged()
	return nil
}

// RefreshModels updates model list for specified provider
func (s *SettingsService) RefreshModels(provider, apiKey string) error {
	fetcher, ok := s.modelFetchers[provider]
	if !ok {
		return fmt.Errorf("no model fetcher for provider: %s", provider)
	}

	s.log.Info("Refreshing model list for: " + provider)
	models, err := fetcher(apiKey)
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to refresh models for %s: %v", provider, err))
		return err
	}

	s.settingsRepo.SetModels(provider, models)

	currentModel := s.settingsRepo.GetSelectedModel(provider)
	isCurrentModelValid := false
	for _, m := range models {
		if m == currentModel {
			isCurrentModelValid = true
			break
		}
	}
	if !isCurrentModelValid && len(models) > 0 {
		s.log.Info(fmt.Sprintf("Selected model '%s' not found in new list. Selecting '%s' by default.", currentModel, models[0]))
		s.settingsRepo.SetSelectedModel(provider, models[0])
	}

	if err := s.settingsRepo.Save(); err != nil {
		return fmt.Errorf("failed to save new model list: %w", err)
	}

	s.log.Info(fmt.Sprintf("Model list for %s refreshed successfully.", provider))
	return nil
}

// OnIgnoreRulesChanged registers callback
func (s *SettingsService) OnIgnoreRulesChanged(callback func() error) {
	s.muCallbacks.Lock()
	defer s.muCallbacks.Unlock()
	s.onIgnoreRulesChangedCallbacks = append(s.onIgnoreRulesChangedCallbacks, callback)
}

func (s *SettingsService) notifyIgnoreRulesChanged() {
	s.muCallbacks.RLock()
	defer s.muCallbacks.RUnlock()
	for _, cb := range s.onIgnoreRulesChangedCallbacks {
		if err := cb(); err != nil {
			s.log.Error(fmt.Sprintf("Error executing OnIgnoreRulesChanged callback: %v", err))
		}
	}
}
