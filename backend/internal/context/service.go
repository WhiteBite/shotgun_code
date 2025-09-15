package context

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Service handles all context management operations with memory-safe streaming by default
type Service struct {
	fileReader    domain.FileContentReader
	tokenCounter  TokenCounter
	eventBus      domain.EventBus
	logger        domain.Logger
	contextDir    string
	
	// Streaming support
	streams    map[string]*Stream
	streamsMu  sync.RWMutex
	
	// Memory limits
	defaultMaxMemoryMB int
	defaultMaxTokens   int
}

// TokenCounter interface for token estimation
type TokenCounter interface {
	CountTokens(text string) int
}

// Stream represents a memory-safe streaming context
type Stream struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Files       []string                  `json:"files"`
	ProjectPath string                    `json:"projectPath"`
	TotalLines  int64                     `json:"totalLines"`
	TotalChars  int64                     `json:"totalChars"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
	TokenCount  int                       `json:"tokenCount"`
	contextPath string                    `json:"-"`
}

// LineRange represents a range of lines from a streaming context
type LineRange struct {
	StartLine int64    `json:"startLine"`
	EndLine   int64    `json:"endLine"`
	Lines     []string `json:"lines"`
}

// BuildOptions controls how context is built
type BuildOptions struct {
	MaxTokens           int  `json:"maxTokens,omitempty"`
	MaxMemoryMB         int  `json:"maxMemoryMB,omitempty"`
	StripComments       bool `json:"stripComments,omitempty"`
	IncludeManifest     bool `json:"includeManifest,omitempty"`
	ForceStream         bool `json:"forceStream,omitempty"`
	EnableProgressEvents bool `json:"enableProgressEvents,omitempty"`
}

// NewService creates a new unified context service
func NewService(
	fileReader domain.FileContentReader, 
	tokenCounter TokenCounter, 
	eventBus domain.EventBus,
	logger domain.Logger,
) *Service {
	homeDir, _ := os.UserHomeDir()
	contextDir := filepath.Join(homeDir, ".shotgun-code", "contexts")
	os.MkdirAll(contextDir, 0755)
	
	return &Service{
		fileReader:         fileReader,
		tokenCounter:       tokenCounter,
		eventBus:           eventBus,
		logger:             logger,
		contextDir:         contextDir,
		streams:            make(map[string]*Stream),
		defaultMaxMemoryMB: 50,  // Strict 50MB default limit
		defaultMaxTokens:   8000, // Strict 8000 token default limit
	}
}

// BuildContext builds a context from project files with memory-safe streaming by default
func (s *Service) BuildContext(ctx context.Context, projectPath string, includedPaths []string, options *BuildOptions) (*domain.Context, error) {
	if options == nil {
		options = &BuildOptions{
			MaxMemoryMB: s.defaultMaxMemoryMB,
			MaxTokens:   s.defaultMaxTokens,
			ForceStream: true, // ALWAYS use streaming for safety
		}
	} else {
		// Apply strict defaults if not set
		if options.MaxMemoryMB <= 0 {
			options.MaxMemoryMB = s.defaultMaxMemoryMB
		}
		if options.MaxTokens <= 0 {
			options.MaxTokens = s.defaultMaxTokens
		}
		// ALWAYS force streaming regardless of options
		options.ForceStream = true
	}
	
	// Validate limits before proceeding
	if err := s.validateLimits(options); err != nil {
		return nil, err
	}
	
	// Always use streaming for memory safety
	return s.buildStreamingContext(ctx, projectPath, includedPaths, options)
}

// GenerateContextAsync generates context asynchronously with progress events (maintains backward compatibility)
func (s *Service) GenerateContextAsync(ctx context.Context, rootDir string, includedPaths []string) {
	go s.generateContextSafe(ctx, rootDir, includedPaths)
}

// CreateStream creates a memory-safe streaming context
func (s *Service) CreateStream(ctx context.Context, projectPath string, includedPaths []string, options *BuildOptions) (*Stream, error) {
	if options == nil {
		options = &BuildOptions{
			MaxMemoryMB: s.defaultMaxMemoryMB,
			MaxTokens:   s.defaultMaxTokens,
		}
	}
	
	// Validate limits
	if err := s.validateLimits(options); err != nil {
		return nil, err
	}
	
	s.logger.Info(fmt.Sprintf("Creating streaming context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Pre-check file sizes to prevent memory issues
	totalSize, oversizedFiles, err := s.estimateTotalSize(projectPath, includedPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate file sizes: %w", err)
	}
	
	// Check memory limit
	if options.MaxMemoryMB > 0 {
		maxBytes := int64(options.MaxMemoryMB) * 1024 * 1024
		if totalSize > maxBytes {
			return nil, fmt.Errorf("context would exceed memory limit: %d MB > %d MB. Oversized files: %v", 
				totalSize/(1024*1024), options.MaxMemoryMB, oversizedFiles)
		}
	}
	
	// Read file contents with progress tracking
	contents, err := s.fileReader.ReadContents(ctx, includedPaths, projectPath, func(current, total int64) {
		if options.EnableProgressEvents && s.eventBus != nil {
			select {
			case <-ctx.Done():
				return
			default:
				s.eventBus.Emit("shotgunContextGenerationProgress", map[string]interface{}{
					"current": current,
					"total":   total,
				})
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}
	
	// Create context content and save to file
	contextID := fmt.Sprintf("stream_%s", uuid.New().String())
	contextPath := filepath.Join(s.contextDir, contextID+".ctx")
	
	file, err := os.Create(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create context file: %w", err)
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	
	var totalLines int64
	var totalChars int64
	var actualFiles []string
	var tokenCount int
	
	// Write header if manifest requested
	if options.IncludeManifest {
		header := fmt.Sprintf("# Streaming Context\nProject Path: %s\nGenerated: %s\n\n", projectPath, time.Now().Format(time.RFC3339))
		writer.WriteString(header)
		totalLines += int64(strings.Count(header, "\n"))
		totalChars += int64(len(header))
	}
	
	// Write file contents with token counting
	for _, filePath := range includedPaths {
		content, exists := contents[filePath]
		if !exists {
			continue
		}
		
		actualFiles = append(actualFiles, filePath)
		
		// Process content based on options
		if options.StripComments {
			content = s.stripComments(content, filePath)
		}
		
		// Count tokens for this file
		fileTokens := s.tokenCounter.CountTokens(content)
		tokenCount += fileTokens
		
		// Check token limit
		if options.MaxTokens > 0 && tokenCount > options.MaxTokens {
			// Clean up partial file
			file.Close()
			os.Remove(contextPath)
			return nil, fmt.Errorf("context would exceed token limit: %d > %d", tokenCount, options.MaxTokens)
		}
		
		// Write file content to context file
		fileHeader := fmt.Sprintf("## File: %s\n\n", filePath)
		writer.WriteString(fileHeader)
		totalLines += int64(strings.Count(fileHeader, "\n"))
		totalChars += int64(len(fileHeader))
		
		writer.WriteString("```")
		if ext := filepath.Ext(filePath); len(ext) > 1 {
			writer.WriteString(ext[1:])
		}
		writer.WriteString("\n")
		totalLines += 1
		totalChars += 4
		
		writer.WriteString(content)
		totalLines += int64(strings.Count(content, "\n")) + 1
		totalChars += int64(len(content))
		
		writer.WriteString("\n```\n\n")
		totalLines += 3
		totalChars += 6
		
		// Memory safety check - flush if getting too large
		if options.MaxMemoryMB > 0 && totalChars > int64(options.MaxMemoryMB*1024*1024)/2 {
			if err := writer.Flush(); err != nil {
				return nil, fmt.Errorf("failed to flush writer: %w", err)
			}
		}
	}
	
	now := time.Now()
	
	// Create context stream object
	stream := &Stream{
		ID:          contextID,
		Name:        s.generateContextName(projectPath, actualFiles),
		Description: fmt.Sprintf("Streaming context with %d files from %s", len(actualFiles), filepath.Base(projectPath)),
		Files:       actualFiles,
		ProjectPath: projectPath,
		TotalLines:  totalLines,
		TotalChars:  totalChars,
		CreatedAt:   now,
		UpdatedAt:   now,
		TokenCount:  tokenCount,
		contextPath: contextPath,
	}
	
	// Store stream reference
	s.streamsMu.Lock()
	s.streams[contextID] = stream
	s.streamsMu.Unlock()
	
	s.logger.Info(fmt.Sprintf("Created streaming context %s with %d lines, %d tokens", contextID, totalLines, tokenCount))
	return stream, nil
}

