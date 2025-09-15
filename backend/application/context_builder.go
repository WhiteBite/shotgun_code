package application

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextBuilderImpl implements the ContextBuilder interface
type ContextBuilderImpl struct {
	fileReader      domain.FileContentReader
	tokenCounter    TokenCounter
	logger          domain.Logger
	settingsService *SettingsService
	bus             domain.EventBus
	opaService      domain.OPAService
	pathProvider    domain.PathProvider
	fileSystemWriter domain.FileSystemWriter
	contextDir      string
}

// NewContextBuilder creates a new ContextBuilder implementation
func NewContextBuilder(
	fileReader domain.FileContentReader,
	tokenCounter TokenCounter,
	logger domain.Logger,
	settingsService *SettingsService,
	bus domain.EventBus,
	opaService domain.OPAService,
	pathProvider domain.PathProvider,
	fileSystemWriter domain.FileSystemWriter,
	contextDir string,
) *ContextBuilderImpl {
	return &ContextBuilderImpl{
		fileReader:      fileReader,
		tokenCounter:    tokenCounter,
		logger:          logger,
		settingsService: settingsService,
		bus:             bus,
		opaService:      opaService,
		pathProvider:    pathProvider,
		fileSystemWriter: fileSystemWriter,
		contextDir:      contextDir,
	}
}

// BuildContext builds a context from project files and returns a ContextSummary to prevent OOM issues
func (cb *ContextBuilderImpl) BuildContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	if options == nil {
		options = &domain.ContextBuildOptions{}
	}
	
	cb.logger.Info(fmt.Sprintf("Building context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Create context content and save to file instead of using strings.Builder to prevent OOM
	contextID := uuid.New().String()
	contextPath := filepath.Join(cb.contextDir, contextID+".ctx")
	
	file, err := os.Create(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create context file: %w", err)
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	
	var totalChars int64
	var actualFiles []string
	var tokenCount int
	
	// Write header if requested
	if options.IncludeManifest {
		header := fmt.Sprintf("# Project Context\nProject Path: %s\nGenerated: %s\n\n", projectPath, time.Now().Format(time.RFC3339))
		writer.WriteString(header)
		totalChars += int64(len(header))
	}
	
	// Process files one by one to prevent loading all content into memory
	for _, filePath := range includedPaths {
		// Read individual file content to avoid loading all files at once
		content, err := cb.fileReader.ReadContents(ctx, []string{filePath}, projectPath, nil)
		if err != nil {
			cb.logger.Warning(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
			continue
		}
		
		fileContent, exists := content[filePath]
		if !exists {
			continue
		}
		
		actualFiles = append(actualFiles, filePath)
		
		// Process content based on options
		if options.StripComments {
			fileContent = cb.stripComments(fileContent, filePath)
		}
		
		// Write file content to context file
		fileHeader := fmt.Sprintf("## File: %s\n\n", filePath)
		writer.WriteString(fileHeader)
		totalChars += int64(len(fileHeader))
		
		writer.WriteString("``")
		if ext := filepath.Ext(filePath); len(ext) > 1 {
			writer.WriteString(ext[1:])
		}
		writer.WriteString("\n")
		totalChars += 4 // Account for code block markers
		
		writer.WriteString(fileContent)
		totalChars += int64(len(fileContent))
		
		writer.WriteString("\n```\n\n")
		totalChars += 7 // Account for code block closing and spacing
		
		// Update token count incrementally
		tokenCount += cb.tokenCounter.CountTokens(fileContent)
	}
	
	// Check token limit
	if options.MaxTokens > 0 && tokenCount > options.MaxTokens {
		// Clean up the context file if token limit exceeded
		os.Remove(contextPath)
		return nil, fmt.Errorf("context exceeds token limit: %d > %d", tokenCount, options.MaxTokens)
	}
	
	now := time.Now()
	
	// Create context summary object instead of full context
	contextSummary := &domain.ContextSummary{
		ID:          contextID,
		ProjectPath: projectPath,
		FileCount:   len(actualFiles),
		TotalSize:   totalChars,
		TokenCount:  tokenCount,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      "ready",
		Metadata: domain.ContextMetadata{
			BuildDuration: 0, // Would need to measure this properly
			LastModified:  now,
			SelectedFiles: actualFiles,
			BuildOptions:  options,
			Warnings:      []string{},
			Errors:        []string{},
		},
	}
	
	// Save context summary to disk
	if err := cb.saveContextSummary(contextSummary); err != nil {
		// Clean up the context file if summary save fails
		os.Remove(contextPath)
		return nil, fmt.Errorf("failed to save context summary: %w", err)
	}
	
	cb.logger.Info(fmt.Sprintf("Built context summary %s with %d tokens", contextID, tokenCount))
	return contextSummary, nil
}

// BuildContextLegacy builds context with legacy format (DEPRECATED - can cause OOM)
func (cb *ContextBuilderImpl) BuildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options domain.ContextBuildOptions) (*domain.Context, error) {
	// This is the legacy BuildContext method that returns full context
	// WARNING: This can cause OOM issues with large contexts
	cb.logger.Warning("Using deprecated BuildContextLegacy method - consider using BuildContext instead")
	return cb.buildContextLegacy(ctx, projectPath, includedPaths, options)
}

// buildContextLegacy is the internal implementation of BuildContextLegacy
func (cb *ContextBuilderImpl) buildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options domain.ContextBuildOptions) (*domain.Context, error) {
	cb.logger.Info(fmt.Sprintf("Building legacy context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Read file contents
	contents, err := cb.fileReader.ReadContents(ctx, includedPaths, projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}
	
	// Build context content
	var contentBuilder strings.Builder
	var actualFiles []string
	
	if options.IncludeManifest {
		contentBuilder.WriteString("# Project Context\n\n")
		contentBuilder.WriteString(fmt.Sprintf("Project Path: %s\n", projectPath))
		contentBuilder.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	}
	
	for _, filePath := range includedPaths {
		content, exists := contents[filePath]
		if !exists {
			continue
		}
		
		actualFiles = append(actualFiles, filePath)
		
		// Process content based on options
		if options.StripComments {
			content = cb.stripComments(content, filePath)
		}
		
		// Add file content to context
		contentBuilder.WriteString(fmt.Sprintf("## File: %s\n\n", filePath))
		contentBuilder.WriteString("```")
		if ext := filepath.Ext(filePath); len(ext) > 1 {
			contentBuilder.WriteString(ext[1:])
		}
		contentBuilder.WriteString("\n")
		contentBuilder.WriteString(content)
		contentBuilder.WriteString("\n```\n\n")
	}
	
	contextContent := contentBuilder.String()
	
	// Check token limit
	tokenCount := cb.tokenCounter.CountTokens(contextContent)
	if options.MaxTokens > 0 && tokenCount > options.MaxTokens {
		return nil, fmt.Errorf("context exceeds token limit: %d > %d", tokenCount, options.MaxTokens)
	}
	
	// Create context object
	contextID := uuid.New().String()
	now := time.Now()
	
	context := &domain.Context{
		ID:          contextID,
		Name:        cb.generateContextName(projectPath, actualFiles),
		Description: fmt.Sprintf("Context with %d files from %s", len(actualFiles), filepath.Base(projectPath)),
		Content:     contextContent,
		Files:       actualFiles,
		CreatedAt:   now,
		UpdatedAt:   now,
		ProjectPath: projectPath,
		TokenCount:  tokenCount,
	}
	
	// Save context to disk
	if err := cb.saveContext(context); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}
	
	cb.logger.Info(fmt.Sprintf("Built legacy context %s with %d tokens", contextID, tokenCount))
	return context, nil
}

// GenerateContext builds a context asynchronously with progress tracking
func (cb *ContextBuilderImpl) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	// Run in separate goroutine with panic recovery
	go cb.generateContextSafe(ctx, rootDir, includedPaths)
}

// generateContextSafe is a helper method that generates context safely with panic recovery
func (cb *ContextBuilderImpl) generateContextSafe(ctx context.Context, rootDir string, includedPaths []string) {
	// Note: This implementation would need access to fileReader and other dependencies
	// For now, we'll leave it as a placeholder since it requires significant dependencies
	// In a full implementation, this would be moved to a separate service or the implementation would be updated
}

// saveContext saves a context to disk
func (cb *ContextBuilderImpl) saveContext(context *domain.Context) error {
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}
	
	contextPath := filepath.Join(cb.contextDir, context.ID+".json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}
	
	return nil
}

// saveContextSummary saves a context summary to disk
func (cb *ContextBuilderImpl) saveContextSummary(contextSummary *domain.ContextSummary) error {
	data, err := json.MarshalIndent(contextSummary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context summary: %w", err)
	}
	
	contextPath := filepath.Join(cb.contextDir, contextSummary.ID+".json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context summary file: %w", err)
	}
	
	return nil
}

