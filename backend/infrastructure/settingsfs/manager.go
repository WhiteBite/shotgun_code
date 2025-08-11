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
	"github.com/zalando/go-keyring"
)

const keyringService = "shotgun-code"

// AppSettings stores settings that are safe to write to a JSON file.
// API keys are handled separately via the system's keyring.
type AppSettings struct {
	CustomIgnoreRules string              `json:"customIgnoreRules"`
	CustomPromptRules string              `json:"customPromptRules"`
	UseGitignore      bool                `json:"useGitignore"`
	UseCustomIgnore   bool                `json:"useCustomIgnore"`
	LocalAIHost       string              `json:"localAIHost,omitempty"`
	LocalAIModelName  string              `json:"localAIModelName,omitempty"`
	SelectedProvider  string              `json:"selectedProvider"`
	ProviderModels    map[string]string   `json:"providerModels"`
	AvailableModels   map[string][]string `json:"availableModels"`
}

type Manager struct {
	settings           AppSettings
	configPath         string
	defaultIgnoreRules string
	defaultPromptRules string
	log                domain.Logger
	mu                 sync.RWMutex
	// In-memory cache for API keys
	openAIAPIKey     string
	geminiAPIKey     string
	openRouterAPIKey string
	localAIAPIKey    string
}

func New(logger domain.Logger, defaultIgnore, defaultPrompt string) (domain.SettingsRepository, error) {
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
	if err := m.load(); err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	return m, nil
}

// Save saves settings to JSON and API keys to keyring.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Save API keys to keyring
	keyring.Set(keyringService, "openai", m.openAIAPIKey)
	keyring.Set(keyringService, "gemini", m.geminiAPIKey)
	keyring.Set(keyringService, "openrouter", m.openRouterAPIKey)
	keyring.Set(keyringService, "localai", m.localAIAPIKey)

	// Save other settings to file
	return m.saveToFile()
}

func (m *Manager) GetOpenAIKey() string { m.mu.RLock(); defer m.mu.RUnlock(); return m.openAIAPIKey }
func (m *Manager) GetGeminiKey() string { m.mu.RLock(); defer m.mu.RUnlock(); return m.geminiAPIKey }
func (m *Manager) GetOpenRouterKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.openRouterAPIKey
}
func (m *Manager) GetLocalAIKey() string { m.mu.RLock(); defer m.mu.RUnlock(); return m.localAIAPIKey }

func (m *Manager) SetOpenAIKey(k string)     { m.mu.Lock(); m.openAIAPIKey = k; m.mu.Unlock() }
func (m *Manager) SetGeminiKey(k string)     { m.mu.Lock(); m.geminiAPIKey = k; m.mu.Unlock() }
func (m *Manager) SetOpenRouterKey(k string) { m.mu.Lock(); m.openRouterAPIKey = k; m.mu.Unlock() }
func (m *Manager) SetLocalAIKey(k string)    { m.mu.Lock(); m.localAIAPIKey = k; m.mu.Unlock() }

// --- Passthrough methods for file-based settings ---

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
func (m *Manager) GetUseGitignore() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.UseGitignore
}
func (m *Manager) GetUseCustomIgnore() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.UseCustomIgnore
}
func (m *Manager) GetLocalAIHost() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.LocalAIHost
}
func (m *Manager) GetLocalAIModelName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.LocalAIModelName
}
func (m *Manager) GetSelectedAIProvider() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.SelectedProvider
}
func (m *Manager) GetSelectedModel(p string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.ProviderModels[p]
}
func (m *Manager) GetModels(p string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.AvailableModels[p]
}

