package application

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
)

// SmartContextService collects relevant context for AI tasks based on code analysis
type SmartContextService struct {
	log              domain.Logger
	fileReader       domain.FileContentReader
	symbolGraphSvc   *SymbolGraphService
	callStackAnalyzer CallStackAnalyzerInterface
}

// CallStackAnalyzerInterface defines the interface for call stack analysis
type CallStackAnalyzerInterface interface {
	AnalyzeCallStack(ctx context.Context, projectRoot, filePath, symbolName string, maxDepth int) (*CallStackResult, error)
	GetTransitiveDependencies(ctx context.Context, projectRoot, filePath, symbolName string, maxDepth int) ([]*domain.SymbolNode, error)
}

// CallStackResult mirrors the infrastructure type for decoupling
type CallStackResult struct {
	RootSymbol   *domain.SymbolNode
	Callers      []*domain.SymbolNode
	Callees      []*domain.SymbolNode
	Dependencies []*domain.SymbolNode
	RelatedFiles []string
	TotalSymbols int
}

// SmartContextRequest represents a request for smart context collection
type SmartContextRequest struct {
	ProjectRoot   string   `json:"projectRoot"`
	Task          string   `json:"task"`          // User's task description
	SelectedFiles []string `json:"selectedFiles"` // Files explicitly selected by user
	SelectedCode  string   `json:"selectedCode"`  // Code snippet selected by user
	SourceFile    string   `json:"sourceFile"`    // File where selection was made
	MaxTokens     int      `json:"maxTokens"`     // Maximum tokens for context (default: 900000 for Qwen)
	MaxDepth      int      `json:"maxDepth"`      // Max depth for call stack traversal
	Language      string   `json:"language"`      // Programming language
}

// SmartContextResult contains the collected context
type SmartContextResult struct {
	Context         string                 `json:"context"`         // Full context string
	Files           []ContextFile          `json:"files"`           // Files included in context
	Symbols         []*domain.SymbolNode   `json:"symbols"`         // Symbols analyzed
	CallStack       *CallStackResult       `json:"callStack"`       // Call stack analysis
	TokenEstimate   int                    `json:"tokenEstimate"`   // Estimated token count
	TruncatedFiles  []string               `json:"truncatedFiles"`  // Files that were truncated
	ExcludedFiles   []string               `json:"excludedFiles"`   // Files excluded due to token limit
	RelevanceScores map[string]float64     `json:"relevanceScores"` // File relevance scores
}

// ContextFile represents a file in the context
type ContextFile struct {
	Path      string  `json:"path"`
	Content   string  `json:"content"`
	Tokens    int     `json:"tokens"`
	Relevance float64 `json:"relevance"`
	Reason    string  `json:"reason"` // Why this file was included
}

// NewSmartContextService creates a new smart context service
func NewSmartContextService(
	log domain.Logger,
	fileReader domain.FileContentReader,
	symbolGraphSvc *SymbolGraphService,
	callStackAnalyzer CallStackAnalyzerInterface,
) *SmartContextService {
	return &SmartContextService{
		log:              log,
		fileReader:       fileReader,
		symbolGraphSvc:   symbolGraphSvc,
		callStackAnalyzer: callStackAnalyzer,
	}
}

