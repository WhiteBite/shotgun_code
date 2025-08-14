package domain

type ExportMode string

const (
	ExportModeClipboard ExportMode = "clipboard"
	ExportModeAI        ExportMode = "ai"
	ExportModeHuman     ExportMode = "human"
)

type ExportSettings struct {
	Mode    ExportMode `json:"mode"`
	Context string     `json:"context"`

	// Clipboard
	StripComments   bool   `json:"stripComments"`
	IncludeManifest bool   `json:"includeManifest"`
	ExportFormat    string `json:"exportFormat"` // "plain" | "manifest" | "json"

	// AI
	AIProfile       string `json:"aiProfile"`
	TokenLimit      int    `json:"tokenLimit"`
	FileSizeLimitKB int    `json:"fileSizeLimitKB"`

	// Human
	Theme              string `json:"theme"`
	IncludeLineNumbers bool   `json:"includeLineNumbers"`
	IncludePageNumbers bool   `json:"includePageNumbers"`
}

type ExportResult struct {
	Mode       ExportMode `json:"mode"`
	Text       string     `json:"text,omitempty"`
	FileName   string     `json:"fileName,omitempty"`
	DataBase64 string     `json:"dataBase64,omitempty"`
}
