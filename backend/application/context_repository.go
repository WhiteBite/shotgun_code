package application

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
)

const (
	summarySuffix   = ".summary.json"
	contentSuffix   = ".ctx"
	defaultFilePerm = 0o644
	directoryPerm   = 0o755
)

// ContextRepositoryImpl implements the ContextRepository interface
type ContextRepositoryImpl struct {
	logger     domain.Logger
	contextDir string
}

// NewContextRepository creates a new ContextRepository implementation
func NewContextRepository(logger domain.Logger, contextDir string) *ContextRepositoryImpl {
	return &ContextRepositoryImpl{
		logger:     logger,
		contextDir: contextDir,
	}
}

// SaveContextSummary persists lightweight context metadata on disk
func (cr *ContextRepositoryImpl) SaveContextSummary(summary *domain.ContextSummary) error {
	if summary == nil {
		return errors.New("context summary is nil")
	}

	if summary.ID == "" {
		return errors.New("context summary ID is required")
	}

	if summary.Metadata.ContentPath == "" {
		summary.Metadata.ContentPath = summary.ID + contentSuffix
	}

	if err := os.MkdirAll(cr.contextDir, directoryPerm); err != nil {
		return fmt.Errorf("failed to ensure context directory: %w", err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context summary: %w", err)
	}

	if err := os.WriteFile(cr.summaryPath(summary.ID), data, defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write context summary file: %w", err)
	}

	return nil
}

// GetContextSummary retrieves persisted context metadata by ID
func (cr *ContextRepositoryImpl) GetContextSummary(ctx context.Context, contextID string) (*domain.ContextSummary, error) {
	summaryPath := cr.summaryPath(contextID)
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

// GetProjectContextSummaries lists context metadata for a project path
func (cr *ContextRepositoryImpl) GetProjectContextSummaries(ctx context.Context, projectPath string) ([]*domain.ContextSummary, error) {
	entries, err := os.ReadDir(cr.contextDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*domain.ContextSummary{}, nil
		}
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}

	var summaries []*domain.ContextSummary

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), summarySuffix) {
			continue
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		contextID := strings.TrimSuffix(entry.Name(), summarySuffix)
		summary, err := cr.GetContextSummary(ctx, contextID)
		if err != nil {
			cr.logger.Warning(fmt.Sprintf("Failed to load context summary %s: %v", contextID, err))
			continue
		}

		if summary.ProjectPath == projectPath {
			summaries = append(summaries, summary)
		}
	}

	return summaries, nil
}

// DeleteContext removes context metadata and content from disk
func (cr *ContextRepositoryImpl) DeleteContext(ctx context.Context, contextID string) error {
	summary, err := cr.GetContextSummary(ctx, contextID)
	if err != nil {
		return err
	}

	summaryPath := cr.summaryPath(contextID)
	if err := os.Remove(summaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete context summary: %w", err)
	}

	contentPath := cr.resolveContentPath(summary)
	if err := os.Remove(contentPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete context content: %w", err)
	}

	cr.logger.Info(fmt.Sprintf("Deleted context %s", contextID))
	return nil
}

// ReadContextChunk returns a memory-safe chunk of context content
func (cr *ContextRepositoryImpl) ReadContextChunk(ctx context.Context, contextID string, startLine int, lineCount int) (*domain.ContextChunk, error) {
	if lineCount <= 0 {
		return nil, errors.New("lineCount must be greater than zero")
	}

	if startLine < 1 {
		startLine = 1
	}

	summary, err := cr.GetContextSummary(ctx, contextID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(cr.resolveContentPath(summary))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("context content not found: %s", contextID)
		}
		return nil, fmt.Errorf("failed to open context content: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 128*1024)
	scanner.Buffer(buf, 1024*1024)

	var (
		currentLine = 0
		lines       []string
		hasMore     bool
	)

	for scanner.Scan() {
		currentLine++

		if currentLine < startLine {
			continue
		}

		if len(lines) < lineCount {
			lines = append(lines, scanner.Text())
			continue
		}

		hasMore = true
		break
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("context line exceeds maximum supported length: %w", err)
		}
		return nil, fmt.Errorf("failed to scan context content: %w", err)
	}

	chunk := &domain.ContextChunk{
		Lines:     lines,
		StartLine: startLine,
		EndLine:   startLine + len(lines) - 1,
		HasMore:   hasMore,
		ChunkID:   fmt.Sprintf("%s:%d", contextID, startLine),
		ContextID: contextID,
	}

	if len(lines) == 0 {
		chunk.EndLine = startLine - 1
	}

	return chunk, nil
}

// ReadContextContent returns the full context content as a string
func (cr *ContextRepositoryImpl) ReadContextContent(ctx context.Context, contextID string) (string, error) {
	summary, err := cr.GetContextSummary(ctx, contextID)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(cr.resolveContentPath(summary))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("context content not found: %s", contextID)
		}
		return "", fmt.Errorf("failed to read context content: %w", err)
	}

	return string(data), nil
}

func (cr *ContextRepositoryImpl) summaryPath(contextID string) string {
	return filepath.Join(cr.contextDir, contextID+summarySuffix)
}

func (cr *ContextRepositoryImpl) resolveContentPath(summary *domain.ContextSummary) string {
	contentName := summary.Metadata.ContentPath
	if contentName == "" {
		contentName = summary.ID + contentSuffix
	}
	if filepath.IsAbs(contentName) {
		return contentName
	}
	return filepath.Join(cr.contextDir, contentName)
}
