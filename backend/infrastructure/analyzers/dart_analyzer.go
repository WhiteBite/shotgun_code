package analyzers

import (
	"context"
	"regexp"
	"shotgun_code/domain/analysis"
	"strings"
)

// DartAnalyzer analyzes Dart/Flutter files
type DartAnalyzer struct {
	classRe    *regexp.Regexp
	widgetRe   *regexp.Regexp
	functionRe *regexp.Regexp
	enumRe     *regexp.Regexp
	importRe   *regexp.Regexp
}

func NewDartAnalyzer() *DartAnalyzer {
	return &DartAnalyzer{
		classRe:    regexp.MustCompile(`(?m)^[\t ]*(abstract\s+)?class\s+(\w+)`),
		widgetRe:   regexp.MustCompile(`(?m)^[\t ]*class\s+(\w+)\s+extends\s+(StatelessWidget|StatefulWidget)`),
		functionRe: regexp.MustCompile(`(?m)^([\w<>?]+)\s+(\w+)\s*\([^)]*\)\s*(?:async\s*)?\{`),
		enumRe:     regexp.MustCompile(`(?m)^[\t ]*enum\s+(\w+)`),
		importRe:   regexp.MustCompile(`(?m)^import\s+['"]([^'"]+)['"]`),
	}
}

func (a *DartAnalyzer) Language() string     { return "dart" }
func (a *DartAnalyzer) Extensions() []string { return []string{".dart"} }
func (a *DartAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".dart")
}

func (a *DartAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]analysis.Symbol, error) {
	var symbols []analysis.Symbol
	text := string(content)
	lines := strings.Split(text, "\n")

	// Widgets first
	widgetNames := make(map[string]bool)
	for _, match := range a.widgetRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[2]:match[3]]
		widgetType := text[match[4]:match[5]]
		widgetNames[name] = true
		endLine := findDartBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{
			Name:      name,
			Kind:      analysis.KindWidget,
			Language:  "dart",
			FilePath:  filePath,
			StartLine: line,
			EndLine:   endLine,
			Extra:     map[string]string{"widgetType": widgetType},
		})
	}

	// Classes (excluding widgets)
	for _, match := range a.classRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		if widgetNames[name] {
			continue
		}
		endLine := findDartBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindClass, Language: "dart", FilePath: filePath, StartLine: line, EndLine: endLine})
	}

	// Enums
	for _, match := range a.enumRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[2]:match[3]]
		endLine := findDartBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindEnum, Language: "dart", FilePath: filePath, StartLine: line, EndLine: endLine})
	}

	return symbols, nil
}

// findDartBlockEndLine finds the end line of a block by matching braces
func findDartBlockEndLine(lines []string, startLineIdx int) int {
	if startLineIdx < 0 || startLineIdx >= len(lines) {
		return startLineIdx + 1
	}

	braceCount := 0
	started := false

	for i := startLineIdx; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				braceCount++
				started = true
			} else if ch == '}' {
				braceCount--
			}
		}
		if started && braceCount == 0 {
			return i + 1
		}
	}

	endLine := startLineIdx + 20
	if endLine > len(lines) {
		endLine = len(lines)
	}
	return endLine
}

func (a *DartAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]analysis.Import, error) {
	var imports []analysis.Import
	for _, match := range a.importRe.FindAllStringSubmatch(string(content), -1) {
		path := match[1]
		isLocal := strings.HasPrefix(path, "package:") || strings.HasPrefix(path, "./")
		imports = append(imports, analysis.Import{Path: path, IsLocal: isLocal})
	}
	return imports, nil
}
