package analyzers

import (
	"context"
	"regexp"
	"shotgun_code/domain/analysis"
	"strings"
)

// stripCommentsAndStrings removes comments and string literals from code
// to prevent false positives in regex-based parsing
func stripCommentsAndStrings(code string) string {
	var result strings.Builder
	lines := strings.Split(code, "\n")
	inBlockComment := false

	for _, line := range lines {
		processedLine := processLineForComments(line, &inBlockComment)
		result.WriteString(processedLine)
		result.WriteString("\n")
	}

	return result.String()
}

// processLineForComments processes a single line, removing comments and strings
func processLineForComments(line string, inBlockComment *bool) string {
	var result strings.Builder
	i := 0
	inString := false
	stringChar := byte(0)

	for i < len(line) {
		// Handle block comment continuation
		if *inBlockComment {
			if i+1 < len(line) && line[i] == '*' && line[i+1] == '/' {
				*inBlockComment = false
				i += 2
				continue
			}
			i++
			continue
		}

		// Handle string literals
		if inString {
			if line[i] == '\\' && i+1 < len(line) {
				// Skip escaped character
				i += 2
				continue
			}
			if line[i] == stringChar {
				inString = false
				result.WriteByte(line[i])
			}
			i++
			continue
		}

		// Check for string start
		if line[i] == '"' || line[i] == '\'' || line[i] == '`' {
			inString = true
			stringChar = line[i]
			result.WriteByte(line[i])
			i++
			continue
		}

		// Check for line comment
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			// Rest of line is comment
			break
		}

		// Check for block comment start
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			*inBlockComment = true
			i += 2
			continue
		}

		result.WriteByte(line[i])
		i++
	}

	return result.String()
}

// getLineMapping creates a mapping from stripped line numbers to original line numbers
func getLineMapping(original, stripped string) map[int]int {
	origLines := strings.Split(original, "\n")
	strippedLines := strings.Split(stripped, "\n")
	mapping := make(map[int]int)

	strippedIdx := 0
	for origIdx := 0; origIdx < len(origLines) && strippedIdx < len(strippedLines); origIdx++ {
		// Simple heuristic: match by non-empty content
		if strings.TrimSpace(strippedLines[strippedIdx]) != "" {
			mapping[strippedIdx+1] = origIdx + 1
		}
		strippedIdx++
	}

	return mapping
}

// TypeScriptAnalyzer analyzes TypeScript files
type TypeScriptAnalyzer struct {
	classRe     *regexp.Regexp
	interfaceRe *regexp.Regexp
	typeRe      *regexp.Regexp
	functionRe  *regexp.Regexp
	enumRe      *regexp.Regexp
	importRe    *regexp.Regexp
}

func NewTypeScriptAnalyzer() *TypeScriptAnalyzer {
	return &TypeScriptAnalyzer{
		classRe:     regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(abstract\s+)?class\s+(\w+)`),
		interfaceRe: regexp.MustCompile(`(?m)^[\t ]*(export\s+)?interface\s+(\w+)`),
		typeRe:      regexp.MustCompile(`(?m)^[\t ]*(export\s+)?type\s+(\w+)`),
		functionRe:  regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(async\s+)?function\s+(\w+)`),
		enumRe:      regexp.MustCompile(`(?m)^[\t ]*(export\s+)?enum\s+(\w+)`),
		importRe:    regexp.MustCompile(`(?m)^import\s+.*from\s+['"]([^'"]+)['"]`),
	}
}

func (a *TypeScriptAnalyzer) Language() string     { return "typescript" }
func (a *TypeScriptAnalyzer) Extensions() []string { return []string{".ts", ".tsx"} }
func (a *TypeScriptAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx")
}

func (a *TypeScriptAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]analysis.Symbol, error) {
	var symbols []analysis.Symbol
	text := string(content)
	lines := strings.Split(text, "\n")

	// Strip comments and strings to avoid false positives
	strippedText := stripCommentsAndStrings(text)
	strippedLines := strings.Split(strippedText, "\n")

	// Helper to check if a match position is in actual code (not comment/string)
	isValidMatch := func(matchStart int) bool {
		lineNum := countLines(content[:matchStart])
		if lineNum > 0 && lineNum <= len(strippedLines) && lineNum <= len(lines) {
			strippedLine := strings.TrimSpace(strippedLines[lineNum-1])
			originalLine := strings.TrimSpace(lines[lineNum-1])
			// If stripped line is empty but original is not, the match is in a comment/string
			if len(strippedLine) == 0 && len(originalLine) > 0 {
				return false
			}
			// If stripped line is significantly shorter, the match might be in a comment
			if len(originalLine) > 0 && len(strippedLine) < len(originalLine)/3 {
				return false
			}
		}
		return true
	}

	for _, match := range a.classRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindClass, Language: "typescript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	for _, match := range a.interfaceRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindInterface, Language: "typescript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	for _, match := range a.typeRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		endLine := findTypeEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindType, Language: "typescript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	for _, match := range a.functionRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindFunction, Language: "typescript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	for _, match := range a.enumRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindEnum, Language: "typescript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	return symbols, nil
}

