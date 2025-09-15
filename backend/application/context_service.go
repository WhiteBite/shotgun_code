package application

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// ContextService handles all context management operations
// Consolidated from ContextService, ContextGenerationService, StreamingContextService, and ContextAnalysisService
type ContextService struct {
	fileReader      domain.FileContentReader
	tokenCounter    TokenCounter
	logger          domain.Logger
	contextDir      string
	aiService       *AIService
	settingsService *SettingsService
	bus             domain.EventBus
	opaService      domain.OPAService      // Add OPA service
	pathProvider    domain.PathProvider    // Add path provider
	fileSystemWriter domain.FileSystemWriter // Add file system writer
	// Streaming context support
	streams    map[string]*domain.ContextStream
	streamPaths map[string]string  // Map context ID to file path
	streamsMu  sync.RWMutex
}

// TokenCounter interface for token estimation
type TokenCounter interface {
	CountTokens(text string) int
}

// ContextLineRange represents a range of lines from a context (for streaming)
type ContextLineRange struct {
	StartLine int64
	EndLine   int64
	Lines     []string
}

// TaskAnalysis contains analysis results for a task
type TaskAnalysis struct {
	Type         string
	Priority     string
	Technologies []string
	FileTypes    []string
	Keywords     []string
	Reasoning    string
}

// ContextAnalysisResult contains the results of context analysis
type ContextAnalysisResult struct {
	Task            string
	TaskType        string
	Priority        string
	SelectedFiles   []*domain.FileNode
	DependencyFiles []*domain.FileNode
	Context         string
	AnalysisTime    time.Duration
	Recommendations []string
	EstimatedTokens int
	Confidence      float64
}

// NewContextService creates a new consolidated context service
func NewContextService(
	fileReader domain.FileContentReader,
	tokenCounter TokenCounter,
	logger domain.Logger,
	aiService *AIService,
	settingsService *SettingsService,
	bus domain.EventBus,
	opaService domain.OPAService, // Add OPA service
	pathProvider domain.PathProvider, // Add path provider
	fileSystemWriter domain.FileSystemWriter, // Add file system writer
) *ContextService {
	homeDir, _ := os.UserHomeDir()
	contextDir := filepath.Join(homeDir, ".shotgun-code", "contexts")
	os.MkdirAll(contextDir, 0755)
	
	return &ContextService{
		fileReader:      fileReader,
		tokenCounter:    tokenCounter,
		logger:          logger,
		contextDir:      contextDir,
		aiService:       aiService,
		settingsService: settingsService,
		bus:             bus,
		opaService:      opaService,
		pathProvider:    pathProvider,
		fileSystemWriter: fileSystemWriter,
		streams:         make(map[string]*domain.ContextStream), // Initialize the streams map
		streamPaths:     make(map[string]string), // Initialize the stream paths map
	}
}

