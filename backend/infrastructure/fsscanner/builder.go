package fsscanner

import (
	"io/fs"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"sync"
	"time"

	gitignore "github.com/sabhiram/go-gitignore"
)

type fileTreeBuilder struct {
	settingsRepo domain.SettingsRepository
	log          domain.Logger

	mu          sync.RWMutex
	giCache     map[string]*gitignore.GitIgnore // per-project .gitignore cache
	customCache *gitignore.GitIgnore            // compiled custom rules
	customHash  string                          // hash of custom rules content for cache invalidation
	
	// Cache for file trees with timestamps for invalidation
	treeCache        map[string]*cachedTree
	cacheAccessTimes map[string]time.Time
	cacheMutex       sync.RWMutex
	cacheDuration    time.Duration
	cacheSize        int64
	cacheHits        int64
	cacheMisses      int64
}

type cachedTree struct {
	nodes   []*domain.FileNode
	modTime time.Time
	size    int64
}

const (
	maxTreeCacheEntries = 5
	maxTreeCacheSizeMB  = 20
)

func New(settingsRepo domain.SettingsRepository, log domain.Logger) domain.TreeBuilder {
	return &fileTreeBuilder{
		settingsRepo:     settingsRepo,
		log:              log,
		giCache:          make(map[string]*gitignore.GitIgnore),
		treeCache:        make(map[string]*cachedTree),
		cacheAccessTimes: make(map[string]time.Time),
		cacheDuration:    2 * time.Minute, // Cache for 2 minutes (reduced from 5)
	}
}

func (b *fileTreeBuilder) BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	// Check if we have a cached tree for this directory
	if cached := b.getCachedTree(dirPath); cached != nil {
		b.log.Debug("Using cached tree for: " + dirPath)
		return cached, nil
	}

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

		// Пропускаем файлы, которые должны игнорироваться
		if !d.IsDir() && (isGi || isCi) {
			return nil
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
			IsIgnored:       isGi || isCi, // Вычисляем IsIgnored
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

	// Cache the result
	result := []*domain.FileNode{root}
	b.setCachedTree(dirPath, result)

	return result, nil
}

// getCachedTree retrieves a cached tree if it's still valid
func (b *fileTreeBuilder) getCachedTree(dirPath string) []*domain.FileNode {
	// Только чтение под RLock
	b.cacheMutex.RLock()
	cached, exists := b.treeCache[dirPath]
	b.cacheMutex.RUnlock()

	if !exists {
		// Cache miss
		b.cacheMutex.Lock()
		b.cacheMisses++
		b.cacheMutex.Unlock()
		return nil
	}

	// Check if cache is still valid
	if time.Since(cached.modTime) < b.cacheDuration {
		// Cache hit - обновляем метрики под Lock
		b.cacheMutex.Lock()
		// Повторная проверка после получения Lock
		if stillCached, stillExists := b.treeCache[dirPath]; stillExists {
			if time.Since(stillCached.modTime) < b.cacheDuration {
				b.cacheAccessTimes[dirPath] = time.Now()
				b.cacheHits++
				b.cacheMutex.Unlock()
				return stillCached.nodes
			}
		}
		b.cacheMutex.Unlock()
	}

	// Cache expired - удаляем под Lock
	b.cacheMutex.Lock()
	if expiredCached, stillExists := b.treeCache[dirPath]; stillExists {
		b.cacheSize -= expiredCached.size
		delete(b.treeCache, dirPath)
		delete(b.cacheAccessTimes, dirPath)
	}
	b.cacheMisses++
	b.cacheMutex.Unlock()
	
	return nil
}

// setCachedTree stores a tree in the cache with LRU eviction
func (b *fileTreeBuilder) setCachedTree(dirPath string, nodes []*domain.FileNode) {
	b.cacheMutex.Lock()
	defer b.cacheMutex.Unlock()

	treeSize := b.estimateTreeSize(nodes)
	b.evictOldestCacheEntry(treeSize)

	b.treeCache[dirPath] = &cachedTree{
		nodes:   nodes,
		modTime: time.Now(),
		size:    treeSize,
	}
	b.cacheAccessTimes[dirPath] = time.Now()
	b.cacheSize += treeSize
}

// estimateTreeSize оценивает размер дерева в байтах
func (b *fileTreeBuilder) estimateTreeSize(nodes []*domain.FileNode) int64 {
	var size int64
	for _, node := range nodes {
		size += int64(len(node.Name) + len(node.Path) + len(node.RelPath) + 100) // Базовый размер структуры
		if node.Children != nil {
			size += b.estimateTreeSize(node.Children)
		}
	}
	return size
}

// evictOldestCacheEntry удаляет старые записи при превышении лимитов
// ВАЖНО: должен вызываться только под b.cacheMutex.Lock()
func (b *fileTreeBuilder) evictOldestCacheEntry(newTreeSize int64) {
	maxSize := int64(maxTreeCacheSizeMB * 1024 * 1024)
	
	for len(b.treeCache) >= maxTreeCacheEntries || (b.cacheSize+newTreeSize) > maxSize {
		if len(b.treeCache) == 0 {
			break
		}
		
		// Находим самую старую запись
		var oldestKey string
		var oldestTime time.Time = time.Now()
		
		for key, accessTime := range b.cacheAccessTimes {
			// Проверяем, что ключ существует в обоих map
			if _, exists := b.treeCache[key]; !exists {
				// Очищаем мусор из cacheAccessTimes
				delete(b.cacheAccessTimes, key)
				continue
			}
			
			if accessTime.Before(oldestTime) {
				oldestTime = accessTime
				oldestKey = key
			}
		}
		
		if oldestKey == "" {
			// Если не нашли ключ с accessTime, берём любой из treeCache
			for key := range b.treeCache {
				oldestKey = key
				break
			}
		}
		
		if oldestKey == "" {
			break
		}
		
		// Удаляем старую запись
		if cached, exists := b.treeCache[oldestKey]; exists {
			b.cacheSize -= cached.size
		}
		delete(b.treeCache, oldestKey)
		delete(b.cacheAccessTimes, oldestKey)
		
		b.log.Debug("Evicted file tree from cache: " + oldestKey)
	}
}

// GetCacheStats возвращает статистику кэша
func (b *fileTreeBuilder) GetCacheStats() map[string]interface{} {
	b.cacheMutex.RLock()
	defer b.cacheMutex.RUnlock()
	
	return map[string]interface{}{
		"cached_trees":  len(b.treeCache),
		"cache_size_mb": b.cacheSize / (1024 * 1024),
		"cache_hits":    b.cacheHits,
		"cache_misses":  b.cacheMisses,
	}
}

// InvalidateCache clears the tree cache
func (b *fileTreeBuilder) InvalidateCache() {
	b.cacheMutex.Lock()
	defer b.cacheMutex.Unlock()
	
	b.treeCache = make(map[string]*cachedTree)
	b.cacheAccessTimes = make(map[string]time.Time)
	b.cacheSize = 0
	b.cacheHits = 0
	b.cacheMisses = 0
}

// InvalidateCacheForPath clears the cache for a specific path
func (b *fileTreeBuilder) InvalidateCacheForPath(path string) {
	b.cacheMutex.Lock()
	defer b.cacheMutex.Unlock()
	
	if cached, exists := b.treeCache[path]; exists {
		b.cacheSize -= cached.size
	}
	delete(b.treeCache, path)
	delete(b.cacheAccessTimes, path)
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
