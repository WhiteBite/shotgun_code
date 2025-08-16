package fsscanner

import (
	"io/fs"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"

	gitignore "github.com/sabhiram/go-gitignore"
)

type fileTreeBuilder struct {
	settingsRepo domain.SettingsRepository
	log          domain.Logger

	mu          sync.RWMutex
	giCache     map[string]*gitignore.GitIgnore // per-project .gitignore cache
	customCache *gitignore.GitIgnore            // compiled custom rules
	customHash  string                          // hash of custom rules content for cache invalidation
}

func New(settingsRepo domain.SettingsRepository, log domain.Logger) domain.TreeBuilder {
	return &fileTreeBuilder{
		settingsRepo: settingsRepo,
		log:          log,
		giCache:      make(map[string]*gitignore.GitIgnore),
	}
}

func (b *fileTreeBuilder) BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	var gi *gitignore.GitIgnore
	var ci *gitignore.GitIgnore

	if useGitignore {
		gi = b.getGitignore(dirPath)
	}
	if useCustomIgnore {
		ci = b.getCustomIgnore()
	}

	nodesMap := make(map[string]*domain.FileNode)

	root := &domain.FileNode{
		Name:     filepath.Base(dirPath),
		Path:     dirPath,
		RelPath:  ".",
		IsDir:    true,
		Children: []*domain.FileNode{},
	}
	nodesMap[dirPath] = root

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dirPath {
			return nil
		}
		relPath, _ := filepath.Rel(dirPath, path)
		matchPath := relPath
		if d.IsDir() && !strings.HasSuffix(matchPath, string(filepath.Separator)) {
			matchPath += string(filepath.Separator)
		}

		isGi := gi != nil && gi.MatchesPath(matchPath)
		isCi := ci != nil && ci.MatchesPath(matchPath)

		if d.IsDir() && (isGi || isCi) {
			return fs.SkipDir
		}

		var fsize int64
		if !d.IsDir() {
			if info, e := d.Info(); e == nil {
				fsize = info.Size()
			}
		}

		node := &domain.FileNode{
			Name:            d.Name(),
			Path:            path,
			RelPath:         relPath,
			IsDir:           d.IsDir(),
			IsGitignored:    isGi,
			IsCustomIgnored: isCi,
			Children:        []*domain.FileNode{},
			Size:            fsize,
		}
		nodesMap[path] = node

		parent := filepath.Dir(path)
		if p, ok := nodesMap[parent]; ok {
			p.Children = append(p.Children, node)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, n := range nodesMap {
		sort.Slice(n.Children, func(i, j int) bool {
			if n.Children[i].IsDir != n.Children[j].IsDir {
				return n.Children[i].IsDir
			}
			return strings.ToLower(n.Children[i].Name) < strings.ToLower(n.Children[j].Name)
		})
	}

	return []*domain.FileNode{root}, nil
}

func (b *fileTreeBuilder) getGitignore(root string) *gitignore.GitIgnore {
	b.mu.RLock()
	if gi, ok := b.giCache[root]; ok {
		b.mu.RUnlock()
		return gi
	}
	b.mu.RUnlock()

	ig, err := gitignore.CompileIgnoreFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return nil
	}

	b.mu.Lock()
	b.giCache[root] = ig
	b.mu.Unlock()
	return ig
}

func (b *fileTreeBuilder) getCustomIgnore() *gitignore.GitIgnore {
	rules := strings.ReplaceAll(b.settingsRepo.GetCustomIgnoreRules(), "\r\n", "\n")
	trimmed := []string{}
	for _, line := range strings.Split(rules, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		trimmed = append(trimmed, line)
	}
	hash := strings.Join(trimmed, "\n")

	b.mu.RLock()
	if b.customCache != nil && b.customHash == hash {
		cc := b.customCache
		b.mu.RUnlock()
		return cc
	}
	b.mu.RUnlock()

	if len(trimmed) == 0 {
		return nil
	}
	ci := gitignore.CompileIgnoreLines(trimmed...)

	b.mu.Lock()
	b.customCache = ci
	b.customHash = hash
	b.mu.Unlock()
	return ci
}
