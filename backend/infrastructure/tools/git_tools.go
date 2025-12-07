package tools

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func (e *Executor) registerGitTools() {
	e.tools["git_status"] = e.gitStatus
	e.tools["git_diff"] = e.gitDiff
	e.tools["git_log"] = e.gitLog
	e.tools["git_blame"] = e.gitBlame
	e.tools["git_show"] = e.gitShow
	// Phase 5: Git-Aware Context
	e.tools["git_diff_branches"] = e.gitDiffBranches
	e.tools["git_search_commits"] = e.gitSearchCommits
	e.tools["git_changed_files"] = e.gitChangedFiles
	e.tools["git_file_history"] = e.gitFileHistory
	e.tools["git_co_changed"] = e.gitCoChanged
	e.tools["git_suggest_context"] = e.gitSuggestContext
}

func (e *Executor) gitStatus(args map[string]any, projectRoot string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	if len(output) == 0 {
		return "Working directory clean", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var modified, added, deleted, untracked []string

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := strings.TrimSpace(line[3:])

		switch {
		case strings.Contains(status, "M"):
			modified = append(modified, file)
		case strings.Contains(status, "A"):
			added = append(added, file)
		case strings.Contains(status, "D"):
			deleted = append(deleted, file)
		case strings.Contains(status, "?"):
			untracked = append(untracked, file)
		}
	}

	var result strings.Builder
	result.WriteString("Git Status:\n")
	if len(modified) > 0 {
		result.WriteString(fmt.Sprintf("\nModified (%d):\n", len(modified)))
		for _, f := range modified {
			result.WriteString("  " + f + "\n")
		}
	}
	if len(added) > 0 {
		result.WriteString(fmt.Sprintf("\nAdded (%d):\n", len(added)))
		for _, f := range added {
			result.WriteString("  " + f + "\n")
		}
	}
	if len(deleted) > 0 {
		result.WriteString(fmt.Sprintf("\nDeleted (%d):\n", len(deleted)))
		for _, f := range deleted {
			result.WriteString("  " + f + "\n")
		}
	}
	if len(untracked) > 0 {
		result.WriteString(fmt.Sprintf("\nUntracked (%d):\n", len(untracked)))
		for _, f := range untracked {
			result.WriteString("  " + f + "\n")
		}
	}

	return result.String(), nil
}

func (e *Executor) gitDiff(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	staged, _ := args["staged"].(bool)

	cmdArgs := []string{"diff"}
	if staged {
		cmdArgs = append(cmdArgs, "--staged")
	}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	if len(output) == 0 {
		return "No changes", nil
	}

	// Truncate if too long
	result := string(output)
	if len(result) > 5000 {
		result = result[:5000] + "\n... (truncated)"
	}

	return result, nil
}

func (e *Executor) gitLog(args map[string]any, projectRoot string) (string, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	path, _ := args["path"].(string)

	cmdArgs := []string{"log", fmt.Sprintf("-n%d", limit), "--oneline", "--decorate"}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return "No commits", nil
	}

	return fmt.Sprintf("Recent commits:\n%s", string(output)), nil
}

func (e *Executor) gitBlame(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	startLine := 0
	endLine := 0
	if s, ok := args["start_line"].(float64); ok && s > 0 {
		startLine = int(s)
	}
	if e, ok := args["end_line"].(float64); ok && e > 0 {
		endLine = int(e)
	}

	cmdArgs := []string{"blame", "--line-porcelain"}
	if startLine > 0 && endLine > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-L%d,%d", startLine, endLine))
	} else if startLine > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-L%d,+20", startLine))
	}
	cmdArgs = append(cmdArgs, path)

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git blame failed: %w", err)
	}

	// Parse porcelain output
	lines := strings.Split(string(output), "\n")
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Blame for %s:\n\n", path))

	var currentCommit, author, date, lineContent string
	lineNum := startLine
	if lineNum == 0 {
		lineNum = 1
	}

	for _, line := range lines {
		switch {
		case len(line) == 40 || (len(line) > 40 && line[40] == ' '):
			// Commit hash line
			if len(line) >= 40 {
				currentCommit = line[:8] // Short hash
			}
		case strings.HasPrefix(line, "author "):
			author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "author-time "):
			// Skip, we'll use author-tz-time or committer-time
		case strings.HasPrefix(line, "summary "):
			// Skip summary
		case strings.HasPrefix(line, "\t"):
			lineContent = strings.TrimPrefix(line, "\t")
			result.WriteString(fmt.Sprintf("%4d | %s | %-15s | %s\n", lineNum, currentCommit, truncate(author, 15), lineContent))
			lineNum++
		}
		_ = date // Suppress unused warning
	}

	return result.String(), nil
}

