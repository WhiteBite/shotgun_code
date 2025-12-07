package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	createFile(t, tmpDir, "main.go", "package main")
	createFile(t, tmpDir, "user_service.go", "package service")
	createFile(t, tmpDir, "user_test.go", "package service")
	createDir(t, tmpDir, "internal")
	createFile(t, tmpDir, "internal/handler.go", "package internal")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerFileTool()

	tests := []struct {
		name        string
		args        map[string]any
		wantContain string
		wantCount   int
	}{
		{
			name:        "search by extension",
			args:        map[string]any{"pattern": "*.go"},
			wantContain: "main.go",
			wantCount:   4,
		},
		{
			name:        "search by partial name",
			args:        map[string]any{"pattern": "user"},
			wantContain: "user_service.go",
			wantCount:   2,
		},
		{
			name:        "search in subdirectory",
			args:        map[string]any{"pattern": "*.go", "directory": "internal"},
			wantContain: "handler.go",
			wantCount:   1,
		},
		{
			name:        "no matches",
			args:        map[string]any{"pattern": "*.py"},
			wantContain: "No files found",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.searchFiles(tt.args, tmpDir)
			if err != nil {
				t.Fatalf("searchFiles failed: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("result should contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestSearchFiles_EmptyPattern(t *testing.T) {
	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerFileTool()

	_, err := e.searchFiles(map[string]any{}, t.TempDir())
	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	content := "line1\nline2\nline3\nline4\nline5"
	createFile(t, tmpDir, "test.txt", content)

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerFileTool()

	tests := []struct {
		name        string
		args        map[string]any
		wantContain string
		wantErr     bool
	}{
		{
			name:        "read entire file",
			args:        map[string]any{"path": "test.txt"},
			wantContain: "line3",
		},
		{
			name:        "read line range",
			args:        map[string]any{"path": "test.txt", "start_line": float64(2), "end_line": float64(4)},
			wantContain: "line2",
		},
		{
			name:    "file not found",
			args:    map[string]any{"path": "nonexistent.txt"},
			wantErr: true,
		},
		{
			name:    "empty path",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.readFile(tt.args, tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("readFile failed: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("result should contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	createFile(t, tmpDir, "file1.go", "")
	createFile(t, tmpDir, "file2.go", "")
	createDir(t, tmpDir, "subdir")
	createFile(t, tmpDir, "subdir/nested.go", "")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerFileTool()

	tests := []struct {
		name        string
		args        map[string]any
		wantContain string
	}{
		{
			name:        "list root",
			args:        map[string]any{},
			wantContain: "file1.go",
		},
		{
			name:        "list subdirectory",
			args:        map[string]any{"path": "subdir"},
			wantContain: "nested.go",
		},
		{
			name:        "recursive listing",
			args:        map[string]any{"recursive": true},
			wantContain: "subdir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.listDirectory(tt.args, tmpDir)
			if err != nil {
				t.Fatalf("listDirectory failed: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("result should contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

func TestSearchContent(t *testing.T) {
	tmpDir := t.TempDir()
	createFile(t, tmpDir, "main.go", "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}")
	createFile(t, tmpDir, "service.go", "package service\n\nfunc Hello() string {\n\treturn \"hello\"\n}")

	e := &Executor{tools: make(map[string]ToolHandler)}
	e.registerFileTool()

	tests := []struct {
		name        string
		args        map[string]any
		wantContain string
		wantErr     bool
	}{
		{
			name:        "search text",
			args:        map[string]any{"pattern": "hello"},
			wantContain: "hello",
		},
		{
			name:        "search with file filter",
			args:        map[string]any{"pattern": "func", "file_pattern": "main.go"},
			wantContain: "main.go",
		},
		{
			name:        "no matches",
			args:        map[string]any{"pattern": "nonexistent_pattern_xyz"},
			wantContain: "No matches found",
		},
		{
			name:    "empty pattern",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.searchContent(tt.args, tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("searchContent failed: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("result should contain %q, got: %s", tt.wantContain, result)
			}
		})
	}
}

// Helper functions
func createDir(t *testing.T, base, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(base, path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
}

func createFile(t *testing.T, base, path, content string) {
	t.Helper()
	fullPath := filepath.Join(base, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
}