// BuildContext builds a context from project files and returns a ContextSummary to prevent OOM issues
func (s *ContextService) BuildContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	if options == nil {
		options = &domain.ContextBuildOptions{}
	}
	
	s.logger.Info(fmt.Sprintf("Building context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Create context content and save to file instead of using strings.Builder to prevent OOM
	contextID := uuid.New().String()
	contextPath := filepath.Join(s.contextDir, contextID+".ctx")
	
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
		content, err := s.fileReader.ReadContents(ctx, []string{filePath}, projectPath, nil)
		if err != nil {
			s.logger.Warning(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
			continue
		}
		
		fileContent, exists := content[filePath]
		if !exists {
			continue
		}
		
		actualFiles = append(actualFiles, filePath)
		
		// Process content based on options
		if options.StripComments {
			fileContent = s.stripComments(fileContent, filePath)
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
		tokenCount += s.tokenCounter.CountTokens(fileContent)
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
	if err := s.SaveContextSummary(contextSummary); err != nil {
		// Clean up the context file if summary save fails
		os.Remove(contextPath)
		return nil, fmt.Errorf("failed to save context summary: %w", err)
	}
	
	s.logger.Info(fmt.Sprintf("Built context summary %s with %d tokens", contextID, tokenCount))
	return contextSummary, nil
}

// GetContext retrieves a context by ID
func (s *ContextService) GetContext(ctx context.Context, contextID string) (*domain.Context, error) {
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

// GetProjectContexts lists all contexts for a project
func (s *ContextService) GetProjectContexts(ctx context.Context, projectPath string) ([]*domain.Context, error) {
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
func (s *ContextService) DeleteContext(ctx context.Context, contextID string) error {
	contextPath := filepath.Join(s.contextDir, contextID+".json")
	
	if err := os.Remove(contextPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("context not found: %s", contextID)
		}
		return fmt.Errorf("failed to delete context file: %w", err)
	}
	
	s.logger.Info(fmt.Sprintf("Deleted context %s", contextID))
	return nil
}

// SaveContext saves a context to disk
func (s *ContextService) SaveContext(context *domain.Context) error {
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

// SaveContextSummary saves a context summary to disk
func (s *ContextService) SaveContextSummary(contextSummary *domain.ContextSummary) error {
	data, err := json.MarshalIndent(contextSummary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context summary: %w", err)
	}
	
	summaryPath := filepath.Join(s.contextDir, contextSummary.ID+".json")
	if err := os.WriteFile(summaryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context summary file: %w", err)
	}
	
	return nil
}

// generateContextName generates a descriptive name for the context
func (s *ContextService) generateContextName(projectPath string, files []string) string {
	projectName := filepath.Base(projectPath)
	
	if len(files) == 1 {
		fileName := filepath.Base(files[0])
		return fmt.Sprintf("%s - %s", projectName, fileName)
	}
	
	return fmt.Sprintf("%s - %d files", projectName, len(files))
}

// stripComments removes comments from code content
func (s *ContextService) stripComments(content, filePath string) string {
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

// stripCStyleComments removes C-style comments (// and /* */)
func (s *ContextService) stripCStyleComments(content string) string {
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
func (s *ContextService) stripHashComments(content string) string {
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
func (s *ContextService) stripXMLComments(content string) string {
	for strings.Contains(content, "<!--") && strings.Contains(content, "-->") {
		start := strings.Index(content, "<!--")
		end := strings.Index(content, "-->")
		if end > start {
			content = content[:start] + content[end+3:]
		}
	}
	return content
}

// SimpleTokenCounter provides a simple token counting implementation
type SimpleTokenCounter struct{}

func (c *SimpleTokenCounter) CountTokens(text string) int {
	// Simple approximation: ~4 characters per token
	return len(text) / 4
}

// CONSOLIDATED METHODS FROM OTHER CONTEXT SERVICES

// GenerateContext builds a context asynchronously with progress tracking
// Consolidated from ContextGenerationService
func (s *ContextService) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	// Run in separate goroutine with panic recovery
	go s.generateContextSafe(ctx, rootDir, includedPaths)
}

func (s *ContextService) generateContextSafe(ctx context.Context, rootDir string, includedPaths []string) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			s.logger.Error(fmt.Sprintf("PANIC recovered in GenerateContext: %v\nStack: %s", r, stack))
			if s.bus != nil {
				s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", r))
				s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", r))
			}
		}
	}()

	// Send generation start event
	if s.bus != nil {
		s.bus.Emit("shotgunContextGenerationStarted", map[string]interface{}{
			"fileCount": len(includedPaths),
			"rootDir":   rootDir,
		})
	}

	// Ensure we always have a non-nil context to avoid panics
	if ctx == nil {
		ctx = context.Background()
	}

	// Add timeout to prevent infinite loading
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info(fmt.Sprintf("Starting context generation for %d files", len(includedPaths)))

	// Sort paths for deterministic processing
	sortedPaths := make([]string, len(includedPaths))
	copy(sortedPaths, includedPaths)
	sort.Strings(sortedPaths)

	// Use errgroup for controlled concurrency
	g, gctx := errgroup.WithContext(ctx)

	// Read file contents with progress tracking
	var contents map[string]string
	var readErr error

	g.Go(func() error {
		var err error
		contents, err = s.fileReader.ReadContents(gctx, sortedPaths, rootDir, func(current, total int64) {
			select {
			case <-gctx.Done():
				return
			default:
				if s.bus != nil {
					s.bus.Emit("shotgunContextGenerationProgress", map[string]interface{}{
						"current": current,
						"total":   total,
					})
				}
			}
		})
		readErr = err
		return err
	})

	if err := g.Wait(); err != nil {
		if err == context.DeadlineExceeded {
			s.logger.Error("Context generation timed out")
			if s.bus != nil {
				s.bus.Emit("app:error", "Context generation timed out after 30 seconds")
				s.bus.Emit("shotgunContextGenerationTimeout")
			}
		} else {
			s.logger.Error(fmt.Sprintf("Failed to read file contents: %v", err))
			if s.bus != nil {
				s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", err))
				s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", err))
			}
		}
		return
	}

	if readErr != nil {
		s.logger.Error(fmt.Sprintf("Failed to read file contents: %v", readErr))
		if s.bus != nil {
			s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", readErr))
			s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", readErr))
		}
		return
	}

	// Build context string deterministically
	var contextBuilder strings.Builder

	// Add manifest header
	contextBuilder.WriteString("Manifest:\n")
	manifestPaths := make([]string, 0, len(contents))
	for path := range contents {
		manifestPaths = append(manifestPaths, path)
	}
	sort.Strings(manifestPaths) // Ensure deterministic order

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
	s.logger.Info(fmt.Sprintf("Context generation completed. Length: %d characters", len(finalContext)))

	s.logger.Info("Emitting shotgunContextGenerated event")
	if s.bus != nil {
		s.bus.Emit("shotgunContextGenerated", finalContext)
	}
	s.logger.Info("Event emitted successfully")
}

func (s *ContextService) buildSimpleTree(paths []string) string {
	root := &treeNode{
		name:     ".",
		children: make(map[string]*treeNode),
	}

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
					isFile:   i == len(parts)-1, // Last part is file
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

func (s *ContextService) walkTree(node *treeNode, prefix string, isLast bool, builder *strings.Builder) {
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
		s.walkTree(child, prefix, i == len(childNames)-1, builder)
	}
}

// STREAMING CONTEXT METHODS
// Consolidated from StreamingContextService

// CreateStreamingContext creates a streaming context from project files
func (s *ContextService) CreateStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextStream, error) {
	if options == nil {
		options = &domain.ContextBuildOptions{}
	}
	
	s.logger.Info(fmt.Sprintf("Creating streaming context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Read file contents
	contents, err := s.fileReader.ReadContents(ctx, includedPaths, projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}
	
	// Create context content and save to file
	contextID := fmt.Sprintf("stream_%d", len(s.streams)+1)
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
	
	// Write header
	header := fmt.Sprintf("# Streaming Context\nProject Path: %s\nGenerated: %s\n\n", projectPath, time.Now().Format(time.RFC3339))
	writer.WriteString(header)
	totalLines += int64(strings.Count(header, "\n"))
	totalChars += int64(len(header))
	
	// Write file contents
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
		
		// Write file content to context file
		fileHeader := fmt.Sprintf("## File: %s\n\n", filePath)
		writer.WriteString(fileHeader)
		totalLines += int64(strings.Count(fileHeader, "\n"))
		totalChars += int64(len(fileHeader))
		
		writer.WriteString("```\n")
		totalLines += 1
		totalChars += 4
		
		writer.WriteString(content)
		totalLines += int64(strings.Count(content, "\n")) + 1
		totalChars += int64(len(content))
		
		writer.WriteString("\n```\n\n")
		totalLines += 3
		totalChars += 6
	}
	
	// Create context stream object
	stream := &domain.ContextStream{
		ID:          contextID,
		Name:        s.generateContextName(projectPath, actualFiles),
		Description: fmt.Sprintf("Streaming context with %d files from %s", len(actualFiles), filepath.Base(projectPath)),
		Files:       actualFiles,
		ProjectPath: projectPath,
		TotalLines:  totalLines,
		TotalChars:  totalChars,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
		TokenCount:  int(totalChars / 4), // Simple approximation
	}
	
	// Store stream reference and path
	s.streamsMu.Lock()
	s.streams[contextID] = stream
	s.streamPaths[contextID] = contextPath
	s.streamsMu.Unlock()
	
	s.logger.Info(fmt.Sprintf("Created streaming context %s with %d lines", contextID, totalLines))
	return stream, nil
}

// CONTEXT ANALYSIS METHODS
// Consolidated from ContextAnalysisService

// AnalyzeTaskAndCollectContext analyzes a task and automatically collects relevant context
func (s *ContextService) AnalyzeTaskAndCollectContext(
	ctx context.Context,
	task string,
	allFiles []*domain.FileNode,
	rootDir string,
) (*ContextAnalysisResult, error) {
	startTime := time.Now()
	s.logger.Info(fmt.Sprintf("Starting task analysis: %s", task))

	// 1. Analyze task and determine type
	taskAnalysis, err := s.analyzeTaskType(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("task type analysis error: %w", err)
	}

	// 2. Determine priority files based on analysis
	priorityFiles, err := s.determinePriorityFiles(ctx, task, taskAnalysis, allFiles)
	if err != nil {
		return nil, fmt.Errorf("priority files determination error: %w", err)
	}

	// 3. Collect context from priority files
	context, err := s.collectContextFromFiles(ctx, priorityFiles, rootDir, taskAnalysis)
	if err != nil {
		return nil, fmt.Errorf("context collection error: %w", err)
	}

	// 4. Analyze dependencies and add related files
	dependencyFiles, err := s.analyzeDependencies(ctx, priorityFiles, allFiles, rootDir)
	if err != nil {
		s.logger.Warning(fmt.Sprintf("Dependency analysis error: %v", err))
	}

	// 5. Form final result
	result := &ContextAnalysisResult{
		Task:            task,
		TaskType:        taskAnalysis.Type,
		Priority:        taskAnalysis.Priority,
		SelectedFiles:   priorityFiles,
		DependencyFiles: dependencyFiles,
		Context:         context,
		AnalysisTime:    time.Since(startTime),
		Recommendations: s.generateRecommendations(taskAnalysis, priorityFiles),
		EstimatedTokens: s.estimateTokens(context),
		Confidence:      s.calculateConfidence(taskAnalysis, priorityFiles),
	}

	s.logger.Info(fmt.Sprintf("Analysis completed in %v, selected %d files", result.AnalysisTime, len(priorityFiles)))
	return result, nil
}

// analyzeTaskType analyzes task type and determines context collection strategy
func (s *ContextService) analyzeTaskType(ctx context.Context, task string) (*TaskAnalysis, error) {
	if s.aiService == nil {
		// Fallback to simple analysis if AI service is not available
		return s.simpleTaskAnalysis(task), nil
	}

	systemPrompt := `You are an expert in development task analysis. Analyze the task and determine:
1. Task type (bug_fix, feature, refactor, test, documentation, optimization)
2. Priority (low, normal, high, critical)
3. Key technologies/frameworks
4. File types to search for
5. Keywords for search

Respond in JSON format:
{
  "type": "task_type",
  "priority": "priority",
  "technologies": ["tech1", "tech2"],
  "fileTypes": [".go", ".ts", ".vue"],
  "keywords": ["keyword1", "keyword2"],
  "reasoning": "explanation of choice"
}`

	response, err := s.aiService.GenerateCode(ctx, systemPrompt, task)
	if err != nil {
		return nil, fmt.Errorf("task analysis error: %w", err)
	}

	// Parse response (simplified version)
	analysis := &TaskAnalysis{
		Type:         s.extractTaskType(task, response),
		Priority:     s.extractPriority(task, response),
		Technologies: s.extractTechnologies(task, response),
		FileTypes:    s.extractFileTypes(task, response),
		Keywords:     s.extractKeywordsForAnalysis(task, response),
		Reasoning:    response,
	}

	return analysis, nil
}

// simpleTaskAnalysis provides fallback analysis when AI is not available
func (s *ContextService) simpleTaskAnalysis(task string) *TaskAnalysis {
	taskLower := strings.ToLower(task)
	
	// Simple keyword-based classification
	taskType := "feature"
	if strings.Contains(taskLower, "bug") || strings.Contains(taskLower, "fix") || strings.Contains(taskLower, "error") {
		taskType = "bug_fix"
	} else if strings.Contains(taskLower, "test") {
		taskType = "test"
	} else if strings.Contains(taskLower, "refactor") || strings.Contains(taskLower, "cleanup") {
		taskType = "refactor"
	} else if strings.Contains(taskLower, "doc") {
		taskType = "documentation"
	}
	
	priority := "normal"
	if strings.Contains(taskLower, "critical") || strings.Contains(taskLower, "urgent") {
		priority = "critical"
	} else if strings.Contains(taskLower, "high") {
		priority = "high"
	}
	
	return &TaskAnalysis{
		Type:         taskType,
		Priority:     priority,
		Technologies: []string{"go", "typescript", "vue"},
		FileTypes:    []string{".go", ".ts", ".vue", ".js"},
		Keywords:     strings.Fields(taskLower),
		Reasoning:    "Simple keyword-based analysis",
	}
}

// Helper methods for extracting information from AI response
func (s *ContextService) extractTaskType(task, response string) string {
	// Simple extraction - in real implementation would parse JSON
	if strings.Contains(strings.ToLower(response), "bug_fix") {
		return "bug_fix"
	}
	return "feature"
}

func (s *ContextService) extractPriority(task, response string) string {
	if strings.Contains(strings.ToLower(response), "critical") {
		return "critical"
	}
	return "normal"
}

func (s *ContextService) extractTechnologies(task, response string) []string {
	return []string{"go", "typescript", "vue"}
}

func (s *ContextService) extractFileTypes(task, response string) []string {
	return []string{".go", ".ts", ".vue", ".js"}
}

func (s *ContextService) extractKeywordsForAnalysis(task, response string) []string {
	return strings.Fields(strings.ToLower(task))
}

// Stub implementations for remaining methods
func (s *ContextService) determinePriorityFiles(ctx context.Context, task string, analysis *TaskAnalysis, allFiles []*domain.FileNode) ([]*domain.FileNode, error) {
	// Simple implementation - take first few files of matching types
	var priorityFiles []*domain.FileNode
	for _, file := range allFiles {
		if len(priorityFiles) >= 5 {
			break
		}
		for _, fileType := range analysis.FileTypes {
			if strings.HasSuffix(file.RelPath, fileType) {
				priorityFiles = append(priorityFiles, file)
				break
			}
		}
	}
	return priorityFiles, nil
}

func (s *ContextService) collectContextFromFiles(ctx context.Context, files []*domain.FileNode, rootDir string, analysis *TaskAnalysis) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files for context collection")
	}

	// Get file paths
	var filePaths []string
	for _, file := range files {
		filePaths = append(filePaths, file.RelPath)
	}

	// Read file contents
	contents, err := s.fileReader.ReadContents(ctx, filePaths, rootDir, nil)
	if err != nil {
		return "", fmt.Errorf("file reading error: %w", err)
	}

	// Form context
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("// Task analysis: %s\n", analysis.Type))
	contextBuilder.WriteString(fmt.Sprintf("// Priority: %s\n", analysis.Priority))
	contextBuilder.WriteString(fmt.Sprintf("// Technologies: %s\n", strings.Join(analysis.Technologies, ", ")))
	contextBuilder.WriteString("// Project context:\n\n")

	for _, file := range files {
		if content, exists := contents[file.RelPath]; exists {
			contextBuilder.WriteString(fmt.Sprintf("// File: %s\n", file.RelPath))
			contextBuilder.WriteString(content)
			contextBuilder.WriteString("\n\n")
		}
	}

	return contextBuilder.String(), nil
}

func (s *ContextService) analyzeDependencies(ctx context.Context, priorityFiles []*domain.FileNode, allFiles []*domain.FileNode, rootDir string) ([]*domain.FileNode, error) {
	// Simple dependency analysis - return empty for now
	return []*domain.FileNode{}, nil
}

func (s *ContextService) generateRecommendations(analysis *TaskAnalysis, files []*domain.FileNode) []string {
	return []string{
		fmt.Sprintf("Task type: %s", analysis.Type),
		fmt.Sprintf("Selected %d files", len(files)),
	}
}

func (s *ContextService) estimateTokens(context string) int {
	return len(context) / 4
}

func (s *ContextService) calculateConfidence(analysis *TaskAnalysis, files []*domain.FileNode) float64 {
	if len(files) > 0 {
		return 0.8
	}
	return 0.5
}

// ADDITIONAL MISSING METHODS FOR BACKWARD COMPATIBILITY

// BuildContextLegacy builds context with legacy format (DEPRECATED - can cause OOM)
func (s *ContextService) BuildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options domain.ContextBuildOptions) (*domain.Context, error) {
	// This is the legacy BuildContext method that returns full context
	// WARNING: This can cause OOM issues with large contexts
	s.logger.Warning("Using deprecated BuildContextLegacy method - consider using BuildContext instead")
	
	// Create a context summary first
	summary, err := s.BuildContext(ctx, projectPath, includedPaths, &options)
	if err != nil {
		return nil, err
	}
	
	// Convert summary to full context (this is a simplified version)
	context := &domain.Context{
		ID:          summary.ID,
		Name:        s.generateContextName(projectPath, summary.Metadata.SelectedFiles),
		Description: fmt.Sprintf("Context with files from %s", filepath.Base(projectPath)),
		Content:     "", // In a real implementation, we would load the full content
		Files:       summary.Metadata.SelectedFiles,
		CreatedAt:   summary.CreatedAt,
		UpdatedAt:   summary.UpdatedAt,
		ProjectPath: summary.ProjectPath,
		TokenCount:  summary.TokenCount,
	}
	
	return context, nil
}