// GetContextLines retrieves a range of lines from a streaming context
func (s *Service) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*LineRange, error) {
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	file, err := os.Open(stream.contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open context file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	var lines []string
	var currentLine int64 = 0
	
	// Skip to start line
	for currentLine < startLine && scanner.Scan() {
		currentLine++
	}
	
	// Read lines in range
	for currentLine >= startLine && currentLine <= endLine && scanner.Scan() {
		lines = append(lines, scanner.Text())
		currentLine++
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading context file: %w", err)
	}
	
	return &LineRange{
		StartLine: startLine,
		EndLine:   endLine,
		Lines:     lines,
	}, nil
}

// GetContext retrieves a context by ID (backward compatibility)
func (s *Service) GetContext(ctx context.Context, contextID string) (*domain.Context, error) {
	contextPath := filepath.Join(s.contextDir, contextID+".json")
	
	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("context not found: %s", contextID)
		}
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}
	
	var context domain.Context
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}
	
	return &context, nil
}

// GetStream retrieves a streaming context by ID
func (s *Service) GetStream(ctx context.Context, contextID string) (*Stream, error) {
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	return stream, nil
}

// ListProjectContexts lists all contexts for a project
func (s *Service) ListProjectContexts(ctx context.Context, projectPath string) ([]*domain.Context, error) {
	entries, err := os.ReadDir(s.contextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}
	
	var contexts []*domain.Context
	
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		
		contextID := strings.TrimSuffix(entry.Name(), ".json")
		context, err := s.GetContext(ctx, contextID)
		if err != nil {
			s.logger.Warning(fmt.Sprintf("Failed to load context %s: %v", contextID, err))
			continue
		}
		
		if context.ProjectPath == projectPath {
			contexts = append(contexts, context)
		}
	}
	
	return contexts, nil
}

