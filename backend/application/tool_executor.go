package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"shotgun_code/infrastructure/analyzers"
	"shotgun_code/infrastructure/git"
	"strings"
)

// ToolExecutorImpl implements the ToolExecutor interface
type ToolExecutorImpl struct {
	logger                  domain.Logger
	fileReader              domain.FileContentReader
	registry                analysis.AnalyzerRegistry
	symbolIndex             analysis.SymbolIndex
	callGraph               *analyzers.CallGraphBuilderImpl
	gitContext              *git.ContextBuilder
	contextMemory           domain.ContextMemory
	hasSemanticSearch       bool // flag indicating if semantic search is available
	semanticSearcherService SemanticSearcher
}

// SemanticSearcher interface for semantic search service
type SemanticSearcher interface {
	Search(query string, limit int) ([]SemanticSearchResult, error)
}

// SemanticSearchResult represents a semantic search result
type SemanticSearchResult struct {
	FilePath string
	Score    float64
	Snippet  string
	Line     int
}

// NewToolExecutor creates a new ToolExecutor
func NewToolExecutor(logger domain.Logger, fileReader domain.FileContentReader) *ToolExecutorImpl {
	registry := analyzers.NewAnalyzerRegistry()
	return &ToolExecutorImpl{
		logger:      logger,
		fileReader:  fileReader,
		registry:    registry,
		symbolIndex: analyzers.NewSymbolIndex(registry),
		callGraph:   analyzers.NewCallGraphBuilder(registry),
	}
}

// SetGitContext sets the git context builder for git-related tools
func (te *ToolExecutorImpl) SetGitContext(projectRoot string) {
	te.gitContext = git.NewContextBuilder(projectRoot)
}

// SetContextMemory sets the context memory for memory-related tools
func (te *ToolExecutorImpl) SetContextMemory(cm domain.ContextMemory) {
	te.contextMemory = cm
}

// SetSemanticSearch sets the semantic search service
func (te *ToolExecutorImpl) SetSemanticSearch(ss SemanticSearcher) {
	te.semanticSearcherService = ss
	te.hasSemanticSearch = ss != nil
}

// GetAvailableTools returns all available tools
func (te *ToolExecutorImpl) GetAvailableTools() []domain.Tool {
	return []domain.Tool{
		{
			Name:        "search_files",
			Description: "Search for files by name pattern (glob). Returns list of matching file paths.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"pattern": {
						Type:        "string",
						Description: "Glob pattern or partial filename to search for (e.g., '*.go', 'auth', 'user_service')",
					},
					"directory": {
						Type:        "string",
						Description: "Directory to search in (relative to project root). Empty for entire project.",
					},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "search_content",
			Description: "Search for text/regex pattern in file contents. Returns matching lines with context.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"pattern": {
						Type:        "string",
						Description: "Text or regex pattern to search for in file contents",
					},
					"file_pattern": {
						Type:        "string",
						Description: "Glob pattern to filter files (e.g., '*.go', '*.ts'). Empty for all files.",
					},
					"max_results": {
						Type:        "integer",
						Description: "Maximum number of results to return",
						Default:     20,
					},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file. Use this to examine code in detail.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file (relative to project root)",
					},
					"start_line": {
						Type:        "integer",
						Description: "Starting line number (1-based). Omit to read from beginning.",
					},
					"end_line": {
						Type:        "integer",
						Description: "Ending line number (inclusive). Omit to read to end.",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "list_directory",
			Description: "List files and directories in a path. Use to explore project structure.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Directory path (relative to project root). Empty for root.",
					},
					"recursive": {
						Type:        "boolean",
						Description: "Whether to list recursively",
						Default:     false,
					},
					"max_depth": {
						Type:        "integer",
						Description: "Maximum depth for recursive listing",
						Default:     2,
					},
				},
			},
		},
		{
			Name:        "get_file_info",
			Description: "Get metadata about a file (size, type, last modified).",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file (relative to project root)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "list_functions",
			Description: "List all functions/methods in a file. Works for Go, TypeScript, JavaScript.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the source file",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "git_status",
			Description: "Get git status - list of modified, added, deleted files.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "git_diff",
			Description: "Get git diff for a file or all changes.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to file (optional, empty for all changes)",
					},
					"staged": {
						Type:        "boolean",
						Description: "Show staged changes only",
						Default:     false,
					},
				},
			},
		},
		{
			Name:        "git_log",
			Description: "Get recent git commits.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"limit": {
						Type:        "integer",
						Description: "Number of commits to show",
						Default:     10,
					},
					"path": {
						Type:        "string",
						Description: "Filter by file path (optional)",
					},
				},
			},
		},
		{
			Name:        "list_symbols",
			Description: "List all symbols (classes, functions, interfaces, types) in a file. Supports Go, Java, Kotlin, Dart, TypeScript, JavaScript, Vue.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the source file",
					},
					"kind": {
						Type:        "string",
						Description: "Filter by symbol kind: class, function, interface, type, method, enum (optional)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "get_imports",
			Description: "Get all imports/dependencies of a file. Supports Go, Java, Kotlin, Dart, TypeScript, JavaScript, Vue.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the source file",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "search_symbols",
			Description: "Search for symbols (classes, functions, types) across the entire project by name.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"query": {
						Type:        "string",
						Description: "Symbol name to search for (partial match supported)",
					},
					"kind": {
						Type:        "string",
						Description: "Filter by kind: class, function, interface, type, method, enum (optional)",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "find_definition",
			Description: "Find where a symbol is defined in the project.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"name": {
						Type:        "string",
						Description: "Exact symbol name to find",
					},
					"kind": {
						Type:        "string",
						Description: "Symbol kind: class, function, interface, type (optional)",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "find_references",
			Description: "Find all references to a symbol across the project. Returns locations where the symbol is used.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"name": {
						Type:        "string",
						Description: "Symbol name to find references for",
					},
					"kind": {
						Type:        "string",
						Description: "Symbol kind: class, function, interface, type (optional, helps filter results)",
					},
					"include_definition": {
						Type:        "boolean",
						Description: "Whether to include the definition in results (default: true)",
						Default:     true,
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_symbol_info",
			Description: "Get detailed information about a symbol including signature, documentation, modifiers, parent class/interface, and source code.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"name": {
						Type:        "string",
						Description: "Symbol name to get info for",
					},
					"kind": {
						Type:        "string",
						Description: "Symbol kind: class, function, interface, type, method (optional)",
					},
					"file_path": {
						Type:        "string",
						Description: "File path to narrow search (optional)",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_class_hierarchy",
			Description: "Get class inheritance hierarchy - parent classes, implemented interfaces, and subclasses. Works for Java, Kotlin, Dart, TypeScript.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"class_name": {
						Type:        "string",
						Description: "Class name to analyze",
					},
					"direction": {
						Type:        "string",
						Description: "Direction: 'up' (ancestors), 'down' (descendants), 'both' (default: both)",
					},
				},
				Required: []string{"class_name"},
			},
		},
		{
			Name:        "get_widget_tree",
			Description: "Get Flutter widget tree structure from a Dart file. Shows widget hierarchy and their properties.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"file_path": {
						Type:        "string",
						Description: "Path to Dart file containing widgets",
					},
					"widget_name": {
						Type:        "string",
						Description: "Specific widget class name to analyze (optional, analyzes all if not specified)",
					},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "get_dependent_files",
			Description: "Get all files that depend on or are depended by the specified file. Useful for understanding impact of changes.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"file_path": {
						Type:        "string",
						Description: "Path to the file",
					},
					"direction": {
						Type:        "string",
						Description: "Direction: 'imports' (files this file imports), 'importers' (files that import this), 'both' (default: both)",
					},
					"depth": {
						Type:        "integer",
						Description: "Depth of transitive dependencies (default: 1, max: 3)",
					},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "get_change_risk",
			Description: "Estimate risk score for changing a file or function based on number of dependents, test coverage, and change frequency.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"file_path": {
						Type:        "string",
						Description: "Path to the file",
					},
					"function_name": {
						Type:        "string",
						Description: "Function name (optional, analyzes whole file if not specified)",
					},
				},
				Required: []string{"file_path"},
			},
		},
		// Project Structure Tools (Phase 6)
		{
			Name:        "detect_architecture",
			Description: "Detect the architecture pattern of the project (Clean Architecture, Hexagonal, MVC, MVVM, Layered, DDD). Returns architecture type, confidence, layers, and indicators.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "detect_frameworks",
			Description: "Detect frameworks and libraries used in the project (Vue, React, Gin, Express, Spring, etc.). Returns framework names, versions, categories, and best practices.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "detect_conventions",
			Description: "Detect coding conventions in the project (naming style, folder structure, test conventions, import style, code style).",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "get_project_structure",
			Description: "Get complete project structure analysis including architecture, frameworks, conventions, languages, and build systems.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "get_related_layers",
			Description: "Get architectural layers related to a specific file. Useful for understanding dependencies between layers.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file (relative to project root)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "suggest_related_files",
			Description: "Suggest related files based on architecture patterns. Finds files in related layers, test files, and files with similar names.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"path": {
						Type:        "string",
						Description: "Path to the file (relative to project root)",
					},
				},
				Required: []string{"path"},
			},
		},
		// Call Graph Tools (Phase 3)
		{
			Name:        "get_callers",
			Description: "Get functions that call the specified function",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"function_id": {Type: "string", Description: "Function identifier (e.g., 'pkg.FunctionName')"},
				},
				Required: []string{"function_id"},
			},
		},
		{
			Name:        "get_callees",
			Description: "Get functions called by the specified function",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"function_id": {Type: "string", Description: "Function identifier (e.g., 'pkg.FunctionName')"},
				},
				Required: []string{"function_id"},
			},
		},
		{
			Name:        "get_impact",
			Description: "Get all functions affected if the specified function changes",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"function_id": {Type: "string", Description: "Function identifier"},
					"max_depth":   {Type: "integer", Description: "Maximum depth to search (default: 3)"},
				},
				Required: []string{"function_id"},
			},
		},
		{
			Name:        "get_call_chain",
			Description: "Find call chain between two functions",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"start_function": {Type: "string", Description: "Starting function identifier"},
					"end_function":   {Type: "string", Description: "Target function identifier"},
					"max_depth":      {Type: "integer", Description: "Maximum depth (default: 5)"},
				},
				Required: []string{"start_function", "end_function"},
			},
		},
		// Git Context Tools (Phase 5)
		{
			Name:        "git_changed_files",
			Description: "Get recently changed files from git history",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"since":       {Type: "string", Description: "Time period (e.g., '1 week ago', '2024-01-01')"},
					"path_filter": {Type: "string", Description: "Filter by path pattern"},
				},
			},
		},
		{
			Name:        "git_co_changed",
			Description: "Get files that are often changed together with the specified file",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"file_path": {Type: "string", Description: "Path to the file"},
					"limit":     {Type: "integer", Description: "Maximum results (default: 10)"},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "git_suggest_context",
			Description: "Suggest files to include in context based on git history and task description",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"task":          {Type: "string", Description: "Task description"},
					"current_files": {Type: "array", Description: "Currently selected files"},
					"limit":         {Type: "integer", Description: "Maximum suggestions (default: 10)"},
				},
			},
		},
		// Memory Tools (Phase 7)
		{
			Name:        "save_context",
			Description: "Save current context for later retrieval",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"topic":   {Type: "string", Description: "Topic/name for the context"},
					"summary": {Type: "string", Description: "Brief summary"},
					"files":   {Type: "array", Description: "List of file paths"},
				},
				Required: []string{"topic"},
			},
		},
		{
			Name:        "find_context",
			Description: "Find saved contexts by topic",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"topic": {Type: "string", Description: "Topic to search for"},
				},
				Required: []string{"topic"},
			},
		},
		{
			Name:        "get_recent_contexts",
			Description: "Get recently accessed contexts",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"limit": {Type: "integer", Description: "Maximum results (default: 10)"},
				},
			},
		},
		// Semantic Search Tool (Phase 4)
		{
			Name:        "semantic_search",
			Description: "Search code by meaning/intent using AI embeddings. Use for conceptual queries like 'error handling code' or 'user authentication logic'.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"query": {Type: "string", Description: "Natural language description of what you're looking for"},
					"limit": {Type: "integer", Description: "Maximum results (default: 10)"},
					"file_pattern": {Type: "string", Description: "Glob pattern to filter files (e.g., '*.go')"},
				},
				Required: []string{"query"},
			},
		},
	}
}

