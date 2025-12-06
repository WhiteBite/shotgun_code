package tools

import (
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"shotgun_code/infrastructure/analyzers"
)

// Executor executes tools
type Executor struct {
	logger       domain.Logger
	fileReader   domain.FileContentReader
	registry     analysis.AnalyzerRegistry
	symbolIndex  analysis.SymbolIndex
	refFinder    *analyzers.ReferenceFinder
	callGraph    *analyzers.CallGraphBuilderImpl
	tools        map[string]ToolHandler
}

// ToolHandler is a function that handles a tool call
type ToolHandler func(args map[string]any, projectRoot string) (string, error)

// NewExecutor creates a new tool executor
func NewExecutor(
	logger domain.Logger,
	fileReader domain.FileContentReader,
	registry analysis.AnalyzerRegistry,
	symbolIndex analysis.SymbolIndex,
) *Executor {
	e := &Executor{
		logger:       logger,
		fileReader:   fileReader,
		registry:     registry,
		symbolIndex:  symbolIndex,
		refFinder:    analyzers.NewReferenceFinder(registry),
		callGraph:    analyzers.NewCallGraphBuilder(registry),
		tools:        make(map[string]ToolHandler),
	}

	// Register file tools
	e.registerFileTool()
	// Register analysis tools
	e.registerAnalysisTools()
	// Register call graph tools
	e.registerCallGraphTools()
	// Register git tools
	e.registerGitTools()
	// Register memory tools (Phase 7)
	e.registerMemoryTools()

	return e
}

