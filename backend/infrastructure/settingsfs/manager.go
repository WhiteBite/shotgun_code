package settingsfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"sync"

	"github.com/adrg/xdg"
)

type AppSettings struct {
	CustomIgnoreRules string              `json:"customIgnoreRules"`
	CustomPromptRules string              `json:"customPromptRules"`
	OpenAIAPIKey      string              `json:"openAIAPIKey,omitempty"`
	GeminiAPIKey      string              `json:"geminiAPIKey,omitempty"`
	LocalAIAPIKey     string              `json:"localAIAPIKey,omitempty"`
	LocalAIHost       string              `json:"localAIHost,omitempty"`
	LocalAIModelName  string              `json:"localAIModelName,omitempty"`
	SelectedProvider  string              `json:"selectedProvider"`
	ProviderModels    map[string]string   `json:"providerModels"`  // provider -> selected model
	AvailableModels   map[string][]string `json:"availableModels"` // provider -> available models
}

type Manager struct {
	settings           AppSettings
	configPath         string
	defaultIgnoreRules string
	defaultPromptRules string
	log                domain.Logger
	mu                 sync.RWMutex
}

func New(logger domain.Logger, defaultIgnore, defaultPrompt string) (*Manager, error) {
	configFilePath, err := xdg.ConfigFile("shotgun-code/settings.json")
	if err != nil {
		logger.Error(fmt.Sprintf("Не удалось получить путь к файлу конфигурации: %v", err))
	}
	m := &Manager{
		log:                logger,
		configPath:         configFilePath,
		defaultIgnoreRules: defaultIgnore,
		defaultPromptRules: defaultPrompt,
	}
	m.load()
	return m, nil
}

func (m *Manager) GetCustomIgnoreRules() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.CustomIgnoreRules
}

func (m *Manager) GetCustomPromptRules() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.CustomPromptRules
}

func (m *Manager) SetCustomIgnoreRules(rules string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.CustomIgnoreRules = rules
	return m.save()
}

func (m *Manager) SetCustomPromptRules(rules string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.CustomPromptRules = rules
	return m.save()
}

func (m *Manager) GetOpenAIKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.OpenAIAPIKey
}

func (m *Manager) SetOpenAIKey(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.OpenAIAPIKey = key
	return m.save()
}

func (m *Manager) GetGeminiKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.GeminiAPIKey
}

func (m *Manager) SetGeminiKey(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.GeminiAPIKey = key
	return m.save()
}

func (m *Manager) GetLocalAIKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.LocalAIAPIKey
}

func (m *Manager) SetLocalAIKey(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.LocalAIAPIKey = key
	return m.save()
}

func (m *Manager) GetLocalAIHost() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.LocalAIHost
}

func (m *Manager) SetLocalAIHost(host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.LocalAIHost = host
	return m.save()
}

func (m *Manager) GetLocalAIModelName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.LocalAIModelName
}

func (m *Manager) SetLocalAIModelName(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.LocalAIModelName = name
	return m.save()
}

func (m *Manager) GetSelectedAIProvider() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.SelectedProvider
}

func (m *Manager) SetSelectedAIProvider(provider string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.SelectedProvider = provider
	return m.save()
}

func (m *Manager) GetSelectedModel(provider string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.ProviderModels[provider]
}

func (m *Manager) SetSelectedModel(provider string, model string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.settings.ProviderModels == nil {
		m.settings.ProviderModels = make(map[string]string)
	}
	m.settings.ProviderModels[provider] = model
	return m.save()
}

func (m *Manager) GetModels(provider string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.AvailableModels[provider]
}

func (m *Manager) SetModels(provider string, models []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.settings.AvailableModels == nil {
		m.settings.AvailableModels = make(map[string][]string)
	}
	m.settings.AvailableModels[provider] = models
	return m.save()
}

