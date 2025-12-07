package memory

import (
	"encoding/json"
	"strings"
)

// PreferenceKey defines known preference keys
type PreferenceKey string

const (
	PrefExcludeTests      PreferenceKey = "exclude_tests"
	PrefExcludeVendor     PreferenceKey = "exclude_vendor"
	PrefExcludeGenerated  PreferenceKey = "exclude_generated"
	PrefPreferredLanguage PreferenceKey = "preferred_language"
	PrefMaxContextFiles   PreferenceKey = "max_context_files"
	PrefIncludeComments   PreferenceKey = "include_comments"
	PrefCodeStyle         PreferenceKey = "code_style"
	PrefTestFramework     PreferenceKey = "test_framework"
	PrefExcludePatterns   PreferenceKey = "exclude_patterns"
	PrefIncludePatterns   PreferenceKey = "include_patterns"
)

// UserPreferences manages user preferences with parsing
type UserPreferences struct {
	memory ContextMemoryInterface
	cache  map[string]string
}

// ContextMemoryInterface defines the interface for context memory
type ContextMemoryInterface interface {
	SetPreference(key, value string) error
	GetPreference(key string) (string, error)
	GetAllPreferences() (map[string]string, error)
}

// NewUserPreferences creates a new user preferences manager
func NewUserPreferences(memory ContextMemoryInterface) *UserPreferences {
	return &UserPreferences{
		memory: memory,
		cache:  make(map[string]string),
	}
}

// Set sets a preference
func (up *UserPreferences) Set(key PreferenceKey, value string) error {
	up.cache[string(key)] = value
	return up.memory.SetPreference(string(key), value)
}

// Get gets a preference
func (up *UserPreferences) Get(key PreferenceKey) string {
	if v, ok := up.cache[string(key)]; ok {
		return v
	}
	v, _ := up.memory.GetPreference(string(key))
	up.cache[string(key)] = v
	return v
}

const boolTrue = "true"

// GetBool gets a boolean preference
func (up *UserPreferences) GetBool(key PreferenceKey) bool {
	v := strings.ToLower(up.Get(key))
	return v == boolTrue || v == "yes" || v == "1"
}

// GetInt gets an integer preference
func (up *UserPreferences) GetInt(key PreferenceKey, defaultVal int) int {
	v := up.Get(key)
	if v == "" {
		return defaultVal
	}
	var i int
	for _, c := range v {
		if c >= '0' && c <= '9' {
			i = i*10 + int(c-'0')
		}
	}
	if i == 0 {
		return defaultVal
	}
	return i
}

// GetStringList gets a list preference
func (up *UserPreferences) GetStringList(key PreferenceKey) []string {
	v := up.Get(key)
	if v == "" {
		return nil
	}
	var list []string
	_ = json.Unmarshal([]byte(v), &list)
	return list
}

// SetStringList sets a list preference
func (up *UserPreferences) SetStringList(key PreferenceKey, list []string) error {
	b, _ := json.Marshal(list)
	return up.Set(key, string(b))
}

// LoadAll loads all preferences into cache
func (up *UserPreferences) LoadAll() error {
	prefs, err := up.memory.GetAllPreferences()
	if err != nil {
		return err
	}
	for k, v := range prefs {
		up.cache[k] = v
	}
	return nil
}

// ShouldExcludeFile checks if a file should be excluded based on preferences
func (up *UserPreferences) ShouldExcludeFile(filePath string) bool {
	filePath = strings.ToLower(filePath)

	// Check test exclusion
	if up.GetBool(PrefExcludeTests) {
		if strings.Contains(filePath, "_test.") ||
			strings.Contains(filePath, ".test.") ||
			strings.Contains(filePath, ".spec.") ||
			strings.Contains(filePath, "/test/") ||
			strings.Contains(filePath, "/tests/") ||
			strings.Contains(filePath, "/__tests__/") ||
			strings.HasPrefix(filePath, "test/") ||
			strings.HasPrefix(filePath, "tests/") ||
			strings.HasPrefix(filePath, "__tests__/") {
			return true
		}
	}

	// Check vendor exclusion
	if up.GetBool(PrefExcludeVendor) {
		if strings.Contains(filePath, "/vendor/") ||
			strings.Contains(filePath, "/node_modules/") ||
			strings.HasPrefix(filePath, "vendor/") ||
			strings.HasPrefix(filePath, "node_modules/") {
			return true
		}
	}

	// Check generated exclusion
	if up.GetBool(PrefExcludeGenerated) {
		if strings.Contains(filePath, ".gen.") ||
			strings.Contains(filePath, ".generated.") ||
			strings.Contains(filePath, "_gen.") ||
			strings.HasSuffix(filePath, ".pb.go") {
			return true
		}
	}

	// Check custom exclude patterns
	excludePatterns := up.GetStringList(PrefExcludePatterns)
	for _, pattern := range excludePatterns {
		pattern = strings.ToLower(pattern)
		// Handle glob-like patterns
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(filePath, suffix) {
				return true
			}
		} else if strings.Contains(filePath, pattern) {
			return true
		}
	}

	return false
}

// ShouldIncludeFile checks if a file matches include patterns
func (up *UserPreferences) ShouldIncludeFile(filePath string) bool {
	includePatterns := up.GetStringList(PrefIncludePatterns)
	if len(includePatterns) == 0 {
		return true // No filter, include all
	}

	filePath = strings.ToLower(filePath)
	for _, pattern := range includePatterns {
		if strings.Contains(filePath, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// ParsePreferenceFromMessage extracts preferences from user message
func ParsePreferenceFromMessage(message string) (key PreferenceKey, value string, found bool) {
	message = strings.ToLower(message)

	// Patterns for preference extraction
	patterns := map[string]PreferenceKey{
		"не включай тесты":     PrefExcludeTests,
		"exclude tests":        PrefExcludeTests,
		"без тестов":           PrefExcludeTests,
		"ignore tests":         PrefExcludeTests,
		"не включай vendor":    PrefExcludeVendor,
		"exclude vendor":       PrefExcludeVendor,
		"без сгенерированного": PrefExcludeGenerated,
		"exclude generated":    PrefExcludeGenerated,
		"максимум файлов":      PrefMaxContextFiles,
		"max files":            PrefMaxContextFiles,
		"включай комментарии":  PrefIncludeComments,
		"include comments":     PrefIncludeComments,
	}

	for pattern, prefKey := range patterns {
		if strings.Contains(message, pattern) {
			// Determine value
			if strings.Contains(message, "не ") || strings.Contains(message, "exclude") ||
				strings.Contains(message, "без") || strings.Contains(message, "ignore") {
				return prefKey, boolTrue, true
			}
			if strings.Contains(message, "включай") || strings.Contains(message, "include") {
				return prefKey, boolTrue, true
			}
		}
	}

	return "", "", false
}
