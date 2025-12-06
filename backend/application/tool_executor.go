package application

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"shotgun_code/infrastructure/analyzers"
	"strings"
)

// ToolExecutorImpl implements the ToolExecutor interface
type ToolExecutorImpl struct {
	logger      domain.Logger
	fileReader  domain.FileContentReader
	registry    analysis.AnalyzerRegistry
	symbolIndex analysis.SymbolIndex
}

// NewToolExecutor creates a new ToolExecutor
func NewToolExecutor(logger domain.Logger, fileReader domain.FileContentReader) *ToolExecutorImpl {
	registry := analyzers.NewAnalyzerRegistry()
	return &ToolExecutorImpl{
		logger:      logger,
		fileReader:  fileReader,
		registry:    registry,
		symbolIndex: analyzers.NewSymbolIndex(registry),
	}
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
	}
}

// ExecuteTool executes a tool and returns the result
func (te *ToolExecutorImpl) ExecuteTool(call domain.ToolCall, projectRoot string) domain.ToolResult {
	te.logger.Info(fmt.Sprintf("Executing tool: %s with args: %v", call.Name, call.Arguments))

	var content string
	var err error

	switch call.Name {
	case "search_files":
		content, err = te.searchFiles(call.Arguments, projectRoot)
	case "search_content":
		content, err = te.searchContent(call.Arguments, projectRoot)
	case "read_file":
		content, err = te.readFile(call.Arguments, projectRoot)
	case "list_directory":
		content, err = te.listDirectory(call.Arguments, projectRoot)
	case "get_file_info":
		content, err = te.getFileInfo(call.Arguments, projectRoot)
	case "list_functions":
		content, err = te.listFunctions(call.Arguments, projectRoot)
	case "git_status":
		content, err = te.gitStatus(projectRoot)
	case "git_diff":
		content, err = te.gitDiff(call.Arguments, projectRoot)
	case "git_log":
		content, err = te.gitLog(call.Arguments, projectRoot)
	case "list_symbols":
		content, err = te.listSymbols(call.Arguments, projectRoot)
	case "get_imports":
		content, err = te.getImports(call.Arguments, projectRoot)
	case "search_symbols":
		content, err = te.searchSymbols(call.Arguments, projectRoot)
	case "find_definition":
		content, err = te.findDefinition(call.Arguments, projectRoot)
	case "find_references":
		content, err = te.findReferences(call.Arguments, projectRoot)
	// Project Structure Tools (Phase 6)
	case "detect_architecture":
		content, err = te.detectArchitecture(projectRoot)
	case "detect_frameworks":
		content, err = te.detectFrameworks(projectRoot)
	case "detect_conventions":
		content, err = te.detectConventions(projectRoot)
	case "get_project_structure":
		content, err = te.getProjectStructure(projectRoot)
	case "get_related_layers":
		content, err = te.getRelatedLayers(call.Arguments, projectRoot)
	case "suggest_related_files":
		content, err = te.suggestRelatedFiles(call.Arguments, projectRoot)
	default:
		err = fmt.Errorf("unknown tool: %s", call.Name)
	}

	result := domain.ToolResult{
		ToolCallID: call.ID,
		Content:    content,
	}
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

	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		// Fall back to literal search
		regex = regexp.MustCompile(regexp.QuoteMeta(pattern))
	}

	var results []string
	count := 0

	err = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip ignored directories
		if strings.Contains(path, "node_modules") || strings.Contains(path, ".git") {
			return nil
		}

		// Check file pattern
		if filePattern != "" {
			if matched, _ := filepath.Match(filePattern, info.Name()); !matched {
				return nil
			}
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(projectRoot, path)
		lines := strings.Split(string(content), "\n")

		for i, line := range lines {
			if regex.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", relPath, i+1, strings.TrimSpace(line)))
				count++
				if count >= maxResults {
					return fmt.Errorf("limit reached")
				}
			}
		}
		return nil
	})

	if len(results) == 0 {
		return "No matches found for: " + pattern, nil
	}

	return fmt.Sprintf("Found %d matches:\n%s", len(results), strings.Join(results, "\n")), nil
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
		baseDepth := strings.Count(targetDir, string(os.PathSeparator))
		
		filepath.Walk(targetDir, func(p string, info os.FileInfo, err error) error {
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
			
			relPath, _ := filepath.Rel(projectRoot, p)
			if relPath == "." {
				return nil
			}
			
			// Skip ignored
			if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			
			prefix := strings.Repeat("  ", depth)
			if info.IsDir() {
				entries = append(entries, fmt.Sprintf("%s📁 %s/", prefix, info.Name()))
			} else {
				entries = append(entries, fmt.Sprintf("%s📄 %s (%d bytes)", prefix, info.Name(), info.Size()))
			}
			
			return nil
		})
	} else {
		files, err := os.ReadDir(targetDir)
		if err != nil {
			return "", err
		}
		
		for _, f := range files {
			info, _ := f.Info()
			if f.IsDir() {
				entries = append(entries, fmt.Sprintf("📁 %s/", f.Name()))
			} else if info != nil {
				entries = append(entries, fmt.Sprintf("📄 %s (%d bytes)", f.Name(), info.Size()))
			}
		}
	}

	if len(entries) == 0 {
		return "Directory is empty", nil
	}

	return strings.Join(entries, "\n"), nil
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
	var functions []string

	switch ext {
	case ".go":
		// Simple regex for Go functions
		re := regexp.MustCompile(`func\s+(\([^)]+\)\s+)?(\w+)\s*\(`)
		matches := re.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			if len(m) >= 3 {
				functions = append(functions, m[2])
			}
		}
	case ".ts", ".js", ".tsx", ".jsx":
		// Simple regex for JS/TS functions
		patterns := []string{
			`function\s+(\w+)\s*\(`,           // function name()
			`(\w+)\s*=\s*(?:async\s+)?function`, // name = function
			`(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`, // name = () =>
			`(?:async\s+)?(\w+)\s*\([^)]*\)\s*{`, // method name() {
		}
		for _, p := range patterns {
			re := regexp.MustCompile(p)
			matches := re.FindAllStringSubmatch(string(content), -1)
			for _, m := range matches {
				if len(m) >= 2 && m[1] != "" {
					functions = append(functions, m[1])
				}
			}
		}
	case ".vue":
		// Extract from <script> section
		scriptRe := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`)
		scriptMatch := scriptRe.FindStringSubmatch(string(content))
		if len(scriptMatch) >= 2 {
			script := scriptMatch[1]
			funcRe := regexp.MustCompile(`(?:function|const|let|var)\s+(\w+)`)
			matches := funcRe.FindAllStringSubmatch(script, -1)
			for _, m := range matches {
				if len(m) >= 2 {
					functions = append(functions, m[1])
				}
			}
		}
	default:
		return "Unsupported file type for function extraction: " + ext, nil
	}

	if len(functions) == 0 {
		return "No functions found in " + path, nil
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, f := range functions {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	return fmt.Sprintf("Functions in %s:\n- %s", path, strings.Join(unique, "\n- ")), nil
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
	var result []string
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

	symbols, err := analyzer.ExtractSymbols(nil, path, content)
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
	var lines []string
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

	imports, err := analyzer.GetImports(nil, path, content)
	if err != nil {
		return "", err
	}

	if len(imports) == 0 {
		return fmt.Sprintf("No imports found in %s", path), nil
	}

	// Format output
	var lines []string
	lines = append(lines, fmt.Sprintf("Imports in %s:", path))

	var localImports, externalImports []analysis.Import
	for _, imp := range imports {
		if imp.IsLocal {
			localImports = append(localImports, imp)
		} else {
			externalImports = append(externalImports, imp)
		}
	}

	if len(localImports) > 0 {
		lines = append(lines, "\nLocal imports:")
		for _, imp := range localImports {
			line := "  " + imp.Path
			if imp.Alias != "" {
				line += " as " + imp.Alias
			}
			if len(imp.Names) > 0 {
				line += " { " + strings.Join(imp.Names, ", ") + " }"
			}
			lines = append(lines, line)
		}
	}

	if len(externalImports) > 0 {
		lines = append(lines, "\nExternal imports:")
		for _, imp := range externalImports {
			line := "  " + imp.Path
			if imp.Alias != "" {
				line += " as " + imp.Alias
			}
			if len(imp.Names) > 0 {
				line += " { " + strings.Join(imp.Names, ", ") + " }"
			}
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n"), nil
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
		if err := te.symbolIndex.IndexProject(nil, projectRoot); err != nil {
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
	var lines []string
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
		if err := te.symbolIndex.IndexProject(nil, projectRoot); err != nil {
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

	refs, err := refFinder.FindReferences(nil, projectRoot, name, kind)
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