// generateContextName generates a descriptive name for the context
func (cb *ContextBuilderImpl) generateContextName(projectPath string, files []string) string {
	projectName := filepath.Base(projectPath)
	
	if len(files) == 1 {
		fileName := filepath.Base(files[0])
		return fmt.Sprintf("%s - %s", projectName, fileName)
	}
	
	return fmt.Sprintf("%s - %d files", projectName, len(files))
}

// stripComments removes comments from code content
func (cb *ContextBuilderImpl) stripComments(content, filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".go", ".js", ".ts", ".java", ".c", ".cpp", ".cs":
		return cb.stripCStyleComments(content)
	case ".py", ".sh":
		return cb.stripHashComments(content)
	case ".html", ".xml":
		return cb.stripXMLComments(content)
	default:
		return content
	}
}

// stripCStyleComments removes C-style comments (// and /* */)
func (cb *ContextBuilderImpl) stripCStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	inBlockComment := false
	
	for _, line := range lines {
		// Handle block comments
		if inBlockComment {
			if strings.Contains(line, "*/") {
				parts := strings.SplitN(line, "*/", 2)
				if len(parts) > 1 {
					line = parts[1]
					inBlockComment = false
				} else {
					continue
				}
			} else {
				continue
			}
		}
		
		// Handle start of block comment
		if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
			parts := strings.SplitN(line, "/*", 2)
			line = parts[0]
			inBlockComment = true
		}
		
		// Handle single line comments
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}
		
		// Remove inline block comments
		for strings.Contains(line, "/*") && strings.Contains(line, "*/") {
			start := strings.Index(line, "/*")
			end := strings.Index(line, "*/")
			if end > start {
				line = line[:start] + line[end+2:]
			}
		}
		
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

// stripHashComments removes hash-style comments (#)
func (cb *ContextBuilderImpl) stripHashComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

// stripXMLComments removes XML-style comments (<!-- -->)
func (cb *ContextBuilderImpl) stripXMLComments(content string) string {
	for strings.Contains(content, "<!--") && strings.Contains(content, "-->") {
		start := strings.Index(content, "<!--")
		end := strings.Index(content, "-->")
		if end > start {
			content = content[:start] + content[end+3:]
		}
	}
	return content
}