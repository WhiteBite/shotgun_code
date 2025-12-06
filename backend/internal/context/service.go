package context

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Buffer pool for memory efficiency - prevents GC pressure
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 64*1024) // 64KB initial capacity
		return &buf
	},
}

// String builder pool for context building
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// Service handles all context management operations with memory-safe streaming by default
type Service struct {
	fileReader   domain.FileContentReader
	tokenCounter TokenCounter
	eventBus     domain.EventBus
	logger       domain.Logger
	contextDir   string

	// Streaming support with RWMutex for concurrent reads
	streams   map[string]*Stream
	streamsMu sync.RWMutex

	// Memory limits (reduced for safety)
	defaultMaxMemoryMB int
	defaultMaxTokens   int

	// Cleanup tracking
	lastCleanup time.Time

	// Worker pool for file scanning (fixed goroutine count)
	workerCount int
	
	// Shutdown coordination
	shutdownCh   chan struct{}
	shutdownOnce sync.Once
	wg           sync.WaitGroup
	
	// Metrics for monitoring
	activeOperations int64
	totalOperations  int64
	totalBytesRead   int64
}

// TokenCounter interface for token estimation
type TokenCounter interface {
	CountTokens(text string) int
}

// Stream represents a memory-safe streaming context
type Stream struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Files       []string  `json:"files"`
	ProjectPath string    `json:"projectPath"`
	TotalLines  int64     `json:"totalLines"`
	TotalChars  int64     `json:"totalChars"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TokenCount  int       `json:"tokenCount"`
	contextPath string    `json:"-"`
}

// LineRange represents a range of lines from a streaming context
type LineRange struct {
	StartLine int64    `json:"startLine"`
	EndLine   int64    `json:"endLine"`
	Lines     []string `json:"lines"`
}

// OutputFormat defines the format for context output
type OutputFormat string

const (
	FormatMarkdown OutputFormat = "markdown" // Default: ## File: path\n```lang\ncontent\n```
	FormatXML      OutputFormat = "xml"      // <file path="..."><content>...</content></file>
	FormatJSON     OutputFormat = "json"     // {"files": [{"path": "...", "content": "..."}]}
	FormatPlain    OutputFormat = "plain"    // --- File: path ---\ncontent
)

// BuildOptions controls how context is built
type BuildOptions struct {
	MaxTokens            int          `json:"maxTokens,omitempty"`
	MaxMemoryMB          int          `json:"maxMemoryMB,omitempty"`
	StripComments        bool         `json:"stripComments,omitempty"`
	IncludeManifest      bool         `json:"includeManifest,omitempty"`
	ForceStream          bool         `json:"forceStream,omitempty"`
	EnableProgressEvents bool         `json:"enableProgressEvents,omitempty"`
	OutputFormat         OutputFormat `json:"outputFormat,omitempty"`
}

// NewService creates a new unified context service
func NewService(
	fileReader domain.FileContentReader,
	tokenCounter TokenCounter,
	eventBus domain.EventBus,
	logger domain.Logger,
) (*Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	contextDir := filepath.Join(homeDir, ".shotgun-code", "contexts")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// Calculate optimal worker count based on CPU cores (max 16)
	workerCount := runtime.NumCPU()
	if workerCount > 16 {
		workerCount = 16
	}
	if workerCount < 2 {
		workerCount = 2
	}

	svc := &Service{
		fileReader:         fileReader,
		tokenCounter:       tokenCounter,
		eventBus:           eventBus,
		logger:             logger,
		contextDir:         contextDir,
		streams:            make(map[string]*Stream),
		defaultMaxMemoryMB: 30,   // Reduced from 50MB to 30MB for safety
		defaultMaxTokens:   5000, // Reduced from 8000 to 5000 for safety
		lastCleanup:        time.Now(),
		workerCount:        workerCount,
		shutdownCh:         make(chan struct{}),
	}

	// Start periodic cleanup with proper shutdown handling
	svc.wg.Add(1)
	go svc.periodicCleanup()

	return svc, nil
}

// Shutdown gracefully stops the service and waits for all operations to complete
// Safe to call multiple times - uses sync.Once to prevent double-close panic
func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down context service...")
	
	// Signal shutdown (safe to call multiple times)
	s.shutdownOnce.Do(func() {
		close(s.shutdownCh)
	})
	
	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		s.logger.Info("Context service shutdown complete")
		return nil
	case <-ctx.Done():
		s.logger.Warning("Context service shutdown timed out")
		return ctx.Err()
	}
}

// getBuffer gets a buffer from the pool
func getBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *[]byte) {
	*buf = (*buf)[:0] // Reset length but keep capacity
	bufferPool.Put(buf)
}

// getStringBuilder gets a string builder from the pool
func getStringBuilder() *strings.Builder {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// putStringBuilder returns a string builder to the pool
func putStringBuilder(sb *strings.Builder) {
	sb.Reset()
	stringBuilderPool.Put(sb)
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
func (s *Service) CreateStream(ctx context.Context, projectPath string, includedPaths []string, options *BuildOptions) (stream *Stream, err error) {
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
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close context file: %w", closeErr)
			} else {
				s.logger.Warning(fmt.Sprintf("failed to close context file: %v", closeErr))
			}
		}
	}()

	writer := bufio.NewWriter(file)
	defer func() {
		if flushErr := writer.Flush(); flushErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to flush writer: %w", flushErr)
			} else {
				s.logger.Warning(fmt.Sprintf("failed to flush writer: %v", flushErr))
			}
		}
	}()

	writeString := func(value string) error {
		if _, writeErr := writer.WriteString(value); writeErr != nil {
			return fmt.Errorf("failed to write streaming context: %w", writeErr)
		}
		return nil
	}

	var totalLines int64
	var totalChars int64
	var actualFiles []string
	var tokenCount int

	// Write header if manifest requested
	if options.IncludeManifest {
		header := fmt.Sprintf("# Streaming Context\nProject Path: %s\nGenerated: %s\n\n", projectPath, time.Now().Format(time.RFC3339))
		if err := writeString(header); err != nil {
			return nil, err
		}
		totalLines += int64(strings.Count(header, "\n"))
		totalChars += int64(len(header))
	}

	// Write file contents with token counting
	for _, filePath := range includedPaths {
		content, exists := contents[filePath]
		if !exists {
			s.logger.Warning(fmt.Sprintf("[CreateStream] File not found in contents: %s", filePath))
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
			if err := file.Close(); err != nil {
				s.logger.Warning(fmt.Sprintf("Failed to close partial context file: %v", err))
			}
			if err := os.Remove(contextPath); err != nil {
				s.logger.Warning(fmt.Sprintf("Failed to remove partial context file: %v", err))
			}
			return nil, fmt.Errorf("context would exceed token limit: %d > %d", tokenCount, options.MaxTokens)
		}

		// Determine output format (default to markdown)
		format := options.OutputFormat
		if format == "" {
			format = FormatMarkdown
		}

		// Escape content if needed
		escapedContent := s.escapeForFormat(content, format)

		// Write file header
		fileHeader := s.formatFileHeader(filePath, format)
		if err := writeString(fileHeader); err != nil {
			return nil, err
		}
		totalLines += int64(strings.Count(fileHeader, "\n"))
		totalChars += int64(len(fileHeader))

		// Write content
		if err := writeString(escapedContent); err != nil {
			return nil, err
		}
		totalLines += int64(strings.Count(escapedContent, "\n")) + 1
		totalChars += int64(len(escapedContent))

		// Write file footer
		fileFooter := s.formatFileFooter(format)
		if err := writeString(fileFooter); err != nil {
			return nil, err
		}
		totalLines += int64(strings.Count(fileFooter, "\n"))
		totalChars += int64(len(fileFooter))

		// Memory safety check - flush if getting too large
		if options.MaxMemoryMB > 0 && totalChars > int64(options.MaxMemoryMB*1024*1024)/2 {
			if err := writer.Flush(); err != nil {
				return nil, fmt.Errorf("failed to flush writer: %w", err)
			}
		}
	}

	now := time.Now()

	// Create context stream object
	stream = &Stream{
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

// GetContextLines retrieves a range of lines from a streaming context with limits
func (s *Service) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*LineRange, error) {
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("streaming context not found: %s", contextID)
	}
	
	// Limit maximum lines per request
	const maxLinesPerRequest = 10000
	if endLine-startLine > maxLinesPerRequest {
		return nil, fmt.Errorf("requested line range too large: %d lines (max: %d)", endLine-startLine, maxLinesPerRequest)
	}

	file, err := os.Open(stream.contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open context file: %w", err)
	}
	defer file.Close()

	// Use buffered reader for better performance
	reader := bufio.NewReaderSize(file, 64*1024) // 64KB buffer
	scanner := bufio.NewScanner(reader)
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

// CleanupOldStreams удаляет старые streaming контексты
func (s *Service) CleanupOldStreams(maxAge time.Duration) error {
	s.streamsMu.Lock()
	defer s.streamsMu.Unlock()
	
	now := time.Now()
	for id, stream := range s.streams {
		if now.Sub(stream.CreatedAt) > maxAge {
			// Удаляем файл контекста
			if err := os.Remove(stream.contextPath); err != nil && !os.IsNotExist(err) {
				s.logger.Warning(fmt.Sprintf("Failed to remove old context file %s: %v", stream.contextPath, err))
			}
			delete(s.streams, id)
			s.logger.Info(fmt.Sprintf("Cleaned up old streaming context: %s", id))
		}
	}
	
	// Ограничиваем количество активных streams
	const maxActiveStreams = 10
	if len(s.streams) > maxActiveStreams {
		// Удаляем самые старые
		type streamAge struct {
			id  string
			age time.Time
		}
		var ages []streamAge
		for id, stream := range s.streams {
			ages = append(ages, streamAge{id: id, age: stream.CreatedAt})
		}
		sort.Slice(ages, func(i, j int) bool {
			return ages[i].age.Before(ages[j].age)
		})
		
		// Удаляем лишние
		for i := 0; i < len(ages)-maxActiveStreams; i++ {
			id := ages[i].id
			if stream, exists := s.streams[id]; exists {
				os.Remove(stream.contextPath)
				delete(s.streams, id)
			}
		}
	}
	
	return nil
}

// periodicCleanup запускает периодическую очистку старых контекстов
func (s *Service) periodicCleanup() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCh:
			s.logger.Info("Periodic cleanup stopped due to shutdown")
			return
		case <-ticker.C:
			if err := s.CleanupOldStreams(24 * time.Hour); err != nil {
				s.logger.Warning(fmt.Sprintf("Failed to cleanup old streams: %v", err))
			}
			s.lastCleanup = time.Now()
			
			// Force GC after cleanup to release memory
			runtime.GC()
		}
	}
}

// GetMemoryStats возвращает статистику использования памяти
func (s *Service) GetMemoryStats() map[string]interface{} {
	s.streamsMu.RLock()
	defer s.streamsMu.RUnlock()

	var totalSize int64
	for _, stream := range s.streams {
		if info, err := os.Stat(stream.contextPath); err == nil {
			totalSize += info.Size()
		}
	}

	// Get runtime memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"active_streams":       len(s.streams),
		"total_disk_size_mb":   totalSize / (1024 * 1024),
		"last_cleanup":         s.lastCleanup,
		"active_operations":    atomic.LoadInt64(&s.activeOperations),
		"total_operations":     atomic.LoadInt64(&s.totalOperations),
		"total_bytes_read":     atomic.LoadInt64(&s.totalBytesRead),
		"worker_count":         s.workerCount,
		"heap_alloc_mb":        memStats.HeapAlloc / (1024 * 1024),
		"heap_sys_mb":          memStats.HeapSys / (1024 * 1024),
		"num_gc":               memStats.NumGC,
		"goroutines":           runtime.NumGoroutine(),
	}
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

	if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete JSON context: %w", err)
	}

	if err := os.Remove(streamPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete streaming context: %w", err)
	}

	// Remove from streams map
	s.streamsMu.Lock()
	delete(s.streams, contextID)
	s.streamsMu.Unlock()

	s.logger.Info(fmt.Sprintf("Deleted context %s", contextID))
	return nil
}

// Private helper methods

func (s *Service) validateLimits(options *BuildOptions) error {
	// Strict memory limit validation
	if options.MaxMemoryMB > 0 && options.MaxMemoryMB > 500 {
		return fmt.Errorf("memory limit cannot exceed 500MB for safety")
	}

	// Token limit validation - very permissive, frontend controls the actual limit
	// This is just a sanity check to prevent accidental huge requests
	const maxTokenLimit = 10000000 // 10M tokens - matches frontend max
	if options.MaxTokens > 0 && options.MaxTokens > maxTokenLimit {
		return fmt.Errorf("token limit cannot exceed %d (requested: %d). Please adjust settings on frontend.", maxTokenLimit, options.MaxTokens)
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
		TotalLines:  stream.TotalLines,
		TotalChars:  stream.TotalChars,
	}

	return domainContext, nil
}

// Legacy method removed - always use streaming for memory safety

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

// formatFileHeader returns the header for a file based on output format
func (s *Service) formatFileHeader(filePath string, format OutputFormat) string {
	switch format {
	case FormatXML:
		return fmt.Sprintf("<file path=\"%s\">\n<content>\n", filePath)
	case FormatJSON:
		return "" // JSON is handled separately
	case FormatPlain:
		return fmt.Sprintf("--- File: %s ---\n", filePath)
	default: // FormatMarkdown
		ext := filepath.Ext(filePath)
		lang := ""
		if len(ext) > 1 {
			lang = ext[1:]
		}
		return fmt.Sprintf("## File: %s\n\n```%s\n", filePath, lang)
	}
}

