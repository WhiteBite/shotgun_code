package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupGitRepo creates a temporary git repository for testing
func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")

	// Create initial commit
	createFile(t, tmpDir, "README.md", "# Test Project")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	return tmpDir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func TestGitStatus(t *testing.T) {
	tmpDir := setupGitRepo(t)

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	// Clean status
	result, err := e.gitStatus(nil, tmpDir)
	if err != nil {
		t.Fatalf("gitStatus failed: %v", err)
	}
	if !strings.Contains(result, "clean") {
		t.Errorf("expected clean status, got: %s", result)
	}

	// Modified file
	createFile(t, tmpDir, "new_file.go", "package main")
	result, err = e.gitStatus(nil, tmpDir)
	if err != nil {
		t.Fatalf("gitStatus failed: %v", err)
	}
	if !strings.Contains(result, "new_file.go") {
		t.Errorf("expected new_file.go in status, got: %s", result)
	}
}

func TestGitDiff(t *testing.T) {
	tmpDir := setupGitRepo(t)

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	// No changes
	result, err := e.gitDiff(map[string]any{}, tmpDir)
	if err != nil {
		t.Fatalf("gitDiff failed: %v", err)
	}
	if !strings.Contains(result, "No changes") {
		t.Errorf("expected no changes, got: %s", result)
	}

	// With changes
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Modified"), 0o644)
	result, err = e.gitDiff(map[string]any{}, tmpDir)
	if err != nil {
		t.Fatalf("gitDiff failed: %v", err)
	}
	if !strings.Contains(result, "Modified") || !strings.Contains(result, "README") {
		t.Errorf("expected diff content, got: %s", result)
	}
}

func TestGitLog(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Add more commits
	createFile(t, tmpDir, "file1.go", "package main")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Add file1")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitLog(map[string]any{"limit": float64(5)}, tmpDir)
	if err != nil {
		t.Fatalf("gitLog failed: %v", err)
	}
	if !strings.Contains(result, "Add file1") {
		t.Errorf("expected commit message, got: %s", result)
	}
}

func TestGitBlame(t *testing.T) {
	tmpDir := setupGitRepo(t)

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	// Test with existing file
	result, err := e.gitBlame(map[string]any{"path": "README.md"}, tmpDir)
	if err != nil {
		t.Fatalf("gitBlame failed: %v", err)
	}
	if !strings.Contains(result, "Test User") || !strings.Contains(result, "README") {
		t.Errorf("expected blame info, got: %s", result)
	}

	// Test missing path
	_, err = e.gitBlame(map[string]any{}, tmpDir)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestGitShow(t *testing.T) {
	tmpDir := setupGitRepo(t)

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitShow(map[string]any{"path": "README.md"}, tmpDir)
	if err != nil {
		t.Fatalf("gitShow failed: %v", err)
	}
	if !strings.Contains(result, "Test Project") {
		t.Errorf("expected file content, got: %s", result)
	}

	// Test missing path
	_, err = e.gitShow(map[string]any{}, tmpDir)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

// Phase 5 Git-Aware Context Tests

func TestGitDiffBranches(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create a branch with changes
	runGit(t, tmpDir, "checkout", "-b", "feature")
	createFile(t, tmpDir, "feature.go", "package feature")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Add feature")
	runGit(t, tmpDir, "checkout", "master")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitDiffBranches(map[string]any{
		"base":    "master",
		"compare": "feature",
	}, tmpDir)
	if err != nil {
		t.Fatalf("gitDiffBranches failed: %v", err)
	}
	if !strings.Contains(result, "feature.go") {
		t.Errorf("expected feature.go in diff, got: %s", result)
	}
}

func TestGitSearchCommits(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Add commits with specific messages
	createFile(t, tmpDir, "auth.go", "package auth")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "feat: add authentication")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitSearchCommits(map[string]any{"query": "authentication"}, tmpDir)
	if err != nil {
		t.Fatalf("gitSearchCommits failed: %v", err)
	}
	if !strings.Contains(result, "authentication") {
		t.Errorf("expected commit with authentication, got: %s", result)
	}

	// Test missing params
	_, err = e.gitSearchCommits(map[string]any{}, tmpDir)
	if err == nil {
		t.Error("expected error for missing params")
	}
}

func TestGitChangedFiles(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Add multiple commits changing same file
	for i := 0; i < 3; i++ {
		os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Version "+string(rune('0'+i))), 0o644)
		runGit(t, tmpDir, "add", ".")
		runGit(t, tmpDir, "commit", "-m", "Update README")
	}

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitChangedFiles(map[string]any{"since": "1 year ago"}, tmpDir)
	if err != nil {
		t.Fatalf("gitChangedFiles failed: %v", err)
	}
	if !strings.Contains(result, "README.md") {
		t.Errorf("expected README.md in changed files, got: %s", result)
	}
}

func TestGitFileHistory(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Multiple commits on same file
	for i := 0; i < 3; i++ {
		os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Version "+string(rune('0'+i))), 0o644)
		runGit(t, tmpDir, "add", ".")
		runGit(t, tmpDir, "commit", "-m", "Update v"+string(rune('0'+i)))
	}

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitFileHistory(map[string]any{"path": "README.md"}, tmpDir)
	if err != nil {
		t.Fatalf("gitFileHistory failed: %v", err)
	}
	if !strings.Contains(result, "Update") {
		t.Errorf("expected commit history, got: %s", result)
	}

	// Test missing path
	_, err = e.gitFileHistory(map[string]any{}, tmpDir)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestGitCoChanged(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create files that change together
	createFile(t, tmpDir, "service.go", "package main")
	createFile(t, tmpDir, "service_test.go", "package main")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Add service")

	// Change them together again
	os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte("package main\n// v2"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "service_test.go"), []byte("package main\n// v2"), 0o644)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Update service")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitCoChanged(map[string]any{"path": "service.go"}, tmpDir)
	if err != nil {
		t.Fatalf("gitCoChanged failed: %v", err)
	}
	if !strings.Contains(result, "service_test.go") {
		t.Errorf("expected service_test.go as co-changed, got: %s", result)
	}

	// Test missing path
	_, err = e.gitCoChanged(map[string]any{}, tmpDir)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestGitSuggestContext(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create related files
	createFile(t, tmpDir, "auth.go", "package auth")
	createFile(t, tmpDir, "auth_test.go", "package auth")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "feat: add auth module")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerGitTools()

	result, err := e.gitSuggestContext(map[string]any{
		"task":          "fix auth bug",
		"current_files": []any{"auth.go"},
	}, tmpDir)
	if err != nil {
		t.Fatalf("gitSuggestContext failed: %v", err)
	}
	// May or may not find suggestions depending on history
	t.Logf("Suggestions: %s", result)
}