// toolHandler is a function type for tool execution
type toolHandler func(te *ToolExecutorImpl, args map[string]any, projectRoot string) (string, error)

// toolHandlers maps tool names to their handlers
var toolHandlers = map[string]toolHandler{
	"search_files": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.searchFiles(args, pr)
	},
	"search_content": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.searchContent(args, pr)
	},
	"read_file": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.readFile(args, pr)
	},
	"list_directory": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.listDirectory(args, pr)
	},
	"get_file_info": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getFileInfo(args, pr)
	},
	"list_functions": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.listFunctions(args, pr)
	},
	"git_status": func(te *ToolExecutorImpl, _ map[string]any, pr string) (string, error) { return te.gitStatus(pr) },
	"git_diff": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.gitDiff(args, pr)
	},
	"git_log": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.gitLog(args, pr) },
	"list_symbols": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.listSymbols(args, pr)
	},
	"get_imports": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getImports(args, pr)
	},
	"search_symbols": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.searchSymbols(args, pr)
	},
	"find_definition": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.findDefinition(args, pr)
	},
	"find_references": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.findReferences(args, pr)
	},
	"get_symbol_info": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getSymbolInfo(args, pr)
	},
	"get_class_hierarchy": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getClassHierarchy(args, pr)
	},
	"get_widget_tree": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getWidgetTree(args, pr)
	},
	"get_dependent_files": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getDependentFiles(args, pr)
	},
	"get_change_risk": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getChangeRisk(args, pr)
	},
	"detect_architecture": func(te *ToolExecutorImpl, _ map[string]any, pr string) (string, error) {
		return te.detectArchitecture(pr)
	},
	"detect_frameworks": func(te *ToolExecutorImpl, _ map[string]any, pr string) (string, error) {
		return te.detectFrameworks(pr)
	},
	"detect_conventions": func(te *ToolExecutorImpl, _ map[string]any, pr string) (string, error) {
		return te.detectConventions(pr)
	},
	"get_project_structure": func(te *ToolExecutorImpl, _ map[string]any, pr string) (string, error) {
		return te.getProjectStructure(pr)
	},
	"get_related_layers": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.getRelatedLayers(args, pr)
	},
	"suggest_related_files": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) {
		return te.suggestRelatedFiles(args, pr)
	},
	// Call Graph Tools (Phase 3)
	"get_callers":    func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.getCallers(args, pr) },
	"get_callees":    func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.getCallees(args, pr) },
	"get_impact":     func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.getImpact(args, pr) },
	"get_call_chain": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.getCallChain(args, pr) },
	// Git Context Tools (Phase 5)
	"git_changed_files":   func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.gitChangedFiles(args, pr) },
	"git_co_changed":      func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.gitCoChanged(args, pr) },
	"git_suggest_context": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.gitSuggestContext(args, pr) },
	// Memory Tools (Phase 7)
	"save_context":        func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.saveContext(args, pr) },
	"find_context":        func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.findContext(args, pr) },
	"get_recent_contexts": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.getRecentContexts(args, pr) },
	// Semantic Search Tool (Phase 4)
	"semantic_search": func(te *ToolExecutorImpl, args map[string]any, pr string) (string, error) { return te.doSemanticSearch(args, pr) },
}

// ExecuteTool executes a tool and returns the result
func (te *ToolExecutorImpl) ExecuteTool(call domain.ToolCall, projectRoot string) domain.ToolResult {
	te.logger.Info(fmt.Sprintf("Executing tool: %s with args: %v", call.Name, call.Arguments))

	var content string
	var err error

	if handler, ok := toolHandlers[call.Name]; ok {
		content, err = handler(te, call.Arguments, projectRoot)
	} else {
		err = fmt.Errorf("unknown tool: %s", call.Name)
	}

	result := domain.ToolResult{ToolCallID: call.ID, Content: content}
	if err != nil {
		result.Error = err.Error()
		result.Content = fmt.Sprintf("Error: %s", err.Error())
	}
	return result
}

