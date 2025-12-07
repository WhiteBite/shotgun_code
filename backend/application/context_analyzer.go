package application

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
)

// File extension constants
const (
	extGo  = ".go"
	extTS  = ".ts"
	extJS  = ".js"
	extVue = ".vue"
	extTSX = ".tsx"
	extJSX = ".jsx"
)

// Task type constants
const (
	taskTypeFeature       = "feature"
	taskTypeBugFix        = "bug_fix"
	taskTypeTest          = "test"
	taskTypeRefactor      = "refactor"
	taskTypeDocumentation = "documentation"
)

// ContextAnalyzerImpl implements the ContextAnalyzer interface
type ContextAnalyzerImpl struct {
	logger    domain.Logger
	aiService *AIService
}

// NewContextAnalyzer creates a new ContextAnalyzer implementation
func NewContextAnalyzer(logger domain.Logger, aiService *AIService) *ContextAnalyzerImpl {
	return &ContextAnalyzerImpl{
		logger:    logger,
		aiService: aiService,
	}
}

// AnalyzeTaskAndCollectContext analyzes a task and automatically collects relevant context
func (ca *ContextAnalyzerImpl) AnalyzeTaskAndCollectContext(
	ctx context.Context,
	task string,
	allFiles []*domain.FileNode,
	rootDir string,
) (*domain.ContextAnalysisResult, error) {
	startTime := time.Now()
	ca.logger.Info(fmt.Sprintf("Starting smart context analysis: %s", task))

	// 1. Extract keywords from task
	keywords := ca.extractKeywords(task, "")
	ca.logger.Info(fmt.Sprintf("Extracted keywords: %v", keywords))

	// 2. Score all files by relevance
	ca.logger.Info(fmt.Sprintf("Input files count: %d", len(allFiles)))
	scoredFiles := ca.scoreFilesByRelevance(allFiles, keywords, task)
	ca.logger.Info(fmt.Sprintf("Scored files count: %d", len(scoredFiles)))

	// 3. Sort by score and take top files
	sort.Slice(scoredFiles, func(i, j int) bool {
		return scoredFiles[i].Score > scoredFiles[j].Score
	})

	// Take top 15 files with score > 0.1
	selectedFiles := make([]domain.ScoredFile, 0, 15)
	for i, sf := range scoredFiles {
		if i >= 15 || sf.Score < 0.1 {
			break
		}
		selectedFiles = append(selectedFiles, domain.ScoredFile{
			RelPath:   sf.File.RelPath,
			Name:      sf.File.Name,
			Size:      sf.File.Size,
			Relevance: sf.Score,
		})
	}

	// If no files found with keywords, take top code files
	if len(selectedFiles) == 0 && len(scoredFiles) > 0 {
		for i, sf := range scoredFiles {
			if i >= 10 {
				break
			}
			selectedFiles = append(selectedFiles, domain.ScoredFile{
				RelPath:   sf.File.RelPath,
				Name:      sf.File.Name,
				Size:      sf.File.Size,
				Relevance: sf.Score,
			})
		}
	}

	// 4. Calculate estimated tokens
	estimatedTokens := 0
	for _, f := range selectedFiles {
		// Rough estimate: 50 tokens per KB
		estimatedTokens += int(f.Size / 20)
	}

	// 5. Form result with relevance scores
	result := &domain.ContextAnalysisResult{
		Task:            task,
		TaskType:        ca.detectTaskType(task),
		Priority:        "normal",
		SelectedFiles:   selectedFiles,
		DependencyFiles: []*domain.FileNode{},
		Context:         "",
		AnalysisTime:    time.Since(startTime),
		Recommendations: []string{fmt.Sprintf("Found %d relevant files", len(selectedFiles))},
		EstimatedTokens: estimatedTokens,
		Confidence:      ca.calculateConfidenceFromScores(scoredFiles, len(selectedFiles)),
	}

	ca.logger.Info(fmt.Sprintf("Smart analysis completed in %v, selected %d files", result.AnalysisTime, len(selectedFiles)))
	return result, nil
}

// ScoredFile holds a file with its relevance score
type ScoredFile struct {
	File  *domain.FileNode
	Score float64
}