// DeleteContext deletes a context by ID
func (s *Service) DeleteContext(ctx context.Context, contextID string) error {
	// Try to delete both JSON and streaming contexts
	jsonPath := filepath.Join(s.contextDir, contextID+".json")
	streamPath := filepath.Join(s.contextDir, contextID+".ctx")
	
	var errors []string
	
	if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("failed to delete JSON context: %v", err))
	}
	
	if err := os.Remove(streamPath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("failed to delete streaming context: %v", err))
	}
	
	// Remove from streams map
	s.streamsMu.Lock()
	delete(s.streams, contextID)
	s.streamsMu.Unlock()
	
	if len(errors) > 0 && !(len(errors) == 2 && strings.Contains(errors[0], "no such file") && strings.Contains(errors[1], "no such file")) {
		return fmt.Errorf("context deletion errors: %s", strings.Join(errors, "; "))
	}
	
	s.logger.Info(fmt.Sprintf("Deleted context %s", contextID))
	return nil
}

// Private helper methods

func (s *Service) validateLimits(options *BuildOptions) error {
	// Strict memory limit validation
	if options.MaxMemoryMB > 0 && options.MaxMemoryMB > 100 {
		return fmt.Errorf("memory limit cannot exceed 100MB for safety")
	}
	
	// Strict token limit validation
	if options.MaxTokens > 0 && options.MaxTokens > 16000 {
		return fmt.Errorf("token limit cannot exceed 16000 for safety")
	}
	
	return nil
}

func (s *Service) estimateTotalSize(projectPath string, includedPaths []string) (int64, []string, error) {
	var totalSize int64
	var oversizedFiles []string
	
	for _, filePath := range includedPaths {
		fullPath := filepath.Join(projectPath, filePath)
		if info, err := os.Stat(fullPath); err == nil {
			totalSize += info.Size()
			// Flag files larger than 1MB
			if info.Size() > 1024*1024 {
				oversizedFiles = append(oversizedFiles, filePath)
			}
		}
	}
	
	return totalSize, oversizedFiles, nil
}

