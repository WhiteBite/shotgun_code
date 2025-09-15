package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"shotgun_code/domain"
	"strings"
	"time"
)

type Repository struct {
	log domain.Logger
}

func New(log domain.Logger) domain.GitRepository {
	return &Repository{
		log: log,
	}
}

func (r *Repository) IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func (r *Repository) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	var files []domain.FileStatus
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= 3 {
			status := strings.TrimSpace(line[:2])
			path := strings.TrimSpace(line[3:])
			// Handle renamed files output: "R old_path -> new_path"
			if strings.HasPrefix(status, "R") {
				parts := strings.Split(path, " -> ")
				if len(parts) == 2 {
					path = parts[1] // Use the new path for renamed files
				}
			}

			// Map git status codes to simpler ones for frontend display
			simpleStatus := status
			if strings.HasPrefix(status, "M") || strings.Contains(status, "M") {
				simpleStatus = "M" // Modified
			} else if strings.HasPrefix(status, "A") {
				simpleStatus = "A" // Added
			} else if strings.HasPrefix(status, "D") {
				simpleStatus = "D" // Deleted
			} else if strings.HasPrefix(status, "R") {
				simpleStatus = "R" // Renamed
			} else if strings.HasPrefix(status, "C") {
				simpleStatus = "C" // Copied
			} else if status == "??" {
				simpleStatus = "U" // Untracked
			} else if strings.HasPrefix(status, "U") {
				// Special handling for unmerged (conflict) status
				// Using "UM" to distinguish from "C" (Copied)
				simpleStatus = "UM" // Unmerged (Conflict)
			}

			files = append(files, domain.FileStatus{
				Path:   path,
				Status: simpleStatus,
			})
		}
	}

	r.log.Info(fmt.Sprintf("Found %d uncommitted files in %s.", len(files), projectRoot))
	return files, nil
}

func (r *Repository) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	args := []string{
		"log",
		"--pretty=format:COMMIT %H %P%n%s%n%an%n%cI", // Include Author Name (%an) and Committer Date, ISO 8601 format (%cI)
		"--name-status",
		"--topo-order",
		fmt.Sprintf("--max-count=%d", limit),
	}

	if branchName != "" {
		args = append(args, branchName)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	commits, parseErr := ParseRichLogOutput(string(output))
	if parseErr != nil {
		return nil, parseErr
	}
	r.log.Info(fmt.Sprintf("Loaded and parsed %d commits for %s.", len(commits), projectRoot))
	return commits, nil
}

func (r *Repository) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commitHash, filePath))
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file content at commit %s:%s: %w", commitHash, filePath, err)
	}

	return string(output), nil
}

func (r *Repository) GetGitignoreContent(projectRoot string) (string, error) {
	cmd := exec.Command("git", "show", "HEAD:.gitignore")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		// .gitignore might not exist or not be tracked, which is not an error for us.
		r.log.Debug(fmt.Sprintf("Could not get .gitignore content for %s (might not exist or be tracked): %v", projectRoot, err))
		return "", nil
	}

	return string(output), nil
}

// ParseRichLogOutput parses the output of git log with --pretty and --name-status
func ParseRichLogOutput(gitOutput string) ([]domain.CommitWithFiles, error) {
	var commits []domain.CommitWithFiles
	scanner := bufio.NewScanner(strings.NewReader(gitOutput))

	var currentCommit *domain.CommitWithFiles
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "COMMIT ") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue // Skip malformed lines
			}

			hash := parts[1]
			isMerge := len(parts) > 3 // More than one parent means merge

			currentCommit = &domain.CommitWithFiles{
				Hash:    hash,
				IsMerge: isMerge,
				Files:   []string{},
			}
		} else if currentCommit != nil {
			// Check if this is a file status line first
			if strings.Contains(line, "\t") {
				// This is a file status line: "M\tfilename" or "A\tfilename"
				parts := strings.Split(line, "\t")
				if len(parts) >= 2 {
					filename := parts[1]
					if filename != "" {
						currentCommit.Files = append(currentCommit.Files, filename)
					}
				}
			} else if currentCommit.Subject == "" {
				// First non-file line is the subject
				currentCommit.Subject = line
			} else if currentCommit.Author == "" {
				// Second non-file line is the author
				currentCommit.Author = line
			} else if currentCommit.Date == "" {
				// Third non-file line is the date
				currentCommit.Date = line
				// Optional: parse date into a more human-readable format if needed,
				// or store as is for frontend formatting.
				// For now, storing as is from git's %cI format.
				t, err := time.Parse(time.RFC3339, line)
				if err == nil {
					currentCommit.Date = t.Format("2006-01-02 15:04") // Format for display
				}
			}
		}
	}

	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	return commits, nil
}

// GetBranches returns all git branches
func (r *Repository) GetBranches(projectRoot string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches: %w", err)
	}

	var branches []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Remove the current branch indicator (*)
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		} else if strings.HasPrefix(line, "  ") {
			line = strings.TrimPrefix(line, "  ")
		}

		// Skip remote tracking references for now
		if strings.HasPrefix(line, "remotes/") {
			continue
		}

		// Skip empty lines and special references
		if line != "" && !strings.Contains(line, "HEAD") {
			branches = append(branches, line)
		}
	}

	r.log.Info(fmt.Sprintf("Found %d branches in %s.", len(branches), projectRoot))
	return branches, nil
}

// GetCurrentBranch returns the current git branch
func (r *Repository) GetCurrentBranch(projectRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current git branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	r.log.Info(fmt.Sprintf("Current branch in %s: %s", projectRoot, branch))
	return branch, nil
}