// GetAvailableTools returns all available tools
func (e *Executor) GetAvailableTools() []Tool {
	return []Tool{
		// File tools
		{
			Name:        "search_files",
			Description: "Search for files by name pattern (glob)",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"pattern":   {Type: "string", Description: "Glob pattern or filename"},
					"directory": {Type: "string", Description: "Directory to search in"},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read file contents",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":       {Type: "string", Description: "File path"},
					"start_line": {Type: "integer", Description: "Start line"},
					"end_line":   {Type: "integer", Description: "End line"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "list_directory",
			Description: "List directory contents",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":      {Type: "string", Description: "Directory path"},
					"recursive": {Type: "boolean", Description: "Recursive listing"},
				},
			},
		},
		{
			Name:        "search_content",
			Description: "Search for text in files",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"pattern":      {Type: "string", Description: "Search pattern"},
					"file_pattern": {Type: "string", Description: "File filter"},
				},
				Required: []string{"pattern"},
			},
		},
		// Analysis tools
		{
			Name:        "list_symbols",
			Description: "List symbols in a file (classes, functions, etc.)",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
					"kind": {Type: "string", Description: "Filter by kind"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "search_symbols",
			Description: "Search symbols across project",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"query": {Type: "string", Description: "Symbol name to search"},
					"kind":  {Type: "string", Description: "Filter by kind"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "find_definition",
			Description: "Find where a symbol is defined",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"name": {Type: "string", Description: "Symbol name"},
					"kind": {Type: "string", Description: "Symbol kind"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_imports",
			Description: "Get imports of a file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "find_references",
			Description: "Find all references to a symbol across the project",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"name": {Type: "string", Description: "Symbol name to find references for"},
					"kind": {Type: "string", Description: "Symbol kind (optional)"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_function",
			Description: "Get the full source code of a function/method by name",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
					"name": {Type: "string", Description: "Function/method name"},
				},
				Required: []string{"path", "name"},
			},
		},
		{
			Name:        "get_exports",
			Description: "Get exported symbols from a module/file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
				},
				Required: []string{"path"},
			},
		},
		// Call graph tools
		{
			Name:        "get_callers",
			Description: "Get functions that call a given function",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"function": {Type: "string", Description: "Function name"},
				},
				Required: []string{"function"},
			},
		},
		{
			Name:        "get_callees",
			Description: "Get functions called by a given function",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"function": {Type: "string", Description: "Function name"},
				},
				Required: []string{"function"},
			},
		},
		{
			Name:        "get_impact",
			Description: "Get all functions affected if a function changes",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"function":  {Type: "string", Description: "Function name"},
					"max_depth": {Type: "integer", Description: "Max depth to traverse"},
				},
				Required: []string{"function"},
			},
		},
		{
			Name:        "get_call_chain",
			Description: "Find call chain between two functions",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"from":      {Type: "string", Description: "Source function name"},
					"to":        {Type: "string", Description: "Target function name"},
					"max_depth": {Type: "integer", Description: "Max depth to search"},
				},
				Required: []string{"from", "to"},
			},
		},
		{
			Name:        "get_file_dependencies",
			Description: "Get files that a file imports/depends on",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "get_file_dependents",
			Description: "Get files that import/depend on a file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path": {Type: "string", Description: "File path"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "find_cyclic_dependencies",
			Description: "Find cyclic dependencies in the project",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
			},
		},
		{
			Name:        "export_call_graph",
			Description: "Export call graph or dependency graph as Mermaid diagram",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"type":      {Type: "string", Description: "Graph type: 'call' or 'dependency'"},
					"format":    {Type: "string", Description: "Output format: 'mermaid'"},
					"max_nodes": {Type: "integer", Description: "Max nodes to include"},
				},
			},
		},
		// Git tools
		{
			Name:        "git_status",
			Description: "Get git status",
			Parameters:  ToolParameters{Type: "object", Properties: map[string]ToolProperty{}},
		},
		{
			Name:        "git_diff",
			Description: "Get git diff",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":   {Type: "string", Description: "File path"},
					"staged": {Type: "boolean", Description: "Staged only"},
				},
			},
		},
		{
			Name:        "git_log",
			Description: "Get git log",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"limit": {Type: "integer", Description: "Number of commits"},
					"path":  {Type: "string", Description: "Filter by path"},
				},
			},
		},
		{
			Name:        "git_blame",
			Description: "Show who changed each line of a file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":       {Type: "string", Description: "File path"},
					"start_line": {Type: "integer", Description: "Start line number"},
					"end_line":   {Type: "integer", Description: "End line number"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_show",
			Description: "Show file content at a specific commit",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":   {Type: "string", Description: "File path"},
					"commit": {Type: "string", Description: "Commit hash or ref (default: HEAD)"},
				},
				Required: []string{"path"},
			},
		},
		// Git-Aware Context (Phase 5)
		{
			Name:        "git_diff_branches",
			Description: "Show differences between two branches",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"base":    {Type: "string", Description: "Base branch (default: main)"},
					"compare": {Type: "string", Description: "Compare branch (default: HEAD)"},
					"path":    {Type: "string", Description: "Filter by path"},
				},
			},
		},
		{
			Name:        "git_search_commits",
			Description: "Search commits by message, author, or date",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"query":  {Type: "string", Description: "Search in commit messages"},
					"author": {Type: "string", Description: "Filter by author"},
					"since":  {Type: "string", Description: "Since date (e.g., '1 week ago', '2024-01-01')"},
					"path":   {Type: "string", Description: "Filter by file path"},
					"limit":  {Type: "integer", Description: "Max results (default: 20)"},
				},
			},
		},
		{
			Name:        "git_changed_files",
			Description: "Get files changed in a time period, sorted by change frequency",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"since":  {Type: "string", Description: "Since when (default: '1 week ago')"},
					"author": {Type: "string", Description: "Filter by author"},
					"path":   {Type: "string", Description: "Filter by path pattern"},
				},
			},
		},
		{
			Name:        "git_file_history",
			Description: "Show detailed commit history for a specific file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":  {Type: "string", Description: "File path"},
					"limit": {Type: "integer", Description: "Max commits (default: 10)"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_co_changed",
			Description: "Find files that are often changed together with a given file",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"path":  {Type: "string", Description: "File path"},
					"limit": {Type: "integer", Description: "Max results (default: 10)"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_suggest_context",
			Description: "Suggest files to include in context based on git history and task description",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"task":          {Type: "string", Description: "Task description to find related commits"},
					"current_files": {Type: "array", Description: "Currently selected files"},
					"limit":         {Type: "integer", Description: "Max suggestions (default: 10)"},
				},
			},
		},
		// Memory & Preferences (Phase 7)
		{
			Name:        "save_context",
			Description: "Save current conversation context for later retrieval",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"topic":   {Type: "string", Description: "Topic/name for this context"},
					"summary": {Type: "string", Description: "Brief summary of the conversation"},
					"files":   {Type: "array", Description: "Files involved in this context"},
				},
				Required: []string{"topic"},
			},
		},
		{
			Name:        "find_context",
			Description: "Find previously saved contexts by topic",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"topic": {Type: "string", Description: "Topic to search for"},
				},
				Required: []string{"topic"},
			},
		},
		{
			Name:        "get_recent_contexts",
			Description: "Get recently accessed conversation contexts",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"limit": {Type: "integer", Description: "Max contexts to return (default: 10)"},
				},
			},
		},
		{
			Name:        "set_preference",
			Description: "Set a user preference for context building",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"key":   {Type: "string", Description: "Preference key (exclude_tests, exclude_vendor, max_context_files, etc.)"},
					"value": {Type: "string", Description: "Preference value"},
				},
				Required: []string{"key", "value"},
			},
		},
		{
			Name:        "get_preferences",
			Description: "Get all user preferences",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
			},
		},
	}
}

// Execute executes a tool call
func (e *Executor) Execute(call ToolCall, projectRoot string) ToolResult {
	e.logger.Info(fmt.Sprintf("Executing tool: %s", call.Name))

	handler, ok := e.tools[call.Name]
	if !ok {
		return ToolResult{
			ToolCallID: call.ID,
			Error:      fmt.Sprintf("Unknown tool: %s", call.Name),
		}
	}

	content, err := handler(call.Arguments, projectRoot)
	if err != nil {
		return ToolResult{
			ToolCallID: call.ID,
			Error:      err.Error(),
		}
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    content,
	}
}