func (m *Manager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	factoryDefaults := AppSettings{
		CustomIgnoreRules: m.defaultIgnoreRules,
		CustomPromptRules: m.defaultPromptRules,
		SelectedProvider:  "openai",
		LocalAIHost:       "http://localhost:1234/v1",
		ProviderModels: map[string]string{
			"openai":  "gpt-4o",
			"gemini":  "gemini-1.5-pro-latest",
			"localai": "local-model",
		},
		AvailableModels: map[string][]string{
			"openai":  {"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"},
			"gemini":  {"gemini-1.5-pro-latest", "gemini-1.5-flash-latest"},
			"localai": {"local-model"}, // Placeholder
		},
	}

	m.settings = factoryDefaults

	if m.configPath == "" {
		m.log.Warning("Путь к конфигурации не задан, используются настройки по умолчанию.")
		return
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.log.Info("Файл настроек не найден. Используются и сохраняются настройки по умолчанию.")
			if errSave := m.save(); errSave != nil {
				m.log.Error(fmt.Sprintf("Не удалось сохранить настройки по умолчанию: %v", errSave))
			}
		} else {
			m.log.Error(fmt.Sprintf("Ошибка чтения файла настроек: %v", err))
		}
		return
	}

	var userSettings AppSettings
	if err := json.Unmarshal(data, &userSettings); err != nil {
		m.log.Error(fmt.Sprintf("Ошибка десериализации настроек: %v, будут применены и сохранены настройки по умолчанию.", err))
		m.save()
		return
	}

	if userSettings.OpenAIAPIKey != "" {
		m.settings.OpenAIAPIKey = userSettings.OpenAIAPIKey
	}
	if userSettings.GeminiAPIKey != "" {
		m.settings.GeminiAPIKey = userSettings.GeminiAPIKey
	}
	if userSettings.LocalAIAPIKey != "" {
		m.settings.LocalAIAPIKey = userSettings.LocalAIAPIKey
	}
	if userSettings.LocalAIHost != "" {
		m.settings.LocalAIHost = userSettings.LocalAIHost
	}
	if userSettings.LocalAIModelName != "" {
		m.settings.LocalAIModelName = userSettings.LocalAIModelName
	}
	if userSettings.SelectedProvider != "" {
		m.settings.SelectedProvider = userSettings.SelectedProvider
	}
	if strings.TrimSpace(userSettings.CustomIgnoreRules) != "" {
		m.settings.CustomIgnoreRules = userSettings.CustomIgnoreRules
	}
	if strings.TrimSpace(userSettings.CustomPromptRules) != "" {
		m.settings.CustomPromptRules = userSettings.CustomPromptRules
	}

	if userSettings.ProviderModels != nil {
		for provider, userSelectedModel := range userSettings.ProviderModels {
			if _, ok := m.settings.AvailableModels[provider]; ok {
				m.settings.ProviderModels[provider] = userSelectedModel
			}
		}
	}

	if userSettings.AvailableModels != nil {
		if _, ok := userSettings.AvailableModels["localai"]; ok {
			m.settings.AvailableModels["localai"] = userSettings.AvailableModels["localai"]
		}
	}

	m.log.Info("Настройки успешно загружены и объединены.")

	if errSave := m.save(); errSave != nil {
		m.log.Error(fmt.Sprintf("Не удалось сохранить объединенные настройки: %v", errSave))
	}
}

func (m *Manager) save() error {
	if m.configPath == "" {
		return fmt.Errorf("путь к файлу конфигурации не задан, сохранение невозможно")
	}
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		m.log.Error(fmt.Sprintf("Ошибка сериализации настроек: %v", err))
		return err
	}
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		m.log.Error(fmt.Sprintf("Ошибка создания директории конфигурации: %v", err))
		return err
	}
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		m.log.Error(fmt.Sprintf("Ошибка записи файла настроек: %v", err))
		return err
	}
	m.log.Info("Настройки успешно сохранены.")
	return nil
}