// CollectContext collects smart context based on the request
func (s *SmartContextService) CollectContext(ctx context.Context, req SmartContextRequest) (*SmartContextResult, error) {
	s.log.Info(fmt.Sprintf("Collecting smart context for task: %s", truncateString(req.Task, 50)))

	// Set defaults
	if req.MaxTokens == 0 {
		req.MaxTokens = 900000 // Leave room for response in 1M context
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}
	if req.Language == "" {
		req.Language = s.detectLanguage(req.ProjectRoot, req.SelectedFiles)
	}

	result := &SmartContextResult{
		Files:           make([]ContextFile, 0),
		Symbols:         make([]*domain.SymbolNode, 0),
		TruncatedFiles:  make([]string, 0),
		ExcludedFiles:   make([]string, 0),
		RelevanceScores: make(map[string]float64),
	}

	// Step 1: Analyze selected code to find symbols
	var callStackResult *CallStackResult
	if req.SelectedCode != "" && req.SourceFile != "" {
		symbols := s.extractSymbolsFromCode(req.SelectedCode)
		for _, symbolName := range symbols {
			csr, err := s.callStackAnalyzer.AnalyzeCallStack(ctx, req.ProjectRoot, req.SourceFile, symbolName, req.MaxDepth)
			if err != nil {
				s.log.Warning(fmt.Sprintf("Failed to analyze call stack for %s: %v", symbolName, err))
				continue
			}
			if callStackResult == nil {
				callStackResult = csr
			} else {
				// Merge results
				callStackResult.Callers = append(callStackResult.Callers, csr.Callers...)
				callStackResult.Callees = append(callStackResult.Callees, csr.Callees...)
				callStackResult.Dependencies = append(callStackResult.Dependencies, csr.Dependencies...)
				callStackResult.RelatedFiles = append(callStackResult.RelatedFiles, csr.RelatedFiles...)
			}
		}
	}
	result.CallStack = callStackResult

	// Step 2: Collect files with relevance scores
	filesToInclude := s.collectRelevantFiles(ctx, req, callStackResult)

	// Step 3: Read file contents and build context
	currentTokens := 0
	var contextBuilder strings.Builder

	// Add task description
	taskSection := fmt.Sprintf("# Task\n%s\n\n", req.Task)
	contextBuilder.WriteString(taskSection)
	currentTokens += s.estimateTokens(taskSection)

	// Add selected code if present
	if req.SelectedCode != "" {
		selectedSection := fmt.Sprintf("# Selected Code (from %s)\n```\n%s\n```\n\n", req.SourceFile, req.SelectedCode)
		contextBuilder.WriteString(selectedSection)
		currentTokens += s.estimateTokens(selectedSection)
	}

	// Sort files by relevance
	sort.Slice(filesToInclude, func(i, j int) bool {
		return filesToInclude[i].Relevance > filesToInclude[j].Relevance
	})

	// Add files to context
	contextBuilder.WriteString("# Project Files\n\n")
	currentTokens += s.estimateTokens("# Project Files\n\n")

	for _, file := range filesToInclude {
		// Read file content
		contents, err := s.fileReader.ReadContents(ctx, []string{file.Path}, req.ProjectRoot, nil)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to read file %s: %v", file.Path, err))
			continue
		}

		content, ok := contents[file.Path]
		if !ok {
			continue
		}

		fileTokens := s.estimateTokens(content)
		fileHeader := fmt.Sprintf("## %s\n```%s\n", file.Path, s.getFileExtension(file.Path))
		fileFooter := "\n```\n\n"
		headerTokens := s.estimateTokens(fileHeader + fileFooter)

		// Check if we can fit this file
		if currentTokens+fileTokens+headerTokens > req.MaxTokens {
			// Try to truncate
			availableTokens := req.MaxTokens - currentTokens - headerTokens - 100 // Buffer
			if availableTokens > 500 { // Minimum useful content
				truncatedContent := s.truncateToTokens(content, availableTokens)
				file.Content = truncatedContent
				file.Tokens = s.estimateTokens(truncatedContent)
				result.TruncatedFiles = append(result.TruncatedFiles, file.Path)

				contextBuilder.WriteString(fileHeader)
				contextBuilder.WriteString(truncatedContent)
				contextBuilder.WriteString("\n// ... truncated ...\n")
				contextBuilder.WriteString(fileFooter)
				currentTokens += file.Tokens + headerTokens
			} else {
				result.ExcludedFiles = append(result.ExcludedFiles, file.Path)
				continue
			}
		} else {
			file.Content = content
			file.Tokens = fileTokens

			contextBuilder.WriteString(fileHeader)
			contextBuilder.WriteString(content)
			contextBuilder.WriteString(fileFooter)
			currentTokens += fileTokens + headerTokens
		}

		result.Files = append(result.Files, file)
		result.RelevanceScores[file.Path] = file.Relevance
	}

	result.Context = contextBuilder.String()
	result.TokenEstimate = currentTokens

	// Collect symbols from call stack
	if callStackResult != nil {
		result.Symbols = append(result.Symbols, callStackResult.Callers...)
		result.Symbols = append(result.Symbols, callStackResult.Callees...)
		result.Symbols = append(result.Symbols, callStackResult.Dependencies...)
	}

	s.log.Info(fmt.Sprintf("Smart context collected: %d files, ~%d tokens", len(result.Files), result.TokenEstimate))

	return result, nil
}

