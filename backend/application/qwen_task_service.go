package application

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"time"
)

// QwenTaskService orchestrates task execution with Qwen's large context window
type QwenTaskService struct {
	log                 domain.Logger
	aiService           *AIService
	smartContextService *SmartContextService
	settingsService     *SettingsService
}

// NewQwenTaskService creates a new Qwen task service
func NewQwenTaskService(
	log domain.Logger,
	aiService *AIService,
	smartContextService *SmartContextService,
	settingsService *SettingsService,
) *QwenTaskService {
	return &QwenTaskService{
		log:                 log,
		aiService:           aiService,
		smartContextService: smartContextService,
		settingsService:     settingsService,
	}
}

// TaskRequest represents a task to be executed with Qwen
type TaskRequest struct {
	Task          string   `json:"task"`          // Task description (e.g., "implement authentication")
	ProjectRoot   string   `json:"projectRoot"`   // Project root path
	SelectedFiles []string `json:"selectedFiles"` // Files to include in context
	SelectedCode  string   `json:"selectedCode"`  // Selected code snippet
	SourceFile    string   `json:"sourceFile"`    // File where code was selected
	Model         string   `json:"model"`         // Model to use (default: qwen-coder-plus-latest)
	MaxTokens     int      `json:"maxTokens"`     // Max tokens for context
	Temperature   float64  `json:"temperature"`   // Temperature for generation
}

// TaskResponse contains the result of task execution
type TaskResponse struct {
	Content        string            `json:"content"`        // Generated response
	Model          string            `json:"model"`          // Model used
	TokensUsed     int               `json:"tokensUsed"`     // Tokens consumed
	ProcessingTime time.Duration     `json:"processingTime"` // Time taken
	ContextSummary ContextSummaryDTO `json:"contextSummary"` // Summary of context used
	Success        bool              `json:"success"`
	Error          string            `json:"error,omitempty"`
}

// ContextSummaryDTO provides info about the context sent to the model
type ContextSummaryDTO struct {
	TotalFiles     int      `json:"totalFiles"`
	TotalTokens    int      `json:"totalTokens"`
	IncludedFiles  []string `json:"includedFiles"`
	TruncatedFiles []string `json:"truncatedFiles"`
	ExcludedFiles  []string `json:"excludedFiles"`
}

// ExecuteTask executes a task using Qwen with smart context collection
func (s *QwenTaskService) ExecuteTask(ctx context.Context, req TaskRequest) (*TaskResponse, error) {
	startTime := time.Now()
	s.log.Info(fmt.Sprintf("Executing task with Qwen: %s", truncateString(req.Task, 100)))

	// Set defaults
	if req.Model == "" {
		req.Model = "qwen-coder-plus-latest"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 900000 // Leave room for response
	}
	if req.Temperature == 0 {
		req.Temperature = 0.3 // Lower temperature for code generation
	}

	// Collect smart context
	contextReq := SmartContextRequest{
		ProjectRoot:   req.ProjectRoot,
		Task:          req.Task,
		SelectedFiles: req.SelectedFiles,
		SelectedCode:  req.SelectedCode,
		SourceFile:    req.SourceFile,
		MaxTokens:     req.MaxTokens,
		MaxDepth:      3,
	}

	smartContext, err := s.smartContextService.CollectContext(ctx, contextReq)
	if err != nil {
		return &TaskResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to collect context: %v", err),
		}, err
	}

	// Build the prompt
	systemPrompt := s.buildSystemPrompt(req)
	userPrompt := s.buildUserPrompt(req, smartContext)

	// Execute with Qwen
	response, err := s.aiService.GenerateCodeWithOptions(ctx, systemPrompt, userPrompt, CodeGenerationOptions{
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   32000, // Max output tokens
		Timeout:     5 * time.Minute,
	})

	if err != nil {
		return &TaskResponse{
			Success: false,
			Error:   fmt.Sprintf("AI generation failed: %v", err),
			ContextSummary: ContextSummaryDTO{
				TotalFiles:     len(smartContext.Files),
				TotalTokens:    smartContext.TokenEstimate,
				IncludedFiles:  s.getFilePaths(smartContext.Files),
				TruncatedFiles: smartContext.TruncatedFiles,
				ExcludedFiles:  smartContext.ExcludedFiles,
			},
		}, err
	}

	return &TaskResponse{
		Content:        response,
		Model:          req.Model,
		TokensUsed:     smartContext.TokenEstimate,
		ProcessingTime: time.Since(startTime),
		Success:        true,
		ContextSummary: ContextSummaryDTO{
			TotalFiles:     len(smartContext.Files),
			TotalTokens:    smartContext.TokenEstimate,
			IncludedFiles:  s.getFilePaths(smartContext.Files),
			TruncatedFiles: smartContext.TruncatedFiles,
			ExcludedFiles:  smartContext.ExcludedFiles,
		},
	}, nil
}

