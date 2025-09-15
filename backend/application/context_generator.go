package application

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// ContextGenerator handles asynchronous context generation with progress tracking
type ContextGenerator struct {
	fileReader domain.FileContentReader
	logger     domain.Logger
	bus        domain.EventBus
	contextDir string
}

// NewContextGenerator creates a new ContextGenerator
func NewContextGenerator(
	fileReader domain.FileContentReader,
	logger domain.Logger,
	bus domain.EventBus,
	contextDir string,
) *ContextGenerator {
	return &ContextGenerator{
		fileReader: fileReader,
		logger:     logger,
		bus:        bus,
		contextDir: contextDir,
	}
}

// GenerateContext builds a context asynchronously with progress tracking
func (cg *ContextGenerator) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	// Run in separate goroutine with panic recovery
	go cg.generateContextSafe(ctx, rootDir, includedPaths)
}

func (cg *ContextGenerator) generateContextSafe(ctx context.Context, rootDir string, includedPaths []string) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			cg.logger.Error(fmt.Sprintf("PANIC recovered in GenerateContext: %v\nStack: %s", r, stack))
			if cg.bus != nil {
				cg.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", r))
				cg.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", r))
			}
		}
	}()

	// Send generation start event
	if cg.bus != nil {
		cg.bus.Emit("shotgunContextGenerationStarted", map[string]interface{}{
			"fileCount": len(includedPaths),
			"rootDir":   rootDir,
		})
	}

	// Ensure we always have a non-nil context to avoid panics
	if ctx == nil {
		ctx = context.Background()
	}

	// Add timeout to prevent infinite loading
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cg.logger.Info(fmt.Sprintf("Starting context generation for %d files", len(includedPaths)))

	// Sort paths for deterministic processing
	sortedPaths := make([]string, len(includedPaths))
	copy(sortedPaths, includedPaths)
	sort.Strings(sortedPaths)

	// Use errgroup for controlled concurrency
	g, gctx := errgroup.WithContext(ctx)

	// Read file contents with progress tracking
	var contents map[string]string
	var readErr error

	g.Go(func() error {
		var err error
		contents, err = cg.fileReader.ReadContents(gctx, sortedPaths, rootDir, func(current, total int64) {
			select {
			case <-gctx.Done():
				return
			default:
				if cg.bus != nil {
					cg.bus.Emit("shotgunContextGenerationProgress", map[string]interface{}{
						"current": current,
						"total":   total,
					})
				}
			}
		})
		readErr = err
		return err
	})

	if err := g.Wait(); err != nil {
		if err == context.DeadlineExceeded {
			cg.logger.Error("Context generation timed out")
			if cg.bus != nil {
				cg.bus.Emit("app:error", "Context generation timed out after 30 seconds")
				cg.bus.Emit("shotgunContextGenerationTimeout")
			}
		} else {
			cg.logger.Error(fmt.Sprintf("Failed to read file contents: %v", err))
			if cg.bus != nil {
				cg.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", err))
				cg.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", err))
			}
		}
		return
	}

	if readErr != nil {
		cg.logger.Error(fmt.Sprintf("Failed to read file contents: %v", readErr))
		if cg.bus != nil {
			cg.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", readErr))
			cg.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", readErr))
		}
		return
	}

	// Build context string deterministically
	var contextBuilder strings.Builder

	// Add manifest header
	contextBuilder.WriteString("Manifest:\n")
	manifestPaths := make([]string, 0, len(contents))
	for path := range contents {
		manifestPaths = append(manifestPaths, path)
	}
	sort.Strings(manifestPaths) // Ensure deterministic order

	// Build simple tree structure
	contextBuilder.WriteString(cg.buildSimpleTree(manifestPaths))
	contextBuilder.WriteString("\n")

	// Add file contents in sorted order
	for _, relPath := range manifestPaths {
		content, exists := contents[relPath]
		if !exists {
			continue
		}

		contextBuilder.WriteString(fmt.Sprintf("--- File: %s ---\n", relPath))
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n")
	}

	finalContext := strings.TrimSpace(contextBuilder.String())
	cg.logger.Info(fmt.Sprintf("Context generation completed. Length: %d characters", len(finalContext)))

	cg.logger.Info("Emitting shotgunContextGenerated event")
	if cg.bus != nil {
		cg.bus.Emit("shotgunContextGenerated", finalContext)
	}
	cg.logger.Info("Event emitted successfully")
}

type treeNode struct {
	name     string
	children map[string]*treeNode
	isFile   bool
}

func (cg *ContextGenerator) buildSimpleTree(paths []string) string {
	root := &treeNode{name: ".", children: make(map[string]*treeNode)}

	// Build tree structure
	for _, path := range paths {
		parts := strings.Split(filepath.ToSlash(path), "/")
		current := root

		for i, part := range parts {
			if part == "" || part == "." {
				continue
			}

			if _, exists := current.children[part]; !exists {
				current.children[part] = &treeNode{
					name:     part,
					children: make(map[string]*treeNode),
					isFile:   i == len(parts)-1, // Last part is file
				}
			}
			current = current.children[part]
		}
	}

	// Generate tree string
	var builder strings.Builder
	cg.walkTree(root, "", true, &builder)
	return builder.String()
}

func (cg *ContextGenerator) walkTree(node *treeNode, prefix string, isLast bool, builder *strings.Builder) {
	if node.name != "." {
		if isLast {
			builder.WriteString(prefix + "└─ " + node.name + "\n")
			prefix += "   "
		} else {
			builder.WriteString(prefix + "├─ " + node.name + "\n")
			prefix += "│  "
		}
	}

	// Sort children for deterministic output
	childNames := make([]string, 0, len(node.children))
	for name := range node.children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	for i, name := range childNames {
		child := node.children[name]
		cg.walkTree(child, prefix, i == len(childNames)-1, builder)
	}
}