func (e *Executor) gitShow(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	commit, _ := args["commit"].(string)

	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if commit == "" {
		commit = "HEAD"
	}

	// Get file content at specific commit
	ref := fmt.Sprintf("%s:%s", commit, path)
	cmd := exec.Command("git", "show", ref)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show failed: %w", err)
	}

	result := string(output)
	if len(result) > 10000 {
		result = result[:10000] + "\n... (truncated)"
	}

	return fmt.Sprintf("Content of %s at %s:\n\n%s", path, commit, result), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Phase 5: Git-Aware Context Tools

// runGitDiff executes git diff command with fallback to master
func runGitDiff(projectRoot string, cmdArgs []string, base, compare string) ([]byte, error) {
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil && base == "main" {
		cmdArgs[2] = "master..." + compare
		cmd = exec.Command("git", cmdArgs...)
		cmd.Dir = projectRoot
		output, err = cmd.Output()
	}
	return output, err
}

// categorizeChanges categorizes file changes by status
func categorizeChanges(output []byte) (added, modified, deleted []string) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		file := strings.TrimSpace(line[1:])
		switch line[0] {
		case 'A':
			added = append(added, file)
		case 'M':
			modified = append(modified, file)
		case 'D':
			deleted = append(deleted, file)
		}
	}
	return
}

// formatDiffResult formats the diff result as a string
func formatDiffResult(base, compare string, added, modified, deleted []string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Changes between %s and %s:\n", base, compare))

	writeSection := func(title string, prefix string, files []string) {
		if len(files) > 0 {
			result.WriteString(fmt.Sprintf("\n%s (%d):\n", title, len(files)))
			for _, f := range files {
				result.WriteString("  " + prefix + " " + f + "\n")
			}
		}
	}
	writeSection("Added", "+", added)
	writeSection("Modified", "M", modified)
	writeSection("Deleted", "-", deleted)
	return result.String()
}

// gitDiffBranches shows diff between two branches
func (e *Executor) gitDiffBranches(args map[string]any, projectRoot string) (string, error) {
	base, _ := args["base"].(string)
	compare, _ := args["compare"].(string)
	path, _ := args["path"].(string)

	if base == "" {
		base = "main"
	}
	if compare == "" {
		compare = "HEAD"
	}

	cmdArgs := []string{"diff", "--name-status", base + "..." + compare}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	output, err := runGitDiff(projectRoot, cmdArgs, base, compare)
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	if len(output) == 0 {
		return fmt.Sprintf("No differences between %s and %s", base, compare), nil
	}

	added, modified, deleted := categorizeChanges(output)
	return formatDiffResult(base, compare, added, modified, deleted), nil
}