// flattenFiles recursively flattens the file tree into a list of files
func (ca *ContextAnalyzerImpl) flattenFiles(nodes []*domain.FileNode) []*domain.FileNode {
	var result []*domain.FileNode
	for _, node := range nodes {
		if node.IsDir {
			if node.Children != nil {
				result = append(result, ca.flattenFiles(node.Children)...)
			}
		} else {
			result = append(result, node)
		}
	}
	return result
}

// codeExtensions returns a map of code file extensions
func codeExtensions() map[string]bool {
	return map[string]bool{
		extGo: true, extTS: true, extJS: true, extVue: true, extTSX: true, extJSX: true,
		".py": true, ".java": true, ".rs": true, ".cpp": true, ".c": true, ".h": true,
	}
}

// scoreKeywordMatches scores based on keyword matches in path/name
func scoreKeywordMatches(nameLower, pathLower string, keywords []string) float64 {
	score := 0.0
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		if len(kwLower) < 2 {
			continue
		}
		if strings.Contains(nameLower, kwLower) {
			score += 0.4
		} else if strings.Contains(pathLower, kwLower) {
			score += 0.2
		}
	}
	return score
}

// fileTypeRule defines a scoring rule for file types
type fileTypeRule struct {
	namePatterns []string
	taskPatterns []string
	score        float64
}

var fileTypeRules = []fileTypeRule{
	{[]string{"handler", "controller"}, []string{"api", "endpoint", "route", "http"}, 0.3},
	{[]string{"auth", "login", "jwt", "token"}, []string{"auth", "login", "авториз", "jwt"}, 0.5},
	{[]string{"user", "account"}, []string{"user", "пользовател", "account", "profile"}, 0.4},
	{[]string{"config", "setting"}, []string{"config", "настройк"}, 0.3},
}

// containsAny checks if s contains any of the patterns
func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// scoreByFileType scores based on file type and task context
func scoreByFileType(nameLower, pathLower, ext, taskLower string) float64 {
	score := 0.0

	for _, rule := range fileTypeRules {
		if containsAny(nameLower, rule.namePatterns) && containsAny(taskLower, rule.taskPatterns) {
			score += rule.score
		}
	}

	// Component files
	if ext == extVue || ext == extTSX || ext == extJSX {
		if containsAny(taskLower, []string{"ui", "компонент", "component", "страниц"}) {
			score += 0.2
		}
	}

	// Test files
	if strings.Contains(nameLower, "test") || strings.Contains(nameLower, "spec") {
		if containsAny(taskLower, []string{"test", "тест"}) {
			score += 0.3
		} else {
			score -= 0.2
		}
	}

	return score
}

// scoreByPath scores based on path patterns
func scoreByPath(pathLower string) float64 {
	score := 0.0
	if strings.Contains(pathLower, "domain") || strings.Contains(pathLower, "model") ||
		strings.Contains(pathLower, "entity") {
		score += 0.1
	}
	if strings.Contains(pathLower, "service") || strings.Contains(pathLower, "application") {
		score += 0.1
	}
	return score
}

// scoreByName scores based on file name patterns
func scoreByName(nameLower string) float64 {
	score := 0.0
	if strings.Contains(nameLower, "store") || strings.Contains(nameLower, "state") {
		score += 0.15
	}
	if nameLower == "main.go" || nameLower == "app.go" || nameLower == "index.ts" ||
		nameLower == "main.ts" || nameLower == "app.vue" {
		score += 0.15
	}
	return score
}

// isExcludedPath checks if path should be excluded
func isExcludedPath(pathLower string) bool {
	return strings.Contains(pathLower, "node_modules") || strings.Contains(pathLower, "vendor") ||
		strings.Contains(pathLower, "dist") || strings.Contains(pathLower, ".git")
}

