// Package analysis provides code analysis services.
package analysis

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"shotgun_code/application/tools"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"shotgun_code/infrastructure/analyzers"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/memory"
	"shotgun_code/infrastructure/projectstructure"
)

// Container manages analysis services with lazy initialization and caching.
// It provides a centralized way to access analysis tools that are expensive to create.
//
// Thread-safe: All methods can be called concurrently.
// Project-scoped: Call SetProject() when switching projects to invalidate caches.
type Container struct {
	mu          sync.RWMutex
	logger      domain.Logger
	projectRoot string
	cacheDir    string

	// Lazy-initialized services
	registry         analysis.AnalyzerRegistry
	symbolIndex      analysis.SymbolIndex
	callGraph        *analyzers.CallGraphBuilderImpl
	gitContext       *git.ContextBuilder
	contextMemory    *memory.ContextMemoryImpl
	projectStructure *projectstructure.Detector
	preferences      *memory.UserPreferences

	// Optional services (may be nil)
	semanticSearch tools.SemanticSearcher

	// Initialization flags
	symbolIndexBuilt bool
	callGraphBuilt   bool
}

// NewContainer creates a new analysis container.
func NewContainer(logger domain.Logger) *Container {
	registry := analyzers.NewAnalyzerRegistry()

	homeDir, err := os.UserHomeDir()
	cacheDir := ""
	if err == nil {
		cacheDir = filepath.Join(homeDir, ".shotgun-code", "analysis")
		_ = os.MkdirAll(cacheDir, 0o755)
	}

	return &Container{
		logger:   logger,
		registry: registry,
		cacheDir: cacheDir,
	}
}

// SetProject sets the current project root and invalidates project-specific caches.
func (c *Container) SetProject(projectRoot string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.projectRoot == projectRoot {
		return
	}

	c.logger.Info("AnalysisContainer: switching project to " + projectRoot)
	c.projectRoot = projectRoot
	c.invalidateCachesLocked()
}

// GetProjectRoot returns the current project root.
func (c *Container) GetProjectRoot() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.projectRoot
}

// GetRegistry returns the analyzer registry.
func (c *Container) GetRegistry() analysis.AnalyzerRegistry {
	return c.registry
}

// GetSymbolIndex returns the symbol index, creating it if needed.
func (c *Container) GetSymbolIndex() analysis.SymbolIndex {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.symbolIndex == nil {
		c.symbolIndex = analyzers.NewSymbolIndex(c.registry)
	}
	return c.symbolIndex
}

// GetCallGraph returns the call graph builder, creating it if needed.
func (c *Container) GetCallGraph() *analyzers.CallGraphBuilderImpl {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.callGraph == nil {
		c.callGraph = analyzers.NewCallGraphBuilder(c.registry)
	}
	return c.callGraph
}

// GetGitContext returns the git context builder for the current project.
func (c *Container) GetGitContext() *git.ContextBuilder {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.projectRoot == "" {
		return nil
	}

	if c.gitContext == nil {
		c.gitContext = git.NewContextBuilder(c.projectRoot)
	}
	return c.gitContext
}

// GetContextMemory returns the context memory service, creating it if needed.
func (c *Container) GetContextMemory() domain.ContextMemory {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.contextMemory == nil {
		cacheDir := c.cacheDir
		if cacheDir == "" {
			cacheDir = os.TempDir()
		}
		contextDir := filepath.Join(cacheDir, "contexts")
		var err error
		c.contextMemory, err = memory.NewContextMemory(contextDir)
		if err != nil {
			c.logger.Warning("Failed to create context memory: " + err.Error())
			return nil
		}
	}
	return c.contextMemory
}

// GetPreferences returns the user preferences service, creating it if needed.
func (c *Container) GetPreferences() *memory.UserPreferences {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.preferences == nil && c.contextMemory != nil {
		c.preferences = memory.NewUserPreferences(c.contextMemory)
	}
	return c.preferences
}

// GetProjectStructure returns the project structure detector, creating it if needed.
func (c *Container) GetProjectStructure() *projectstructure.Detector {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.projectStructure == nil {
		c.projectStructure = projectstructure.NewDetector()
	}
	return c.projectStructure
}

// SetSemanticSearch sets the semantic search service (optional).
func (c *Container) SetSemanticSearch(ss tools.SemanticSearcher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.semanticSearch = ss
}

// GetSemanticSearch returns the semantic search service (may be nil).
func (c *Container) GetSemanticSearch() tools.SemanticSearcher {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.semanticSearch
}

// EnsureSymbolIndexBuilt ensures the symbol index is built for the current project.
func (c *Container) EnsureSymbolIndexBuilt(ctx context.Context) error {
	c.mu.Lock()
	if c.symbolIndexBuilt || c.projectRoot == "" {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	index := c.GetSymbolIndex()
	if err := index.IndexProject(ctx, c.projectRoot); err != nil {
		return err
	}

	c.mu.Lock()
	c.symbolIndexBuilt = true
	c.mu.Unlock()

	return nil
}

// EnsureCallGraphBuilt ensures the call graph is built for the current project.
func (c *Container) EnsureCallGraphBuilt(projectRoot string) error {
	c.mu.Lock()
	if c.callGraphBuilt {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	cg := c.GetCallGraph()
	if _, err := cg.Build(projectRoot); err != nil {
		return err
	}

	c.mu.Lock()
	c.callGraphBuilt = true
	c.mu.Unlock()

	return nil
}

// InvalidateCache invalidates all caches.
func (c *Container) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.invalidateCachesLocked()
}

func (c *Container) invalidateCachesLocked() {
	if c.symbolIndex != nil {
		c.symbolIndex = nil
	}
	c.symbolIndexBuilt = false

	if c.callGraph != nil {
		c.callGraph = nil
	}
	c.callGraphBuilt = false

	c.gitContext = nil
	c.projectStructure = nil
}

// Stats returns statistics about the container state.
func (c *Container) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"projectRoot":       c.projectRoot,
		"symbolIndexBuilt":  c.symbolIndexBuilt,
		"callGraphBuilt":    c.callGraphBuilt,
		"hasGitContext":     c.gitContext != nil,
		"hasContextMemory":  c.contextMemory != nil,
		"hasSemanticSearch": c.semanticSearch != nil,
	}
}
