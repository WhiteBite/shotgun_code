package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
)

// ContextGenerationService is responsible for generating the final context string from project files.
type ContextGenerationService struct {
	log        domain.Logger
	bus        domain.EventBus
	fileReader domain.FileContentReader
}

// NewContextGenerationService creates a new ContextGenerationService.
func NewContextGenerationService(log domain.Logger, bus domain.EventBus, reader domain.FileContentReader) *ContextGenerationService {
	return &ContextGenerationService{
		log:        log,
		bus:        bus,
		fileReader: reader,
	}
}

// Generate builds the context string from the given file paths and emits progress events.
func (s *ContextGenerationService) Generate(ctx context.Context, rootDir string, includedPaths []string) {
	s.log.Info(fmt.Sprintf("Starting context generation for %d files.", len(includedPaths)))

	progressCallback := func(current, total int64) {
		progressData := map[string]int64{"current": current, "total": total}
		s.bus.Emit("shotgunContextGenerationProgress", progressData)
	}

	fileContents, err := s.fileReader.ReadContents(ctx, includedPaths, rootDir, progressCallback)
	if err != nil {
		if err == context.Canceled {
			s.log.Info("Context generation was canceled.")
			return
		}
		s.log.Error(fmt.Sprintf("Failed to read file contents: %v", err))
		s.bus.Emit("app:error", fmt.Sprintf("Context generation failed: %v", err))
		return
	}

	var contextBuilder strings.Builder
	for path, content := range fileContents {
		select {
		case <-ctx.Done():
			s.log.Info("Context generation was canceled during string building.")
			return
		default:
		}
		contextBuilder.WriteString("--- File: " + path + " ---\n")
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n")
	}

	finalContext := contextBuilder.String()
	s.log.Info(fmt.Sprintf("Context generation complete. Total size: %d bytes.", len(finalContext)))
	s.bus.Emit("shotgunContextGenerated", finalContext)
}
