package fsscanner

import (
	"context"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"

	gitignore "github.com/sabhiram/go-gitignore"
)

type TreeBuilder struct {
	settingsRepo     domain.SettingsRepository
	mu               sync.RWMutex
	projectGitignore *gitignore.GitIgnore
}

func New(settingsRepo domain.SettingsRepository) *TreeBuilder {
	return &TreeBuilder{
		settingsRepo: settingsRepo,
	}
}

func (tb *TreeBuilder) BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	tb.mu.Lock()
	tb.projectGitignore = nil
	tb.mu.Unlock()
	var gitIgn *gitignore.GitIgnore
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		gitIgn, err = gitignore.CompileIgnoreFile(gitignorePath)
		if err == nil {
			tb.mu.Lock()
			tb.projectGitignore = gitIgn
			tb.mu.Unlock()
		}
	}

	var customIgn *gitignore.GitIgnore
	if useCustomIgnore {
		customRulesText := tb.settingsRepo.GetCustomIgnoreRules()
		if strings.TrimSpace(customRulesText) != "" {
			lines := strings.Split(strings.ReplaceAll(customRulesText, "\r\n", "\n"), "\n")
			customIgn = gitignore.CompileIgnoreLines(lines...)
		}
	}

	rootNode := &domain.FileNode{
		Name:    filepath.Base(dirPath),
		Path:    dirPath,
		RelPath: ".",
		IsDir:   true,
	}
	children, err := tb.buildRecursive(context.Background(), dirPath, dirPath, useGitignore, useCustomIgnore, customIgn)
	if err != nil {
		return nil, err
	}
	rootNode.Children = children
	return []*domain.FileNode{rootNode}, nil
}

func (tb *TreeBuilder) buildRecursive(ctx context.Context, currentPath, rootPath string, useGitignore, useCustomIgnore bool, customIgn *gitignore.GitIgnore) ([]*domain.FileNode, error) {
	tb.mu.RLock()
	gitIgn := tb.projectGitignore
	tb.mu.RUnlock()
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}
	var nodes []*domain.FileNode
	for _, entry := range entries {
		nodePath := filepath.Join(currentPath, entry.Name())
		relPath, _ := filepath.Rel(rootPath, nodePath)
		pathToMatch := relPath
		if entry.IsDir() {
			pathToMatch = strings.TrimSuffix(pathToMatch, string(os.PathSeparator)) + string(os.PathSeparator)
		}
		isGitignored := useGitignore && gitIgn != nil && gitIgn.MatchesPath(pathToMatch)
		isCustomIgnored := useCustomIgnore && customIgn != nil && customIgn.MatchesPath(pathToMatch)
		node := &domain.FileNode{
			Name:            entry.Name(),
			Path:            nodePath,
			RelPath:         relPath,
			IsDir:           entry.IsDir(),
			IsGitignored:    isGitignored,
			IsCustomIgnored: isCustomIgnored,
		}
		if entry.IsDir() && !(isGitignored || isCustomIgnored) {
			children, err := tb.buildRecursive(ctx, nodePath, rootPath, useGitignore, useCustomIgnore, customIgn)
			if err == nil {
				node.Children = children
			}
		}
		nodes = append(nodes, node)
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
	return nodes, nil
}
