package application

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"shotgun_code/domain"
	"strings"
	"sync"
)

const (
	// MaxContextSize is the maximum allowed context size (50 MB)
	MaxContextSize = 50 * 1024 * 1024 // 50 MB
	// DefaultChunkSize is the default number of lines per chunk
	DefaultChunkSize = 1000
)

// ContextService manages context lifecycle with OOM safety
type ContextService struct {
	contexts map[string]string // contextId -> content
	mu       sync.RWMutex
}

// NewContextService creates a new ContextService instance
func NewContextService() *ContextService {
	return &ContextService{
		contexts: make(map[string]string),
	}
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
	s.contexts[contextId] = content.String()
	s.mu.Unlock()

	summary := &domain.ContextSummaryInfo{
		FileCount:      len(filePaths),
		TotalSize:      int64(totalSize),
		LineCount:      totalLines,
	}

	return summary, nil
}

// GetLines retrieves a range of lines from a context
func (s *ContextService) GetLines(contextId string, start, count int) (string, error) {
	s.mu.RLock()
	content, exists := s.contexts[contextId]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("context not found: %s", contextId)
	}

	// Split into lines
	lines := strings.Split(content, "\n")

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
	return result, nil
}

// GetFullContext retrieves the entire context
func (s *ContextService) GetFullContext(contextId string) (string, error) {
	s.mu.RLock()
	content, exists := s.contexts[contextId]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("context not found: %s", contextId)
	}

	return content, nil
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

func (s *ContextService) SuggestFiles(ctx context.Context, taskDescription string, files []*domain.FileNode) ([]string, error) {
	return []string{}, nil
}

func (s *ContextService) CreateStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextStream, error) {
	return nil, nil
}

func (s *ContextService) GetContextContent(ctx context.Context, contextID string, startLine int, lineCount int) (interface{}, error) {
    return nil, nil
}

func (s *ContextService) GetContext(ctx context.Context, contextID string) (*domain.Context, error) {
	return nil, nil
}

func (s *ContextService) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*domain.ContextLineRange, error) {
	return nil, nil
}

func (s *ContextService) GetProjectContexts(ctx context.Context, projectPath string) ([]*domain.Context, error) {
	return nil, nil
}

func (s *ContextService) SaveContext(context *domain.Context) error {
	return nil
}

func (s *ContextService) SaveContextSummary(contextSummary *domain.ContextSummary) error {
	return nil
}

func (s *ContextService) AnalyzeTaskAndCollectContext(ctx context.Context, task string, allFiles []*domain.FileNode, rootDir string) (*domain.ContextAnalysisResult, error) {
	return nil, nil
}

func (s *ContextService) BuildContextLegacy(ctx context.Context, projectPath string, includedPaths []string, options domain.ContextBuildOptions) (*domain.Context, error) {
	return nil, nil
}

func (s *ContextService) GetContextStream(ctx context.Context, contextID string) (*domain.ContextStream, error) {
	return nil, nil
}

func (s *ContextService) CloseContextStream(ctx context.Context, contextID string) error {
	return nil
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
