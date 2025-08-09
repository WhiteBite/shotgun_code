package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"shotgun_code/core"

	"github.com/adrg/xdg"
	gitignore "github.com/sabhiram/go-gitignore"
)

// AppSettings хранит пользовательские настройки приложения.
type AppSettings struct {
	CustomIgnoreRules string `json:"customIgnoreRules"`
	CustomPromptRules string `json:"customPromptRules"`
}

// Manager управляет загрузкой, сохранением и доступом к настройкам приложения.
// Он обеспечивает потокобезопасность операций с настройками.
type Manager struct {
	settings                    AppSettings
	currentCustomIgnorePatterns *gitignore.GitIgnore
	configPath                  string
	defaultIgnoreRules          string
	defaultPromptRules          string
	log                         core.Logger
	mu                          sync.RWMutex
}

// NewManager создает новый менеджер настроек.
func NewManager(logger core.Logger, defaultIgnore, defaultPrompt string) (*Manager, error) {
	configFilePath, err := xdg.ConfigFile("shotgun-code/settings.json")
	if err != nil {
		logger.Error(fmt.Sprintf("Не удалось получить путь к файлу конфигурации: %v", err))
		// Можно продолжать без пути, но сохранение не будет работать.
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

// GetCustomIgnoreRules возвращает пользовательские правила игнорирования в виде строки.
func (m *Manager) GetCustomIgnoreRules() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.CustomIgnoreRules
}

// GetCustomPromptRules возвращает пользовательские правила для промптов в виде строки.
func (m *Manager) GetCustomPromptRules() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.CustomPromptRules
}

// SetCustomIgnoreRules обновляет, компилирует и сохраняет правила игнорирования.
func (m *Manager) SetCustomIgnoreRules(rules string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.CustomIgnoreRules = rules
	m.compileIgnorePatterns() // Компилируем новые правила

	if err := m.save(); err != nil {
		return fmt.Errorf("не удалось сохранить правила игнорирования: %w", err)
	}
	return nil
}

// SetCustomPromptRules обновляет и сохраняет правила для промптов.
func (m *Manager) SetCustomPromptRules(rules string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.CustomPromptRules = rules
	if err := m.save(); err != nil {
		return fmt.Errorf("не удалось сохранить правила для промптов: %w", err)
	}
	return nil
}

// GetCompiledIgnorePatterns возвращает скомпилированный gitignore объект для пользовательских правил.
func (m *Manager) GetCompiledIgnorePatterns() *gitignore.GitIgnore {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentCustomIgnorePatterns
}

// load загружает настройки из файла конфигурации или использует значения по умолчанию.
func (m *Manager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Устанавливаем значения по умолчанию
	m.settings.CustomIgnoreRules = m.defaultIgnoreRules
	m.settings.CustomPromptRules = m.defaultPromptRules

	if m.configPath == "" {
		m.log.Warning("Путь к конфигурации не задан, используются настройки по умолчанию.")
		m.compileIgnorePatterns()
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
	} else {
		var loadedSettings AppSettings
		if err := json.Unmarshal(data, &loadedSettings); err != nil {
			m.log.Error(fmt.Sprintf("Ошибка десериализации настроек: %v", err))
		} else {
			m.settings = loadedSettings
			if strings.TrimSpace(m.settings.CustomIgnoreRules) == "" {
				m.settings.CustomIgnoreRules = m.defaultIgnoreRules
			}
			if strings.TrimSpace(m.settings.CustomPromptRules) == "" {
				m.settings.CustomPromptRules = m.defaultPromptRules
			}
			m.log.Info("Настройки успешно загружены из файла.")
		}
	}
	m.compileIgnorePatterns()
}

// save записывает текущие настройки в файл конфигурации.
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

// compileIgnorePatterns компилирует строку с правилами в объект gitignore.
func (m *Manager) compileIgnorePatterns() {
	if strings.TrimSpace(m.settings.CustomIgnoreRules) == "" {
		m.currentCustomIgnorePatterns = nil
		m.log.Debug("Пользовательские правила игнорирования пусты, паттерны не скомпилированы.")
		return
	}
	lines := strings.Split(strings.ReplaceAll(m.settings.CustomIgnoreRules, "\r\n", "\n"), "\n")
	m.currentCustomIgnorePatterns = gitignore.CompileIgnoreLines(lines...)
	m.log.Info("Пользовательские правила игнорирования успешно скомпилированы.")
}
