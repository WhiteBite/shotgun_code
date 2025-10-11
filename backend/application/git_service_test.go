package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDiff_NoChanges(t *testing.T) {
	// Setup: Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Test
	service := NewGitService()
	diff, err := service.GenerateDiff(tmpDir)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, diff, "Diff should be empty when there are no changes")
}

func TestGenerateDiff_WithChanges(t *testing.T) {
	// Setup: Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Modify the file
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	require.NoError(t, err)

	// Test
	service := NewGitService()
	diff, err := service.GenerateDiff(tmpDir)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, diff, "Diff should not be empty when there are changes")
	assert.Contains(t, diff, "test.txt", "Diff should mention the modified file")
	assert.Contains(t, diff, "Changes in repository", "Diff should contain header")
}

func TestIsGitRepository(t *testing.T) {
	// Setup: Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	service := NewGitService()

	// Test non-git directory
	assert.False(t, service.IsGitRepository(tmpDir), "Should return false for non-git directory")

	// Initialize repository
	_, err = git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Test git directory
	assert.True(t, service.IsGitRepository(tmpDir), "Should return true for git directory")
}

func TestGetCurrentBranch(t *testing.T) {
	// Setup: Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create initial commit (required for branch)
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Test
	service := NewGitService()
	branch, err := service.GetCurrentBranch(tmpDir)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, branch, "Branch name should not be empty")
	// Default branch is usually "master" or "main"
	assert.True(t, branch == "master" || branch == "main", "Branch should be master or main")
}

func TestGetStatus(t *testing.T) {
	// Setup: Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize repository
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create untracked file
	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	err = os.WriteFile(untrackedFile, []byte("untracked"), 0644)
	require.NoError(t, err)

	// Test
	service := NewGitService()
	status, err := service.GetStatus(tmpDir)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Contains(t, status, "untracked.txt", "Status should contain untracked file")
}
