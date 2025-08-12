package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"strings"
)

var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9\/\._-]+$`)

type Repository struct {
	log domain.Logger
}

func New(logger domain.Logger) domain.GitRepository {
	return &Repository{log: logger}
}
func (gs *Repository) CheckAvailability() (bool, error) {
	_, err := exec.LookPath("git")
	if err != nil {
		gs.log.Warning("Git executable not found in PATH.")
		return false, nil
	}
	return true, nil
}
func (gs *Repository) executeGitCommand(projectRoot string, args ...string) ([]byte, error) {
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("could not resolve absolute path for '%s': %w", projectRoot, err)
	}
	cmdArgs := append([]string{"-C", absPath}, args...)
	cmd := exec.Command("git", cmdArgs...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git command failed: %w - %s", err, errOut.String())
	}
	return out.Bytes(), nil
}
func (gs *Repository) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	output, err := gs.executeGitCommand(projectRoot, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	var statuses []domain.FileStatus
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		if strings.HasPrefix(status, "R") {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1]
			}
		}
		simpleStatus := ""
		if strings.HasPrefix(status, "M") || strings.Contains(status, "M") {
			simpleStatus = "M"
		} else if strings.HasPrefix(status, "A") {
			simpleStatus = "A"
		} else if strings.HasPrefix(status, "D") {
			simpleStatus = "D"
		} else if strings.HasPrefix(status, "R") {
			simpleStatus = "R"
		} else if strings.HasPrefix(status, "C") {
			simpleStatus = "C"
		} else if status == "??" {
			simpleStatus = "U"
		} else {
			simpleStatus = status
		}
		statuses = append(statuses, domain.FileStatus{
			Path:   path,
			Status: simpleStatus,
		})
	}
	gs.log.Info(fmt.Sprintf("Found %d uncommitted files.", len(statuses)))
	return statuses, nil
}
func (gs *Repository) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	logArgs := []string{
		"log",
		"--pretty=format:COMMIT %H %P%n%s%n%an%n%cI",
		"--name-status",
		"--topo-order",
		"-n", fmt.Sprintf("%d", limit),
	}
	if branchName != "" {
		if err := validateBranchName(branchName); err != nil {
			return nil, err
		}
		logArgs = append(logArgs, branchName)
	}
	output, err := gs.executeGitCommand(projectRoot, logArgs...)
	if err != nil {
		return nil, err
	}
	commits, parseErr := ParseRichLogOutput(string(output))
	if parseErr != nil {
		return nil, parseErr
	}
	gs.log.Info(fmt.Sprintf("Loaded and parsed %d commits.", len(commits)))
	return commits, nil
}
func (gs *Repository) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	if strings.Contains(filePath, "..") || strings.Contains(commitHash, "..") {
		return "", fmt.Errorf("invalid characters in path or hash")
	}
	output, err := gs.executeGitCommand(projectRoot, "show", fmt.Sprintf("%s:%s", commitHash, filePath))
	if err != nil {
		return "", fmt.Errorf("could not get file content for %s at commit %s: %w", filePath, commitHash, err)
	}
	return string(output), nil
}
func (gs *Repository) GetGitignoreContent(projectRoot string) (string, error) {
	gitignorePath := filepath.Join(projectRoot, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return "", nil
	}
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
func ParseRichLogOutput(output string) ([]domain.CommitWithFiles, error) {
	var commits []domain.CommitWithFiles
	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentCommit *domain.CommitWithFiles
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMIT ") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			hash := parts[1]
			parentHashes := parts[2:]
			currentCommit = &domain.CommitWithFiles{
				Hash:    hash,
				IsMerge: len(parentHashes) > 1,
				Files:   []string{},
			}
			if scanner.Scan() {
				currentCommit.Subject = scanner.Text()
			}
			if scanner.Scan() {
				currentCommit.Author = scanner.Text()
			}
			if scanner.Scan() {
				currentCommit.Date = scanner.Text()
			}
		} else if currentCommit != nil && len(strings.TrimSpace(line)) > 0 {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				filePath := parts[len(parts)-1]
				currentCommit.Files = append(currentCommit.Files, filePath)
			}
		}
	}
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing git log output: %w", err)
	}
	return commits, nil
}
func validateBranchName(name string) error {
	if !validBranchName.MatchString(name) {
		return fmt.Errorf("invalid branch name: %s", name)
	}
	return nil
}
