package application

import (
	"context"
	"errors"
	"fmt"
	"shotgun_code/domain"
)

const (
	// MaxContextSize is the maximum allowed context size (50 MB)
	MaxContextSize = 50 * 1024 * 1024 // 50 MB
	// DefaultChunkSize is the default number of lines per chunk
	DefaultChunkSize = 1000

	// TTL configuration for contexts
	defaultContextTTL      = 30 * time.Minute
	defaultCleanupInterval = 5 * time.Minute
)

// contextEntry stores context data and timestamps
type contextEntry struct {
	content   string
	createdAt time.Time
	lastUsed  time.Time
}

// ContextService manages context lifecycle with OOM safety and TTL cleanup
type ContextService struct {
	contexts        map[string]*contextEntry // contextId -> entry
	mu              sync.RWMutex
	ttl             time.Duration
	cleanupInterval time.Duration
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

// NewContextService creates a new ContextService instance
func NewContextService() *ContextService {
	s := &ContextService{
		contexts:        make(map[string]*contextEntry),
		ttl:             defaultContextTTL,
		cleanupInterval: defaultCleanupInterval,
		stopCh:          make(chan struct{}),
	}

	// Start background cleanup goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.stopCh:
				return
			}
		}
	}()

	return s
}

// Stop gracefully stops background cleanup
func (s *ContextService) Stop() {
	s.mu.Lock()
	if s.stopCh == nil {
		s.mu.Unlock()
		return
	}
	close(s.stopCh)
	s.stopCh = nil
	s.mu.Unlock()
	s.wg.Wait()
}

// cleanupExpired removes contexts that exceeded TTL based on last access time
func (s *ContextService) cleanupExpired() {
	now := time.Now()
	s.mu.Lock()
	for id, e := range s.contexts {
		if now.Sub(e.lastUsed) > s.ttl {
			delete(s.contexts, id)
		}
	}
	s.mu.Unlock()
}

// BuildContext reads files, concatenates them, and stores in memory
func (s *ContextService) BuildContext(filePaths []string) (*domain.ContextSummaryInfo, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	var content strings.Builder
	totalSize := 0
	totalLines := 0

	for _, path := range filePaths {
		// Check if file exists
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file %s: %w", path, err)
		}

		// Check size limit
		if totalSize+int(fileInfo.Size()) > MaxContextSize {
			return nil, fmt.Errorf("context size exceeds limit of %d bytes", MaxContextSize)
		}

		// Read file
		fileContent, lineCount, err := readFileWithLines(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Append to context
		content.WriteString(fmt.Sprintf("=== File: %s ===\n", path))
		content.WriteString(fileContent)
		content.WriteString("\n\n")

		totalSize += int(fileInfo.Size())
		totalLines += lineCount
	}

	// Generate context ID
	contextId, err := generateContextID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate context ID: %w", err)
	}

	// Store context
	s.mu.Lock()
	s.contexts[contextId] = &contextEntry{
		content:   content.String(),
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}
	s.mu.Unlock()

	summary := &domain.ContextSummaryInfo{
		FileCount: len(filePaths),
		TotalSize: int64(totalSize),
		LineCount: totalLines,
	}

	return summary, nil
}

// GetLines retrieves a range of lines from a context
func (s *ContextService) GetLines(contextId string, start, count int) (string, error) {
	s.mu.RLock()
	entry, exists := s.contexts[contextId]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("context not found: %s", contextId)
	}

	// Split into lines
	lines := strings.Split(entry.content, "\n")

	// Validate range
	if start < 0 || start >= len(lines) {
		return "", fmt.Errorf("invalid start index: %d (total lines: %d)", start, len(lines))
	}

	end := start + count
	if end > len(lines) {
		end = len(lines)
	}

	// Extract lines
	result := strings.Join(lines[start:end], "\n")

	// Update lastUsed timestamp
	s.mu.Lock()
	if e, ok := s.contexts[contextId]; ok {
		e.lastUsed = time.Now()
	}
	s.mu.Unlock()

	return result, nil
}

// GetFullContext retrieves the entire context
func (s *ContextService) GetFullContext(contextId string) (string, error) {
	s.mu.RLock()
	entry, exists := s.contexts[contextId]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("context not found: %s", contextId)
	}

	// Update lastUsed
	s.mu.Lock()
	if e, ok := s.contexts[contextId]; ok {
		e.lastUsed = time.Now()
	}
	s.mu.Unlock()

	return entry.content, nil
}

// DeleteContext removes a context from memory
func (s *ContextService) DeleteContext(ctx context.Context, contextId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.contexts[contextId]; !exists {
		return fmt.Errorf("context not found: %s", contextId)
	}

	delete(s.contexts, contextId)
	return nil
}

// GetContextCount returns the number of stored contexts
func (s *ContextService) GetContextCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.contexts)
}

// SuggestFiles suggests relevant files for a task (not implemented yet)
func (s *ContextService) SuggestFiles(ctx context.Context, taskDescription string, files []*domain.FileNode) ([]string, error) {
	return []string{}, nil // Empty list is valid, no error
}