// collectRelevantFiles collects files relevant to the task
func (s *SmartContextService) collectRelevantFiles(
	ctx context.Context,
	req SmartContextRequest,
	callStack *CallStackResult,
) []ContextFile {
	files := make([]ContextFile, 0)
	seen := make(map[string]bool)

	// Add explicitly selected files with highest relevance
	for _, path := range req.SelectedFiles {
		if !seen[path] {
			seen[path] = true
			files = append(files, ContextFile{
				Path:      path,
				Relevance: 1.0,
				Reason:    "explicitly selected",
			})
		}
	}

	// Add source file
	if req.SourceFile != "" && !seen[req.SourceFile] {
		seen[req.SourceFile] = true
		files = append(files, ContextFile{
			Path:      req.SourceFile,
			Relevance: 0.95,
			Reason:    "source of selected code",
		})
	}

	// Add files from call stack analysis
	if callStack != nil {
		for _, path := range callStack.RelatedFiles {
			if !seen[path] {
				seen[path] = true
				files = append(files, ContextFile{
					Path:      path,
					Relevance: 0.8,
					Reason:    "call stack related",
				})
			}
		}

		// Add files containing callers
		for _, symbol := range callStack.Callers {
			if !seen[symbol.Path] {
				seen[symbol.Path] = true
				files = append(files, ContextFile{
					Path:      symbol.Path,
					Relevance: 0.7,
					Reason:    fmt.Sprintf("contains caller: %s", symbol.Name),
				})
			}
		}

		// Add files containing callees
		for _, symbol := range callStack.Callees {
			if !seen[symbol.Path] {
				seen[symbol.Path] = true
				files = append(files, ContextFile{
					Path:      symbol.Path,
					Relevance: 0.75,
					Reason:    fmt.Sprintf("contains callee: %s", symbol.Name),
				})
			}
		}

		// Add files containing dependencies
		for _, symbol := range callStack.Dependencies {
			if !seen[symbol.Path] {
				seen[symbol.Path] = true
				files = append(files, ContextFile{
					Path:      symbol.Path,
					Relevance: 0.6,
					Reason:    fmt.Sprintf("contains dependency: %s", symbol.Name),
				})
			}
		}
	}

	// Add files matching task keywords
	taskKeywords := s.extractKeywords(req.Task)
	for _, kw := range taskKeywords {
		// This would ideally search the project, but for now we rely on call stack
		_ = kw
	}

	return files
}

// extractSymbolsFromCode extracts symbol names from code snippet
func (s *SmartContextService) extractSymbolsFromCode(code string) []string {
	symbols := make([]string, 0)
	seen := make(map[string]bool)

	// Simple extraction: look for function/method names
	// This is a basic implementation - could be enhanced with proper parsing
	words := strings.FieldsFunc(code, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_')
	})

	for _, word := range words {
		// Skip common keywords and short words
		if len(word) < 3 || isCommonKeyword(word) {
			continue
		}
		// Look for PascalCase or camelCase (likely function/type names)
		if (word[0] >= 'A' && word[0] <= 'Z') || strings.Contains(word, "_") {
			if !seen[word] {
				seen[word] = true
				symbols = append(symbols, word)
			}
		}
	}

	return symbols
}

// extractKeywords extracts keywords from task description
func (s *SmartContextService) extractKeywords(task string) []string {
	keywords := make([]string, 0)
	words := strings.Fields(strings.ToLower(task))

	for _, word := range words {
		// Skip common words
		if len(word) < 4 || isStopWord(word) {
			continue
		}
		keywords = append(keywords, word)
	}

	return keywords
}

// detectLanguage detects the primary language of the project
func (s *SmartContextService) detectLanguage(projectRoot string, files []string) string {
	extCount := make(map[string]int)

	for _, f := range files {
		ext := filepath.Ext(f)
		extCount[ext]++
	}

	// Map extensions to languages
	langMap := map[string]string{
		".go":   "go",
		".ts":   "typescript",
		".tsx":  "typescript",
		".js":   "javascript",
		".jsx":  "javascript",
		".py":   "python",
		".java": "java",
		".rs":   "rust",
		".cpp":  "cpp",
		".c":    "c",
	}

	maxCount := 0
	detectedLang := "go" // default

	for ext, count := range extCount {
		if count > maxCount {
			if lang, ok := langMap[ext]; ok {
				maxCount = count
				detectedLang = lang
			}
		}
	}

	return detectedLang
}

// estimateTokens estimates token count for text
func (s *SmartContextService) estimateTokens(text string) int {
	// Approximate: 1 token ≈ 4 characters
	return len(text) / 4
}

// truncateToTokens truncates text to fit within token limit
func (s *SmartContextService) truncateToTokens(text string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}

	// Try to truncate at a line boundary
	truncated := text[:maxChars]
	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > maxChars/2 {
		return truncated[:lastNewline]
	}

	return truncated
}

// getFileExtension returns the file extension for syntax highlighting
func (s *SmartContextService) getFileExtension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return ext[1:] // Remove the dot
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func isCommonKeyword(word string) bool {
	keywords := map[string]bool{
		"func": true, "return": true, "if": true, "else": true,
		"for": true, "range": true, "var": true, "const": true,
		"type": true, "struct": true, "interface": true, "package": true,
		"import": true, "nil": true, "true": true, "false": true,
		"string": true, "int": true, "bool": true, "error": true,
		"make": true, "new": true, "len": true, "append": true,
		"function": true, "class": true, "this": true, "self": true,
		"def": true, "async": true, "await": true, "export": true,
	}
	return keywords[strings.ToLower(word)]
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true,
		"this": true, "that": true, "from": true, "have": true,
		"will": true, "should": true, "would": true, "could": true,
		"need": true, "want": true, "like": true, "make": true,
		"добавить": true, "сделать": true, "создать": true, "нужно": true,
	}
	return stopWords[word]
}
