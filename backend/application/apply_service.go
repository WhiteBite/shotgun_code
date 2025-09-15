package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
)

// ApplyService предоставляет высокоуровневый API для применения правок
type ApplyService struct {
	log    domain.Logger
	engine domain.ApplyEngine
	config *domain.ApplyEngineConfig
}

// NewApplyService создает новый сервис применения
func NewApplyService(
	log domain.Logger,
	config *domain.ApplyEngineConfig,
	engine domain.ApplyEngine,
	formatters map[string]domain.Formatter,
	importFixers map[string]domain.ImportFixer,
) *ApplyService {
	// Register formatters and import fixers with the engine
	for lang, formatter := range formatters {
		engine.RegisterFormatter(lang, formatter)
	}
	
	for lang, fixer := range importFixers {
		engine.RegisterImportFixer(lang, fixer)
	}

	return &ApplyService{
		log:    log,
		engine: engine,
		config: config,
	}
}

// ApplyEdits применяет правки из Edits JSON
func (s *ApplyService) ApplyEdits(ctx context.Context, edits *domain.EditsJSON) ([]*domain.ApplyResult, error) {
	s.log.Info(fmt.Sprintf("Applying %d edits", len(edits.Edits)))

	var operations []*domain.ApplyOperation

	// Конвертируем Edits JSON в операции
	for _, edit := range edits.Edits {
		op := &domain.ApplyOperation{
			ID:        edit.ID,
			Path:      edit.Path,
			Language:  edit.Language,
			Strategy:  domain.ApplyStrategy(edit.Kind),
			Operation: edit.Op,
			Content:   edit.Content,
		}

		// Добавляем якоря для anchor стратегии
		if edit.Kind == "anchorPatch" {
			if anchor, ok := edit.Anchor.(map[string]interface{}); ok {
				if before, ok := anchor["before"].(string); ok {
					op.AnchorBefore = before
				}
				if after, ok := anchor["after"].(string); ok {
					op.AnchorAfter = after
				}
			}
		}

		operations = append(operations, op)
	}

	// Применяем операции
	results, err := s.engine.ApplyOperations(ctx, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to apply operations: %w", err)
	}

	// Логируем результаты
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			s.log.Info(fmt.Sprintf("Successfully applied operation %s to %s", result.OperationID, result.Path))
		} else {
			s.log.Error(fmt.Sprintf("Failed to apply operation %s to %s: %s", result.OperationID, result.Path, result.Error))
		}
	}

	s.log.Info(fmt.Sprintf("Applied %d/%d operations successfully", successCount, len(results)))
	return results, nil
}

// ApplySingleEdit применяет одну правку
func (s *ApplyService) ApplySingleEdit(ctx context.Context, edit *domain.Edit) (*domain.ApplyResult, error) {
	op := &domain.ApplyOperation{
		ID:        edit.ID,
		Path:      edit.Path,
		Language:  edit.Language,
		Strategy:  domain.ApplyStrategy(edit.Kind),
		Operation: edit.Op,
		Content:   edit.Content,
	}

	// Добавляем якоря для anchor стратегии
	if edit.Kind == "anchorPatch" {
		if anchor, ok := edit.Anchor.(map[string]interface{}); ok {
			if before, ok := anchor["before"].(string); ok {
				op.AnchorBefore = before
			}
			if after, ok := anchor["after"].(string); ok {
				op.AnchorAfter = after
			}
		}
	}

	return s.engine.ApplyOperation(ctx, op)
}

// ValidateEdits проверяет корректность правок
func (s *ApplyService) ValidateEdits(ctx context.Context, edits *domain.EditsJSON) error {
	for _, edit := range edits.Edits {
		op := &domain.ApplyOperation{
			ID:        edit.ID,
			Path:      edit.Path,
			Language:  edit.Language,
			Strategy:  domain.ApplyStrategy(edit.Kind),
			Operation: edit.Op,
			Content:   edit.Content,
		}

		if err := s.engine.ValidateOperation(ctx, op); err != nil {
			return fmt.Errorf("validation failed for edit %s: %w", edit.ID, err)
		}
	}

	return nil
}

// RollbackEdits откатывает правки
func (s *ApplyService) RollbackEdits(ctx context.Context, results []*domain.ApplyResult) error {
	for i := len(results) - 1; i >= 0; i-- { // Откатываем в обратном порядке
		result := results[i]
		if result.Success {
			if err := s.engine.RollbackOperation(ctx, result); err != nil {
				s.log.Error(fmt.Sprintf("Failed to rollback operation %s: %v", result.OperationID, err))
				return err
			}
		}
	}

	return nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (s *ApplyService) GetSupportedLanguages() []string {
	return s.config.Languages
}

// GetConfig возвращает конфигурацию
func (s *ApplyService) GetConfig() *domain.ApplyEngineConfig {
	return s.config
}

// UpdateConfig обновляет конфигурацию
func (s *ApplyService) UpdateConfig(config *domain.ApplyEngineConfig) {
	s.config = config
	s.log.Info("Updated apply engine configuration")
}