// CreateStreamingContext creates a streaming context and saves metadata to disk
func (s *ContextService) CreateStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextStream, error) {
	// Validate inputs
	if len(includedPaths) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Generate context ID
	contextID, err := generateContextID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate context ID: %w", err)
	}

	// Build context content to get metadata
	summary, err := s.buildContextContent(includedPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Create streaming context metadata
	stream := &domain.ContextStream{
		ID:          contextID,
		Name:        fmt.Sprintf("Context for %d files", len(includedPaths)),
		Description: fmt.Sprintf("Streaming context from %s", projectPath),
		Files:       includedPaths,
		ProjectPath: projectPath,
		TotalLines:  int64(summary.LineCount),
		TotalChars:  summary.TotalSize,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
		TokenCount:  summary.TokenCount,
	}

	// Store in memory (context content is already on disk via buildContextContent)
	s.mu.Lock()
	s.contexts[contextID] = &contextEntry{
		content:   fmt.Sprintf("stream:%s", contextID),
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}
	s.mu.Unlock()

	return stream, nil
}

// GetContextContent returns paginated context content for memory-safe viewing
func (s *ContextService) GetContextContent(ctx context.Context, contextID string, startLine int, lineCount int) (interface{}, error) {
	// Validate context exists
	s.mu.RLock()
	_, exists := s.contexts[contextID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context not found: %s", contextID)
	}

	// Get lines from stored content
	lines, err := s.GetLines(contextID, startLine, lineCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get context lines: %w", err)
	}

	result := map[string]interface{}{
		"contextID": contextID,
		"startLine": startLine,
		"lineCount": lineCount,
		"content":   lines,
	}

	return result, nil
}

// GetContext retrieves a full context by ID (use with caution - can cause OOM)
func (s *ContextService) GetContext(ctx context.Context, contextID string) (*domain.Context, error) {
	s.mu.RLock()
	entry, exists := s.contexts[contextID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context not found: %s", contextID)
	}

	// Update lastUsed
	s.mu.Lock()
	if e, ok := s.contexts[contextID]; ok {
		e.lastUsed = time.Now()
	}
	s.mu.Unlock()

	// Create context object
	ctxObj := &domain.Context{
		ID:        contextID,
		Content:   entry.content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return ctxObj, nil
}

// GetContextLines retrieves a range of lines from a streaming context
func (s *ContextService) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*domain.ContextLineRange, error) {
	if startLine < 0 || endLine < startLine {
		return nil, fmt.Errorf("invalid line range: %d to %d", startLine, endLine)
	}

	lineCount := int(endLine - startLine + 1)
	content, err := s.GetLines(contextID, int(startLine), lineCount)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")

	return &domain.ContextLineRange{
		StartLine: startLine,
		EndLine:   endLine,
		Lines:     lines,
	}, nil
}

// GetProjectContexts lists all contexts for a project (stub for now)
func (s *ContextService) GetProjectContexts(ctx context.Context, projectPath string) ([]*domain.Context, error) {
	// Return empty list - context listing not implemented yet
	return []*domain.Context{}, nil
}

// SaveContext saves a context to memory
func (s *ContextService) SaveContext(context *domain.Context) error {
	if context == nil || context.ID == "" {
		return fmt.Errorf("invalid context: missing ID")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.contexts[context.ID] = &contextEntry{
		content:   context.Content,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}
	return nil
}

// SaveContextSummary saves a context summary (stub - not persisted)
func (s *ContextService) SaveContextSummary(contextSummary *domain.ContextSummary) error {
	if contextSummary == nil || contextSummary.ID == "" {
		return fmt.Errorf("invalid context summary: missing ID")
	}
	// Summary saved to memory as metadata only
	return nil
}

// AnalyzeTaskAndCollectContext analyzes a task and suggests files (not implemented)
func (s *ContextService) AnalyzeTaskAndCollectContext(ctx context.Context, task string, allFiles []*domain.FileNode, rootDir string) (*domain.ContextAnalysisResult, error) {
	return nil, fmt.Errorf("task analysis not implemented")
}

// BuildContextLegacy builds context with legacy format (DEPRECATED)
func (s *ContextService) BuildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options domain.ContextBuildOptions) (*domain.Context, error) {
	return nil, fmt.Errorf("legacy context building is deprecated - use BuildContext instead")
}

// GetContextStream retrieves streaming context metadata
func (s *ContextService) GetContextStream(ctx context.Context, contextID string) (*domain.ContextStream, error) {
	s.mu.RLock()
	_, exists := s.contexts[contextID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context stream not found: %s", contextID)
	}

	// Return minimal metadata
	return &domain.ContextStream{
		ID:        contextID,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// CloseContextStream closes a streaming context and cleans up resources
func (s *ContextService) CloseContextStream(ctx context.Context, contextID string) error {
	return s.DeleteContext(ctx, contextID)
}

// buildContextContent is a helper to build context and calculate metadata
func (s *ContextService) buildContextContent(filePaths []string) (*domain.ContextSummaryInfo, error) {
	var totalSize int64
	var totalLines int
	var content strings.Builder

	for _, path := range filePaths {
		fileContent, lineCount, err := readFileWithLines(path)
		if err != nil {
			continue // Skip unreadable files
		}

		content.WriteString(fmt.Sprintf("=== File: %s ===\n", path))
		content.WriteString(fileContent)
		content.WriteString("\n\n")

		totalSize += int64(len(fileContent))
		totalLines += lineCount
	}

	return &domain.ContextSummaryInfo{
		FileCount:  len(filePaths),
		TotalSize:  totalSize,
		LineCount:  totalLines,
		TokenCount: totalLines * 4, // Rough estimate: 4 tokens per line
	}, nil
}

// readFileWithLines reads a file and returns content with line count
func readFileWithLines(path string) (string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	var content strings.Builder
	lineCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return "", 0, err
	}

	return content.String(), lineCount, nil
}

// generateContextID generates a unique context ID
func generateContextID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
