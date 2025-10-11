package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContext_Success(t *testing.T) {
	// Setup: Create temporary files
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	err = os.WriteFile(file1, []byte("line 1\nline 2\nline 3"), 0644)
	require.NoError(t, err)

	file2 := filepath.Join(tmpDir, "file2.txt")
	err = os.WriteFile(file2, []byte("line 4\nline 5"), 0644)
	require.NoError(t, err)

	// Test
	service := NewContextService()
	summary, err := service.BuildContext([]string{file1, file2})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.NotEmpty(t, summary.ID, "Context ID should be generated")
	assert.Equal(t, 2, summary.TotalFiles, "Should count 2 files")
	assert.Equal(t, 5, summary.TotalLines, "Should count 5 lines")
	assert.Greater(t, summary.TotalSize, 0, "Total size should be positive")
	assert.Greater(t, summary.TotalChunks, 0, "Should have at least one chunk")

	// Verify context is stored
	assert.Equal(t, 1, service.GetContextCount(), "Should have one context stored")
}

func TestBuildContext_TooLarge(t *testing.T) {
	// Setup: Create a large temporary file
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a file larger than MaxContextSize
	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("x", MaxContextSize+1)
	err = os.WriteFile(largeFile, []byte(largeContent), 0644)
	require.NoError(t, err)

	// Test
	service := NewContextService()
	summary, err := service.BuildContext([]string{largeFile})

	// Assert
	assert.Error(t, err, "Should return error for too large context")
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "exceeds limit", "Error should mention size limit")
}

func TestBuildContext_NoFiles(t *testing.T) {
	service := NewContextService()
	summary, err := service.BuildContext([]string{})

	assert.Error(t, err, "Should return error when no files provided")
	assert.Nil(t, summary)
}

func TestBuildContext_FileNotFound(t *testing.T) {
	service := NewContextService()
	summary, err := service.BuildContext([]string{"/nonexistent/file.txt"})

	assert.Error(t, err, "Should return error for nonexistent file")
	assert.Nil(t, summary)
}

func TestGetLines_ValidRange(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	content := "line 1\nline 2\nline 3\nline 4\nline 5"
	err = os.WriteFile(file1, []byte(content), 0644)
	require.NoError(t, err)

	service := NewContextService()
	summary, err := service.BuildContext([]string{file1})
	require.NoError(t, err)

	// Test: Get lines 1-3
	result, err := service.GetLines(summary.ID, 1, 3)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Should contain lines 1, 2, 3 from original content plus header
	assert.Contains(t, result, "line")
}

func TestGetLines_InvalidRange(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	err = os.WriteFile(file1, []byte("line 1\nline 2"), 0644)
	require.NoError(t, err)

	service := NewContextService()
	summary, err := service.BuildContext([]string{file1})
	require.NoError(t, err)

	// Test: Invalid start index
	result, err := service.GetLines(summary.ID, 1000, 10)

	// Assert
	assert.Error(t, err, "Should return error for invalid range")
	assert.Empty(t, result)
}

func TestGetLines_ContextNotFound(t *testing.T) {
	service := NewContextService()

	result, err := service.GetLines("nonexistent-id", 0, 10)

	assert.Error(t, err, "Should return error for nonexistent context")
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetFullContext(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	err = os.WriteFile(file1, []byte("test content"), 0644)
	require.NoError(t, err)

	service := NewContextService()
	summary, err := service.BuildContext([]string{file1})
	require.NoError(t, err)

	// Test
	content, err := service.GetFullContext(summary.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "test content")
	assert.Contains(t, content, "=== File:")
}

func TestDeleteContext(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	err = os.WriteFile(file1, []byte("test"), 0644)
	require.NoError(t, err)

	service := NewContextService()
	summary, err := service.BuildContext([]string{file1})
	require.NoError(t, err)

	assert.Equal(t, 1, service.GetContextCount())

	// Test
	err = service.DeleteContext(summary.ID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, service.GetContextCount())

	// Try to get deleted context
	_, err = service.GetFullContext(summary.ID)
	assert.Error(t, err)
}

func TestDeleteContext_NotFound(t *testing.T) {
	service := NewContextService()

	err := service.DeleteContext("nonexistent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConcurrentAccess(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	err = os.WriteFile(file1, []byte("concurrent test"), 0644)
	require.NoError(t, err)

	service := NewContextService()
	summary, err := service.BuildContext([]string{file1})
	require.NoError(t, err)

	// Test: Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.GetFullContext(summary.ID)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
