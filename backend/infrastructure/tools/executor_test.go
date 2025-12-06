package tools

import (
	"testing"
)

func TestToolHandler_Registration(t *testing.T) {
	e := &Executor{tools: make(map[string]ToolHandler)}

	// Register file tools
	e.registerFileTool()
	e.registerGitTools()

	expectedTools := []string{
		"search_files", "read_file", "list_directory", "search_content",
		"git_status", "git_diff", "git_log", "git_blame", "git_show",
		"git_diff_branches", "git_search_commits", "git_changed_files",
		"git_file_history", "git_co_changed", "git_suggest_context",
	}

	for _, name := range expectedTools {
		if _, ok := e.tools[name]; !ok {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestGetAvailableTools_Schema(t *testing.T) {
	e := &Executor{}
	tools := e.GetAvailableTools()

	if len(tools) == 0 {
		t.Fatal("no tools returned")
	}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			if tool.Name == "" {
				t.Error("tool name is empty")
			}
			if tool.Description == "" {
				t.Error("tool description is empty")
			}
			if tool.Parameters.Type != "object" {
				t.Errorf("expected parameters type 'object', got %q", tool.Parameters.Type)
			}
		})
	}
}

func TestGetAvailableTools_EssentialTools(t *testing.T) {
	e := &Executor{}
	tools := e.GetAvailableTools()

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	essentialTools := []string{
		// File tools
		"search_files", "read_file", "list_directory", "search_content",
		// Analysis tools
		"list_symbols", "search_symbols", "find_definition", "get_imports", "find_references",
		// Git tools
		"git_status", "git_diff", "git_log", "git_blame", "git_show",
		// Git-Aware Context (Phase 5)
		"git_diff_branches", "git_search_commits", "git_changed_files",
		"git_file_history", "git_co_changed", "git_suggest_context",
		// Call graph tools
		"get_callers", "get_callees", "get_impact",
		// Memory tools (Phase 7)
		"save_context", "find_context", "get_preferences",
	}

	for _, name := range essentialTools {
		if !toolNames[name] {
			t.Errorf("essential tool %q not found", name)
		}
	}
}

func TestGetAvailableTools_RequiredParams(t *testing.T) {
	e := &Executor{}
	tools := e.GetAvailableTools()

	// Tools that should have required params
	toolsWithRequired := map[string][]string{
		"read_file":        {"path"},
		"search_files":     {"pattern"},
		"search_content":   {"pattern"},
		"list_symbols":     {"path"},
		"search_symbols":   {"query"},
		"find_definition":  {"name"},
		"find_references":  {"name"},
		"get_imports":      {"path"},
		"git_blame":        {"path"},
		"git_show":         {"path"},
		"git_file_history": {"path"},
		"git_co_changed":   {"path"},
		"get_callers":      {"function"},
		"get_callees":      {"function"},
		"get_impact":       {"function"},
		"save_context":     {"topic"},
		"find_context":     {"topic"},
	}

	toolMap := make(map[string]Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	for name, expectedRequired := range toolsWithRequired {
		t.Run(name, func(t *testing.T) {
			tool, ok := toolMap[name]
			if !ok {
				t.Fatalf("tool %q not found", name)
			}

			for _, req := range expectedRequired {
				found := false
				for _, r := range tool.Parameters.Required {
					if r == req {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("tool %q should require param %q", name, req)
				}
			}
		})
	}
}

func TestExecute_WithoutHandler(t *testing.T) {
	e := &Executor{
		tools:  make(map[string]ToolHandler),
		logger: &noopLogger{},
	}

	result := e.Execute(ToolCall{
		ID:   "test-1",
		Name: "unknown_tool",
	}, t.TempDir())

	if result.Error == "" {
		t.Error("expected error for unknown tool")
	}
	if result.ToolCallID != "test-1" {
		t.Errorf("expected ToolCallID 'test-1', got %q", result.ToolCallID)
	}
}

func TestExecute_WithHandler(t *testing.T) {
	e := &Executor{
		tools:  make(map[string]ToolHandler),
		logger: &noopLogger{},
	}

	// Register a simple handler
	e.tools["test_tool"] = func(args map[string]any, projectRoot string) (string, error) {
		return "success", nil
	}

	result := e.Execute(ToolCall{
		ID:   "test-1",
		Name: "test_tool",
	}, t.TempDir())

	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Content != "success" {
		t.Errorf("expected content 'success', got %q", result.Content)
	}
}

// noopLogger for testing
type noopLogger struct{}

func (l *noopLogger) Debug(msg string)                                       {}
func (l *noopLogger) Info(msg string)                                        {}
func (l *noopLogger) Warning(msg string)                                     {}
func (l *noopLogger) Error(msg string)                                       {}
func (l *noopLogger) Fatal(msg string)                                       {}
func (l *noopLogger) Debugf(format string, args ...interface{})              {}
func (l *noopLogger) Infof(format string, args ...interface{})               {}
func (l *noopLogger) Warningf(format string, args ...interface{})            {}
func (l *noopLogger) Errorf(format string, args ...interface{})              {}
func (l *noopLogger) Fatalf(format string, args ...interface{})              {}
func (l *noopLogger) WithField(key string, value interface{}) interface{}    { return l }
func (l *noopLogger) WithFields(fields map[string]interface{}) interface{}   { return l }
