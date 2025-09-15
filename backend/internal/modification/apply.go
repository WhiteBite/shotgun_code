package modification

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
)

// ApplyService handles code edit application with multiple strategies
type ApplyService struct {
	applyEngine domain.ApplyEngine
	formatter   domain.Formatter
	log         domain.Logger
}

// NewApplyService creates a new apply service
func NewApplyService(
	applyEngine domain.ApplyEngine,
	formatter domain.Formatter,
	log domain.Logger,
) *ApplyService {
	return &ApplyService{
		applyEngine: applyEngine,
		formatter:   formatter,
		log:         log,
	}
}

// ApplyEdits applies a list of edits to files
func (s *ApplyService) ApplyEdits(ctx context.Context, edits []domain.Edit) error {
	s.log.Info(fmt.Sprintf("Applying %d edits", len(edits)))

	for i, edit := range edits {
		s.log.Info(fmt.Sprintf("Applying edit %d/%d to file: %s", i+1, len(edits), edit.FilePath))
		
		if err := s.applyEngine.ApplyEdit(ctx, edit); err != nil {
			return fmt.Errorf("failed to apply edit to %s: %w", edit.FilePath, err)
		}

		// Format the file after applying the edit if it's a supported language
		if s.shouldFormat(edit.FilePath) {
			if err := s.formatter.FormatFile(ctx, edit.FilePath); err != nil {
				s.log.Warning(fmt.Sprintf("Failed to format file %s: %v", edit.FilePath, err))
				// Don't fail the entire operation just because formatting failed
			} else {
				s.log.Info(fmt.Sprintf("Formatted file: %s", edit.FilePath))
			}
		}
	}

	s.log.Info("All edits applied successfully")
	return nil
}

// ValidateEdits validates that edits can be applied without conflicts
func (s *ApplyService) ValidateEdits(ctx context.Context, edits []domain.Edit) error {
	s.log.Info(fmt.Sprintf("Validating %d edits", len(edits)))

	for i, edit := range edits {
		s.log.Info(fmt.Sprintf("Validating edit %d/%d for file: %s", i+1, len(edits), edit.FilePath))
		
		if err := s.validateEdit(edit); err != nil {
			return fmt.Errorf("validation failed for edit to %s: %w", edit.FilePath, err)
		}
	}

	s.log.Info("All edits validated successfully")
	return nil
}

// RollbackEdits attempts to rollback previously applied edits
func (s *ApplyService) RollbackEdits(ctx context.Context, edits []domain.Edit) error {
	s.log.Info(fmt.Sprintf("Rolling back %d edits", len(edits)))

	// Rollback in reverse order
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		s.log.Info(fmt.Sprintf("Rolling back edit %d/%d for file: %s", len(edits)-i, len(edits), edit.FilePath))
		
		if err := s.rollbackEdit(ctx, edit); err != nil {
			s.log.Error(fmt.Sprintf("Failed to rollback edit for %s: %v", edit.FilePath, err))
			// Continue with other rollbacks even if one fails
		}
	}

	s.log.Info("Rollback completed")
	return nil
}

// shouldFormat determines if a file should be formatted based on its extension
func (s *ApplyService) shouldFormat(filePath string) bool {
	supportedExtensions := []string{".go", ".ts", ".tsx", ".js", ".jsx", ".json"}
	
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(strings.ToLower(filePath), ext) {
			return true
		}
	}
	
	return false
}

// validateEdit performs basic validation on an edit
func (s *ApplyService) validateEdit(edit domain.Edit) error {
	if edit.FilePath == "" {
		return fmt.Errorf("edit has empty file path")
	}

	if edit.Type == "" {
		return fmt.Errorf("edit has empty type")
	}

	switch edit.Type {
	case domain.EditTypeReplace:
		if edit.OldContent == "" {
			return fmt.Errorf("replace edit requires non-empty old content")
		}
		if edit.NewContent == edit.OldContent {
			return fmt.Errorf("replace edit has identical old and new content")
		}
	case domain.EditTypeInsert:
		if edit.NewContent == "" {
			return fmt.Errorf("insert edit requires non-empty new content")
		}
		if edit.Position < 0 {
			return fmt.Errorf("insert edit requires valid position")
		}
	case domain.EditTypeDelete:
		if edit.OldContent == "" {
			return fmt.Errorf("delete edit requires non-empty old content")
		}
	default:
		return fmt.Errorf("unsupported edit type: %s", edit.Type)
	}

	return nil
}

// rollbackEdit attempts to rollback a single edit
func (s *ApplyService) rollbackEdit(ctx context.Context, edit domain.Edit) error {
	// Create inverse edit for rollback
	var inverseEdit domain.Edit
	
	switch edit.Type {
	case domain.EditTypeReplace:
		// For replace, swap old and new content
		inverseEdit = domain.Edit{
			FilePath:    edit.FilePath,
			Type:        domain.EditTypeReplace,
			OldContent:  edit.NewContent,
			NewContent:  edit.OldContent,
			Position:    edit.Position,
		}
	case domain.EditTypeInsert:
		// For insert, create delete edit
		inverseEdit = domain.Edit{
			FilePath:   edit.FilePath,
			Type:       domain.EditTypeDelete,
			OldContent: edit.NewContent,
			Position:   edit.Position,
		}
	case domain.EditTypeDelete:
		// For delete, create insert edit
		inverseEdit = domain.Edit{
			FilePath:    edit.FilePath,
			Type:        domain.EditTypeInsert,
			NewContent:  edit.OldContent,
			Position:    edit.Position,
		}
	default:
		return fmt.Errorf("cannot rollback unsupported edit type: %s", edit.Type)
	}

	return s.applyEngine.ApplyEdit(ctx, inverseEdit)
}