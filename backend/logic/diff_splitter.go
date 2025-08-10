package logic

import (
	"regexp"
	"strings"

	"shotgun_code/domain"
)

type diffSplitter struct {
	log domain.Logger
}

func NewDiffSplitter(log domain.Logger) domain.DiffSplitter {
	return &diffSplitter{log: log}
}

func (ds *diffSplitter) Split(gitDiffText string, approxLineLimit int) ([]string, error) {
	ds.log.Info("Splitting diff with line limit " + string(rune(approxLineLimit)))

	if strings.TrimSpace(gitDiffText) == "" {
		return []string{}, nil
	}

	fileDiffStartRegex := regexp.MustCompile(`(?m)^diff --git `)
	startIndices := fileDiffStartRegex.FindAllStringIndex(gitDiffText, -1)

	var fileDiffBlocks []string
	if len(startIndices) == 0 {
		ds.log.Warning("No 'diff --git' blocks found.")
		if strings.TrimSpace(gitDiffText) != "" {
			fileDiffBlocks = append(fileDiffBlocks, gitDiffText)
		}
	} else {
		for i := 0; i < len(startIndices); i++ {
			start := startIndices[i][0]
			var end int
			if i+1 < len(startIndices) {
				end = startIndices[i+1][0]
			} else {
				end = len(gitDiffText)
			}
			block := gitDiffText[start:end]
			if strings.TrimSpace(block) != "" {
				fileDiffBlocks = append(fileDiffBlocks, block)
			}
		}
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

			var fileHeader strings.Builder
			var hunks []string
			firstHunkIdx := -1
			for i, line := range lines {
				if hunkHeaderRegex.MatchString(line) {
					firstHunkIdx = i
					break
				}
				fileHeader.WriteString(line + "\n")
			}

			if firstHunkIdx != -1 {
				var currentHunk strings.Builder
				for i := firstHunkIdx; i < len(lines); i++ {
					if hunkHeaderRegex.MatchString(lines[i]) && currentHunk.Len() > 0 {
						hunks = append(hunks, currentHunk.String())
						currentHunk.Reset()
					}
					currentHunk.WriteString(lines[i] + "\n")
				}
				hunks = append(hunks, currentHunk.String())

				var tempSplit strings.Builder
				tempSplit.WriteString(fileHeader.String())
				for _, hunk := range hunks {
					hunkLineCount := len(strings.Split(hunk, "\n"))
					if tempSplit.Len() > fileHeader.Len() && len(strings.Split(tempSplit.String(), "\n"))+hunkLineCount > approxLineLimit {
						splitDiffs = append(splitDiffs, strings.TrimSpace(tempSplit.String()))
						tempSplit.Reset()
						tempSplit.WriteString(fileHeader.String())
					}
					tempSplit.WriteString(hunk)
				}
				if tempSplit.Len() > fileHeader.Len() {
					splitDiffs = append(splitDiffs, strings.TrimSpace(tempSplit.String()))
				}
			} else {
				for i := 0; i < len(lines); i += approxLineLimit {
					end := i + approxLineLimit
					if end > len(lines) {
						end = len(lines)
					}
					chunk := strings.Join(lines[i:end], "\n")
					splitDiffs = append(splitDiffs, strings.TrimSpace(chunk))
				}
			}
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