// formatFileFooter returns the footer for a file based on output format
func (s *Service) formatFileFooter(format OutputFormat) string {
	switch format {
	case FormatXML:
		return "\n</content>\n</file>\n\n"
	case FormatJSON:
		return "" // JSON is handled separately
	case FormatPlain:
		return "\n\n"
	default: // FormatMarkdown
		return "\n```\n\n"
	}
}

// formatContextHeader returns the header for the entire context
func (s *Service) formatContextHeader(projectPath string, fileCount int, format OutputFormat) string {
	switch format {
	case FormatXML:
		return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<context project=\"%s\" files=\"%d\">\n", filepath.Base(projectPath), fileCount)
	case FormatJSON:
		return fmt.Sprintf("{\"project\":\"%s\",\"files\":[", filepath.Base(projectPath))
	case FormatPlain:
		return fmt.Sprintf("=== Context: %s (%d files) ===\n\n", filepath.Base(projectPath), fileCount)
	default: // FormatMarkdown
		return fmt.Sprintf("# Context: %s\n\nFiles: %d\n\n---\n\n", filepath.Base(projectPath), fileCount)
	}
}

// formatContextFooter returns the footer for the entire context
func (s *Service) formatContextFooter(format OutputFormat) string {
	switch format {
	case FormatXML:
		return "</context>\n"
	case FormatJSON:
		return "]}"
	case FormatPlain:
		return "=== End of Context ===\n"
	default: // FormatMarkdown
		return "---\n\n*End of context*\n"
	}
}

