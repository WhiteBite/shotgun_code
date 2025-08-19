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

type ContextGenerationService struct {
	log        domain.Logger
	bus        domain.EventBus
	fileReader domain.FileContentReader
}

func NewContextGenerationService(
	log domain.Logger,
	bus domain.EventBus,
	fileReader domain.FileContentReader,
) *ContextGenerationService {
	return &ContextGenerationService{
		log:        log,
		bus:        bus,
		fileReader: fileReader,
	}
}

type treeNode struct {
	name     string
	children map[string]*treeNode
	isFile   bool
}

func (s *ContextGenerationService) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	// Запускаем в отдельной горутине с panic recovery
	go s.generateContextSafe(ctx, rootDir, includedPaths)
}

func (s *ContextGenerationService) generateContextSafe(ctx context.Context, rootDir string, includedPaths []string) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			s.log.Error(fmt.Sprintf("PANIC recovered in GenerateContext: %v\nStack: %s", r, stack))
			s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", r))
			s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", r))
		}
	}()

	// Отправляем событие начала генерации
	s.bus.Emit("shotgunContextGenerationStarted", map[string]interface{}{
		"fileCount": len(includedPaths),
		"rootDir":   rootDir,
	})

	// Ensure we always have a non-nil context to avoid panics in downstream calls
	if ctx == nil {
		ctx = context.Background()
	}

	// Добавляем таймаут для предотвращения бесконечной загрузки
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.log.Info(fmt.Sprintf("Starting context generation for %d files", len(includedPaths)))

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
		contents, err = s.fileReader.ReadContents(gctx, sortedPaths, rootDir, func(current, total int64) {
			select {
			case <-gctx.Done():
				return
			default:
				s.bus.Emit("shotgunContextGenerationProgress", map[string]interface{}{
					"current": current,
					"total":   total,
				})
			}
		})
		readErr = err
		return err
	})

	if err := g.Wait(); err != nil {
		if err == context.DeadlineExceeded {
			s.log.Error("Context generation timed out")
			s.bus.Emit("app:error", "Context generation timed out after 30 seconds")
			s.bus.Emit("shotgunContextGenerationTimeout")
		} else {
			s.log.Error(fmt.Sprintf("Failed to read file contents: %v", err))
			s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", err))
			s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", err))
		}
		return
	}

	if readErr != nil {
		s.log.Error(fmt.Sprintf("Failed to read file contents: %v", readErr))
		s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", readErr))
		s.bus.Emit("shotgunContextGenerationFailed", fmt.Sprintf("%v", readErr))
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
	contextBuilder.WriteString(s.buildSimpleTree(manifestPaths))
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
	s.log.Info(fmt.Sprintf("Context generation completed. Length: %d characters", len(finalContext)))

	s.log.Info("Emitting shotgunContextGenerated event")
	s.bus.Emit("shotgunContextGenerated", finalContext)
	s.log.Info("Event emitted successfully")
}

func (s *ContextGenerationService) buildSimpleTree(paths []string) string {
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
	s.walkTree(root, "", true, &builder)
	return builder.String()
}

func (s *ContextGenerationService) walkTree(node *treeNode, prefix string, isLast bool, builder *strings.Builder) {
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

	for i, childName := range childNames {
		child := node.children[childName]
		s.walkTree(child, prefix, i == len(childNames)-1, builder)
	}
}
