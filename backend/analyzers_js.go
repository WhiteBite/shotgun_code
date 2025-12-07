package application

import (
	"context"
	"regexp"
	"strings"
)

// TypeScriptAnalyzer analyzes TypeScript/TSX files
type TypeScriptAnalyzer struct {
	classRe     *regexp.Regexp
	interfaceRe *regexp.Regexp
	typeRe      *regexp.Regexp
	functionRe  *regexp.Regexp
	arrowFnRe   *regexp.Regexp
	enumRe      *regexp.Regexp
	importRe    *regexp.Regexp
}

func NewTypeScriptAnalyzer() *TypeScriptAnalyzer {
	return &TypeScriptAnalyzer{
		classRe:     regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(abstract\s+)?class\s+(\w+)`),
		interfaceRe: regexp.MustCompile(`(?m)^[\t ]*(export\s+)?interface\s+(\w+)`),
		typeRe:      regexp.MustCompile(`(?m)^[\t ]*(export\s+)?type\s+(\w+)`),
		functionRe:  regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(async\s+)?function\s+(\w+)`),
		arrowFnRe:   regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(const|let)\s+(\w+)\s*=\s*(?:async\s*)?\(`),
		enumRe:      regexp.MustCompile(`(?m)^[\t ]*(export\s+)?enum\s+(\w+)`),
		importRe:    regexp.MustCompile(`(?m)^import\s+.*from\s+['"]([^'"]+)['"]`),
	}
}

func (a *TypeScriptAnalyzer) Language() string     { return "typescript" }
func (a *TypeScriptAnalyzer) Extensions() []string { return []string{".ts", ".tsx"} }
func (a *TypeScriptAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx")
}

func (a *TypeScriptAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]Symbol, error) {
	var symbols []Symbol
	text := string(content)

	for _, match := range a.classRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolClass, Language: "typescript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.interfaceRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolInterface, Language: "typescript", FilePath: filePath, StartLine: line})
	}

	for _, match := range a.typeRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolType, Language: "typescript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.functionRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolFunction, Language: "typescript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.arrowFnRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolFunction, Language: "typescript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.enumRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolEnum, Language: "typescript", FilePath: filePath, StartLine: line})
	}
	return symbols, nil
}

func (a *TypeScriptAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]Import, error) {
	var imports []Import
	for _, match := range a.importRe.FindAllStringSubmatch(string(content), -1) {
		imports = append(imports, Import{Path: match[1], IsLocal: strings.HasPrefix(match[1], ".")})
	}
	return imports, nil
}

// JavaScriptAnalyzer analyzes JS/JSX files
type JavaScriptAnalyzer struct {
	classRe    *regexp.Regexp
	functionRe *regexp.Regexp
	arrowFnRe  *regexp.Regexp
	importRe   *regexp.Regexp
}

func NewJavaScriptAnalyzer() *JavaScriptAnalyzer {
	return &JavaScriptAnalyzer{
		classRe:    regexp.MustCompile(`(?m)^[\t ]*(export\s+)?class\s+(\w+)`),
		functionRe: regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(async\s+)?function\s+(\w+)`),
		arrowFnRe:  regexp.MustCompile(`(?m)^[\t ]*(export\s+)?(const|let)\s+(\w+)\s*=\s*(?:async\s*)?\(`),
		importRe:   regexp.MustCompile(`(?m)^import\s+.*from\s+['"]([^'"]+)['"]`),
	}
}

func (a *JavaScriptAnalyzer) Language() string     { return "javascript" }
func (a *JavaScriptAnalyzer) Extensions() []string { return []string{".js", ".jsx", ".mjs"} }
func (a *JavaScriptAnalyzer) CanAnalyze(filePath string) bool {
	return strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".jsx") || strings.HasSuffix(filePath, ".mjs")
}

func (a *JavaScriptAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]Symbol, error) {
	var symbols []Symbol
	text := string(content)
	for _, match := range a.classRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[4]:match[5]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolClass, Language: "javascript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.functionRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolFunction, Language: "javascript", FilePath: filePath, StartLine: line})
	}
	for _, match := range a.arrowFnRe.FindAllStringSubmatchIndex(text, -1) {
		line := countLines(content[:match[0]])
		name := text[match[6]:match[7]]
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolFunction, Language: "javascript", FilePath: filePath, StartLine: line})
	}
	return symbols, nil
}

func (a *JavaScriptAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]Import, error) {
	var imports []Import
	for _, match := range a.importRe.FindAllStringSubmatch(string(content), -1) {
		imports = append(imports, Import{Path: match[1], IsLocal: strings.HasPrefix(match[1], ".")})
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

func (a *VueAnalyzer) ExtractSymbols(ctx context.Context, filePath string, content []byte) ([]Symbol, error) {
	var symbols []Symbol

	// Get component name from filename
	parts := strings.Split(filePath, "/")
	if len(parts) == 0 {
		parts = strings.Split(filePath, "\\")
	}
	fileName := parts[len(parts)-1]
	componentName := strings.TrimSuffix(fileName, ".vue")

	symbols = append(symbols, Symbol{
		Name:      componentName,
		Kind:      SymbolComponent,
		Language:  "vue",
		FilePath:  filePath,
		StartLine: 1,
	})

	// Extract script content
	text := string(content)
	scriptMatch := a.scriptRe.FindStringSubmatch(text)
	if scriptMatch != nil {
		scriptContent := scriptMatch[1]
		// Find composables
		for _, match := range a.composableRe.FindAllStringSubmatchIndex(scriptContent, -1) {
			line := countLines([]byte(scriptContent[:match[0]]))
			name := scriptContent[match[4]:match[5]]
			symbols = append(symbols, Symbol{
				Name:      name,
				Kind:      SymbolComposable,
				Language:  "vue",
				FilePath:  filePath,
				StartLine: line,
			})
		}
	}
	return symbols, nil
}

func (a *VueAnalyzer) GetImports(ctx context.Context, filePath string, content []byte) ([]Import, error) {
	var imports []Import
	text := string(content)
	scriptMatch := a.scriptRe.FindStringSubmatch(text)
	if scriptMatch == nil {
		return imports, nil
	}
	for _, match := range a.importRe.FindAllStringSubmatch(scriptMatch[1], -1) {
		imports = append(imports, Import{Path: match[1], IsLocal: strings.HasPrefix(match[1], ".") || strings.HasPrefix(match[1], "@/")})
	}
	return imports, nil
}