// escapeForFormat escapes content based on output format
func (s *Service) escapeForFormat(content string, format OutputFormat) string {
	switch format {
	case FormatXML:
		// Escape XML special characters
		content = strings.ReplaceAll(content, "&", "&amp;")
		content = strings.ReplaceAll(content, "<", "&lt;")
		content = strings.ReplaceAll(content, ">", "&gt;")
		return content
	case FormatJSON:
		// JSON escaping handled by json.Marshal
		return content
	default:
		return content
	}
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

// ============ CONTEXT BUILDER INTERFACE IMPLEMENTATION ============
// Implements domain.ContextBuilder interface

// BuildContextSummary implements domain.ContextBuilder.BuildContext - builds context and returns ContextSummary
// This method satisfies the domain.ContextBuilder interface
func (s *Service) BuildContextSummary(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	atomic.AddInt64(&s.activeOperations, 1)
	defer atomic.AddInt64(&s.activeOperations, -1)
	atomic.AddInt64(&s.totalOperations, 1)

	if ctx == nil {
		ctx = context.Background()
	}

	// Add timeout for safety
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	buildOpts := s.convertBuildOptions(options)
	domainCtx, err := s.BuildContext(ctx, projectPath, includedPaths, buildOpts)
	if err != nil {
		return nil, err
	}

	// Convert to ContextSummary
	summary := &domain.ContextSummary{
		ID:          domainCtx.ID,
		ProjectPath: domainCtx.ProjectPath,
		FileCount:   len(domainCtx.Files),
		TotalSize:   domainCtx.TotalChars,
		LineCount:   int(domainCtx.TotalLines),
		TokenCount:  domainCtx.TokenCount,
		CreatedAt:   domainCtx.CreatedAt,
		UpdatedAt:   domainCtx.UpdatedAt,
		Status:      "ready",
		Metadata: domain.ContextMetadata{
			SelectedFiles: domainCtx.Files,
			ProjectPath:   domainCtx.ProjectPath,
		},
	}

	// Save summary to disk for persistence
	if err := s.SaveContextSummary(summary); err != nil {
		s.logger.Warning(fmt.Sprintf("Failed to save context summary: %v", err))
		// Don't fail the operation, just log warning
	}

	return summary, nil
}

func (s *Service) convertBuildOptions(opts *domain.ContextBuildOptions) *BuildOptions {
	if opts == nil {
		return nil
	}
	return &BuildOptions{
		MaxTokens:            opts.MaxTokens,
		MaxMemoryMB:          opts.MaxMemoryMB,
		StripComments:        opts.StripComments,
		IncludeManifest:      opts.IncludeManifest,
		ForceStream:          true, // Always stream
		EnableProgressEvents: true,
	}
}

// ============ CONTEXT REPOSITORY INTERFACE IMPLEMENTATION ============

// SaveContextSummary persists context metadata
func (s *Service) SaveContextSummary(summary *domain.ContextSummary) error {
	if summary == nil {
		return fmt.Errorf("context summary is nil")
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context summary: %w", err)
	}

	summaryPath := filepath.Join(s.contextDir, summary.ID+".summary.json")
	if err := os.WriteFile(summaryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context summary: %w", err)
	}

	return nil
}

// GetContextSummary retrieves context metadata by ID
func (s *Service) GetContextSummary(ctx context.Context, contextID string) (*domain.ContextSummary, error) {
	summaryPath := filepath.Join(s.contextDir, contextID+".summary.json")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("context summary not found: %s", contextID)
		}
		return nil, fmt.Errorf("failed to read context summary: %w", err)
	}

	var summary domain.ContextSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context summary: %w", err)
	}

	return &summary, nil
}

