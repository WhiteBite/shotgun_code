package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"shotgun_code/domain"
	"strings"
)

var validCommitHash = regexp.MustCompile(`^[a-f0-9]{7,40}$`)
var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9\/\._-]+$`)

type Repository struct {
	log domain.Logger
}

func New(logger domain.Logger) *Repository {
	return &Repository{log: logger}
}

func (gs *Repository) CheckAvailability() (bool, error) {
	_, err := exec.LookPath("git")
	if err != nil {
		gs.log.Warning("Исполняемый файл 'git' не найден в PATH.")
		return false, nil
	}
	return true, nil
}

func (gs *Repository) GetUncommittedFiles(projectRoot string) ([]string, error) {
	if err := validatePath(projectRoot); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "-C", projectRoot, "status", "--porcelain")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ошибка git status: %w - %s", err, errOut.String())
	}
	var files []string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	gs.log.Info(fmt.Sprintf("Найдено %d незакоммиченных файлов.", len(files)))
	return files, nil
}

func (gs *Repository) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	if err := validatePath(projectRoot); err != nil {
		return nil, err
	}

	logArgs := []string{
		"-C", projectRoot, "log",
		"--pretty=format:COMMIT %H %P%n%s",
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

	cmd := exec.Command("git", logArgs...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ошибка git log: %w - %s", err, errOut.String())
	}

	commits, err := ParseRichLogOutput(out.String())
	if err != nil {
		return nil, err
	}
	gs.log.Info(fmt.Sprintf("Загружено и обработано %d коммитов.", len(commits)))
	return commits, nil
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
				continue // Corrupted line
			}
			hash := parts[1]
			parentHashes := parts[2:]

			if scanner.Scan() {
				subject := scanner.Text()
				currentCommit = &domain.CommitWithFiles{
					Hash:    hash,
					Subject: subject,
					IsMerge: len(parentHashes) > 1,
					Files:   []string{},
				}
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
		return nil, fmt.Errorf("ошибка парсинга вывода git log: %w", err)
	}

	return commits, nil
}

func validatePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("недопустимый путь: содержит '..'")
	}
	return nil
}

func validateBranchName(name string) error {
	if !validBranchName.MatchString(name) {
		return fmt.Errorf("недопустимое имя ветки: %s", name)
	}
	return nil
}
