package analyzers

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"shotgun_code/domain/analysis"
	"strings"
)

// ReferenceFinder finds references to symbols across the project
type ReferenceFinder struct {
	registry analysis.AnalyzerRegistry
}

// NewReferenceFinder creates a new reference finder
func NewReferenceFinder(registry analysis.AnalyzerRegistry) *ReferenceFinder {
	return &ReferenceFinder{registry: registry}
}

// Reference represents a reference to a symbol
type Reference struct {
	FilePath   string `json:"filePath"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	LineText   string `json:"lineText"`
	Context    string `json:"context"` // surrounding context
	IsDefinition bool `json:"isDefinition"`
}

// FindReferences finds all references to a symbol in the project
func (rf *ReferenceFinder) FindReferences(ctx context.Context, projectRoot string, symbolName string, symbolKind analysis.SymbolKind) ([]Reference, error) {
	var references []Reference

	// Create regex for the symbol name (word boundary match)
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbolName) + `\b`)

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "build" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if we can analyze this file
		analyzer := rf.registry.GetAnalyzer(path)
		if analyzer == nil {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(projectRoot, path)
		lines := strings.Split(string(content), "\n")

		// Find all matches
		for i, line := range lines {
			matches := pattern.FindAllStringIndex(line, -1)
			for _, match := range matches {
				// Get context (surrounding lines)
				contextStart := i - 2
				if contextStart < 0 {
					contextStart = 0
				}
				contextEnd := i + 3
				if contextEnd > len(lines) {
					contextEnd = len(lines)
				}
				contextLines := lines[contextStart:contextEnd]

				ref := Reference{
					FilePath:   relPath,
					Line:       i + 1,
					Column:     match[0] + 1,
					LineText:   strings.TrimSpace(line),
					Context:    strings.Join(contextLines, "\n"),
				}

				// Check if this is the definition
				if symbolKind != "" {
					symbols, _ := analyzer.ExtractSymbols(ctx, relPath, content)
					for _, sym := range symbols {
						if sym.Name == symbolName && sym.StartLine == i+1 {
							ref.IsDefinition = true
							break
						}
					}
				}

				references = append(references, ref)

				// Limit results
				if len(references) >= 50 {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, err
	}

	return references, nil
}

// FindUsages finds where a symbol is used (excluding definition)
func (rf *ReferenceFinder) FindUsages(ctx context.Context, projectRoot string, symbolName string) ([]Reference, error) {
	refs, err := rf.FindReferences(ctx, projectRoot, symbolName, "")
	if err != nil {
		return nil, err
	}

	// Filter out definitions
	var usages []Reference
	for _, ref := range refs {
		if !ref.IsDefinition {
			usages = append(usages, ref)
		}
	}

	return usages, nil
}