// CloseContextStream closes a streaming context and cleans up resources
func (s *ContextService) CloseContextStream(ctx context.Context, contextID string) error {
	s.streamsMu.Lock()
	defer s.streamsMu.Unlock()
	
	_, exists := s.streams[contextID]
	if !exists {
		return fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	// Get the context path
	contextPath, pathExists := s.streamPaths[contextID]
	
	// Clean up the stream file if it exists
	if pathExists && contextPath != "" {
		if err := os.Remove(contextPath); err != nil && !os.IsNotExist(err) {
			s.logger.Warning(fmt.Sprintf("Failed to remove context file %s: %v", contextPath, err))
		}
	}
	
	// Remove from memory
	delete(s.streams, contextID)
	delete(s.streamPaths, contextID)
	
	s.logger.Info(fmt.Sprintf("Closed streaming context %s", contextID))
	return nil
}

// GetContextStream retrieves a context stream by ID
func (s *ContextService) GetContextStream(ctx context.Context, contextID string) (*domain.ContextStream, error) {
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	return stream, nil
}

// GetContextLines retrieves a range of lines from a streaming context
func (s *ContextService) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*domain.ContextLineRange, error) {
	s.streamsMu.RLock()
	_, exists := s.streams[contextID]
	path, pathExists := s.streamPaths[contextID]
	s.streamsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	if !pathExists {
		return nil, fmt.Errorf("streaming context path not found: %s", contextID)
	}
	
	file, err := os.Open(path)
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
	
	return &domain.ContextLineRange{
		StartLine: startLine,
		EndLine:   endLine,
		Lines:     lines,
	}, nil
}

