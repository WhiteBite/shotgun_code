package textutils

import (
	"fmt"
	"regexp"
	"shotgun_code/domain"
	"strings"
)

type diffSplitter struct {
	log domain.Logger
}

func NewDiffSplitter(log domain.Logger) domain.DiffSplitter {
	return &diffSplitter{log: log}
}

// extractDiffBlocks extracts individual file diff blocks from git diff text
func (ds *diffSplitter) extractDiffBlocks(gitDiffText string) []string {
	fileDiffStartRegex := regexp.MustCompile(`(?m)^diff --git `)
	startIndices := fileDiffStartRegex.FindAllStringIndex(gitDiffText, -1)

	if len(startIndices) == 0 {
		ds.log.Warning("No 'diff --git' blocks found.")
		if strings.TrimSpace(gitDiffText) != "" {
			return []string{gitDiffText}
		}
		return nil
	}

	var blocks []string
	for i := 0; i < len(startIndices); i++ {
		start := startIndices[i][0]
		end := len(gitDiffText)
		if i+1 < len(startIndices) {
			end = startIndices[i+1][0]
		}
		if block := gitDiffText[start:end]; strings.TrimSpace(block) != "" {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// extractHunks extracts hunks from a diff block
func (ds *diffSplitter) extractHunks(lines []string, hunkHeaderRegex *regexp.Regexp) (header string, hunks []string, firstHunkIdx int) {
	var fileHeader strings.Builder
	firstHunkIdx = -1

	for i, line := range lines {
		if hunkHeaderRegex.MatchString(line) {
			firstHunkIdx = i
			break
		}
		fileHeader.WriteString(line + "\n")
	}
	header = fileHeader.String()

	if firstHunkIdx == -1 {
		return header, nil, -1
	}

	var currentHunk strings.Builder
	for i := firstHunkIdx; i < len(lines); i++ {
		if hunkHeaderRegex.MatchString(lines[i]) && currentHunk.Len() > 0 {
			hunks = append(hunks, currentHunk.String())
			currentHunk.Reset()
		}
		currentHunk.WriteString(lines[i] + "\n")
	}
	if currentHunk.Len() > 0 {
		hunks = append(hunks, currentHunk.String())
	}

	return header, hunks, firstHunkIdx
}

// splitLargeBlock splits a large diff block by hunks or by line count
func (ds *diffSplitter) splitLargeBlock(lines []string, approxLineLimit int, hunkHeaderRegex *regexp.Regexp) []string {
	var result []string
	header, hunks, firstHunkIdx := ds.extractHunks(lines, hunkHeaderRegex)

	if firstHunkIdx != -1 && len(hunks) > 0 {
		var tempSplit strings.Builder
		tempSplit.WriteString(header)

		for _, hunk := range hunks {
			hunkLineCount := len(strings.Split(hunk, "\n"))
			if tempSplit.Len() > len(header) && len(strings.Split(tempSplit.String(), "\n"))+hunkLineCount > approxLineLimit {
				result = append(result, strings.TrimSpace(tempSplit.String()))
				tempSplit.Reset()
				tempSplit.WriteString(header)
			}
			tempSplit.WriteString(hunk)
		}
		if tempSplit.Len() > len(header) {
			result = append(result, strings.TrimSpace(tempSplit.String()))
		}
	} else {
		for i := 0; i < len(lines); i += approxLineLimit {
			end := i + approxLineLimit
			if end > len(lines) {
				end = len(lines)
			}
			chunk := strings.Join(lines[i:end], "\n")
			result = append(result, strings.TrimSpace(chunk))
		}
	}

	return result
}

func (ds *diffSplitter) Split(gitDiffText string, approxLineLimit int) ([]string, error) {
	ds.log.Info(fmt.Sprintf("Splitting diff with line limit %d", approxLineLimit))

	if strings.TrimSpace(gitDiffText) == "" {
		return []string{}, nil
	}

	fileDiffBlocks := ds.extractDiffBlocks(gitDiffText)
	if len(fileDiffBlocks) == 0 {
		return []string{}, nil
	}

	var splitDiffs []string
	var currentSplit strings.Builder
	currentLines := 0
	hunkHeaderRegex := regexp.MustCompile(`^@@ .* @@`)

	for _, block := range fileDiffBlocks {
		lines := strings.Split(block, "\n")
		blockLineCount := len(lines)

		if blockLineCount > approxLineLimit {
			if currentSplit.Len() > 0 {
				splitDiffs = append(splitDiffs, strings.TrimSpace(currentSplit.String()))
				currentSplit.Reset()
				currentLines = 0
			}
			splitDiffs = append(splitDiffs, ds.splitLargeBlock(lines, approxLineLimit, hunkHeaderRegex)...)
		} else {
			if currentLines > 0 && currentLines+blockLineCount > approxLineLimit {
				splitDiffs = append(splitDiffs, strings.TrimSpace(currentSplit.String()))
				currentSplit.Reset()
				currentLines = 0
			}
			currentSplit.WriteString(block)
			currentLines += blockLineCount
		}
	}

	if currentSplit.Len() > 0 {
		splitDiffs = append(splitDiffs, strings.TrimSpace(currentSplit.String()))
	}

	return splitDiffs, nil
}