func (s *Service) buildStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *BuildOptions) (*domain.Context, error) {
	// Create streaming context
	stream, err := s.CreateStream(ctx, projectPath, includedPaths, options)
	if err != nil {
		return nil, err
	}
	
	// Convert to domain.Context for backward compatibility
	domainContext := &domain.Context{
		ID:          stream.ID,
		Name:        stream.Name,
		Description: stream.Description,
		Content:     fmt.Sprintf("STREAMING_CONTEXT:%s", stream.ID), // Placeholder - actual content accessed via streaming
		Files:       stream.Files,
		CreatedAt:   stream.CreatedAt,
		UpdatedAt:   stream.UpdatedAt,
		ProjectPath: stream.ProjectPath,
		TokenCount:  stream.TokenCount,
	}
	
	return domainContext, nil
}

func (s *Service) buildLegacyContext(ctx context.Context, projectPath string, includedPaths []string, options *BuildOptions) (*domain.Context, error) {
	s.logger.Info(fmt.Sprintf("Building legacy context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Read file contents
	contents, err := s.fileReader.ReadContents(ctx, includedPaths, projectPath, nil)
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
			content = s.stripComments(content, filePath)
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
	tokenCount := s.tokenCounter.CountTokens(contextContent)
	if options.MaxTokens > 0 && tokenCount > options.MaxTokens {
		return nil, fmt.Errorf("context exceeds token limit: %d > %d", tokenCount, options.MaxTokens)
	}
	
	// Create context object
	contextID := uuid.New().String()
	now := time.Now()
	
	context := &domain.Context{
		ID:          contextID,
		Name:        s.generateContextName(projectPath, actualFiles),
		Description: fmt.Sprintf("Context with %d files from %s", len(actualFiles), filepath.Base(projectPath)),
		Content:     contextContent,
		Files:       actualFiles,
		CreatedAt:   now,
		UpdatedAt:   now,
		ProjectPath: projectPath,
		TokenCount:  tokenCount,
	}
	
	// Save context to disk
	if err := s.saveContext(context); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}
	
	s.logger.Info(fmt.Sprintf("Built legacy context %s with %d tokens", contextID, tokenCount))
	return context, nil
}

func (s *Service) generateContextSafe(ctx context.Context, rootDir string, includedPaths []string) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			s.logger.Error(fmt.Sprintf("PANIC recovered in GenerateContext: %v\nStack: %s", r, stack))
			if s.eventBus != nil {
				s.eventBus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", r))
				s.eventBus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", r))
			}
		}
	}()

	// Emit start event
	if s.eventBus != nil {
		s.eventBus.Emit("shotgunContextGenerationStarted", map[string]interface{}{
			"fileCount": len(includedPaths),
			"rootDir":   rootDir,
		})
	}

	// Ensure we always have a non-nil context
	if ctx == nil {
		ctx = context.Background()
	}

	// Add timeout for safety
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info(fmt.Sprintf("Starting async context generation for %d files", len(includedPaths)))

	// Sort paths for deterministic processing
	sortedPaths := make([]string, len(includedPaths))
	copy(sortedPaths, includedPaths)
	sort.Strings(sortedPaths)

	// Use errgroup for controlled concurrency
	g, gctx := errgroup.WithContext(ctx)

	var contents map[string]string

	g.Go(func() error {
		var err error
		contents, err = s.fileReader.ReadContents(gctx, sortedPaths, rootDir, func(current, total int64) {
			select {
			case <-gctx.Done():
				return
			default:
				if s.eventBus != nil {
					s.eventBus.Emit("shotgunContextGenerationProgress", map[string]interface{}{
						"current": current,
						"total":   total,
					})
				}
			}
		})
		return err
	})

	if err := g.Wait(); err != nil {
		if err == context.DeadlineExceeded {
			s.logger.Error("Context generation timed out")
			if s.eventBus != nil {
				s.eventBus.Emit("app:error", "Context generation timed out after 30 seconds")
				s.eventBus.Emit("shotgunContextGenerationTimeout")
			}
		} else {
			s.logger.Error(fmt.Sprintf("Failed to read file contents: %v", err))
			if s.eventBus != nil {
				s.eventBus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", err))
				s.eventBus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", err))
			}
		}
		return
	}

	// Build context content
	var contextBuilder strings.Builder

	// Add manifest header
	contextBuilder.WriteString("Manifest:\n")
	manifestPaths := make([]string, 0, len(contents))
	for path := range contents {
		manifestPaths = append(manifestPaths, path)
	}
	sort.Strings(manifestPaths)

	// Build simple tree structure
	contextBuilder.WriteString(s.buildSimpleTree(manifestPaths))
	contextBuilder.WriteString("\n")

	// Add file contents in sorted order
	for _, relPath := range manifestPaths {
		content, exists := contents[relPath]
		if !exists {
			continue
		}

		contextBuilder.WriteString(fmt.Sprintf("--- File: %s ---\n", relPath))
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n")
	}

	finalContext := strings.TrimSpace(contextBuilder.String())
	s.logger.Info(fmt.Sprintf("Async context generation completed. Length: %d characters", len(finalContext)))

	if s.eventBus != nil {
		s.logger.Info("Emitting shotgunContextGenerated event")
		s.eventBus.Emit("shotgunContextGenerated", finalContext)
		s.logger.Info("Event emitted successfully")
	}
}