// scoreFilesByRelevance scores files based on keyword matching and path analysis
func (ca *ContextAnalyzerImpl) scoreFilesByRelevance(files []*domain.FileNode, keywords []string, task string) []ScoredFile {
	flatFiles := ca.flattenFiles(files)
	var scored []ScoredFile
	taskLower := strings.ToLower(task)
	codeExts := codeExtensions()

	for _, file := range flatFiles {
		if file.IsDir {
			continue
		}

		pathLower := strings.ToLower(file.RelPath)
		nameLower := strings.ToLower(file.Name)
		ext := strings.ToLower(filepath.Ext(file.Name))

		if isExcludedPath(pathLower) {
			continue
		}

		score := 0.0
		if codeExts[ext] {
			score += 0.1
		}

		score += scoreKeywordMatches(nameLower, pathLower, keywords)
		score += scoreByFileType(nameLower, pathLower, ext, taskLower)
		score += scoreByPath(pathLower)
		score += scoreByName(nameLower)

		if score > 0 {
			scored = append(scored, ScoredFile{File: file, Score: score})
		}
	}

	return scored
}

// detectTaskType detects task type from description
func (ca *ContextAnalyzerImpl) detectTaskType(task string) string {
	taskLower := strings.ToLower(task)

	if strings.Contains(taskLower, "bug") || strings.Contains(taskLower, "fix") ||
		strings.Contains(taskLower, "ошибк") || strings.Contains(taskLower, "исправ") {
		return taskTypeBugFix
	}
	if strings.Contains(taskLower, "test") || strings.Contains(taskLower, "тест") {
		return taskTypeTest
	}
	if strings.Contains(taskLower, "refactor") || strings.Contains(taskLower, "рефактор") ||
		strings.Contains(taskLower, "cleanup") || strings.Contains(taskLower, "очист") {
		return taskTypeRefactor
	}
	if strings.Contains(taskLower, "doc") || strings.Contains(taskLower, "документ") {
		return taskTypeDocumentation
	}

	return taskTypeFeature
}

// calculateConfidenceFromScores calculates confidence based on score distribution
func (ca *ContextAnalyzerImpl) calculateConfidenceFromScores(scored []ScoredFile, selectedCount int) float64 {
	if len(scored) == 0 || selectedCount == 0 {
		return 0.3
	}

	// Higher confidence if top scores are significantly higher than others
	if len(scored) >= selectedCount {
		topScore := scored[0].Score
		if topScore > 0.7 {
			return 0.9
		}
		if topScore > 0.5 {
			return 0.75
		}
		if topScore > 0.3 {
			return 0.6
		}
	}

	return 0.5
}

// extractKeywords extracts keywords from task description
func (ca *ContextAnalyzerImpl) extractKeywords(task, _ string) []string {
	taskLower := strings.ToLower(task)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true, "shall": true,
		"this": true, "that": true, "these": true, "those": true, "it": true,
		"i": true, "you": true, "he": true, "she": true, "we": true, "they": true,
		"как": true, "что": true, "это": true, "для": true, "на": true, "из": true,
		"нужно": true, "сделать": true, "добавить": true, "создать": true,
	}

	words := strings.Fields(taskLower)
	keywords := make([]string, 0, len(words))

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}«»")
		if len(word) < 2 {
			continue
		}
		if stopWords[word] {
			continue
		}
		keywords = append(keywords, word)
	}

	return keywords
}

// SuggestFiles implements the domain.ContextAnalyzer interface
func (ca *ContextAnalyzerImpl) SuggestFiles(ctx context.Context, task string, allFiles []*domain.FileNode) ([]string, error) {
	ca.logger.Info(fmt.Sprintf("Starting file suggestion for task: %s", task))

	// Use the same logic as AnalyzeTaskAndCollectContext
	keywords := ca.extractKeywords(task, "")
	scoredFiles := ca.scoreFilesByRelevance(allFiles, keywords, task)

	// Sort by score
	sort.Slice(scoredFiles, func(i, j int) bool {
		return scoredFiles[i].Score > scoredFiles[j].Score
	})

	// Take top 10 files with score > 0.1
	filePaths := make([]string, 0, 10)
	for i, sf := range scoredFiles {
		if i >= 10 || sf.Score < 0.1 {
			break
		}
		filePaths = append(filePaths, sf.File.RelPath)
	}

	// Fallback: if no matches, take any code files
	if len(filePaths) == 0 && len(scoredFiles) > 0 {
		for i, sf := range scoredFiles {
			if i >= 5 {
				break
			}
			filePaths = append(filePaths, sf.File.RelPath)
		}
	}

	ca.logger.Info(fmt.Sprintf("Suggested %d files for the task", len(filePaths)))
	return filePaths, nil
}