func (a *TypeScriptAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]analysis.Import, error) {
	var imports []analysis.Import
	for _, match := range a.importRe.FindAllStringSubmatch(string(content), -1) {
		imports = append(imports, analysis.Import{Path: match[1], IsLocal: strings.HasPrefix(match[1], ".")})
	}
	return imports, nil
}

// JavaScriptAnalyzer analyzes JavaScript files
type JavaScriptAnalyzer struct {
	classRe    *regexp.Regexp
	functionRe *regexp.Regexp
	importRe   *regexp.Regexp
}

func NewJavaScriptAnalyzer() *JavaScriptAnalyzer {
	return &JavaScriptAnalyzer{
		classRe:    regexp.MustCompile(`(?m)^[\t ]*(export\s+)?class\s+(\w+)`),
		functionRe: regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(async\s+)?function\s+(\w+)`),
		importRe:   regexp.MustCompile(`(?m)^import\s+.*from\s+['"]([^'"]+)['"]`),
	}
}

func (a *JavaScriptAnalyzer) Language() string     { return "javascript" }
func (a *JavaScriptAnalyzer) Extensions() []string { return []string{".js", ".jsx", ".mjs"} }
func (a *JavaScriptAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".jsx") || strings.HasSuffix(filePath, ".mjs")
}

func (a *JavaScriptAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]analysis.Symbol, error) {
	var symbols []analysis.Symbol
	text := string(content)
	lines := strings.Split(text, "\n")

	// Strip comments and strings to avoid false positives
	strippedText := stripCommentsAndStrings(text)
	strippedLines := strings.Split(strippedText, "\n")

	// Helper to check if a match position is in actual code (not comment/string)
	isValidMatch := func(matchStart int) bool {
		lineNum := countLines(content[:matchStart])
		if lineNum > 0 && lineNum <= len(strippedLines) && lineNum <= len(lines) {
			strippedLine := strings.TrimSpace(strippedLines[lineNum-1])
			originalLine := strings.TrimSpace(lines[lineNum-1])
			// If stripped line is empty but original is not, the match is in a comment/string
			if len(strippedLine) == 0 && len(originalLine) > 0 {
				return false
			}
			// If stripped line is significantly shorter, the match might be in a comment
			if len(originalLine) > 0 && len(strippedLine) < len(originalLine)/3 {
				return false
			}
		}
		return true
	}

	for _, match := range a.classRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindClass, Language: "javascript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	for _, match := range a.functionRe.FindAllStringSubmatchIndex(text, -1) {
		if !isValidMatch(match[0]) {
			continue
		}
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		endLine := findBlockEndLine(lines, line-1)
		symbols = append(symbols, analysis.Symbol{Name: name, Kind: analysis.KindFunction, Language: "javascript", FilePath: filePath, StartLine: line, EndLine: endLine})
	}
	return symbols, nil
}

func (a *JavaScriptAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]analysis.Import, error) {
	var imports []analysis.Import
	for _, match := range a.importRe.FindAllStringSubmatch(string(content), -1) {
		imports = append(imports, analysis.Import{Path: match[1], IsLocal: strings.HasPrefix(match[1], ".")})
	}
	return imports, nil
}

// VueAnalyzer analyzes Vue SFC files
type VueAnalyzer struct {
	scriptRe     *regexp.Regexp
	composableRe *regexp.Regexp
	importRe     *regexp.Regexp
}

func NewVueAnalyzer() *VueAnalyzer {
	return &VueAnalyzer{
		scriptRe:     regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`),
		composableRe: regexp.MustCompile(`(?m)(const|function)\s+(use\w+)`),
		importRe:     regexp.MustCompile(`(?m)^import\s+.*from\s+['"]([^'"]+)['"]`),
	}
}

func (a *VueAnalyzer) Language() string     { return "vue" }
func (a *VueAnalyzer) Extensions() []string { return []string{".vue"} }
func (a *VueAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".vue")
}

