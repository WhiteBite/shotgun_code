package application

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"sync"
	"time"
)

// ContextStreamerImpl implements the ContextStreamer interface
type ContextStreamerImpl struct {
	fileReader domain.FileContentReader
	tokenCounter domain.TokenCounter
	logger     domain.Logger
	contextDir string
    commentStripper domain.CommentStripper
	// Streaming context support
	streams   map[string]*domain.ContextStream
	streamPaths map[string]string  // Map context ID to file path
	streamsMu sync.RWMutex
}

// NewContextStreamer creates a new ContextStreamer implementation
func NewContextStreamer(
	fileReader domain.FileContentReader,
	tokenCounter domain.TokenCounter,
	logger domain.Logger,
	contextDir string,
    commentStripper domain.CommentStripper,
) *ContextStreamerImpl {
	return &ContextStreamerImpl{
		fileReader:   fileReader,
		tokenCounter: tokenCounter,
		logger:       logger,
		contextDir:   contextDir,
        commentStripper: commentStripper,
		streams:      make(map[string]*domain.ContextStream),
		streamPaths:  make(map[string]string), // Initialize the stream paths map
	}
}

// CreateStreamingContext creates a streaming context from project files
func (cs *ContextStreamerImpl) CreateStreamingContext(ctx context.Context, projectPath string, includedPaths []string, options *domain.ContextBuildOptions) (*domain.ContextStream, error) {
	if options == nil {
		options = &domain.ContextBuildOptions{}
	}
	
	cs.logger.Info(fmt.Sprintf("Creating streaming context for project: %s, files: %d", projectPath, len(includedPaths)))
	
	// Read file contents
	contents, err := cs.fileReader.ReadContents(ctx, includedPaths, projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}
	
	// Create context content and save to file
	contextID := fmt.Sprintf("stream_%d", len(cs.streams)+1)
	contextPath := filepath.Join(cs.contextDir, contextID+".ctx")
	
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
    var totalTokens int
	
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
        if options.StripComments && cs.commentStripper != nil {
            content = cs.commentStripper.Strip(content, filePath)
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
        totalTokens += cs.tokenCounter(content)
		
		writer.WriteString("\n```\n\n")
		totalLines += 3
		totalChars += 6
	}
	
	// Create context stream object
	stream := &domain.ContextStream{
		ID:          contextID,
		Name:        cs.generateContextName(projectPath, actualFiles),
		Description: fmt.Sprintf("Streaming context with %d files from %s", len(actualFiles), filepath.Base(projectPath)),
		Files:       actualFiles,
		ProjectPath: projectPath,
		TotalLines:  totalLines,
		TotalChars:  totalChars,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
        TokenCount:  totalTokens,
	}
	
	// Store stream reference and path
	cs.streamsMu.Lock()
	cs.streams[contextID] = stream
	cs.streamPaths[contextID] = contextPath
	cs.streamsMu.Unlock()
	
	cs.logger.Info(fmt.Sprintf("Created streaming context %s with %d lines", contextID, totalLines))
	return stream, nil
}

// GetContextLines retrieves a range of lines from a streaming context
func (cs *ContextStreamerImpl) GetContextLines(ctx context.Context, contextID string, startLine, endLine int64) (*domain.ContextLineRange, error) {
	cs.streamsMu.RLock()
	_, exists := cs.streams[contextID]
	path, pathExists := cs.streamPaths[contextID]
	cs.streamsMu.RUnlock()
	
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
func (cs *ContextStreamerImpl) GetContextContent(ctx context.Context, contextID string, startLine int, lineCount int) (interface{}, error) {
	// Try to get from streaming context first
	lines, err := cs.GetContextLines(ctx, contextID, int64(startLine), int64(startLine+lineCount-1))
	if err == nil {
		return lines, nil
	}
	
	return nil, fmt.Errorf("context not found: %w", err)
}

// generateContextName generates a descriptive name for the context
func (cs *ContextStreamerImpl) generateContextName(projectPath string, files []string) string {
	projectName := filepath.Base(projectPath)
	
	if len(files) == 1 {
		fileName := filepath.Base(files[0])
		return fmt.Sprintf("%s - %s", projectName, fileName)
	}
	
	return fmt.Sprintf("%s - %d files", projectName, len(files))
}

// Comment stripping helpers removed; delegated to domain.CommentStripper