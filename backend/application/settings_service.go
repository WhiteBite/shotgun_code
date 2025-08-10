package application

import (
	"fmt"
	"shotgun_code/domain"
	"sync"
)

// ModelFetcher - это функция, которая знает, как получить список моделей для провайдера, используя предоставленный API ключ.
type ModelFetcher func(apiKey string) ([]string, error)

// SettingsService отвечает за управление настройками приложения.
type SettingsService struct {
	log                           domain.Logger
	bus                           domain.EventBus
	settingsRepo                  domain.SettingsRepository
	modelFetchers                 map[string]ModelFetcher
	useGitignore                  bool
	useCustomIgnore               bool
	mu                            sync.RWMutex
	onIgnoreRulesChangedCallbacks []func() error
	muCallbacks                   sync.RWMutex
}

// NewSettingsService создает новый экземпляр SettingsService.
func NewSettingsService(
	log domain.Logger,
	bus domain.EventBus,
	settingsRepo domain.SettingsRepository,
	modelFetchers map[string]ModelFetcher,
) (*SettingsService, error) {
	s := &SettingsService{
		log:             log,
		bus:             bus,
		settingsRepo:    settingsRepo,
		modelFetchers:   modelFetchers,
		useGitignore:    true,
		useCustomIgnore: true,
	}
	return s, nil
}

// RefreshModels обновляет список моделей для указанного провайдера, используя зарегистрированный ModelFetcher.
func (s *SettingsService) RefreshModels(provider, apiKey string) error {
	fetcher, ok := s.modelFetchers[provider]
	if !ok {
		return fmt.Errorf("нет способа обновить модели для провайдера: %s", provider)
	}

	s.log.Info("Обновление списка моделей для: " + provider)
	models, err := fetcher(apiKey)
	if err != nil {
		s.log.Error(fmt.Sprintf("Не удалось обновить модели для %s: %v", provider, err))
		return err
	}

	if err := s.settingsRepo.SetModels(provider, models); err != nil {
		return err
	}

	currentModel := s.settingsRepo.GetSelectedModel(provider)
	isCurrentModelValid := false
	for _, m := range models {
		if m == currentModel {
			isCurrentModelValid = true
			break
		}
	}
	if !isCurrentModelValid && len(models) > 0 {
		s.log.Info(fmt.Sprintf("Выбранная модель '%s' не найдена в обновленном списке. Выбираем '%s' по умолчанию.", currentModel, models[0]))
		if err := s.settingsRepo.SetSelectedModel(provider, models[0]); err != nil {
			return err
		}
	}

	s.log.Info(fmt.Sprintf("Список моделей для %s успешно обновлен.", provider))
	return nil
}

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
			s.log.Error(fmt.Sprintf("Ошибка при выполнении коллбэка на смену правил: %v", err))
		}
	}
}

func (s *SettingsService) GetUseGitignore() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.useGitignore
}

func (s *SettingsService) GetUseCustomIgnore() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.useCustomIgnore
}

func (s *SettingsService) SetUseGitignore(enabled bool) error {
	s.mu.Lock()
	s.useGitignore = enabled
	s.mu.Unlock()
	s.log.Info(fmt.Sprintf("Использование .gitignore установлено в: %v", enabled))
	s.notifyIgnoreRulesChanged()
	return nil
}

func (s *SettingsService) SetUseCustomIgnore(enabled bool) error {
	s.mu.Lock()
	s.useCustomIgnore = enabled
	s.mu.Unlock()
	s.log.Info(fmt.Sprintf("Использование кастомных правил установлено в: %v", enabled))
	s.notifyIgnoreRulesChanged()
	return nil
}

func (s *SettingsService) GetCustomIgnoreRules() string { return s.settingsRepo.GetCustomIgnoreRules() }
func (s *SettingsService) SetCustomIgnoreRules(rules string) error {
	if err := s.settingsRepo.SetCustomIgnoreRules(rules); err != nil {
		return err
	}
	s.notifyIgnoreRulesChanged()
	return nil
}
func (s *SettingsService) GetCustomPromptRules() string { return s.settingsRepo.GetCustomPromptRules() }
func (s *SettingsService) SetCustomPromptRules(rules string) error {
	return s.settingsRepo.SetCustomPromptRules(rules)
}

func (s *SettingsService) GetOpenAIKey() string          { return s.settingsRepo.GetOpenAIKey() }
func (s *SettingsService) SetOpenAIKey(key string) error { return s.settingsRepo.SetOpenAIKey(key) }

func (s *SettingsService) GetGeminiKey() string          { return s.settingsRepo.GetGeminiKey() }
func (s *SettingsService) SetGeminiKey(key string) error { return s.settingsRepo.SetGeminiKey(key) }

func (s *SettingsService) GetLocalAIKey() string          { return s.settingsRepo.GetLocalAIKey() }
func (s *SettingsService) SetLocalAIKey(key string) error { return s.settingsRepo.SetLocalAIKey(key) }
func (s *SettingsService) GetLocalAIHost() string         { return s.settingsRepo.GetLocalAIHost() }
func (s *SettingsService) SetLocalAIHost(host string) error {
	return s.settingsRepo.SetLocalAIHost(host)
}
func (s *SettingsService) GetLocalAIModelName() string { return s.settingsRepo.GetLocalAIModelName() }
func (s *SettingsService) SetLocalAIModelName(name string) error {
	return s.settingsRepo.SetLocalAIModelName(name)
}

func (s *SettingsService) GetSelectedAIProvider() string {
	return s.settingsRepo.GetSelectedAIProvider()
}
func (s *SettingsService) SetSelectedAIProvider(provider string) error {
	return s.settingsRepo.SetSelectedAIProvider(provider)
}

func (s *SettingsService) GetModels(provider string) ([]string, error) {
	models := s.settingsRepo.GetModels(provider)
	if models == nil {
		return []string{}, fmt.Errorf("нет моделей для провайдера %s", provider)
	}
	return models, nil
}
func (s *SettingsService) GetSelectedModel(provider string) string {
	return s.settingsRepo.GetSelectedModel(provider)
}
func (s *SettingsService) SetSelectedModel(provider, model string) error {
	return s.settingsRepo.SetSelectedModel(provider, model)
}