func (m *Manager) SetCustomIgnoreRules(r string) {
	m.mu.Lock()
	m.settings.CustomIgnoreRules = r
	m.mu.Unlock()
}
func (m *Manager) SetCustomPromptRules(r string) {
	m.mu.Lock()
	m.settings.CustomPromptRules = r
	m.mu.Unlock()
}
func (m *Manager) SetUseGitignore(e bool) { m.mu.Lock(); m.settings.UseGitignore = e; m.mu.Unlock() }
func (m *Manager) SetUseCustomIgnore(e bool) {
	m.mu.Lock()
	m.settings.UseCustomIgnore = e
	m.mu.Unlock()
}
func (m *Manager) SetLocalAIHost(h string) { m.mu.Lock(); m.settings.LocalAIHost = h; m.mu.Unlock() }
func (m *Manager) SetLocalAIModelName(n string) {
	m.mu.Lock()
	m.settings.LocalAIModelName = n
	m.mu.Unlock()
}
func (m *Manager) SetSelectedAIProvider(p string) {
	m.mu.Lock()
	m.settings.SelectedProvider = p
	m.mu.Unlock()
}
func (m *Manager) SetSelectedModel(p, mdl string) {
	m.mu.Lock()
	if m.settings.ProviderModels == nil {
		m.settings.ProviderModels = make(map[string]string)
	}
	m.settings.ProviderModels[p] = mdl
	m.mu.Unlock()
}
func (m *Manager) SetModels(p string, mdls []string) {
	m.mu.Lock()
	if m.settings.AvailableModels == nil {
		m.settings.AvailableModels = make(map[string][]string)
	}
	m.settings.AvailableModels[p] = mdls
	m.mu.Unlock()
}

// load reads settings from file and API keys from keyring.
func (m *Manager) load() error {
	m.loadDefaults()

	// Load non-sensitive settings from JSON file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.log.Info("Settings file not found. Saving default settings.")
			return m.saveToFile() // Create file with defaults
		}
		return fmt.Errorf("error reading settings file: %w", err)
	}

	var userSettings AppSettings
	if err := json.Unmarshal(data, &userSettings); err != nil {
		m.log.Error(fmt.Sprintf("Error deserializing settings: %v. Using defaults.", err))
		// Keep defaults, but don't overwrite user's broken file
		return nil
	}
	m.mergeWithDefaults(&userSettings)

	// Load sensitive data from keyring
	m.loadKeysFromKeyring()

	return nil
}

func (m *Manager) loadKeysFromKeyring() {
	m.openAIAPIKey, _ = keyring.Get(keyringService, "openai")
	m.geminiAPIKey, _ = keyring.Get(keyringService, "gemini")
	m.openRouterAPIKey, _ = keyring.Get(keyringService, "openrouter")
	m.localAIAPIKey, _ = keyring.Get(keyringService, "localai")
}

func (m *Manager) loadDefaults() {
	m.settings = AppSettings{
		CustomIgnoreRules: m.defaultIgnoreRules,
		CustomPromptRules: m.defaultPromptRules,
		UseGitignore:      true,
		UseCustomIgnore:   true,
		SelectedProvider:  "openai",
		LocalAIHost:       "http://localhost:1234/v1",
		ProviderModels: map[string]string{
			"openai":     "gpt-4o",
			"gemini":     "gemini-1.5-pro-latest",
			"openrouter": "google/gemini-flash-1.5",
			"localai":    "local-model",
		},
		AvailableModels: map[string][]string{
			"openai":     {"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"},
			"gemini":     {"gemini-1.5-pro-latest", "gemini-1.5-flash-latest"},
			"openrouter": {"google/gemini-flash-1.5", "openai/gpt-4o", "meta-llama/llama-3-70b-instruct"},
			"localai":    {"local-model"},
		},
	}
}

func (m *Manager) mergeWithDefaults(userSettings *AppSettings) {
	m.settings.SelectedProvider = userSettings.SelectedProvider
	m.settings.UseGitignore = userSettings.UseGitignore
	m.settings.UseCustomIgnore = userSettings.UseCustomIgnore
	m.settings.LocalAIHost = userSettings.LocalAIHost
	m.settings.LocalAIModelName = userSettings.LocalAIModelName
	if strings.TrimSpace(userSettings.CustomIgnoreRules) != "" {
		m.settings.CustomIgnoreRules = userSettings.CustomIgnoreRules
	}
	if strings.TrimSpace(userSettings.CustomPromptRules) != "" {
		m.settings.CustomPromptRules = userSettings.CustomPromptRules
	}
	if userSettings.ProviderModels != nil {
		for p, model := range userSettings.ProviderModels {
			m.settings.ProviderModels[p] = model
		}
	}
}

// saveToFile persists only non-sensitive settings.
func (m *Manager) saveToFile() error {
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации настроек: %w", err)
	}
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("ошибка создания директории конфигурации: %w", err)
	}
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла настроек: %w", err)
	}
	return nil
}
