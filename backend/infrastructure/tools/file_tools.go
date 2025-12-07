package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const nodeModulesDir = "node_modules"

func (e *Executor) registerFileTool() {
	e.tools["search_files"] = e.searchFiles
	e.tools["read_file"] = e.readFile
	e.tools["list_directory"] = e.listDirectory
	e.tools["search_content"] = e.searchContent
}

func (e *Executor) searchFiles(args map[string]any, projectRoot string) (string, error) {
	pattern, _ := args["pattern"].(string)
	dir, _ := args["directory"].(string)

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	searchDir := projectRoot
	if dir != "" {
		searchDir = filepath.Join(projectRoot, dir)
	}

	var matches []string
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == nodeModulesDir || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(projectRoot, path)
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			matches = append(matches, relPath)
		} else if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
			matches = append(matches, relPath)
		}

		if len(matches) >= 50 {
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && !errors.Is(err, filepath.SkipAll) {
		return "", err
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No files found matching '%s'", pattern), nil
	}

	return fmt.Sprintf("Found %d files:\n%s", len(matches), strings.Join(matches, "\n")), nil
}

func (e *Executor) readFile(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	startLine, _ := args["start_line"].(float64)
	endLine, _ := args["end_line"].(float64)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)

	// Security: prevent path traversal
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project root: %w", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}
	absProjectRoot = filepath.Clean(absProjectRoot)
	absFullPath = filepath.Clean(absFullPath)

	if !strings.HasPrefix(absFullPath, absProjectRoot+string(filepath.Separator)) && absFullPath != absProjectRoot {
		return "", fmt.Errorf("path traversal not allowed")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	start := int(startLine)
	end := int(endLine)
	if start < 1 {
		start = 1
	}
	if end < 1 || end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("File: %s (lines %d-%d of %d)\n\n", path, start, end, len(lines)))
	for i := start - 1; i < end; i++ {
		result.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
	}

	return result.String(), nil
}

func (e *Executor) listDirectory(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	recursive, _ := args["recursive"].(bool)

	dir := projectRoot
	if path != "" {
		dir = filepath.Join(projectRoot, path)
	}

	var entries []string
	if recursive {
		entries = e.listDirRecursive(dir, projectRoot)
	} else {
		var err error
		entries, err = e.listDirFlat(dir)
		if err != nil {
			return "", err
		}
	}
	return strings.Join(entries, "\n"), nil
}

func (e *Executor) listDirRecursive(dir, projectRoot string) []string {
	var entries []string
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && e.shouldSkipDir(info.Name()) {
			return filepath.SkipDir
		}
		relPath, _ := filepath.Rel(projectRoot, p)
		if info.IsDir() {
			entries = append(entries, relPath+"/")
		} else {
			entries = append(entries, relPath)
		}
		if len(entries) >= 100 {
			return filepath.SkipAll
		}
		return nil
	})
	return entries
}

func (e *Executor) listDirFlat(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var entries []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if f.IsDir() {
			entries = append(entries, f.Name()+"/")
		} else {
			entries = append(entries, f.Name())
		}
	}
	return entries, nil
}

func (e *Executor) shouldSkipDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == nodeModulesDir
}

func (e *Executor) searchContent(args map[string]any, projectRoot string) (string, error) {
	pattern, _ := args["pattern"].(string)
	filePattern, _ := args["file_pattern"].(string)

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	re := e.compilePattern(pattern)
	results := e.searchFilesContent(projectRoot, filePattern, re, 30)

	if len(results) == 0 {
		return fmt.Sprintf("No matches found for '%s'", pattern), nil
	}
	return strings.Join(results, "\n"), nil
}

func (e *Executor) compilePattern(pattern string) *regexp.Regexp {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return regexp.MustCompile(regexp.QuoteMeta(pattern))
	}
	return re
}

func (e *Executor) searchFilesContent(projectRoot, filePattern string, re *regexp.Regexp, maxResults int) []string {
	var results []string
	_ = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if e.shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !e.matchesPattern(info.Name(), filePattern) {
			return nil
		}
		matches := e.searchInFile(path, projectRoot, re, maxResults-len(results))
		results = append(results, matches...)
		if len(results) >= maxResults {
			return filepath.SkipAll
		}
		return nil
	})
	return results
}

func (e *Executor) matchesPattern(name, pattern string) bool {
	if pattern == "" {
		return true
	}
	matched, _ := filepath.Match(pattern, name)
	return matched
}

func (e *Executor) searchInFile(path, projectRoot string, re *regexp.Regexp, limit int) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	relPath, _ := filepath.Rel(projectRoot, path)
	var results []string
	for i, line := range strings.Split(string(content), "\n") {
		if re.MatchString(line) {
			results = append(results, fmt.Sprintf("%s:%d: %s", relPath, i+1, strings.TrimSpace(line)))
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}