// PreviewContext returns a preview of what context would be collected
func (s *QwenTaskService) PreviewContext(ctx context.Context, req TaskRequest) (*SmartContextResult, error) {
	if req.MaxTokens == 0 {
		req.MaxTokens = 900000
	}

	contextReq := SmartContextRequest{
		ProjectRoot:   req.ProjectRoot,
		Task:          req.Task,
		SelectedFiles: req.SelectedFiles,
		SelectedCode:  req.SelectedCode,
		SourceFile:    req.SourceFile,
		MaxTokens:     req.MaxTokens,
		MaxDepth:      3,
	}

	return s.smartContextService.CollectContext(ctx, contextReq)
}

// buildSystemPrompt creates the system prompt for Qwen
func (s *QwenTaskService) buildSystemPrompt(req TaskRequest) string {
	return `You are an expert software developer assistant. Your task is to help implement features and fix issues in code.

IMPORTANT RULES:
1. Analyze the provided codebase context carefully before making changes
2. Follow the existing code style and patterns in the project
3. Provide complete, working code - not snippets or pseudocode
4. When modifying existing files, provide the changes in unified diff format
5. When creating new files, provide the complete file content
6. Explain your changes briefly but focus on the code
7. Consider edge cases and error handling
8. Maintain backward compatibility unless explicitly asked to break it

OUTPUT FORMAT:
- For file modifications, use this format:
  --- a/path/to/file.go
  +++ b/path/to/file.go
  @@ -line,count +line,count @@
  -old line
  +new line

- For new files, use this format:
  === NEW FILE: path/to/new_file.go ===
  <complete file content>
  === END FILE ===

- After code changes, provide a brief summary of what was changed and why.`
}

// buildUserPrompt creates the user prompt with context
func (s *QwenTaskService) buildUserPrompt(req TaskRequest, smartContext *SmartContextResult) string {
	var builder strings.Builder

	builder.WriteString("# Task\n")
	builder.WriteString(req.Task)
	builder.WriteString("\n\n")

	if req.SelectedCode != "" {
		builder.WriteString("# Selected Code (focus area)\n")
		builder.WriteString("File: ")
		builder.WriteString(req.SourceFile)
		builder.WriteString("\n```\n")
		builder.WriteString(req.SelectedCode)
		builder.WriteString("\n```\n\n")
	}

	builder.WriteString("# Project Context\n")
	builder.WriteString(fmt.Sprintf("Total files in context: %d\n", len(smartContext.Files)))
	builder.WriteString(fmt.Sprintf("Estimated tokens: %d\n\n", smartContext.TokenEstimate))

	// Add call stack info if available
	if smartContext.CallStack != nil && smartContext.CallStack.RootSymbol != nil {
		builder.WriteString("## Call Stack Analysis\n")
		builder.WriteString(fmt.Sprintf("Root symbol: %s (%s)\n",
			smartContext.CallStack.RootSymbol.Name,
			smartContext.CallStack.RootSymbol.Type))

		if len(smartContext.CallStack.Callers) > 0 {
			builder.WriteString("Called by: ")
			callerNames := make([]string, 0, len(smartContext.CallStack.Callers))
			for _, c := range smartContext.CallStack.Callers {
				callerNames = append(callerNames, c.Name)
			}
			builder.WriteString(strings.Join(callerNames, ", "))
			builder.WriteString("\n")
		}

		if len(smartContext.CallStack.Callees) > 0 {
			builder.WriteString("Calls: ")
			calleeNames := make([]string, 0, len(smartContext.CallStack.Callees))
			for _, c := range smartContext.CallStack.Callees {
				calleeNames = append(calleeNames, c.Name)
			}
			builder.WriteString(strings.Join(calleeNames, ", "))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	// Add file contents
	builder.WriteString("## Files\n\n")
	for _, file := range smartContext.Files {
		builder.WriteString(fmt.Sprintf("### %s\n", file.Path))
		builder.WriteString(fmt.Sprintf("Relevance: %.0f%% - %s\n", file.Relevance*100, file.Reason))
		builder.WriteString("```")
		builder.WriteString(getLanguageFromPath(file.Path))
		builder.WriteString("\n")
		builder.WriteString(file.Content)
		builder.WriteString("\n```\n\n")
	}

	builder.WriteString("# Instructions\n")
	builder.WriteString("Based on the task and the provided context, implement the required changes. ")
	builder.WriteString("Make sure to follow the existing code patterns and style.\n")

	return builder.String()
}

// getFilePaths extracts file paths from context files
func (s *QwenTaskService) getFilePaths(files []ContextFile) []string {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	return paths
}

// extToLang maps file extensions to language identifiers for syntax highlighting
var extToLang = map[string]string{
	".go": langGo, ".ts": langTypeScript, ".tsx": langTypeScript,
	".js": langJavaScript, ".jsx": langJavaScript, ".py": "python",
	".java": "java", ".rs": "rust", ".vue": "vue",
	".css": "css", ".scss": "css", ".html": "html",
	".json": "json", ".yaml": "yaml", ".yml": "yaml",
}

// getLanguageFromPath returns the language identifier for syntax highlighting
func getLanguageFromPath(path string) string {
	ext := filepath.Ext(path)
	if lang, ok := extToLang[ext]; ok {
		return lang
	}
	return ""
}