// GetProjectContextSummaries lists all context summaries for a project
func (s *Service) GetProjectContextSummaries(ctx context.Context, projectPath string) ([]*domain.ContextSummary, error) {
	entries, err := os.ReadDir(s.contextDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*domain.ContextSummary{}, nil
		}
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}

	summaries := make([]*domain.ContextSummary, 0)
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".summary.json") {
			continue
		}

		contextID := strings.TrimSuffix(entry.Name(), ".summary.json")
		summary, err := s.GetContextSummary(ctx, contextID)
		if err != nil {
			s.logger.Warning(fmt.Sprintf("Failed to load context summary %s: %v", contextID, err))
			continue
		}

		if summary.ProjectPath == projectPath {
			summaries = append(summaries, summary)
		}
	}

	return summaries, nil
}

// ReadContextChunk returns a chunk of context content (memory-safe pagination)
func (s *Service) ReadContextChunk(ctx context.Context, contextID string, startLine int, lineCount int) (*domain.ContextChunk, error) {
	if lineCount <= 0 {
		lineCount = 1000
	}
	if startLine < 1 {
		startLine = 1
	}

	// Limit max lines per request
	const maxLinesPerRequest = 10000
	if lineCount > maxLinesPerRequest {
		lineCount = maxLinesPerRequest
	}

	// Try streaming context first
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()

	var contextPath string
	if exists {
		contextPath = stream.contextPath
	} else {
		// Try .ctx file
		contextPath = filepath.Join(s.contextDir, contextID+".ctx")
		if _, err := os.Stat(contextPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("context not found: %s", contextID)
		}
	}

	file, err := os.Open(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open context file: %w", err)
	}
	defer file.Close()

	// Use pooled buffer for better performance
	reader := bufio.NewReaderSize(file, 64*1024)
	scanner := bufio.NewScanner(reader)

	var lines []string
	currentLine := 0
	hasMore := false

	for scanner.Scan() {
		currentLine++
		if currentLine < startLine {
			continue
		}
		if len(lines) >= lineCount {
			hasMore = true
			break
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading context file: %w", err)
	}

	endLine := startLine + len(lines) - 1
	if len(lines) == 0 {
		endLine = startLine - 1
	}

	return &domain.ContextChunk{
		Lines:     lines,
		StartLine: startLine,
		EndLine:   endLine,
		HasMore:   hasMore,
		ChunkID:   fmt.Sprintf("%s:%d", contextID, startLine),
		ContextID: contextID,
	}, nil
}

// ReadContextContent returns full context content as string
func (s *Service) ReadContextContent(ctx context.Context, contextID string) (string, error) {
	// Try streaming context first
	s.streamsMu.RLock()
	stream, exists := s.streams[contextID]
	s.streamsMu.RUnlock()

	var contextPath string
	if exists {
		contextPath = stream.contextPath
	} else {
		contextPath = filepath.Join(s.contextDir, contextID+".ctx")
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("context content not found: %s", contextID)
		}
		return "", fmt.Errorf("failed to read context content: %w", err)
	}

	atomic.AddInt64(&s.totalBytesRead, int64(len(data)))
	return string(data), nil
}

// ============ WORKER POOL FOR PARALLEL FILE PROCESSING ============

type fileResult struct {
	path    string
	content string
	err     error
}

// processFilesParallel processes files using a worker pool
func (s *Service) processFilesParallel(ctx context.Context, projectPath string, paths []string, processor func(string, string) error) error {
	if len(paths) == 0 {
		return nil
	}

	// Create buffered channels
	jobs := make(chan string, len(paths))
	results := make(chan fileResult, len(paths))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case path, ok := <-jobs:
					if !ok {
						return
					}
					// Read file content
					contents, err := s.fileReader.ReadContents(ctx, []string{path}, projectPath, nil)
					if err != nil {
						results <- fileResult{path: path, err: err}
						continue
					}
					content, exists := contents[path]
					if !exists {
						results <- fileResult{path: path, err: fmt.Errorf("file not found: %s", path)}
						continue
					}
					results <- fileResult{path: path, content: content}
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, path := range paths {
			select {
			case <-ctx.Done():
				break
			case jobs <- path:
			}
		}
		close(jobs)
	}()

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errs []error
	for result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}
		if err := processor(result.path, result.content); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("processing errors: %v", errs)
	}
	return nil
}