func (a *VueAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]analysis.Symbol, error) {
	var symbols []analysis.Symbol
	text := string(content)
	totalLines := strings.Count(text, "\n") + 1

	// Component name from filename
	parts := strings.Split(filePath, "/")
	if len(parts) == 0 {
		parts = strings.Split(filePath, "\\")
	}
	fileName := parts[len(parts)-1]
	componentName := strings.TrimSuffix(fileName, ".vue")

	symbols = append(symbols, analysis.Symbol{
		Name:      componentName,
		Kind:      analysis.KindComponent,
		Language:  "vue",
		FilePath:  filePath,
		StartLine: 1,
		EndLine:   totalLines,
	})

	// Extract composables from script
	if match := a.scriptRe.FindStringSubmatch(text); match != nil {
		script := match[1]
		scriptLines := strings.Split(script, "\n")
		for _, m := range a.composableRe.FindAllStringSubmatchIndex(script, -1) {
			line := countLines([]byte(script[:m[0]]))
			name := script[m[4]:m[5]]
			endLine := findBlockEndLine(scriptLines, line-1)
			symbols = append(symbols, analysis.Symbol{
				Name:      name,
				Kind:      analysis.KindComposable,
				Language:  "vue",
				FilePath:  filePath,
				StartLine: line,
				EndLine:   endLine,
			})
		}
	}
	return symbols, nil
}

func (a *VueAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]analysis.Import, error) {
	var imports []analysis.Import
	text := string(content)
	if match := a.scriptRe.FindStringSubmatch(text); match != nil {
		for _, m := range a.importRe.FindAllStringSubmatch(match[1], -1) {
			imports = append(imports, analysis.Import{Path: m[1], IsLocal: strings.HasPrefix(m[1], ".") || strings.HasPrefix(m[1], "@/")})
		}
	}
	return imports, nil
}

// GetExports extracts exported symbols from TypeScript files
func (a *TypeScriptAnalyzer) GetExports(ctx context.Context, filePath string, content []byte) ([]analysis.Export, error) {
	var exports []analysis.Export
	text := string(content)
	lines := strings.Split(text, "\n")

	// export function/class/const/type/interface
	exportDeclRe := regexp.MustCompile(`(?m)^[\t ]*export\s+(default\s+)?(async\s+)?(function|class|const|let|var|type|interface|enum)\s+(\w+)`)
	for i, line := range lines {
		if match := exportDeclRe.FindStringSubmatch(line); match != nil {
			exports = append(exports, analysis.Export{
				Name:      match[4],
				Kind:      match[3],
				IsDefault: match[1] != "",
				Line:      i + 1,
			})
		}
	}

	// export { name1, name2 }
	exportListRe := regexp.MustCompile(`(?m)^[\t ]*export\s*\{([^}]+)\}`)
	for i, line := range lines {
		if match := exportListRe.FindStringSubmatch(line); match != nil {
			names := strings.Split(match[1], ",")
			for _, n := range names {
				n = strings.TrimSpace(n)
				// Handle "name as alias"
				parts := strings.Split(n, " as ")
				name := strings.TrimSpace(parts[0])
				alias := ""
				if len(parts) > 1 {
					alias = strings.TrimSpace(parts[1])
				}
				if name != "" {
					exports = append(exports, analysis.Export{Name: name, Alias: alias, Kind: "named", Line: i + 1})
				}
			}
		}
	}

	// export default
	exportDefaultRe := regexp.MustCompile(`(?m)^[\t ]*export\s+default\s+(\w+)`)
	for i, line := range lines {
		if match := exportDefaultRe.FindStringSubmatch(line); match != nil {
			exports = append(exports, analysis.Export{Name: match[1], Kind: "default", IsDefault: true, Line: i + 1})
		}
	}

	// re-exports: export { x } from './module'
	reExportRe := regexp.MustCompile(`(?m)^[\t ]*export\s*\{([^}]+)\}\s*from\s*['"]([^'"]+)['"]`)
	for i, line := range lines {
		if match := reExportRe.FindStringSubmatch(line); match != nil {
			names := strings.Split(match[1], ",")
			for _, n := range names {
				n = strings.TrimSpace(n)
				parts := strings.Split(n, " as ")
				name := strings.TrimSpace(parts[0])
				if name != "" {
					exports = append(exports, analysis.Export{
						Name:       name,
						Kind:       "reexport",
						IsReExport: true,
						FromPath:   match[2],
						Line:       i + 1,
					})
				}
			}
		}
	}

	return exports, nil
}

// GetFunctionBody returns the full body of a function
func (a *TypeScriptAnalyzer) GetFunctionBody(ctx context.Context, filePath string, content []byte, funcName string) (string, int, int, error) {
	return extractFunctionBody(content, funcName)
}