// GetContextContent returns paginated context content for memory-safe viewing
func (s *ContextService) GetContextContent(ctx context.Context, contextID string, startLine int, lineCount int) (interface{}, error) {
	// Try to get from streaming context first
	lines, err := s.GetContextLines(ctx, contextID, int64(startLine), int64(startLine+lineCount-1))
	if err == nil {
		return lines, nil
	}
	
	return nil, fmt.Errorf("context not found: %w", err)
}

// SuggestFiles анализирует задачу пользователя и текущее состояние проекта (все файлы)
// и возвращает список относительных путей к файлам, которые наиболее релевантны.
func (s *ContextService) SuggestFiles(ctx context.Context, task string, allFiles []*domain.FileNode) ([]string, error) {
	keywords := s.extractKeywordsForSuggestion(task)
	if len(keywords) == 0 {
		s.logger.Warning("Не удалось извлечь ключевые слова из задачи для авто-выбора.")
		return []string{}, nil
	}

	s.logger.Info("Извлеченные ключевые слова для поиска: " + strings.Join(keywords, ", "))

	// Плоский список: original + normalized
	type pair struct {
		orig  string
		lower string
	}
	var files []pair

	var traverse func([]*domain.FileNode)
	traverse = func(nodes []*domain.FileNode) {
		for _, n := range nodes {
			if !n.IsDir {
				orig := n.RelPath
				files = append(files, pair{
					orig:  orig,
					lower: strings.ToLower(strings.ReplaceAll(orig, "\\", "/")),
				})
			}
			if len(n.Children) > 0 {
				traverse(n.Children)
			}
		}
	}
	traverse(allFiles)

	parts := make([]string, 0, len(keywords))
	for _, k := range keywords {
		parts = append(parts, regexp.QuoteMeta(strings.ToLower(k)))
	}
	pattern := "(?i)(" + strings.Join(parts, "|") + ")"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	unique := make(map[string]struct{})
	for _, f := range files {
		if re.MatchString(f.lower) {
			unique[f.orig] = struct{}{}
		}
	}

	res := make([]string, 0, len(unique))
	for k := range unique {
		res = append(res, k)
	}
	s.logger.Info(fmt.Sprintf("Предложено %d релевантных файлов.", len(res)))
	return res, nil
}

// extractKeywordsForSuggestion извлекает потенциальные ключевые слова из строки задачи
func (s *ContextService) extractKeywordsForSuggestion(task string) []string {
	re := regexp.MustCompile(`[^\w\s]`)
	cleanTask := re.ReplaceAllString(task, " ")
	words := strings.Fields(strings.ToLower(cleanTask))
	var keywords []string
	for _, word := range words {
		if len(word) > 3 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}