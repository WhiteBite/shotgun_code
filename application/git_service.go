package application

import (
	"bytes"
	"fmt"
	"shotgun_code/domain"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitServiceImpl реализует GitService
type GitServiceImpl struct {
	log domain.Logger
}

// NewGitService создает новый GitService
func NewGitService(log domain.Logger) domain.GitService {
	return &GitServiceImpl{log: log}
}

// GenerateDiff генерирует git diff для всех изменений в репозитории
func (s *GitServiceImpl) GenerateDiff(projectPath string) (string, error) {
	s.log.Info(fmt.Sprintf("Generating diff for project: %s", projectPath))

	// Открываем репозиторий
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Получаем рабочее дерево
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Получаем HEAD
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Получаем коммит, на ко��орый указывает HEAD
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	// Получаем дерево коммита
	commitTree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get commit tree: %w", err)
	}

	// Получаем статус рабочего дерева (изменения)
	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree status: %w", err)
	}

	// Если изменений нет, возвращаем пустую строку
	if status.IsClean() {
		return "No changes detected.", nil
	}

	// Получаем изменения между деревом коммита и рабочим деревом
	changes, err := commitTree.Diff(worktree)
	if err != nil {
		return "", fmt.Errorf("failed to diff commit tree and worktree: %w", err)
	}

	// Форматируем изменения в виде патча (diff)
	var patchBuff bytes.Buffer
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			return "", fmt.Errorf("failed to get patch for change: %w", err)
		}
		
		patchStr := patch.String()
		if _, err := patchBuff.WriteString(patchStr); err != nil {
			return "", fmt.Errorf("failed to write patch to buffer: %w", err)
		}
	}
	
	return patchBuff.String(), nil
}