type treeNode struct {
	name     string
	children map[string]*treeNode
	isFile   bool
}

func (s *Service) buildSimpleTree(paths []string) string {
	
	root := &treeNode{name: ".", children: make(map[string]*treeNode)}

	// Build tree structure
	for _, path := range paths {
		parts := strings.Split(filepath.ToSlash(path), "/")
		current := root

		for i, part := range parts {
			if part == "" || part == "." {
				continue
			}

			if _, exists := current.children[part]; !exists {
				current.children[part] = &treeNode{
					name:     part,
					children: make(map[string]*treeNode),
					isFile:   i == len(parts)-1,
				}
			}
			current = current.children[part]
		}
	}

	// Generate tree string
	var builder strings.Builder
	s.walkTree(root, "", true, &builder)
	return builder.String()
}

func (s *Service) walkTree(node *treeNode, prefix string, isLast bool, builder *strings.Builder) {
	if node.name != "." {
		if isLast {
			builder.WriteString(prefix + "└─ " + node.name + "\n")
			prefix += "   "
		} else {
			builder.WriteString(prefix + "├─ " + node.name + "\n")
			prefix += "│  "
		}
	}

	// Sort children for deterministic output
	childNames := make([]string, 0, len(node.children))
	for name := range node.children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	for i, name := range childNames {
		child := node.children[name]
		isLastChild := i == len(childNames)-1
		s.walkTree(child, prefix, isLastChild, builder)
	}
}

func (s *Service) saveContext(context *domain.Context) error {
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}
	
	contextPath := filepath.Join(s.contextDir, context.ID+".json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}
	
	return nil
}

func (s *Service) generateContextName(projectPath string, files []string) string {
	projectName := filepath.Base(projectPath)
	
	if len(files) == 1 {
		fileName := filepath.Base(files[0])
		return fmt.Sprintf("%s - %s", projectName, fileName)
	}
	
	return fmt.Sprintf("%s - %d files", projectName, len(files))
}

func (s *Service) estimateTokens(totalChars int64) int {
	// Simple approximation: 1 token ≈ 4 characters
	return int(totalChars / 4)
}

func (s *Service) stripComments(content, filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".go", ".js", ".ts", ".java", ".c", ".cpp", ".cs":
		return s.stripCStyleComments(content)
	case ".py", ".sh":
		return s.stripHashComments(content)
	case ".html", ".xml":
		return s.stripXMLComments(content)
	default:
		return content
	}
}

func (s *Service) stripCStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	inBlockComment := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Handle block comments
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
				// Keep part after */
				parts := strings.SplitN(line, "*/", 2)
				if len(parts) > 1 {
					line = parts[1]
				} else {
					continue
				}
			} else {
				continue
			}
		}
		
		// Handle start of block comments
		if strings.Contains(line, "/*") {
			inBlockComment = true
			parts := strings.SplitN(line, "/*", 2)
			line = parts[0]
			if strings.TrimSpace(line) == "" {
				continue
			}
		}
		
		// Handle line comments
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

func (s *Service) stripHashComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

func (s *Service) stripXMLComments(content string) string {
	// Simple implementation - can be enhanced
	result := content
	for {
		start := strings.Index(result, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "-->")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+3:]
	}
	return result
}