// searchFiles searches for files by pattern
func (te *ToolExecutorImpl) searchFiles(args map[string]any, projectRoot string) (string, error) {
	pattern, _ := args["pattern"].(string)
	directory, _ := args["directory"].(string)

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	searchDir := projectRoot
	if directory != "" {
		searchDir = filepath.Join(projectRoot, directory)
	}

	var matches []string
	patternLower := strings.ToLower(pattern)

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip common ignored directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(projectRoot, path)
		nameLower := strings.ToLower(info.Name())

		// Check if matches pattern
		if strings.Contains(nameLower, patternLower) {
			matches = append(matches, relPath)
		} else if matched, _ := filepath.Match(strings.ToLower(pattern), nameLower); matched {
			matches = append(matches, relPath)
		}

		if len(matches) >= 50 {
			return fmt.Errorf("limit reached")
		}
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return "", err
	}

	if len(matches) == 0 {
		return "No files found matching pattern: " + pattern, nil
	}

	return fmt.Sprintf("Found %d files:\n%s", len(matches), strings.Join(matches, "\n")), nil
}

// searchContent searches for content in files
func (te *ToolExecutorImpl) searchContent(args map[string]any, projectRoot string) (string, error) {
	pattern, _ := args["pattern"].(string)
	filePattern, _ := args["file_pattern"].(string)
	maxResults := 20
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	regex := te.compileSearchPattern(pattern)
	results := te.searchInFiles(projectRoot, filePattern, regex, maxResults)

	if len(results) == 0 {
		return "No matches found for: " + pattern, nil
	}
	return fmt.Sprintf("Found %d matches:\n%s", len(results), strings.Join(results, "\n")), nil
}

// compileSearchPattern compiles a regex pattern with fallback to literal
func (te *ToolExecutorImpl) compileSearchPattern(pattern string) *regexp.Regexp {
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return regexp.MustCompile(regexp.QuoteMeta(pattern))
	}
	return regex
}

// searchInFiles searches for regex matches in files
func (te *ToolExecutorImpl) searchInFiles(projectRoot, filePattern string, regex *regexp.Regexp, maxResults int) []string {
	var results []string
	count := 0

	_ = filepath.Walk(projectRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}
		if te.shouldSkipPath(path) || !te.matchesFilePattern(info.Name(), filePattern) {
			return nil
		}

		matches := te.searchFileContent(path, projectRoot, regex, maxResults-count)
		results = append(results, matches...)
		count += len(matches)
		if count >= maxResults {
			return fmt.Errorf("limit reached")
		}
		return nil
	})
	return results
}

// shouldSkipPath checks if path should be skipped
func (te *ToolExecutorImpl) shouldSkipPath(path string) bool {
	return strings.Contains(path, "node_modules") || strings.Contains(path, ".git")
}

// matchesFilePattern checks if filename matches pattern
func (te *ToolExecutorImpl) matchesFilePattern(name, pattern string) bool {
	if pattern == "" {
		return true
	}
	matched, _ := filepath.Match(pattern, name)
	return matched
}

