package application

import (
	"shotgun_code/domain"
	"sort"
	"strings"
)

// AnalysisService handles task analysis and file suggestions
type AnalysisService struct{}

// NewAnalysisService creates a new AnalysisService instance
func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

// SuggestFiles analyzes task description and suggests relevant files
// This is a simple keyword-based implementation (not reading file contents)
func (s *AnalysisService) SuggestFiles(taskDescription string, files []*domain.FileNode) []domain.SuggestedFile {
	if taskDescription == "" || len(files) == 0 {
		return []domain.SuggestedFile{}
	}

	// Extract keywords from task description
	keywords := extractKeywords(taskDescription)
	if len(keywords) == 0 {
		return []domain.SuggestedFile{}
	}

	// Flatten file tree
	flatFiles := flattenFileTree(files)

	// Score and filter files
	suggestions := []domain.SuggestedFile{}
	for _, file := range flatFiles {
		if file.IsDir {
			continue // Skip directories
		}

		score, matchedKeywords := scoreFile(file, keywords)
		if score > 0 {
			suggestions = append(suggestions, domain.SuggestedFile{
				Path:       file.Path,
				Reason:     buildReason(matchedKeywords),
				Confidence: score,
			})
		}
	}

	// Sort by confidence (descending)
	sortSuggestionsByConfidence(suggestions)

	// Return top suggestions (max 20)
	if len(suggestions) > 20 {
		suggestions = suggestions[:20]
	}

	return suggestions
}

// extractKeywords extracts meaningful keywords from task description
func extractKeywords(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Split by spaces and common separators
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == ';' || r == ':' || r == '\n' || r == '\t'
	})

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"be": true, "have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"i": true, "you": true, "we": true, "they": true, "it": true, "this": true,
		"that": true, "these": true, "those": true, "my": true, "your": true,
	}

	keywords := []string{}
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// flattenFileTree converts tree structure to flat list
func flattenFileTree(nodes []*domain.FileNode) []domain.FileNode {
	result := []domain.FileNode{}

	var traverse func([]*domain.FileNode)
	traverse = func(nodeList []*domain.FileNode) {
		for _, node := range nodeList {
			result = append(result, *node)
			if node.Children != nil {
				traverse(node.Children)
			}
		}
	}

	traverse(nodes)
	return result
}

// scoreFile calculates relevance score for a file based on keywords
func scoreFile(file domain.FileNode, keywords []string) (float64, []string) {
	fileLower := strings.ToLower(file.Path)
	fileName := strings.ToLower(file.Name)

	score := 0.0
	matched := []string{}

	for _, keyword := range keywords {
		// Check path
		if strings.Contains(fileLower, keyword) {
			score += 0.5
			matched = append(matched, keyword)
		}

		// Check filename (higher weight)
		if strings.Contains(fileName, keyword) {
			score += 1.0
			if !containsString(matched, keyword) {
				matched = append(matched, keyword)
			}
		}

		// Check file extension matches
		if strings.HasSuffix(fileName, "."+keyword) {
			score += 0.3
		}
	}

	// Normalize score to 0-1 range
	if score > 0 {
		score /= float64(len(keywords))
		if score > 1.0 {
			score = 1.0
		}
	}

	return score, matched
}

// buildReason creates a human-readable reason string
func buildReason(keywords []string) string {
	if len(keywords) == 0 {
		return "Matches search criteria"
	}
	return "Matches keywords: " + strings.Join(keywords, ", ")
}

// sortSuggestionsByConfidence sorts suggestions by confidence (descending)
func sortSuggestionsByConfidence(suggestions []domain.SuggestedFile) {
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})
}

// containsString checks if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