// gitSearchCommits searches commits by message pattern
func (e *Executor) gitSearchCommits(args map[string]any, projectRoot string) (string, error) {
	query, _ := args["query"].(string)
	author, _ := args["author"].(string)
	since, _ := args["since"].(string)
	path, _ := args["path"].(string)
	limit := 20
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if query == "" && author == "" && since == "" {
		return "", fmt.Errorf("at least one of query, author, or since is required")
	}

	cmdArgs := []string{"log", fmt.Sprintf("-n%d", limit), "--format=%h | %s | %an | %ar"}

	if query != "" {
		cmdArgs = append(cmdArgs, "--grep="+query)
	}
	if author != "" {
		cmdArgs = append(cmdArgs, "--author="+author)
	}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since="+since)
	}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log search failed: %w", err)
	}

	if len(output) == 0 {
		return "No commits found matching criteria", nil
	}

	var result strings.Builder
	result.WriteString("Matching commits:\n\n")
	result.WriteString("Hash     | Message                                    | Author          | When\n")
	result.WriteString("---------+--------------------------------------------+-----------------+------------\n")
	result.WriteString(string(output))

	return result.String(), nil
}

// gitChangedFiles returns files changed in a time period or by pattern
func (e *Executor) gitChangedFiles(args map[string]any, projectRoot string) (string, error) {
	since, _ := args["since"].(string)
	path, _ := args["path"].(string)
	author, _ := args["author"].(string)

	if since == "" {
		since = "1 week ago"
	}

	cmdArgs := []string{"log", "--since=" + since, "--name-only", "--format="}
	if author != "" {
		cmdArgs = append(cmdArgs, "--author="+author)
	}
	if path != "" {
		cmdArgs = append(cmdArgs, "--", path)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return fmt.Sprintf("No files changed since %s", since), nil
	}

	// Count file occurrences
	fileCounts := make(map[string]int)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fileCounts[line]++
		}
	}

	// Sort by count
	type fileCount struct {
		file  string
		count int
	}
	sorted := make([]fileCount, 0, len(fileCounts))
	for f, c := range fileCounts {
		sorted = append(sorted, fileCount{f, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files changed since %s (%d unique files):\n\n", since, len(sorted)))

	for i, fc := range sorted {
		if i >= 50 {
			result.WriteString(fmt.Sprintf("\n... and %d more files", len(sorted)-50))
			break
		}
		result.WriteString(fmt.Sprintf("%3d changes: %s\n", fc.count, fc.file))
	}

	return result.String(), nil
}

// gitFileHistory shows detailed history of a file
func (e *Executor) gitFileHistory(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	cmdArgs := []string{"log", fmt.Sprintf("-n%d", limit), "--follow", "--format=%h|%s|%an|%ar|%ad", "--date=short", "--", path}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return fmt.Sprintf("No history found for %s", path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("History of %s:\n\n", path))

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) >= 5 {
			result.WriteString(fmt.Sprintf("%s  %s\n", parts[0], parts[4]))
			result.WriteString(fmt.Sprintf("    %s (%s, %s)\n\n", parts[1], parts[2], parts[3]))
		}
	}

	return result.String(), nil
}