// searchFileContent searches for matches in a single file
func (te *ToolExecutorImpl) searchFileContent(path, projectRoot string, regex *regexp.Regexp, limit int) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	relPath, _ := filepath.Rel(projectRoot, path)
	var results []string
	for i, line := range strings.Split(string(content), "\n") {
		if regex.MatchString(line) {
			results = append(results, fmt.Sprintf("%s:%d: %s", relPath, i+1, strings.TrimSpace(line)))
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

// readFile reads file contents
func (te *ToolExecutorImpl) readFile(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)

	// Security check - normalize paths before comparison to prevent path traversal
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project root: %w", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}
	// Clean paths to handle .. and . components
	absProjectRoot = filepath.Clean(absProjectRoot)
	absFullPath = filepath.Clean(absFullPath)

	if !strings.HasPrefix(absFullPath, absProjectRoot+string(filepath.Separator)) && absFullPath != absProjectRoot {
		return "", fmt.Errorf("path traversal not allowed")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Handle line range
	startLine := 1
	endLine := len(lines)

	if sl, ok := args["start_line"].(float64); ok && sl > 0 {
		startLine = int(sl)
	}
	if el, ok := args["end_line"].(float64); ok && el > 0 {
		endLine = int(el)
	}

	if startLine > len(lines) {
		return "", fmt.Errorf("start_line %d exceeds file length %d", startLine, len(lines))
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Add line numbers
	var result []string
	for i := startLine - 1; i < endLine; i++ {
		result = append(result, fmt.Sprintf("%4d | %s", i+1, lines[i]))
	}

	header := fmt.Sprintf("=== %s (lines %d-%d of %d) ===\n", path, startLine, endLine, len(lines))
	return header + strings.Join(result, "\n"), nil
}

// listDirectory lists directory contents
func (te *ToolExecutorImpl) listDirectory(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	recursive, _ := args["recursive"].(bool)
	maxDepth := 2
	if md, ok := args["max_depth"].(float64); ok {
		maxDepth = int(md)
	}

	targetDir := projectRoot
	if path != "" {
		targetDir = filepath.Join(projectRoot, path)
	}

	var entries []string
	if recursive {
		entries = te.listDirectoryRecursive(targetDir, projectRoot, maxDepth)
	} else {
		var err error
		entries, err = te.listDirectoryFlat(targetDir)
		if err != nil {
			return "", err
		}
	}

	if len(entries) == 0 {
		return "Directory is empty", nil
	}
	return strings.Join(entries, "\n"), nil
}

// listDirectoryRecursive lists directory contents recursively
func (te *ToolExecutorImpl) listDirectoryRecursive(targetDir, projectRoot string, maxDepth int) []string {
	var entries []string
	baseDepth := strings.Count(targetDir, string(os.PathSeparator))

	_ = filepath.Walk(targetDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		depth := strings.Count(p, string(os.PathSeparator)) - baseDepth
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relPath, _ := filepath.Rel(projectRoot, p); relPath == "." {
			return nil
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		entries = append(entries, te.formatEntry(info, depth))
		return nil
	})
	return entries
}

// listDirectoryFlat lists directory contents non-recursively
func (te *ToolExecutorImpl) listDirectoryFlat(targetDir string) ([]string, error) {
	files, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}
	var entries []string
	for _, f := range files {
		info, _ := f.Info()
		if f.IsDir() {
			entries = append(entries, fmt.Sprintf("📁 %s/", f.Name()))
		} else if info != nil {
			entries = append(entries, fmt.Sprintf("📄 %s (%d bytes)", f.Name(), info.Size()))
		}
	}
	return entries, nil
}

// formatEntry formats a file/directory entry with indentation
func (te *ToolExecutorImpl) formatEntry(info os.FileInfo, depth int) string {
	prefix := strings.Repeat("  ", depth)
	if info.IsDir() {
		return fmt.Sprintf("%s📁 %s/", prefix, info.Name())
	}
	return fmt.Sprintf("%s📄 %s (%d bytes)", prefix, info.Name(), info.Size())
}

// getFileInfo returns file metadata
func (te *ToolExecutorImpl) getFileInfo(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"name":     info.Name(),
		"path":     path,
		"size":     info.Size(),
		"isDir":    info.IsDir(),
		"modified": info.ModTime().Format("2006-01-02 15:04:05"),
		"ext":      filepath.Ext(path),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// listFunctions extracts function names from a file
func (te *ToolExecutorImpl) listFunctions(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	functions := te.extractFunctions(string(content), ext)

	if functions == nil {
		return "Unsupported file type for function extraction: " + ext, nil
	}
	if len(functions) == 0 {
		return "No functions found in " + path, nil
	}

	unique := removeDuplicateFunctions(functions)
	return fmt.Sprintf("Functions in %s:\n- %s", path, strings.Join(unique, "\n- ")), nil
}

// extractFunctions extracts function names based on file extension
func (te *ToolExecutorImpl) extractFunctions(content, ext string) []string {
	switch ext {
	case extGo:
		return te.extractGoFunctions(content)
	case extTS, extJS, ".tsx", ".jsx":
		return te.extractJSFunctions(content)
	case extVue:
		return te.extractVueFunctions(content)
	default:
		return nil
	}
}

// extractGoFunctions extracts Go function names
func (te *ToolExecutorImpl) extractGoFunctions(content string) []string {
	re := regexp.MustCompile(`func\s+(\([^)]+\)\s+)?(\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(content, -1)
	var functions []string
	for _, m := range matches {
		if len(m) >= 3 {
			functions = append(functions, m[2])
		}
	}
	return functions
}

// extractJSFunctions extracts JS/TS function names
func (te *ToolExecutorImpl) extractJSFunctions(content string) []string {
	patterns := []string{
		`function\s+(\w+)\s*\(`,
		`(\w+)\s*=\s*(?:async\s+)?function`,
		`(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`,
		`(?:async\s+)?(\w+)\s*\([^)]*\)\s*{`,
	}
	var functions []string
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		for _, m := range re.FindAllStringSubmatch(content, -1) {
			if len(m) >= 2 && m[1] != "" {
				functions = append(functions, m[1])
			}
		}
	}
	return functions
}

// extractVueFunctions extracts Vue script function names
func (te *ToolExecutorImpl) extractVueFunctions(content string) []string {
	scriptRe := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`)
	scriptMatch := scriptRe.FindStringSubmatch(content)
	if len(scriptMatch) < 2 {
		return []string{}
	}
	funcRe := regexp.MustCompile(`(?:function|const|let|var)\s+(\w+)`)
	var functions []string
	for _, m := range funcRe.FindAllStringSubmatch(scriptMatch[1], -1) {
		if len(m) >= 2 {
			functions = append(functions, m[1])
		}
	}
	return functions
}

// removeDuplicateFunctions removes duplicate function names
func removeDuplicateFunctions(items []string) []string {
	seen := make(map[string]bool)
	var unique []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}
	return unique
}

// gitStatus returns git status
func (te *ToolExecutorImpl) gitStatus(projectRoot string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	if len(output) == 0 {
		return "Working directory clean - no changes", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := strings.TrimSpace(line[2:])

		var statusText string
		switch strings.TrimSpace(status) {
		case "M":
			statusText = "modified"
		case "A":
			statusText = "added"
		case "D":
			statusText = "deleted"
		case "??":
			statusText = "untracked"
		case "MM":
			statusText = "modified (staged + unstaged)"
		default:
			statusText = status
		}
		result = append(result, fmt.Sprintf("%s: %s", statusText, file))
	}

	return fmt.Sprintf("Git status (%d files):\n%s", len(result), strings.Join(result, "\n")), nil
}

// gitDiff returns git diff
func (te *ToolExecutorImpl) gitDiff(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	staged, _ := args["staged"].(bool)

	cmdArgs := []string{"diff"}
	if staged {
		cmdArgs = append(cmdArgs, "--staged")
	}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	if len(output) == 0 {
		return "No differences found", nil
	}

	// Truncate if too long
	result := string(output)
	if len(result) > 5000 {
		result = result[:5000] + "\n... (truncated)"
	}

	return result, nil
}

// gitLog returns recent commits
func (te *ToolExecutorImpl) gitLog(args map[string]any, projectRoot string) (string, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	path, _ := args["path"].(string)

	cmdArgs := []string{"log", fmt.Sprintf("-n%d", limit), "--oneline", "--format=%h %s (%an, %ar)"}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return "No commits found", nil
	}

	return fmt.Sprintf("Recent commits:\n%s", strings.TrimSpace(string(output))), nil
}

// listSymbols uses language analyzers to extract symbols from a file
func (te *ToolExecutorImpl) listSymbols(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	kindFilter, _ := args["kind"].(string)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	analyzer := te.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer available for file type: %s", filepath.Ext(path)), nil
	}

	symbols, err := analyzer.ExtractSymbols(context.Background(), path, content)
	if err != nil {
		return "", err
	}

	// Filter by kind if specified
	if kindFilter != "" {
		var filtered []analysis.Symbol
		for _, s := range symbols {
			if strings.EqualFold(string(s.Kind), kindFilter) {
				filtered = append(filtered, s)
			}
		}
		symbols = filtered
	}

	if len(symbols) == 0 {
		return fmt.Sprintf("No symbols found in %s", path), nil
	}

	// Format output
	lines := make([]string, 0, len(symbols)+1)
	lines = append(lines, fmt.Sprintf("Symbols in %s (%s):", path, analyzer.Language()))
	for _, s := range symbols {
		line := fmt.Sprintf("  [%s] %s", s.Kind, s.Name)
		if s.StartLine > 0 {
			line += fmt.Sprintf(" (line %d)", s.StartLine)
		}
		if s.Parent != "" {
			line += fmt.Sprintf(" <- %s", s.Parent)
		}
		if s.Signature != "" {
			line += fmt.Sprintf("\n       %s", s.Signature)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// getImports uses language analyzers to extract imports from a file
func (te *ToolExecutorImpl) getImports(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	analyzer := te.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer available for file type: %s", filepath.Ext(path)), nil
	}

	imports, err := analyzer.GetImports(context.Background(), path, content)
	if err != nil {
		return "", err
	}

	if len(imports) == 0 {
		return fmt.Sprintf("No imports found in %s", path), nil
	}

	localImports, externalImports := te.separateImports(imports)
	return te.formatImportsOutput(path, localImports, externalImports), nil
}

// separateImports separates imports into local and external
func (te *ToolExecutorImpl) separateImports(imports []analysis.Import) ([]analysis.Import, []analysis.Import) {
	var local, external []analysis.Import
	for _, imp := range imports {
		if imp.IsLocal {
			local = append(local, imp)
		} else {
			external = append(external, imp)
		}
	}
	return local, external
}

// formatImportsOutput formats imports for display
func (te *ToolExecutorImpl) formatImportsOutput(path string, local, external []analysis.Import) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Imports in %s:", path))

	if len(local) > 0 {
		lines = append(lines, "\nLocal imports:")
		lines = append(lines, te.formatImportLines(local)...)
	}
	if len(external) > 0 {
		lines = append(lines, "\nExternal imports:")
		lines = append(lines, te.formatImportLines(external)...)
	}
	return strings.Join(lines, "\n")
}

// formatImportLines formats a slice of imports as lines
func (te *ToolExecutorImpl) formatImportLines(imports []analysis.Import) []string {
	lines := make([]string, 0, len(imports))
	for _, imp := range imports {
		line := "  " + imp.Path
		if imp.Alias != "" {
			line += " as " + imp.Alias
		}
		if len(imp.Names) > 0 {
			line += " { " + strings.Join(imp.Names, ", ") + " }"
		}
		lines = append(lines, line)
	}
	return lines
}

// searchSymbols searches for symbols across the project
func (te *ToolExecutorImpl) searchSymbols(args map[string]any, projectRoot string) (string, error) {
	query, _ := args["query"].(string)
	kindFilter, _ := args["kind"].(string)

	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	// Ensure index is built
	if !te.symbolIndex.IsIndexed() {
		te.logger.Info("Building symbol index...")
		if err := te.symbolIndex.IndexProject(context.Background(), projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
		stats := te.symbolIndex.Stats()
		te.logger.Info(fmt.Sprintf("Indexed %d symbols in %d files", stats["total_symbols"], stats["files"]))
	}

	// Search
	results := te.symbolIndex.SearchByName(query)

	// Filter by kind if specified
	if kindFilter != "" {
		var filtered []analysis.Symbol
		for _, s := range results {
			if strings.EqualFold(string(s.Kind), kindFilter) {
				filtered = append(filtered, s)
			}
		}
		results = filtered
	}

	if len(results) == 0 {
		return fmt.Sprintf("No symbols found matching '%s'", query), nil
	}

	// Limit results
	if len(results) > 30 {
		results = results[:30]
	}

	// Format output
	lines := make([]string, 0, len(results)+1)
	lines = append(lines, fmt.Sprintf("Found %d symbols matching '%s':", len(results), query))
	for _, s := range results {
		line := fmt.Sprintf("  [%s] %s", s.Kind, s.Name)
		if s.FilePath != "" {
			line += fmt.Sprintf(" in %s", s.FilePath)
		}
		if s.StartLine > 0 {
			line += fmt.Sprintf(":%d", s.StartLine)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// findDefinition finds where a symbol is defined
func (te *ToolExecutorImpl) findDefinition(args map[string]any, projectRoot string) (string, error) {
	name, _ := args["name"].(string)
	kindFilter, _ := args["kind"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	// Ensure index is built
	if !te.symbolIndex.IsIndexed() {
		te.logger.Info("Building symbol index...")
		if err := te.symbolIndex.IndexProject(context.Background(), projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
	}

	// Find exact match
	var kind analysis.SymbolKind
	if kindFilter != "" {
		kind = analysis.SymbolKind(kindFilter)
	}

	sym := te.symbolIndex.FindDefinition(name, kind)
	if sym == nil {
		return fmt.Sprintf("No definition found for '%s'", name), nil
	}

	// Read the relevant code
	fullPath := filepath.Join(projectRoot, sym.FilePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Found %s '%s' in %s:%d but couldn't read file", sym.Kind, sym.Name, sym.FilePath, sym.StartLine), nil
	}

	lines := strings.Split(string(content), "\n")
	startLine := sym.StartLine - 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := sym.EndLine
	if endLine == 0 || endLine > len(lines) {
		endLine = startLine + 20 // Show 20 lines by default
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Format output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Definition of %s '%s':\n", sym.Kind, sym.Name))
	result.WriteString(fmt.Sprintf("File: %s\n", sym.FilePath))
	result.WriteString(fmt.Sprintf("Lines: %d-%d\n\n", sym.StartLine, endLine))

	for i := startLine; i < endLine; i++ {
		result.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
	}

	return result.String(), nil
}

// findReferences finds all references to a symbol across the project
func (te *ToolExecutorImpl) findReferences(args map[string]any, projectRoot string) (string, error) {
	name, _ := args["name"].(string)
	kindFilter, _ := args["kind"].(string)
	includeDefinition := true
	if incDef, ok := args["include_definition"].(bool); ok {
		includeDefinition = incDef
	}

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	// Use ReferenceFinder
	refFinder := analyzers.NewReferenceFinder(te.registry)

	var kind analysis.SymbolKind
	if kindFilter != "" {
		kind = analysis.SymbolKind(kindFilter)
	}

	refs, err := refFinder.FindReferences(context.Background(), projectRoot, name, kind)
	if err != nil {
		return "", fmt.Errorf("failed to find references: %w", err)
	}

	// Filter out definitions if requested
	if !includeDefinition {
		var filtered []analyzers.Reference
		for _, ref := range refs {
			if !ref.IsDefinition {
				filtered = append(filtered, ref)
			}
		}
		refs = filtered
	}

	if len(refs) == 0 {
		return fmt.Sprintf("No references found for '%s'", name), nil
	}

	// Format output
	var result strings.Builder
	defCount := 0
	usageCount := 0
	for _, ref := range refs {
		if ref.IsDefinition {
			defCount++
		} else {
			usageCount++
		}
	}

	result.WriteString(fmt.Sprintf("Found %d references to '%s' (%d definitions, %d usages):\n\n",
		len(refs), name, defCount, usageCount))

	for _, ref := range refs {
		marker := "  "
		if ref.IsDefinition {
			marker = "* "
		}
		result.WriteString(fmt.Sprintf("%s%s:%d:%d\n", marker, ref.FilePath, ref.Line, ref.Column))
		result.WriteString(fmt.Sprintf("    %s\n", ref.LineText))
	}

	return result.String(), nil
}

// getSymbolInfo returns detailed information about a symbol
func (te *ToolExecutorImpl) getSymbolInfo(args map[string]any, projectRoot string) (string, error) {
	name, _ := args["name"].(string)
	kindFilter, _ := args["kind"].(string)
	filePath, _ := args["file_path"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	// Ensure index is built
	if !te.symbolIndex.IsIndexed() {
		te.logger.Info("Building symbol index...")
		if err := te.symbolIndex.IndexProject(context.Background(), projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
	}

	// Find symbol
	var kind analysis.SymbolKind
	if kindFilter != "" {
		kind = analysis.SymbolKind(kindFilter)
	}

	sym := te.symbolIndex.FindDefinition(name, kind)
	if sym == nil {
		// Try searching by name
		results := te.symbolIndex.SearchByName(name)
		if len(results) == 0 {
			return fmt.Sprintf("No symbol found: '%s'", name), nil
		}
		// Filter by file path if provided
		if filePath != "" {
			for _, s := range results {
				if s.FilePath == filePath {
					sym = &s
					break
				}
			}
		}
		if sym == nil {
			sym = &results[0]
		}
	}

	// Read source code
	fullPath := filepath.Join(projectRoot, sym.FilePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		content = nil // Continue without source
	}

	// Format output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Symbol: %s\n", sym.Name))
	result.WriteString(fmt.Sprintf("Kind: %s\n", sym.Kind))
	result.WriteString(fmt.Sprintf("File: %s:%d\n", sym.FilePath, sym.StartLine))

	if sym.Parent != "" {
		result.WriteString(fmt.Sprintf("Parent: %s\n", sym.Parent))
	}

	if len(sym.Modifiers) > 0 {
		result.WriteString(fmt.Sprintf("Modifiers: %s\n", strings.Join(sym.Modifiers, ", ")))
	}

	if sym.Signature != "" {
		result.WriteString(fmt.Sprintf("\nSignature:\n  %s\n", sym.Signature))
	}

	if sym.DocComment != "" {
		result.WriteString(fmt.Sprintf("\nDocumentation:\n  %s\n", sym.DocComment))
	}

	// Show children if any
	if len(sym.Children) > 0 {
		result.WriteString(fmt.Sprintf("\nMembers (%d):\n", len(sym.Children)))
		for _, child := range sym.Children {
			result.WriteString(fmt.Sprintf("  [%s] %s", child.Kind, child.Name))
			if child.Signature != "" {
				result.WriteString(fmt.Sprintf(" - %s", child.Signature))
			}
			result.WriteString("\n")
		}
	}

	// Show source code
	if content != nil && sym.StartLine > 0 {
		lines := strings.Split(string(content), "\n")
		startLine := sym.StartLine - 1
		endLine := sym.EndLine
		if endLine == 0 || endLine > len(lines) {
			endLine = startLine + 15
		}
		if endLine > len(lines) {
			endLine = len(lines)
		}

		result.WriteString(fmt.Sprintf("\nSource (lines %d-%d):\n", sym.StartLine, endLine))
		for i := startLine; i < endLine; i++ {
			result.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
		}
	}

	return result.String(), nil
}

// getClassHierarchy returns class inheritance hierarchy
func (te *ToolExecutorImpl) getClassHierarchy(args map[string]any, projectRoot string) (string, error) {
	className, _ := args["class_name"].(string)
	direction, _ := args["direction"].(string)

	if className == "" {
		return "", fmt.Errorf("class_name is required")
	}
	if direction == "" {
		direction = "both"
	}

	// Ensure index is built
	if !te.symbolIndex.IsIndexed() {
		if err := te.symbolIndex.IndexProject(context.Background(), projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
	}

	// Find the class
	sym := te.symbolIndex.FindDefinition(className, analysis.KindClass)
	if sym == nil {
		// Try interface
		sym = te.symbolIndex.FindDefinition(className, analysis.KindInterface)
	}
	if sym == nil {
		return fmt.Sprintf("Class or interface '%s' not found", className), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Class Hierarchy for: %s\n", className))
	result.WriteString(fmt.Sprintf("Type: %s\n", sym.Kind))
	result.WriteString(fmt.Sprintf("File: %s:%d\n\n", sym.FilePath, sym.StartLine))

	// Get parent from symbol
	if (direction == "up" || direction == "both") && sym.Parent != "" {
		result.WriteString("Extends/Implements:\n")
		parents := strings.Split(sym.Parent, ",")
		for _, p := range parents {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			result.WriteString(fmt.Sprintf("  ↑ %s", p))
			// Try to find parent definition
			parentSym := te.symbolIndex.FindDefinition(p, "")
			if parentSym != nil {
				result.WriteString(fmt.Sprintf(" (%s:%d)", parentSym.FilePath, parentSym.StartLine))
			}
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	// Find subclasses (classes that extend this one)
	if direction == "down" || direction == "both" {
		allClasses := te.symbolIndex.GetSymbolsByKind(analysis.KindClass)
		var subclasses []analysis.Symbol
		for _, c := range allClasses {
			if c.Parent != "" && strings.Contains(c.Parent, className) {
				subclasses = append(subclasses, c)
			}
		}

		if len(subclasses) > 0 {
			result.WriteString(fmt.Sprintf("Subclasses (%d):\n", len(subclasses)))
			for _, sub := range subclasses {
				result.WriteString(fmt.Sprintf("  ↓ %s (%s:%d)\n", sub.Name, sub.FilePath, sub.StartLine))
			}
		} else {
			result.WriteString("Subclasses: none found\n")
		}
	}

	return result.String(), nil
}

// getWidgetTree returns Flutter widget tree structure
func (te *ToolExecutorImpl) getWidgetTree(args map[string]any, projectRoot string) (string, error) {
	filePath, _ := args["file_path"].(string)
	widgetName, _ := args["widget_name"].(string)

	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	// Check if it's a Dart file
	if !strings.HasSuffix(filePath, ".dart") {
		return "", fmt.Errorf("file must be a Dart file (.dart)")
	}

	fullPath := filepath.Join(projectRoot, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)

	// Find widget classes (classes extending StatelessWidget or StatefulWidget)
	widgetRe := regexp.MustCompile(`class\s+(\w+)\s+extends\s+(StatelessWidget|StatefulWidget|State<\w+>)`)
	matches := widgetRe.FindAllStringSubmatch(text, -1)

	if len(matches) == 0 {
		return "No Flutter widgets found in this file", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Widget Tree in %s:\n\n", filePath))

	for _, match := range matches {
		name := match[1]
		extends := match[2]

		// Filter by widget name if specified
		if widgetName != "" && name != widgetName {
			continue
		}

		result.WriteString(fmt.Sprintf("📦 %s (extends %s)\n", name, extends))

		// Find build method and extract widget tree
		buildRe := regexp.MustCompile(`Widget\s+build\s*\([^)]*\)\s*\{`)
		buildMatch := buildRe.FindStringIndex(text)
		if buildMatch != nil {
			// Extract widgets used in build method
			widgetUsageRe := regexp.MustCompile(`\b(Scaffold|AppBar|Container|Column|Row|Text|Center|Padding|ListView|GridView|Stack|Positioned|Expanded|Flexible|SizedBox|Card|ListTile|IconButton|FloatingActionButton|ElevatedButton|TextButton|TextField|Image|Icon|Drawer|BottomNavigationBar|TabBar|TabBarView)\s*\(`)
			usages := widgetUsageRe.FindAllStringSubmatch(text, -1)

			if len(usages) > 0 {
				seen := make(map[string]int)
				for _, u := range usages {
					seen[u[1]]++
				}
				result.WriteString("  Used widgets:\n")
				for widget, count := range seen {
					result.WriteString(fmt.Sprintf("    • %s (%dx)\n", widget, count))
				}
			}
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// getDependentFiles returns files that depend on or are depended by the specified file
func (te *ToolExecutorImpl) getDependentFiles(args map[string]any, projectRoot string) (string, error) {
	filePath, _ := args["file_path"].(string)
	direction, _ := args["direction"].(string)
	depth := 1
	if d, ok := args["depth"].(float64); ok && d > 0 {
		depth = int(d)
		if depth > 3 {
			depth = 3
		}
	}

	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}
	if direction == "" {
		direction = "both"
	}

	// Build dependency graph
	if _, err := te.callGraph.BuildDependencyGraph(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build dependency graph: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Dependencies for: %s (depth: %d)\n\n", filePath, depth))

	// Get files this file imports
	if direction == "imports" || direction == "both" {
		deps := te.callGraph.GetFileDependencies(filePath)
		result.WriteString(fmt.Sprintf("Imports (%d files):\n", len(deps)))
		if len(deps) == 0 {
			result.WriteString("  (none)\n")
		}
		for _, dep := range deps {
			result.WriteString(fmt.Sprintf("  → %s\n", dep.FilePath))
		}
		result.WriteString("\n")
	}

	// Get files that import this file
	if direction == "importers" || direction == "both" {
		dependents := te.callGraph.GetFileDependents(filePath)
		result.WriteString(fmt.Sprintf("Imported by (%d files):\n", len(dependents)))
		if len(dependents) == 0 {
			result.WriteString("  (none)\n")
		}
		for _, dep := range dependents {
			result.WriteString(fmt.Sprintf("  ← %s\n", dep.FilePath))
		}
	}

	return result.String(), nil
}

// getChangeRisk estimates risk score for changing a file or function
func (te *ToolExecutorImpl) getChangeRisk(args map[string]any, projectRoot string) (string, error) {
	filePath, _ := args["file_path"].(string)
	functionName, _ := args["function_name"].(string)

	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	// Build graphs
	if _, err := te.callGraph.Build(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build call graph: %w", err)
	}
	if _, err := te.callGraph.BuildDependencyGraph(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build dependency graph: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Change Risk Analysis: %s", filePath))
	if functionName != "" {
		result.WriteString(fmt.Sprintf(" → %s", functionName))
	}
	result.WriteString("\n\n")

	// Calculate risk factors
	riskScore := 0.0
	factors := []string{}

	// Factor 1: Number of dependents
	dependents := te.callGraph.GetFileDependents(filePath)
	dependentCount := len(dependents)
	if dependentCount > 10 {
		riskScore += 30
		factors = append(factors, fmt.Sprintf("High coupling: %d files depend on this", dependentCount))
	} else if dependentCount > 5 {
		riskScore += 20
		factors = append(factors, fmt.Sprintf("Medium coupling: %d files depend on this", dependentCount))
	} else if dependentCount > 0 {
		riskScore += 10
		factors = append(factors, fmt.Sprintf("Low coupling: %d files depend on this", dependentCount))
	}

	// Factor 2: Function callers (if function specified)
	if functionName != "" {
		funcID := filePath + ":" + functionName
		callers := te.callGraph.GetCallers(funcID)
		if len(callers) > 5 {
			riskScore += 25
			factors = append(factors, fmt.Sprintf("Many callers: %d functions call this", len(callers)))
		} else if len(callers) > 0 {
			riskScore += 10
			factors = append(factors, fmt.Sprintf("Some callers: %d functions call this", len(callers)))
		}
	}

	// Factor 3: Check if it's a core/shared file
	if strings.Contains(filePath, "domain") || strings.Contains(filePath, "core") || strings.Contains(filePath, "shared") || strings.Contains(filePath, "common") {
		riskScore += 20
		factors = append(factors, "Core/shared module - changes affect many parts")
	}

	// Factor 4: Check for test coverage
	testFile := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "_test" + filepath.Ext(filePath)
	if _, err := os.Stat(filepath.Join(projectRoot, testFile)); err != nil {
		riskScore += 15
		factors = append(factors, "No test file found - changes harder to verify")
	} else {
		factors = append(factors, "✓ Test file exists")
	}

	// Determine risk level
	var riskLevel string
	if riskScore >= 60 {
		riskLevel = "🔴 HIGH"
	} else if riskScore >= 30 {
		riskLevel = "🟡 MEDIUM"
	} else {
		riskLevel = "🟢 LOW"
	}

	result.WriteString(fmt.Sprintf("Risk Level: %s (score: %.0f/100)\n\n", riskLevel, riskScore))
	result.WriteString("Risk Factors:\n")
	for _, f := range factors {
		result.WriteString(fmt.Sprintf("  • %s\n", f))
	}

	// Recommendations
	result.WriteString("\nRecommendations:\n")
	if riskScore >= 60 {
		result.WriteString("  • Consider breaking down changes into smaller PRs\n")
		result.WriteString("  • Ensure comprehensive test coverage before changes\n")
		result.WriteString("  • Review with team members familiar with dependents\n")
	} else if riskScore >= 30 {
		result.WriteString("  • Add tests for changed functionality\n")
		result.WriteString("  • Check dependent files for compatibility\n")
	} else {
		result.WriteString("  • Standard review process should be sufficient\n")
	}

	return result.String(), nil
}

// Project Structure Tools (Phase 6)

// detectArchitecture detects the architecture pattern of the project
func (te *ToolExecutorImpl) detectArchitecture(projectRoot string) (string, error) {
	service := NewProjectStructureService(te.logger)
	arch, err := service.DetectArchitecture(projectRoot)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Architecture: %s\n", arch.Type))
	result.WriteString(fmt.Sprintf("Confidence: %.0f%%\n", arch.Confidence*100))
	result.WriteString(fmt.Sprintf("Description: %s\n", arch.Description))

	if len(arch.Indicators) > 0 {
		result.WriteString("\nIndicators:\n")
		for _, ind := range arch.Indicators {
			result.WriteString(fmt.Sprintf("  - %s\n", ind))
		}
	}

	if len(arch.Layers) > 0 {
		result.WriteString("\nArchitectural Layers:\n")
		for _, layer := range arch.Layers {
			result.WriteString(fmt.Sprintf("  [%s] %s\n", layer.Name, layer.Path))
			if layer.Description != "" {
				result.WriteString(fmt.Sprintf("    Description: %s\n", layer.Description))
			}
			if len(layer.Dependencies) > 0 {
				result.WriteString(fmt.Sprintf("    Dependencies: %v\n", layer.Dependencies))
			}
			if len(layer.Patterns) > 0 {
				result.WriteString(fmt.Sprintf("    Patterns: %v\n", layer.Patterns))
			}
		}
	}

	return result.String(), nil
}

// detectFrameworks detects frameworks used in the project
func (te *ToolExecutorImpl) detectFrameworks(projectRoot string) (string, error) {
	service := NewProjectStructureService(te.logger)
	frameworks, err := service.DetectFrameworks(projectRoot)
	if err != nil {
		return "", err
	}

	if len(frameworks) == 0 {
		return "No frameworks detected in this project.", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Detected %d frameworks:\n\n", len(frameworks)))

	for _, fw := range frameworks {
		version := ""
		if fw.Version != "" {
			version = " v" + fw.Version
		}
		result.WriteString(fmt.Sprintf("[%s] %s%s\n", fw.Category, fw.Name, version))
		result.WriteString(fmt.Sprintf("  Language: %s\n", fw.Language))

		if len(fw.ConfigFiles) > 0 {
			result.WriteString(fmt.Sprintf("  Config files: %v\n", fw.ConfigFiles))
		}

		if len(fw.Indicators) > 0 {
			result.WriteString(fmt.Sprintf("  Detected by: %s\n", fw.Indicators[0]))
		}

		if len(fw.BestPractices) > 0 {
			result.WriteString("  Best Practices:\n")
			for i, bp := range fw.BestPractices {
				if i >= 3 {
					result.WriteString(fmt.Sprintf("    ... and %d more\n", len(fw.BestPractices)-3))
					break
				}
				result.WriteString(fmt.Sprintf("    • %s\n", bp))
			}
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// detectConventions detects coding conventions in the project
func (te *ToolExecutorImpl) detectConventions(projectRoot string) (string, error) {
	service := NewProjectStructureService(te.logger)
	conventions, err := service.DetectConventions(projectRoot)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString("Project Conventions:\n\n")

	result.WriteString(fmt.Sprintf("Naming Style: %s\n", conventions.NamingStyle))
	result.WriteString(fmt.Sprintf("Folder Structure: %s\n", conventions.FolderStructure))

	result.WriteString("\nFile Naming:\n")
	result.WriteString(fmt.Sprintf("  Style: %s\n", conventions.FileNaming.Style))
	if len(conventions.FileNaming.Suffixes) > 0 {
		result.WriteString(fmt.Sprintf("  Common suffixes: %v\n", conventions.FileNaming.Suffixes))
	}
	if len(conventions.FileNaming.Prefixes) > 0 {
		result.WriteString(fmt.Sprintf("  Common prefixes: %v\n", conventions.FileNaming.Prefixes))
	}

	result.WriteString("\nTest Conventions:\n")
	result.WriteString(fmt.Sprintf("  Location: %s\n", conventions.TestConventions.Location))
	result.WriteString(fmt.Sprintf("  File suffix: %s\n", conventions.TestConventions.FileSuffix))
	if conventions.TestConventions.Framework != "" {
		result.WriteString(fmt.Sprintf("  Framework: %s\n", conventions.TestConventions.Framework))
	}

	result.WriteString("\nImport Style:\n")
	result.WriteString(fmt.Sprintf("  Absolute imports: %v\n", conventions.ImportStyle.AbsoluteImports))
	result.WriteString(fmt.Sprintf("  Aliased imports: %v\n", conventions.ImportStyle.AliasedImports))
	result.WriteString(fmt.Sprintf("  Import order: %v\n", conventions.ImportStyle.ImportOrder))

	result.WriteString("\nCode Style:\n")
	result.WriteString(fmt.Sprintf("  Indent: %s (%d)\n", conventions.CodeStyle.IndentStyle, conventions.CodeStyle.IndentSize))
	if conventions.CodeStyle.ConfigFile != "" {
		result.WriteString(fmt.Sprintf("  Config file: %s\n", conventions.CodeStyle.ConfigFile))
	}

	return result.String(), nil
}

// getProjectStructure returns complete project structure analysis
func (te *ToolExecutorImpl) getProjectStructure(projectRoot string) (string, error) {
	service := NewProjectStructureService(te.logger)
	return service.GetArchitectureSummary(projectRoot)
}

// getRelatedLayers returns layers related to a specific file
func (te *ToolExecutorImpl) getRelatedLayers(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	service := NewProjectStructureService(te.logger)
	layers, err := service.GetRelatedLayers(projectRoot, filepath.Join(projectRoot, path))
	if err != nil {
		return "", err
	}

	if len(layers) == 0 {
		return fmt.Sprintf("No architectural layers found related to %s", path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Layers related to %s:\n\n", path))

	for _, layer := range layers {
		result.WriteString(fmt.Sprintf("[%s] %s\n", layer.Name, layer.Path))
		if layer.Description != "" {
			result.WriteString(fmt.Sprintf("  Description: %s\n", layer.Description))
		}
		if len(layer.Dependencies) > 0 {
			result.WriteString(fmt.Sprintf("  Dependencies: %v\n", layer.Dependencies))
		}
		if len(layer.Patterns) > 0 {
			result.WriteString(fmt.Sprintf("  Patterns: %v\n", layer.Patterns))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// suggestRelatedFiles suggests related files based on architecture
func (te *ToolExecutorImpl) suggestRelatedFiles(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	service := NewProjectStructureService(te.logger)
	suggestions, err := service.SuggestRelatedFiles(projectRoot, filepath.Join(projectRoot, path))
	if err != nil {
		return "", err
	}

	if len(suggestions) == 0 {
		return fmt.Sprintf("No related files found for %s", path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Related files for %s:\n\n", path))

	for _, suggestion := range suggestions {
		result.WriteString(fmt.Sprintf("  - %s\n", suggestion))
	}

	return result.String(), nil
}

// ============================================
// Call Graph Tools (Phase 3)
// ============================================

func (te *ToolExecutorImpl) getCallers(args map[string]any, projectRoot string) (string, error) {
	functionID, _ := args["function_id"].(string)
	if functionID == "" {
		return "", fmt.Errorf("function_id is required")
	}

	if _, err := te.callGraph.Build(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build call graph: %w", err)
	}

	callers := te.callGraph.GetCallers(functionID)
	if len(callers) == 0 {
		return fmt.Sprintf("No callers found for %s", functionID), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Functions that call %s:\n\n", functionID))
	for _, c := range callers {
		result.WriteString(fmt.Sprintf("  - %s (%s:%d)\n", c.Name, c.FilePath, c.Line))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) getCallees(args map[string]any, projectRoot string) (string, error) {
	functionID, _ := args["function_id"].(string)
	if functionID == "" {
		return "", fmt.Errorf("function_id is required")
	}

	if _, err := te.callGraph.Build(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build call graph: %w", err)
	}

	callees := te.callGraph.GetCallees(functionID)
	if len(callees) == 0 {
		return fmt.Sprintf("No callees found for %s", functionID), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Functions called by %s:\n\n", functionID))
	for _, c := range callees {
		result.WriteString(fmt.Sprintf("  - %s (%s:%d)\n", c.Name, c.FilePath, c.Line))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) getImpact(args map[string]any, projectRoot string) (string, error) {
	functionID, _ := args["function_id"].(string)
	if functionID == "" {
		return "", fmt.Errorf("function_id is required")
	}
	maxDepth := 3
	if d, ok := args["max_depth"].(float64); ok {
		maxDepth = int(d)
	}

	if _, err := te.callGraph.Build(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build call graph: %w", err)
	}

	affected := te.callGraph.GetImpact(functionID, maxDepth)
	if len(affected) == 0 {
		return fmt.Sprintf("No impact found for %s", functionID), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Functions affected by changes to %s (depth %d):\n\n", functionID, maxDepth))
	for _, a := range affected {
		result.WriteString(fmt.Sprintf("  - %s (%s:%d)\n", a.Name, a.FilePath, a.Line))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) getCallChain(args map[string]any, projectRoot string) (string, error) {
	startID, _ := args["start_function"].(string)
	endID, _ := args["end_function"].(string)
	if startID == "" || endID == "" {
		return "", fmt.Errorf("start_function and end_function are required")
	}
	maxDepth := 5
	if d, ok := args["max_depth"].(float64); ok {
		maxDepth = int(d)
	}

	if _, err := te.callGraph.Build(projectRoot); err != nil {
		return "", fmt.Errorf("failed to build call graph: %w", err)
	}

	chains := te.callGraph.GetCallChain(startID, endID, maxDepth)
	if len(chains) == 0 {
		return fmt.Sprintf("No call chain found from %s to %s", startID, endID), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Call chains from %s to %s:\n\n", startID, endID))
	for i, chain := range chains {
		result.WriteString(fmt.Sprintf("  Chain %d: %s\n", i+1, strings.Join(chain, " → ")))
	}
	return result.String(), nil
}

// ============================================
// Git Context Tools (Phase 5)
// ============================================

func (te *ToolExecutorImpl) gitChangedFiles(args map[string]any, projectRoot string) (string, error) {
	if te.gitContext == nil {
		te.gitContext = git.NewContextBuilder(projectRoot)
	}

	since, _ := args["since"].(string)
	pathFilter, _ := args["path_filter"].(string)

	changes, err := te.gitContext.GetRecentChanges(since, pathFilter)
	if err != nil {
		return "", fmt.Errorf("failed to get recent changes: %w", err)
	}

	if len(changes) == 0 {
		return "No recent changes found", nil
	}

	var result strings.Builder
	result.WriteString("Recently changed files:\n\n")
	for _, c := range changes {
		result.WriteString(fmt.Sprintf("  - %s (%d changes, last: %s)\n", c.FilePath, c.ChangeCount, c.LastChanged.Format("2006-01-02")))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) gitCoChanged(args map[string]any, projectRoot string) (string, error) {
	if te.gitContext == nil {
		te.gitContext = git.NewContextBuilder(projectRoot)
	}

	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	coChanged, err := te.gitContext.GetCoChangedFiles(filePath, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get co-changed files: %w", err)
	}

	if len(coChanged) == 0 {
		return fmt.Sprintf("No co-changed files found for %s", filePath), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files often changed with %s:\n\n", filePath))
	for i, f := range coChanged {
		result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, f))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) gitSuggestContext(args map[string]any, projectRoot string) (string, error) {
	if te.gitContext == nil {
		te.gitContext = git.NewContextBuilder(projectRoot)
	}

	task, _ := args["task"].(string)
	var currentFiles []string
	if files, ok := args["current_files"].([]any); ok {
		for _, f := range files {
			if s, ok := f.(string); ok {
				currentFiles = append(currentFiles, s)
			}
		}
	}
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	suggestions, err := te.gitContext.SuggestContextFiles(task, currentFiles, limit)
	if err != nil {
		return "", fmt.Errorf("failed to suggest context files: %w", err)
	}

	if len(suggestions) == 0 {
		return "No context suggestions found", nil
	}

	var result strings.Builder
	result.WriteString("Suggested files for context:\n\n")
	for i, f := range suggestions {
		result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, f))
	}
	return result.String(), nil
}

// ============================================
// Memory Tools (Phase 7)
// ============================================

func (te *ToolExecutorImpl) saveContext(args map[string]any, projectRoot string) (string, error) {
	if te.contextMemory == nil {
		return "", fmt.Errorf("context memory not initialized")
	}

	topic, _ := args["topic"].(string)
	summary, _ := args["summary"].(string)
	var files []string
	if f, ok := args["files"].([]any); ok {
		for _, file := range f {
			if s, ok := file.(string); ok {
				files = append(files, s)
			}
		}
	}

	ctx := &domain.ConversationContext{
		ProjectRoot: projectRoot,
		Topic:       topic,
		Summary:     summary,
		Files:       files,
	}

	if err := te.contextMemory.SaveContext(ctx); err != nil {
		return "", fmt.Errorf("failed to save context: %w", err)
	}

	return fmt.Sprintf("Context saved: %s", topic), nil
}

func (te *ToolExecutorImpl) findContext(args map[string]any, projectRoot string) (string, error) {
	if te.contextMemory == nil {
		return "", fmt.Errorf("context memory not initialized")
	}

	topic, _ := args["topic"].(string)
	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}

	contexts, err := te.contextMemory.FindContextByTopic(projectRoot, topic)
	if err != nil {
		return "", fmt.Errorf("failed to find context: %w", err)
	}

	if len(contexts) == 0 {
		return fmt.Sprintf("No contexts found for topic: %s", topic), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contexts matching '%s':\n\n", topic))
	for _, c := range contexts {
		result.WriteString(fmt.Sprintf("  - %s: %s (%d files)\n", c.Topic, c.Summary, len(c.Files)))
	}
	return result.String(), nil
}

func (te *ToolExecutorImpl) getRecentContexts(args map[string]any, projectRoot string) (string, error) {
	if te.contextMemory == nil {
		return "", fmt.Errorf("context memory not initialized")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	contexts, err := te.contextMemory.GetRecentContexts(projectRoot, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get recent contexts: %w", err)
	}

	if len(contexts) == 0 {
		return "No recent contexts found", nil
	}

	var result strings.Builder
	result.WriteString("Recent contexts:\n\n")
	for _, c := range contexts {
		result.WriteString(fmt.Sprintf("  - %s: %s (last accessed: %s)\n", c.Topic, c.Summary, c.LastAccessed.Format("2006-01-02")))
	}
	return result.String(), nil
}

// ============================================
// Semantic Search Tool (Phase 4)
// ============================================

func (te *ToolExecutorImpl) doSemanticSearch(args map[string]any, projectRoot string) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	filePattern, _ := args["file_pattern"].(string)

	// Use semantic search service if available
	if te.hasSemanticSearch && te.semanticSearcherService != nil {
		results, err := te.semanticSearcherService.Search(query, limit)
		if err != nil {
			te.logger.Warning(fmt.Sprintf("Semantic search failed, falling back to keyword search: %v", err))
		} else if len(results) > 0 {
			return te.formatSemanticSearchResults(results, query, filePattern), nil
		}
	}

	// Fallback to keyword-based search using symbols and content
	return te.fallbackSemanticSearch(query, filePattern, limit, projectRoot)
}

func (te *ToolExecutorImpl) formatSemanticSearchResults(results []SemanticSearchResult, query, filePattern string) string {
	var filtered []SemanticSearchResult
	for _, r := range results {
		if filePattern == "" || matchesPattern(r.FilePath, filePattern) {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return fmt.Sprintf("No semantic matches found for: %s", query)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Semantic search results for '%s':\n\n", query))
	for i, r := range filtered {
		result.WriteString(fmt.Sprintf("%d. %s (score: %.2f)\n", i+1, r.FilePath, r.Score))
		if r.Snippet != "" {
			result.WriteString(fmt.Sprintf("   %s\n", truncateSnippet(r.Snippet, 100)))
		}
		result.WriteString("\n")
	}
	return result.String()
}

func (te *ToolExecutorImpl) fallbackSemanticSearch(query, filePattern string, limit int, projectRoot string) (string, error) {
	// Extract keywords from query
	keywords := extractSemanticKeywords(query)
	if len(keywords) == 0 {
		return "Could not extract meaningful keywords from query", nil
	}

	var allResults []string
	seen := make(map[string]bool)

	// Search symbols by keywords
	for _, kw := range keywords {
		symbols := te.symbolIndex.SearchByName(kw)
		for _, sym := range symbols {
			if filePattern != "" && !matchesPattern(sym.FilePath, filePattern) {
				continue
			}
			key := fmt.Sprintf("%s:%s", sym.FilePath, sym.Name)
			if !seen[key] {
				seen[key] = true
				allResults = append(allResults, fmt.Sprintf("  - %s (%s) in %s:%d", sym.Name, sym.Kind, sym.FilePath, sym.StartLine))
			}
		}
	}

	// Also search content
	for _, kw := range keywords {
		contentResults := te.searchInFiles(projectRoot, filePattern, te.compileSearchPattern(kw), limit)
		for _, cr := range contentResults {
			if !seen[cr] {
				seen[cr] = true
				allResults = append(allResults, fmt.Sprintf("  - %s", cr))
			}
		}
	}

	if len(allResults) == 0 {
		return fmt.Sprintf("No results found for: %s", query), nil
	}

	// Limit results
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for '%s' (keywords: %s):\n\n", query, strings.Join(keywords, ", ")))
	for _, r := range allResults {
		result.WriteString(r + "\n")
	}
	return result.String(), nil
}

func extractSemanticKeywords(query string) []string {
	// Simple keyword extraction - split by spaces and filter common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "and": true, "or": true, "with": true, "that": true,
		"this": true, "it": true, "from": true, "by": true, "as": true,
		"all": true, "any": true, "find": true, "show": true, "get": true,
		"where": true, "how": true, "what": true, "which": true, "code": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var keywords []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?\"'")
		if len(w) > 2 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}

func matchesPattern(path, pattern string) bool {
	if pattern == "" {
		return true
	}
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}

func truncateSnippet(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
