package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain/analysis"
	"shotgun_code/infrastructure/analyzers"
	"strings"
)

func (e *Executor) registerAnalysisTools() {
	e.tools["list_symbols"] = e.listSymbols
	e.tools["search_symbols"] = e.searchSymbols
	e.tools["find_definition"] = e.findDefinition
	e.tools["get_imports"] = e.getImports
	e.tools["find_references"] = e.findReferences
	e.tools["get_function"] = e.getFunction
	e.tools["get_exports"] = e.getExports
}

func (e *Executor) listSymbols(args map[string]any, projectRoot string) (string, error) {
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

	analyzer := e.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer for file type: %s", filepath.Ext(path)), nil
	}

	symbols, err := analyzer.ExtractSymbols(nil, path, content)
	if err != nil {
		return "", err
	}

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
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

func (e *Executor) searchSymbols(args map[string]any, projectRoot string) (string, error) {
	query, _ := args["query"].(string)
	kindFilter, _ := args["kind"].(string)

	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	// Build index if needed
	if !e.symbolIndex.IsIndexed() {
		e.logger.Info("Building symbol index...")
		if err := e.symbolIndex.IndexProject(nil, projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
	}

	results := e.symbolIndex.SearchByName(query)

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

	if len(results) > 30 {
		results = results[:30]
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Found %d symbols matching '%s':", len(results), query))
	for _, s := range results {
		line := fmt.Sprintf("  [%s] %s in %s", s.Kind, s.Name, s.FilePath)
		if s.StartLine > 0 {
			line += fmt.Sprintf(":%d", s.StartLine)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

func (e *Executor) findDefinition(args map[string]any, projectRoot string) (string, error) {
	name, _ := args["name"].(string)
	kindFilter, _ := args["kind"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	if !e.symbolIndex.IsIndexed() {
		e.logger.Info("Building symbol index...")
		if err := e.symbolIndex.IndexProject(nil, projectRoot); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}
	}

	var kind analysis.SymbolKind
	if kindFilter != "" {
		kind = analysis.SymbolKind(kindFilter)
	}

	sym := e.symbolIndex.FindDefinition(name, kind)
	if sym == nil {
		return fmt.Sprintf("No definition found for '%s'", name), nil
	}

	// Read code
	fullPath := filepath.Join(projectRoot, sym.FilePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Found %s '%s' in %s:%d", sym.Kind, sym.Name, sym.FilePath, sym.StartLine), nil
	}

	lines := strings.Split(string(content), "\n")
	startLine := sym.StartLine - 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := sym.EndLine
	if endLine == 0 || endLine > len(lines) {
		endLine = startLine + 15
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Definition of %s '%s':\n", sym.Kind, sym.Name))
	result.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n\n", sym.FilePath, sym.StartLine, endLine))

	for i := startLine; i < endLine; i++ {
		result.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
	}

	return result.String(), nil
}

func (e *Executor) getImports(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	analyzer := e.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer for file type: %s", filepath.Ext(path)), nil
	}

	imports, err := analyzer.GetImports(nil, path, content)
	if err != nil {
		return "", err
	}

	if len(imports) == 0 {
		return fmt.Sprintf("No imports in %s", path), nil
	}

	var local, external []string
	for _, imp := range imports {
		if imp.IsLocal {
			local = append(local, imp.Path)
		} else {
			external = append(external, imp.Path)
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Imports in %s:", path))
	if len(local) > 0 {
		lines = append(lines, "\nLocal:")
		for _, p := range local {
			lines = append(lines, "  "+p)
		}
	}
	if len(external) > 0 {
		lines = append(lines, "\nExternal:")
		for _, p := range external {
			lines = append(lines, "  "+p)
		}
	}

	return strings.Join(lines, "\n"), nil
}


func (e *Executor) findReferences(args map[string]any, projectRoot string) (string, error) {
	name, _ := args["name"].(string)
	kindFilter, _ := args["kind"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	var kind analysis.SymbolKind
	if kindFilter != "" {
		kind = analysis.SymbolKind(kindFilter)
	}

	refs, err := e.refFinder.FindReferences(nil, projectRoot, name, kind)
	if err != nil {
		return "", err
	}

	if len(refs) == 0 {
		return fmt.Sprintf("No references found for '%s'", name), nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Found %d references to '%s':", len(refs), name))

	// Group by file
	byFile := make(map[string][]analyzers.Reference)
	for _, ref := range refs {
		byFile[ref.FilePath] = append(byFile[ref.FilePath], ref)
	}

	for file, fileRefs := range byFile {
		lines = append(lines, fmt.Sprintf("\n%s:", file))
		for _, ref := range fileRefs {
			marker := ""
			if ref.IsDefinition {
				marker = " [DEFINITION]"
			}
			lines = append(lines, fmt.Sprintf("  Line %d:%d%s: %s", ref.Line, ref.Column, marker, ref.LineText))
		}
	}

	return strings.Join(lines, "\n"), nil
}


func (e *Executor) getFunction(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	name, _ := args["name"].(string)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if name == "" {
		return "", fmt.Errorf("function name is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	analyzer := e.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer for file type: %s", filepath.Ext(path)), nil
	}

	body, startLine, endLine, err := analyzer.GetFunctionBody(nil, path, content, name)
	if err != nil {
		return "", err
	}

	if body == "" {
		return fmt.Sprintf("Function '%s' not found in %s", name, path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Function '%s' in %s (lines %d-%d):\n\n", name, path, startLine, endLine))
	result.WriteString(body)

	return result.String(), nil
}

func (e *Executor) getExports(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	fullPath := filepath.Join(projectRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	analyzer := e.registry.GetAnalyzer(path)
	if analyzer == nil {
		return fmt.Sprintf("No analyzer for file type: %s", filepath.Ext(path)), nil
	}

	exports, err := analyzer.GetExports(nil, path, content)
	if err != nil {
		return "", err
	}

	if len(exports) == 0 {
		return fmt.Sprintf("No exports in %s", path), nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Exports in %s (%d):", path, len(exports)))

	for _, exp := range exports {
		line := fmt.Sprintf("  [%s] %s", exp.Kind, exp.Name)
		if exp.Alias != "" {
			line += fmt.Sprintf(" as %s", exp.Alias)
		}
		if exp.IsDefault {
			line += " (default)"
		}
		if exp.IsReExport {
			line += fmt.Sprintf(" from '%s'", exp.FromPath)
		}
		if exp.Line > 0 {
			line += fmt.Sprintf(" (line %d)", exp.Line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}