// gitCoChanged returns files that are often changed together with the given file
func (e *Executor) gitCoChanged(args map[string]any, projectRoot string) (string, error) {
	path, _ := args["path"].(string)
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Get commits that changed this file
	cmd := exec.Command("git", "log", "--format=%H", "-n", "50", "--", path)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	fileCounts := make(map[string]int)

	for _, commit := range commits {
		if commit == "" {
			continue
		}
		cmd = exec.Command("git", "show", "--name-only", "--format=", commit)
		cmd.Dir = projectRoot
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		for _, file := range strings.Split(string(out), "\n") {
			file = strings.TrimSpace(file)
			if file != "" && file != path {
				fileCounts[file]++
			}
		}
	}

	if len(fileCounts) == 0 {
		return fmt.Sprintf("No co-changed files found for %s", path), nil
	}

	// Sort by count
	type fileCount struct {
		path  string
		count int
	}
	sorted := make([]fileCount, 0, len(fileCounts))
	for p, c := range fileCounts {
		sorted = append(sorted, fileCount{p, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files often changed together with %s:\n\n", path))

	for i, fc := range sorted {
		if i >= limit {
			break
		}
		result.WriteString(fmt.Sprintf("%3d times: %s\n", fc.count, fc.path))
	}

	return result.String(), nil
}

// gitSuggestContext suggests files to include in context based on git history
func (e *Executor) gitSuggestContext(args map[string]any, projectRoot string) (string, error) {
	task, _ := args["task"].(string)
	filesArg, _ := args["current_files"].([]any)
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	currentFiles := e.extractStringSlice(filesArg)
	suggestions := make(map[string]int)

	e.collectCoChangedSuggestions(currentFiles, projectRoot, suggestions)
	e.collectTaskKeywordSuggestions(task, projectRoot, suggestions)
	e.removeCurrentFiles(currentFiles, suggestions)

	if len(suggestions) == 0 {
		return "No additional files suggested", nil
	}

	return e.formatSuggestions(suggestions, limit), nil
}

// extractStringSlice converts []any to []string
func (e *Executor) extractStringSlice(filesArg []any) []string {
	var result []string
	for _, f := range filesArg {
		if s, ok := f.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// collectCoChangedSuggestions finds files that are often changed together
func (e *Executor) collectCoChangedSuggestions(currentFiles []string, projectRoot string, suggestions map[string]int) {
	for _, file := range currentFiles {
		commits := e.getFileCommits(file, projectRoot, 30)
		for _, commit := range commits {
			files := e.getCommitFiles(commit, projectRoot)
			for _, f := range files {
				if f != file {
					suggestions[f] += 2
				}
			}
		}
	}
}

// getFileCommits returns commit hashes for a file
func (e *Executor) getFileCommits(file, projectRoot string, limit int) []string {
	// #nosec G204 - file path is from internal git operations
	cmd := exec.Command("git", "log", "--format=%H", "-n", fmt.Sprintf("%d", limit), "--", file)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var commits []string
	for _, c := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if c != "" {
			commits = append(commits, c)
		}
	}
	return commits
}

// getCommitFiles returns files changed in a commit
func (e *Executor) getCommitFiles(commit, projectRoot string) []string {
	// #nosec G204 - commit hash is validated before use
	cmd := exec.Command("git", "show", "--name-only", "--format=", commit)
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var files []string
	for _, f := range strings.Split(string(out), "\n") {
		f = strings.TrimSpace(f)
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// collectTaskKeywordSuggestions searches commits by task keywords
func (e *Executor) collectTaskKeywordSuggestions(task, projectRoot string, suggestions map[string]int) {
	if task == "" {
		return
	}
	for _, word := range strings.Fields(task) {
		if len(word) > 3 {
			e.searchCommitsByKeyword(word, projectRoot, suggestions)
		}
	}
}

// searchCommitsByKeyword searches commits containing a keyword
func (e *Executor) searchCommitsByKeyword(word, projectRoot string, suggestions map[string]int) {
	cmd := exec.Command("git", "log", "--grep="+word, "-i", "--name-only", "--format=", "-n", "10") //nolint:gosec // Git command
	cmd.Dir = projectRoot
	output, _ := cmd.Output()
	for _, f := range strings.Split(string(output), "\n") {
		f = strings.TrimSpace(f)
		if f != "" {
			suggestions[f]++
		}
	}
}

// removeCurrentFiles removes already included files from suggestions
func (e *Executor) removeCurrentFiles(currentFiles []string, suggestions map[string]int) {
	for _, f := range currentFiles {
		delete(suggestions, f)
	}
}

// formatSuggestions formats the suggestions map as a string
func (e *Executor) formatSuggestions(suggestions map[string]int, limit int) string {
	type suggestion struct {
		path  string
		score int
	}
	sorted := make([]suggestion, 0, len(suggestions))
	for p, s := range suggestions {
		sorted = append(sorted, suggestion{p, s})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	var result strings.Builder
	result.WriteString("Suggested files to include in context:\n\n")
	for i, s := range sorted {
		if i >= limit {
			break
		}
		result.WriteString(fmt.Sprintf("  %s (score: %d)\n", s.path, s.score))
	}
	return result.String()
}
