package application

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextBuilderImpl implements the ContextBuilder interface
type ContextBuilderImpl struct {
	fileReader       domain.FileContentReader
	tokenCounter     domain.TokenCounter
	logger           domain.Logger
	settingsService  *SettingsService
	bus              domain.EventBus
	opaService       domain.OPAService
	pathProvider     domain.PathProvider
	fileSystemWriter domain.FileSystemWriter
	commentStripper  domain.CommentStripper
	repository       domain.ContextRepository
	contextDir       string
}

// NewContextBuilder creates a new ContextBuilder implementation
func NewContextBuilder(
	fileReader domain.FileContentReader,
	tokenCounter domain.TokenCounter,
	logger domain.Logger,
	settingsService *SettingsService,
	bus domain.EventBus,
	opaService domain.OPAService,
	pathProvider domain.PathProvider,
	fileSystemWriter domain.FileSystemWriter,
	commentStripper domain.CommentStripper,
	repository domain.ContextRepository,
	contextDir string,
) *ContextBuilderImpl {
	return &ContextBuilderImpl{
		fileReader:       fileReader,
		tokenCounter:     tokenCounter,
		logger:           logger,
		settingsService:  settingsService,
		bus:              bus,
		opaService:       opaService,
		pathProvider:     pathProvider,
		fileSystemWriter: fileSystemWriter,
		commentStripper:  commentStripper,
		repository:       repository,
		contextDir:       contextDir,
	}
}

// BuildContext builds a context from project files and returns a ContextSummary to prevent OOM issues
func (cb *ContextBuilderImpl) BuildContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextSummary, error) {
	if options == nil {
		options = &domain.ContextBuildOptions{}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()
	cb.logger.Info(fmt.Sprintf("Building context for project: %s, files: %d", projectPath, len(includedPaths)))

	if err := cb.fileSystemWriter.MkdirAll(cb.contextDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to prepare context directory: %w", err)
	}

	sortedPaths := make([]string, len(includedPaths))
	copy(sortedPaths, includedPaths)
	sort.Strings(sortedPaths)

	contents, err := cb.fileReader.ReadContents(ctx, sortedPaths, projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read project files: %w", err)
	}

	contextID := uuid.New().String()
	contextFileName := contextID + ".ctx"
	contextPath := filepath.Join(cb.contextDir, contextFileName)

	file, err := os.Create(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create context file: %w", err)
	}

	writer := bufio.NewWriter(file)
	cleanup := func() {
		writer.Flush()
		file.Close()
	}
	defer cleanup()

	var (
		totalBytes  int64
		totalTokens int
		totalLines  int
		processed   []string
	)

	if options.IncludeManifest {
		manifest := cb.buildManifest(projectPath, sortedPaths)
		if _, err := writer.WriteString(manifest); err != nil {
			cb.cleanupContextFile(contextPath)
			return nil, fmt.Errorf("failed to write manifest: %w", err)
		}
		totalBytes += int64(len(manifest))
	}

	for _, relPath := range sortedPaths {
		content, ok := contents[relPath]
		if !ok {
			cb.logger.Warning(fmt.Sprintf("Skipping missing file content: %s", relPath))
			continue
		}

		if options.StripComments && cb.commentStripper != nil {
			content = cb.commentStripper.Strip(content, relPath)
		}

		if err := cb.writeFileSection(writer, relPath, content); err != nil {
			cb.cleanupContextFile(contextPath)
			return nil, err
		}

		totalBytes += int64(len(content))
		totalTokens += cb.tokenCounter(content)
		totalLines += countLines(content)
		processed = append(processed, relPath)
	}

	if err := writer.Flush(); err != nil {
		cb.cleanupContextFile(contextPath)
		return nil, fmt.Errorf("failed to flush context file: %w", err)
	}

	if options.MaxTokens > 0 && totalTokens > options.MaxTokens {
		cb.cleanupContextFile(contextPath)
		return nil, fmt.Errorf("context exceeds token limit: %d > %d", totalTokens, options.MaxTokens)
	}

	now := time.Now()
	summary := &domain.ContextSummary{
		ID:          contextID,
		ProjectPath: projectPath,
		FileCount:   len(processed),
		TotalSize:   totalBytes,
		TokenCount:  totalTokens,
		LineCount:   totalLines,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      "ready",
		Metadata: domain.ContextMetadata{
			BuildDuration: time.Since(start).Milliseconds(),
			LastModified:  now,
			SelectedFiles: processed,
			BuildOptions:  copyBuildOptions(options),
			Warnings:      []string{},
			Errors:        []string{},
			ContentPath:   contextFileName,
		},
	}

	if err := cb.repository.SaveContextSummary(summary); err != nil {
		cb.cleanupContextFile(contextPath)
		return nil, fmt.Errorf("failed to persist context summary: %w", err)
	}

	if cb.bus != nil {
		cb.bus.Emit("shotgunContextBuilt", summary)
	}

	cb.logger.Info(fmt.Sprintf("Built context summary %s with %d files", contextID, len(processed)))
	return summary, nil
}
func (cb *ContextBuilderImpl) buildManifest(projectPath string, files []string) string {
	if len(files) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("# Project Context\n")
	builder.WriteString(fmt.Sprintf("Project Path: %s\n", projectPath))
	builder.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	builder.WriteString("## Included Files\n")
	for _, file := range files {
		builder.WriteString("- ")
		builder.WriteString(file)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	return builder.String()
}

func (cb *ContextBuilderImpl) writeFileSection(writer *bufio.Writer, relPath, content string) error {
	if relPath == "" {
		return fmt.Errorf("file path is empty")
	}

	if _, err := writer.WriteString(fmt.Sprintf("## File: %s\n\n", relPath)); err != nil {
		return fmt.Errorf("failed to write file header: %w", err)
	}

	lang := strings.TrimPrefix(filepath.Ext(relPath), ".")
	if _, err := writer.WriteString("```"); err != nil {
		return fmt.Errorf("failed to start code fence: %w", err)
	}
	if lang != "" {
		if _, err := writer.WriteString(lang); err != nil {
			return fmt.Errorf("failed to write code fence language: %w", err)
		}
	}
	if _, err := writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to terminate code fence header: %w", err)
	}

	if _, err := writer.WriteString(content); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	if !strings.HasSuffix(content, "\n") {
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to append newline to file content: %w", err)
		}
	}

	if _, err := writer.WriteString("```\n\n"); err != nil {
		return fmt.Errorf("failed to close code fence: %w", err)
	}

	return nil
}

func (cb *ContextBuilderImpl) cleanupContextFile(path string) {
	if path == "" {
		return
	}
	if err := cb.fileSystemWriter.Remove(path); err != nil {
		cb.logger.Warning(fmt.Sprintf("failed to cleanup context file %s: %v", path, err))
	}
}

func copyBuildOptions(options *domain.ContextBuildOptions) *domain.ContextBuildOptions {
	if options == nil {
		return nil
	}
	clone := *options
	return &clone
}

func countLines(content string) int {
	if content == "" {
		return 0
	}

	lines := strings.Count(content, "\n")
	if !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}
