package git

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ContextBuilder builds context from git history
type ContextBuilder struct {
	projectRoot string
}

// NewContextBuilder creates a new git context builder
func NewContextBuilder(projectRoot string) *ContextBuilder {
	return &ContextBuilder{projectRoot: projectRoot}
}

// RecentChange represents a recently changed file
type RecentChange struct {
	FilePath    string
	ChangeCount int
	LastChanged time.Time
	Authors     []string
}

// GetRecentChanges returns files changed recently, sorted by relevance
func (b *ContextBuilder) GetRecentChanges(since string, pathFilter string) ([]RecentChange, error) {
	if since == "" {
		since = "1 week ago"
	}

	cmdArgs := []string{"log", "--since=" + since, "--name-only", "--format=%H|%an|%at"}
	if pathFilter != "" {
		cmdArgs = append(cmdArgs, "--", pathFilter)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = b.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse output
	changes := make(map[string]*RecentChange)
	lines := strings.Split(string(output), "\n")

	var currentAuthor string
	var currentTime time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				currentAuthor = parts[1]
				// Parse unix timestamp
				var ts int64
				if _, err := parseUnixTime(parts[2], &ts); err == nil {
					currentTime = time.Unix(ts, 0)
				}
			}
			continue
		}

		// It's a file path
		filePath := line
		if change, exists := changes[filePath]; exists {
			change.ChangeCount++
			if currentTime.After(change.LastChanged) {
				change.LastChanged = currentTime
			}
			if !containsString(change.Authors, currentAuthor) {
				change.Authors = append(change.Authors, currentAuthor)
			}
		} else {
			changes[filePath] = &RecentChange{
				FilePath:    filePath,
				ChangeCount: 1,
				LastChanged: currentTime,
				Authors:     []string{currentAuthor},
			}
		}
	}

	// Convert to slice and sort
	result := make([]RecentChange, 0, len(changes))
	for _, c := range changes {
		result = append(result, *c)
	}

	sort.Slice(result, func(i, j int) bool {
		// Sort by change count descending, then by last changed descending
		if result[i].ChangeCount != result[j].ChangeCount {
			return result[i].ChangeCount > result[j].ChangeCount
		}
		return result[i].LastChanged.After(result[j].LastChanged)
	})

	return result, nil
}


// GetRelatedByAuthor returns files frequently changed by the same author
func (b *ContextBuilder) GetRelatedByAuthor(filePath string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get authors who changed this file
	cmd := exec.Command("git", "log", "--format=%an", "--", filePath)
	cmd.Dir = b.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	authors := make(map[string]int)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			authors[line]++
		}
	}

	// Find top author
	var topAuthor string
	var maxCount int
	for author, count := range authors {
		if count > maxCount {
			topAuthor = author
			maxCount = count
		}
	}

	if topAuthor == "" {
		return nil, nil
	}

	// Get files changed by this author
	cmd = exec.Command("git", "log", "--author="+topAuthor, "--name-only", "--format=", "-n", "100")
	cmd.Dir = b.projectRoot
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	fileCounts := make(map[string]int)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != filePath {
			fileCounts[line]++
		}
	}

	// Sort by count
	type fileCount struct {
		path  string
		count int
	}
	var sorted []fileCount
	for path, count := range fileCounts {
		sorted = append(sorted, fileCount{path, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	result := make([]string, 0, limit)
	for i, fc := range sorted {
		if i >= limit {
			break
		}
		result = append(result, fc.path)
	}

	return result, nil
}

// GetCoChangedFiles returns files that are often changed together with the given file
func (b *ContextBuilder) GetCoChangedFiles(filePath string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get commits that changed this file
	cmd := exec.Command("git", "log", "--format=%H", "-n", "50", "--", filePath)
	cmd.Dir = b.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	fileCounts := make(map[string]int)

	for _, commit := range commits {
		if commit == "" {
			continue
		}
		// Get files changed in this commit
		cmd = exec.Command("git", "show", "--name-only", "--format=", commit)
		cmd.Dir = b.projectRoot
		output, err = cmd.Output()
		if err != nil {
			continue
		}

		for _, file := range strings.Split(string(output), "\n") {
			file = strings.TrimSpace(file)
			if file != "" && file != filePath {
				fileCounts[file]++
			}
		}
	}

	// Sort by count
	type fileCount struct {
		path  string
		count int
	}
	var sorted []fileCount
	for path, count := range fileCounts {
		sorted = append(sorted, fileCount{path, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	result := make([]string, 0, limit)
	for i, fc := range sorted {
		if i >= limit {
			break
		}
		result = append(result, fc.path)
	}

	return result, nil
}

// SuggestContextFiles suggests files to include in context based on git history
func (b *ContextBuilder) SuggestContextFiles(taskDescription string, currentFiles []string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	suggestions := make(map[string]int)

	// 1. Add co-changed files for current files
	for _, file := range currentFiles {
		coChanged, err := b.GetCoChangedFiles(file, 5)
		if err == nil {
			for _, f := range coChanged {
				suggestions[f] += 3
			}
		}
	}

	// 2. Add recently changed files in same directories
	for _, file := range currentFiles {
		dir := filepath.Dir(file)
		recent, err := b.GetRecentChanges("2 weeks ago", dir)
		if err == nil {
			for _, r := range recent {
				if r.FilePath != file {
					suggestions[r.FilePath] += r.ChangeCount
				}
			}
		}
	}

	// 3. Search commits by task keywords
	keywords := extractKeywords(taskDescription)
	for _, kw := range keywords {
		cmd := exec.Command("git", "log", "--grep="+kw, "-i", "--name-only", "--format=", "-n", "20")
		cmd.Dir = b.projectRoot
		output, _ := cmd.Output()
		for _, file := range strings.Split(string(output), "\n") {
			file = strings.TrimSpace(file)
			if file != "" {
				suggestions[file] += 2
			}
		}
	}

	// Remove current files from suggestions
	for _, f := range currentFiles {
		delete(suggestions, f)
	}

	// Sort by score
	type suggestion struct {
		path  string
		score int
	}
	var sorted []suggestion
	for path, score := range suggestions {
		sorted = append(sorted, suggestion{path, score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	result := make([]string, 0, limit)
	for i, s := range sorted {
		if i >= limit {
			break
		}
		result = append(result, s.path)
	}

	return result, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func parseUnixTime(s string, ts *int64) (int, error) {
	*ts = 0 // Reset to avoid accumulation from previous calls
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			*ts = *ts*10 + int64(c-'0')
			n++
		}
	}
	return n, nil
}

func extractKeywords(text string) []string {
	// Simple keyword extraction - split by spaces and filter
	words := strings.Fields(strings.ToLower(text))
	var keywords []string
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"to": true, "of": true, "in": true, "for": true, "on": true,
		"with": true, "as": true, "at": true, "by": true, "from": true,
		"this": true, "that": true, "it": true, "i": true, "we": true,
		"add": true, "fix": true, "update": true, "change": true,
	}

	for _, w := range words {
		if len(w) > 3 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}

	if len(keywords) > 5 {
		keywords = keywords[:5]
	}
	return keywords
}
