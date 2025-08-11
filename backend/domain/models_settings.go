package domain

// SettingsDTO - это объект для передачи данных (Data Transfer Object) настроек
// между бэкендом и фронтендом. Он скрывает детали реализации хранения.
type SettingsDTO struct {
	CustomIgnoreRules string              `json:"customIgnoreRules"`
	CustomPromptRules string              `json:"customPromptRules"`
	OpenAIAPIKey      string              `json:"openAIAPIKey"`
	GeminiAPIKey      string              `json:"geminiAPIKey"`
	OpenRouterAPIKey  string              `json:"openRouterAPIKey"`
	LocalAIAPIKey     string              `json:"localAIAPIKey"`
	LocalAIHost       string              `json:"localAIHost"`
	LocalAIModelName  string              `json:"localAIModelName"`
	SelectedProvider  string              `json:"selectedProvider"`
	SelectedModels    map[string]string   `json:"selectedModels"`  // provider -> selected model
	AvailableModels   map[string][]string `json:"availableModels"` // provider -> available models
	UseGitignore      bool                `json:"useGitignore"`
	UseCustomIgnore   bool                `json:"useCustomIgnore"`
}