// GetExports for JavaScript
func (a *JavaScriptAnalyzer) GetExports(ctx context.Context, filePath string, content []byte) ([]analysis.Export, error) {
	// Reuse TypeScript implementation
	ts := NewTypeScriptAnalyzer()
	return ts.GetExports(ctx, filePath, content)
}

// GetFunctionBody for JavaScript
func (a *JavaScriptAnalyzer) GetFunctionBody(ctx context.Context, filePath string, content []byte, funcName string) (string, int, int, error) {
	return extractFunctionBody(content, funcName)
}

// GetExports for Vue
func (a *VueAnalyzer) GetExports(ctx context.Context, filePath string, content []byte) ([]analysis.Export, error) {
	text := string(content)
	if match := a.scriptRe.FindStringSubmatch(text); match != nil {
		ts := NewTypeScriptAnalyzer()
		return ts.GetExports(ctx, filePath, []byte(match[1]))
	}
	return nil, nil
}

// GetFunctionBody for Vue
func (a *VueAnalyzer) GetFunctionBody(ctx context.Context, filePath string, content []byte, funcName string) (string, int, int, error) {
	text := string(content)
	if match := a.scriptRe.FindStringSubmatch(text); match != nil {
		return extractFunctionBody([]byte(match[1]), funcName)
	}
	return "", 0, 0, nil
}

// findBlockEndLine finds the end line of a block (class, function, etc.) by matching braces
// startLineIdx is 0-based index
func findBlockEndLine(lines []string, startLineIdx int) int {
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
			return i + 1 // Convert to 1-based line number
		}
	}

	// If no closing brace found, estimate based on next empty line or reasonable default
	for i := startLineIdx + 1; i < len(lines) && i < startLineIdx+100; i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Check for next top-level declaration
		if strings.HasPrefix(trimmed, "export ") || 
		   strings.HasPrefix(trimmed, "class ") || 
		   strings.HasPrefix(trimmed, "function ") ||
		   strings.HasPrefix(trimmed, "interface ") ||
		   strings.HasPrefix(trimmed, "type ") {
			return i // Return line before next declaration
		}
	}

	// Default: return start + 20 lines or end of file
	endLine := startLineIdx + 20
	if endLine > len(lines) {
		endLine = len(lines)
	}
	return endLine
}

// findTypeEndLine finds the end line of a type alias (usually ends with semicolon or next line)
func findTypeEndLine(lines []string, startLineIdx int) int {
	if startLineIdx < 0 || startLineIdx >= len(lines) {
		return startLineIdx + 1
	}

	// Check if type is on single line
	if strings.Contains(lines[startLineIdx], ";") {
		return startLineIdx + 1
	}

	// For multi-line types, look for semicolon or closing brace
	braceCount := 0
	for i := startLineIdx; i < len(lines) && i < startLineIdx+50; i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' || ch == '<' {
				braceCount++
			} else if ch == '}' || ch == '>' {
				braceCount--
			}
		}
		if strings.Contains(line, ";") && braceCount <= 0 {
			return i + 1
		}
	}

	return startLineIdx + 1
}

// extractFunctionBody is a helper to extract function body from JS/TS code
func extractFunctionBody(content []byte, funcName string) (string, int, int, error) {
	text := string(content)
	lines := strings.Split(text, "\n")

	// Patterns to find function start
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(async\s+)?function\s+` + regexp.QuoteMeta(funcName) + `\s*\(`),
		regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(const|let|var)\s+` + regexp.QuoteMeta(funcName) + `\s*=\s*(async\s*)?\(`),
		regexp.MustCompile(`(?m)^[\t ]*` + regexp.QuoteMeta(funcName) + `\s*\([^)]*\)\s*\{`),
		regexp.MustCompile(`(?m)^[\t ]*` + regexp.QuoteMeta(funcName) + `\s*:\s*(async\s*)?\(`),
	}

	startLine := -1
	for i, line := range lines {
		for _, p := range patterns {
			if p.MatchString(line) {
				startLine = i
				break
			}
		}
		if startLine >= 0 {
			break
		}
	}

	if startLine < 0 {
		return "", 0, 0, nil
	}

	// Find matching braces
	braceCount := 0
	started := false
	endLine := startLine

	for i := startLine; i < len(lines); i++ {
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
			endLine = i
			break
		}
	}

	// Extract body
	var body strings.Builder
	for i := startLine; i <= endLine && i < len(lines); i++ {
		body.WriteString(lines[i])
		if i < endLine {
			body.WriteString("\n")
		}
	}

	return body.String(), startLine + 1, endLine + 1, nil
}
