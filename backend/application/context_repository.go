package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
)

// ContextRepositoryImpl implements the ContextRepository interface
type ContextRepositoryImpl struct {
	logger     domain.Logger
	contextDir string
}

// NewContextRepository creates a new ContextRepository implementation
func NewContextRepository(logger domain.Logger, contextDir string) *ContextRepositoryImpl {
	return &ContextRepositoryImpl{
		logger:     logger,
		contextDir: contextDir,
	}
}

// GetContext retrieves a context by ID
func (cr *ContextRepositoryImpl) GetContext(ctx context.Context, contextID string) (*domain.Context, error) {
	contextPath := filepath.Join(cr.contextDir, contextID+".json")
	
	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("context not found: %s", contextID)
		}
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}
	
	var context domain.Context
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}
	
	return &context, nil
}

// GetProjectContexts lists all contexts for a project
func (cr *ContextRepositoryImpl) GetProjectContexts(ctx context.Context, projectPath string) ([]*domain.Context, error) {
	entries, err := os.ReadDir(cr.contextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}
	
	var contexts []*domain.Context
	
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		
		contextID := strings.TrimSuffix(entry.Name(), ".json")
		context, err := cr.GetContext(ctx, contextID)
		if err != nil {
			cr.logger.Warning(fmt.Sprintf("Failed to load context %s: %v", contextID, err))
			continue
		}
		
		if context.ProjectPath == projectPath {
			contexts = append(contexts, context)
		}
	}
	
	return contexts, nil
}

// DeleteContext deletes a context by ID
func (cr *ContextRepositoryImpl) DeleteContext(ctx context.Context, contextID string) error {
	contextPath := filepath.Join(cr.contextDir, contextID+".json")
	
	if err := os.Remove(contextPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("context not found: %s", contextID)
		}
		return fmt.Errorf("failed to delete context file: %w", err)
	}
	
	cr.logger.Info(fmt.Sprintf("Deleted context %s", contextID))
	return nil
}

// SaveContext saves a context to disk
func (cr *ContextRepositoryImpl) SaveContext(context *domain.Context) error {
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}
	
	contextPath := filepath.Join(cr.contextDir, context.ID+".json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}
	
	return nil
}

// SaveContextSummary saves a context summary to disk
func (cr *ContextRepositoryImpl) SaveContextSummary(contextSummary *domain.ContextSummary) error {
	data, err := json.MarshalIndent(contextSummary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context summary: %w", err)
	}
	
	contextPath := filepath.Join(cr.contextDir, contextSummary.ID+".json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context summary file: %w", err)
	}
	
	return nil
}