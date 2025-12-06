package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
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
	var selectedFiles []domain.ScoredFile
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

// scoreFilesByRelevance scores files based on keyword matching and path analysis
func (ca *ContextAnalyzerImpl) scoreFilesByRelevance(files []*domain.FileNode, keywords []string, task string) []ScoredFile {
	// Flatten tree structure first
	flatFiles := ca.flattenFiles(files)

	var scored []ScoredFile
	taskLower := strings.ToLower(task)

	for _, file := range flatFiles {
		if file.IsDir {
			continue
		}

		score := 0.0
		pathLower := strings.ToLower(file.RelPath)
		nameLower := strings.ToLower(file.Name)
		ext := strings.ToLower(filepath.Ext(file.Name))

		// Base score for code files
		codeExts := map[string]bool{
			".go": true, ".ts": true, ".js": true, ".vue": true, ".tsx": true, ".jsx": true,
			".py": true, ".java": true, ".rs": true, ".cpp": true, ".c": true, ".h": true,
		}
		if codeExts[ext] {
			score += 0.1 // Base score for any code file
		}

		// Score by keyword matches in path/name
		for _, kw := range keywords {
			kwLower := strings.ToLower(kw)
			if len(kwLower) < 2 {
				continue
			}
			if strings.Contains(nameLower, kwLower) {
				score += 0.4 // Strong match in filename
			} else if strings.Contains(pathLower, kwLower) {
				score += 0.2 // Match in path
			}
		}

		// Handler/Controller files
		if strings.Contains(nameLower, "handler") || strings.Contains(nameLower, "controller") {
			if strings.Contains(taskLower, "api") || strings.Contains(taskLower, "endpoint") ||
				strings.Contains(taskLower, "route") || strings.Contains(taskLower, "http") {
				score += 0.3
			}
		}

		// Auth-related
		if strings.Contains(nameLower, "auth") || strings.Contains(nameLower, "login") ||
			strings.Contains(nameLower, "jwt") || strings.Contains(nameLower, "token") {
			if strings.Contains(taskLower, "auth") || strings.Contains(taskLower, "login") ||
				strings.Contains(taskLower, "авториз") || strings.Contains(taskLower, "jwt") {
				score += 0.5
			}
		}

		// User-related
		if strings.Contains(nameLower, "user") || strings.Contains(nameLower, "account") {
			if strings.Contains(taskLower, "user") || strings.Contains(taskLower, "пользовател") ||
				strings.Contains(taskLower, "account") || strings.Contains(taskLower, "profile") {
				score += 0.4
			}
		}

		// Domain/Model files
		if strings.Contains(pathLower, "domain") || strings.Contains(pathLower, "model") ||
			strings.Contains(pathLower, "entity") {
			score += 0.1
		}

		// Service/Application layer
		if strings.Contains(pathLower, "service") || strings.Contains(pathLower, "application") {
			score += 0.1
		}

		// Store/State files (frontend)
		if strings.Contains(nameLower, "store") || strings.Contains(nameLower, "state") {
			score += 0.15
		}

		// Component files
		if ext == ".vue" || ext == ".tsx" || ext == ".jsx" {
			if strings.Contains(taskLower, "ui") || strings.Contains(taskLower, "компонент") ||
				strings.Contains(taskLower, "component") || strings.Contains(taskLower, "страниц") {
				score += 0.2
			}
		}

		// Config files
		if strings.Contains(nameLower, "config") || strings.Contains(nameLower, "setting") {
			if strings.Contains(taskLower, "config") || strings.Contains(taskLower, "настройк") {
				score += 0.3
			}
		}

		// Test files - lower priority unless task is about tests
		if strings.Contains(nameLower, "test") || strings.Contains(nameLower, "spec") {
			if strings.Contains(taskLower, "test") || strings.Contains(taskLower, "тест") {
				score += 0.3
			} else {
				score -= 0.2
			}
		}

		// Penalize generated/vendor files
		if strings.Contains(pathLower, "node_modules") || strings.Contains(pathLower, "vendor") ||
			strings.Contains(pathLower, "dist") || strings.Contains(pathLower, ".git") {
			score = 0
		}

		// Boost main entry points
		if nameLower == "main.go" || nameLower == "app.go" || nameLower == "index.ts" ||
			nameLower == "main.ts" || nameLower == "app.vue" {
			score += 0.15
		}

		if score > 0 {
			scored = append(scored, ScoredFile{File: file, Score: score})
		}
	}

	return scored
}

// refineWithAI uses AI to refine file selection
func (ca *ContextAnalyzerImpl) refineWithAI(ctx context.Context, task string, candidates []*domain.FileNode, allFiles []*domain.FileNode) ([]*domain.FileNode, error) {
	if ca.aiService == nil {
		return candidates, nil
	}

	// Build file list for AI
	var fileList []string
	for _, f := range candidates {
		fileList = append(fileList, f.RelPath)
	}

	prompt := fmt.Sprintf(`Given the task: "%s"

And these candidate files:
%s

Return a JSON array of the most relevant file paths (max 10) for this task.
Only return the JSON array, nothing else.
Example: ["path/to/file1.go", "path/to/file2.ts"]`, task, strings.Join(fileList, "\n"))

	response, err := ca.aiService.GenerateCode(ctx, "You are a code analysis assistant. Return only valid JSON.", prompt)
	if err != nil {
		return candidates, err
	}

	// Parse response
	var selectedPaths []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(response)), &selectedPaths); err != nil {
		// Try to extract JSON from response
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(response[start:end+1]), &selectedPaths); err != nil {
				return candidates, nil
			}
		} else {
			return candidates, nil
		}
	}

	// Map paths to files
	pathToFile := make(map[string]*domain.FileNode)
	for _, f := range candidates {
		pathToFile[f.RelPath] = f
	}

	var refined []*domain.FileNode
	for _, p := range selectedPaths {
		if f, ok := pathToFile[p]; ok {
			refined = append(refined, f)
		}
	}

	if len(refined) == 0 {
		return candidates, nil
	}

	return refined, nil
}

// detectTaskType detects task type from description
func (ca *ContextAnalyzerImpl) detectTaskType(task string) string {
	taskLower := strings.ToLower(task)

	if strings.Contains(taskLower, "bug") || strings.Contains(taskLower, "fix") ||
		strings.Contains(taskLower, "ошибк") || strings.Contains(taskLower, "исправ") {
		return "bug_fix"
	}
	if strings.Contains(taskLower, "test") || strings.Contains(taskLower, "тест") {
		return "test"
	}
	if strings.Contains(taskLower, "refactor") || strings.Contains(taskLower, "рефактор") ||
		strings.Contains(taskLower, "cleanup") || strings.Contains(taskLower, "очист") {
		return "refactor"
	}
	if strings.Contains(taskLower, "doc") || strings.Contains(taskLower, "документ") {
		return "documentation"
	}

	return "feature"
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
	var keywords []string

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
	var filePaths []string
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
