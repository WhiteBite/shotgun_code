package settingsfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const keyringService = "shotgun-code"

// storage handles the low-level reading/writing of settings to files and the system keyring.
type storage struct {
	settingsFilePath string
}

// newStorage initializes a new storage helper.
func newStorage() (*storage, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}
	appSpecificDir := filepath.Join(appDataDir, "shotgun-code")
	if err := os.MkdirAll(appSpecificDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create application data directory: %w", err)
	}
	return &storage{
		settingsFilePath: filepath.Join(appSpecificDir, "settings.json"),
	}, nil
}

// loadFromFile reads and unmarshals the settings from the JSON file.
func (s *storage) loadFromFile(settings *appSettings) error {
	data, err := os.ReadFile(s.settingsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File not existing is not an error on first run
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	return json.Unmarshal(data, settings)
}

// saveToFile marshals and writes the settings to the JSON file.
func (s *storage) saveToFile(settings *appSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	return os.WriteFile(s.settingsFilePath, data, 0644)
}

// loadKeysFromKeyring populates the API key fields from the system's keyring.
func (s *storage) loadKeysFromKeyring(settings *secureSettings) {
	settings.openAIAPIKey, _ = keyring.Get(keyringService, "openai")
	settings.geminiAPIKey, _ = keyring.Get(keyringService, "gemini")
	settings.openRouterAPIKey, _ = keyring.Get(keyringService, "openrouter")
	settings.localAIAPIKey, _ = keyring.Get(keyringService, "localai")
}

// saveKeysToKeyring persists the API keys to the system's keyring.
func (s *storage) saveKeysToKeyring(settings *secureSettings) {
	keyring.Set(keyringService, "openai", settings.openAIAPIKey)
	keyring.Set(keyringService, "gemini", settings.geminiAPIKey)
	keyring.Set(keyringService, "openrouter", settings.openRouterAPIKey)
	keyring.Set(keyringService, "localai", settings.localAIAPIKey)
}
