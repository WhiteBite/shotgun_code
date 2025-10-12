package modification

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/diffengine"
)

// DiffService provides high-level API for diff operations
type DiffService struct {
	log    domain.Logger
	engine domain.DiffEngine
}

// NewDiffService creates a new diff service
func NewDiffService(log domain.Logger) *DiffService {
	engine := diffengine.NewDiffEngine(log)

	return &DiffService{
		log:    log,
		engine: engine,
	}
}

// GenerateDiff generates diff between two states
func (s *DiffService) GenerateDiff(ctx context.Context, beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	s.log.Info(fmt.Sprintf("Generating diff between %s and %s", beforePath, afterPath))

	return s.engine.GenerateDiff(ctx, beforePath, afterPath, format)
}

// GenerateDiffFromResults generates diff from apply results
func (s *DiffService) GenerateDiffFromResults(ctx context.Context, results []*domain.ApplyResult, format domain.DiffFormat) (*domain.DiffResult, error) {
	s.log.Info(fmt.Sprintf("Generating diff from %d apply results", len(results)))

	return s.engine.GenerateDiffFromResults(ctx, results, format)
}

// GenerateDiffFromEdits generates diff from Edits JSON
func (s *DiffService) GenerateDiffFromEdits(ctx context.Context, edits *domain.EditsJSON, format domain.DiffFormat) (*domain.DiffResult, error) {
	s.log.Info(fmt.Sprintf("Generating diff from %d edits", len(edits.Edits)))

	return s.engine.GenerateDiffFromEdits(ctx, edits, format)
}

// PublishDiff publishes a diff
func (s *DiffService) PublishDiff(ctx context.Context, diff *domain.DiffResult) error {
	s.log.Info(fmt.Sprintf("Publishing diff %s", diff.ID))

	return s.engine.PublishDiff(ctx, diff)
}

// GenerateAndPublishDiff generates and publishes diff
func (s *DiffService) GenerateAndPublishDiff(ctx context.Context, beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	// Generate diff
	diff, err := s.GenerateDiff(ctx, beforePath, afterPath, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff: %w", err)
	}

	// Publish diff
	if err := s.PublishDiff(ctx, diff); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to publish diff: %v", err))
	}

	return diff, nil
}

// GenerateAndPublishDiffFromResults generates and publishes diff from results
func (s *DiffService) GenerateAndPublishDiffFromResults(ctx context.Context, results []*domain.ApplyResult, format domain.DiffFormat) (*domain.DiffResult, error) {
	// Generate diff
	diff, err := s.GenerateDiffFromResults(ctx, results, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff from results: %w", err)
	}

	// Publish diff
	if err := s.PublishDiff(ctx, diff); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to publish diff: %v", err))
	}

	return diff, nil
}

// GenerateAndPublishDiffFromEdits generates and publishes diff from edits
func (s *DiffService) GenerateAndPublishDiffFromEdits(ctx context.Context, edits *domain.EditsJSON, format domain.DiffFormat) (*domain.DiffResult, error) {
	// Generate diff
	diff, err := s.GenerateDiffFromEdits(ctx, edits, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff from edits: %w", err)
	}

	// Publish diff
	if err := s.PublishDiff(ctx, diff); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to publish diff: %v", err))
	}

	return diff, nil
}

// GetSupportedFormats returns supported diff formats
func (s *DiffService) GetSupportedFormats() []domain.DiffFormat {
	return []domain.DiffFormat{
		domain.DiffFormatGit,
		domain.DiffFormatUnified,
		domain.DiffFormatJSON,
		domain.DiffFormatHTML,
	}
}
