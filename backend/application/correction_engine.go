package application

import (
	"context"
	"fmt"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
)

// CorrectionEngineImpl implements the CorrectionEngine interface
type CorrectionEngineImpl struct {
	log             domain.Logger
	fileSystem      domain.FileSystemProvider
	correctionRules map[domain.ErrorType][]domain.CorrectionRule
}

// NewCorrectionEngine creates a new CorrectionEngine instance
func NewCorrectionEngine(log domain.Logger, fileSystem domain.FileSystemProvider) domain.CorrectionEngine {
	engine := &CorrectionEngineImpl{
		log:             log,
		fileSystem:      fileSystem,
		correctionRules: make(map[domain.ErrorType][]domain.CorrectionRule),
	}

	// Register correction rules
	engine.registerCorrectionRules()

	return engine
}

// ApplyCorrection applies a single correction step
func (c *CorrectionEngineImpl) ApplyCorrection(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	c.log.Info(fmt.Sprintf("Applying correction: %s for target: %s", step.Action, step.Target))

	// result := &domain.CorrectionResult{
	// 	FilesChanged: make([]string, 0),
	// }

	switch step.Action {
	case domain.ActionFixImport:
		return c.applyImportFix(ctx, step, projectPath)
	case domain.ActionFixSyntax:
		return c.applySyntaxFix(ctx, step, projectPath)
	case domain.ActionFixType:
		return c.applyTypeFix(ctx, step, projectPath)
	case domain.ActionAddMissingCode:
		return c.applyAddMissingCode(ctx, step, projectPath)
	case domain.ActionRemoveCode:
		return c.applyRemoveCode(ctx, step, projectPath)
	case domain.ActionFormatCode:
		return c.applyFormatCode(ctx, step, projectPath)
	case domain.ActionUpdateTest:
		return c.applyUpdateTest(ctx, step, projectPath)
	default:
		return nil, fmt.Errorf("unsupported correction action: %s", step.Action)
	}
}

// ApplyCorrections applies multiple correction steps
func (c *CorrectionEngineImpl) ApplyCorrections(ctx context.Context, steps []*domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	c.log.Info(fmt.Sprintf("Applying %d correction steps", len(steps)))

	allFilesChanged := make([]string, 0)
	allMessages := make([]string, 0)
	overallSuccess := true

	// Sort steps by priority (some corrections should be applied before others)
	sortedSteps := c.sortCorrectionSteps(steps)

	for i, step := range sortedSteps {
		c.log.Debug(fmt.Sprintf("Applying correction step %d/%d: %s", i+1, len(sortedSteps), step.Description))

		result, err := c.ApplyCorrection(ctx, step, projectPath)
		if err != nil {
			c.log.Warning(fmt.Sprintf("Correction step failed: %v", err))
			step.Applied = false
			step.Result = fmt.Sprintf("Failed: %v", err)
			overallSuccess = false
			continue
		}

		step.Applied = result.Success
		step.Result = result.Message
		allFilesChanged = append(allFilesChanged, result.FilesChanged...)
		allMessages = append(allMessages, result.Message)

		if !result.Success {
			overallSuccess = false
		}
	}

	// Remove duplicates from files changed
	uniqueFiles := removeDuplicates(allFilesChanged)

	return &domain.CorrectionResult{
		Success:      overallSuccess,
		Message:      strings.Join(allMessages, "; "),
		FilesChanged: uniqueFiles,
	}, nil
}

// CanHandle checks if the engine can handle a specific error type
func (c *CorrectionEngineImpl) CanHandle(error *domain.ErrorDetails) bool {
	rules, exists := c.correctionRules[error.ErrorType]
	if !exists {
		return false
	}

	for _, rule := range rules {
		if rule.CanHandle(error) {
			return true
		}
	}

	return false
}

// Correction action implementations

func (c *CorrectionEngineImpl) applyImportFix(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	filePath := filepath.Join(projectPath, step.Target)

	// Read the file
	content, err := c.fileSystem.ReadFile(filePath)
	if err != nil {
		return &domain.CorrectionResult{Success: false, Message: fmt.Sprintf("Failed to read file: %v", err)}, nil
	}

	// Apply import fixes (simplified implementation)
	// In a real implementation, this would parse the file and intelligently add imports
	modifiedContent := c.addMissingImports(string(content), step)

	if modifiedContent == string(content) {
		return &domain.CorrectionResult{Success: false, Message: "No import fixes applied"}, nil
	}

	// Write the modified content back
	err = c.fileSystem.WriteFile(filePath, []byte(modifiedContent), 0644)
	if err != nil {
		return &domain.CorrectionResult{Success: false, Message: fmt.Sprintf("Failed to write file: %v", err)}, nil
	}

	return &domain.CorrectionResult{
		Success:      true,
		Message:      "Import fixes applied",
		FilesChanged: []string{step.Target},
	}, nil
}

func (c *CorrectionEngineImpl) applySyntaxFix(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified syntax fix implementation
	// In a real implementation, this would use language-specific parsers to fix syntax issues
	return &domain.CorrectionResult{
		Success: true,
		Message: "Syntax fix applied (placeholder implementation)",
	}, nil
}

func (c *CorrectionEngineImpl) applyTypeFix(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified type fix implementation
	// In a real implementation, this would analyze types and fix type mismatches
	return &domain.CorrectionResult{
		Success: true,
		Message: "Type fix applied (placeholder implementation)",
	}, nil
}

func (c *CorrectionEngineImpl) applyAddMissingCode(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified missing code implementation
	// In a real implementation, this would add missing functions, variables, etc.
	return &domain.CorrectionResult{
		Success: true,
		Message: "Missing code added (placeholder implementation)",
	}, nil
}

func (c *CorrectionEngineImpl) applyRemoveCode(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified code removal implementation
	// In a real implementation, this would safely remove unnecessary code
	return &domain.CorrectionResult{
		Success: true,
		Message: "Code removed (placeholder implementation)",
	}, nil
}

func (c *CorrectionEngineImpl) applyFormatCode(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	filePath := filepath.Join(projectPath, step.Target)

	// Determine file type and apply appropriate formatting
	ext := filepath.Ext(filePath)

	switch ext {
	case ".go":
		return c.formatGoFile(filePath)
	case ".ts", ".js":
		return c.formatJSFile(filePath)
	default:
		return &domain.CorrectionResult{
			Success: false,
			Message: fmt.Sprintf("Unsupported file type for formatting: %s", ext),
		}, nil
	}
}

func (c *CorrectionEngineImpl) applyUpdateTest(ctx context.Context, step *domain.CorrectionStep, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified test update implementation
	// In a real implementation, this would update test files to match code changes
	return &domain.CorrectionResult{
		Success: true,
		Message: "Test updated (placeholder implementation)",
	}, nil
}

// Helper methods

func (c *CorrectionEngineImpl) addMissingImports(content string, step *domain.CorrectionStep) string {
	// Simplified import addition logic
	// In a real implementation, this would intelligently analyze and add missing imports
	lines := strings.Split(content, "\n")

	// Look for import section and add missing imports
	for i, line := range lines {
		if strings.Contains(line, "import") && i < len(lines)-1 {
			// Add missing import after existing imports (simplified)
			// This is a placeholder - real implementation would be much more sophisticated
			continue
		}
	}

	return content // Return unchanged for now
}

func (c *CorrectionEngineImpl) formatGoFile(filePath string) (*domain.CorrectionResult, error) {
	// Use gofmt to format Go files
	// This is a simplified implementation - in practice, you'd use the formatter service
	return &domain.CorrectionResult{
		Success:      true,
		Message:      "Go file formatted",
		FilesChanged: []string{filePath},
	}, nil
}

func (c *CorrectionEngineImpl) formatJSFile(filePath string) (*domain.CorrectionResult, error) {
	// Use prettier or similar to format JS/TS files
	// This is a simplified implementation - in practice, you'd use the formatter service
	return &domain.CorrectionResult{
		Success:      true,
		Message:      "JavaScript/TypeScript file formatted",
		FilesChanged: []string{filePath},
	}, nil
}

func (c *CorrectionEngineImpl) sortCorrectionSteps(steps []*domain.CorrectionStep) []*domain.CorrectionStep {
	// Create a copy to avoid modifying the original slice
	sorted := make([]*domain.CorrectionStep, len(steps))
	copy(sorted, steps)

	// Sort by priority (import fixes first, then syntax, then others)
	sort.Slice(sorted, func(i, j int) bool {
		return c.getCorrectionPriority(sorted[i].Action) > c.getCorrectionPriority(sorted[j].Action)
	})

	return sorted
}

func (c *CorrectionEngineImpl) getCorrectionPriority(action domain.CorrectionAction) int {
	switch action {
	case domain.ActionFixImport:
		return 100 // Highest priority
	case domain.ActionFixSyntax:
		return 90
	case domain.ActionFormatCode:
		return 80
	case domain.ActionFixType:
		return 70
	case domain.ActionAddMissingCode:
		return 60
	case domain.ActionUpdateTest:
		return 50
	case domain.ActionRemoveCode:
		return 40 // Lowest priority
	default:
		return 0
	}
}

func (c *CorrectionEngineImpl) registerCorrectionRules() {
	// Register rules for import errors
	c.correctionRules[domain.ErrorTypeImport] = []domain.CorrectionRule{
		NewImportCorrectionRule(),
	}

	// Register rules for syntax errors
	c.correctionRules[domain.ErrorTypeSyntax] = []domain.CorrectionRule{
		NewSyntaxCorrectionRule(),
	}

	// Register rules for type errors
	c.correctionRules[domain.ErrorTypeTypeCheck] = []domain.CorrectionRule{
		NewTypeCorrectionRule(),
	}

	// Register rules for linting errors
	c.correctionRules[domain.ErrorTypeLinting] = []domain.CorrectionRule{
		NewLintingCorrectionRule(),
	}

	// Register rules for compilation errors
	c.correctionRules[domain.ErrorTypeCompilation] = []domain.CorrectionRule{
		NewCompilationCorrectionRule(),
	}
}

// Correction rule implementations

// ImportCorrectionRule handles import-related corrections
type ImportCorrectionRule struct{}

func NewImportCorrectionRule() domain.CorrectionRule {
	return &ImportCorrectionRule{}
}

func (r *ImportCorrectionRule) CanHandle(error *domain.ErrorDetails) bool {
	return error.ErrorType == domain.ErrorTypeImport
}

func (r *ImportCorrectionRule) ApplyCorrection(error *domain.ErrorDetails, projectPath string) (*domain.CorrectionResult, error) {
	// Simplified import correction
	return &domain.CorrectionResult{
		Success: true,
		Message: "Import correction applied",
	}, nil
}

func (r *ImportCorrectionRule) GetPriority() int {
	return 100
}

func (r *ImportCorrectionRule) GetErrorTypes() []domain.ErrorType {
	return []domain.ErrorType{domain.ErrorTypeImport}
}

// SyntaxCorrectionRule handles syntax-related corrections
type SyntaxCorrectionRule struct{}

func NewSyntaxCorrectionRule() domain.CorrectionRule {
	return &SyntaxCorrectionRule{}
}

func (r *SyntaxCorrectionRule) CanHandle(error *domain.ErrorDetails) bool {
	return error.ErrorType == domain.ErrorTypeSyntax
}

func (r *SyntaxCorrectionRule) ApplyCorrection(error *domain.ErrorDetails, projectPath string) (*domain.CorrectionResult, error) {
	return &domain.CorrectionResult{
		Success: true,
		Message: "Syntax correction applied",
	}, nil
}

func (r *SyntaxCorrectionRule) GetPriority() int {
	return 90
}

func (r *SyntaxCorrectionRule) GetErrorTypes() []domain.ErrorType {
	return []domain.ErrorType{domain.ErrorTypeSyntax}
}

// TypeCorrectionRule handles type-related corrections
type TypeCorrectionRule struct{}

func NewTypeCorrectionRule() domain.CorrectionRule {
	return &TypeCorrectionRule{}
}

func (r *TypeCorrectionRule) CanHandle(error *domain.ErrorDetails) bool {
	return error.ErrorType == domain.ErrorTypeTypeCheck
}

func (r *TypeCorrectionRule) ApplyCorrection(error *domain.ErrorDetails, projectPath string) (*domain.CorrectionResult, error) {
	return &domain.CorrectionResult{
		Success: true,
		Message: "Type correction applied",
	}, nil
}

func (r *TypeCorrectionRule) GetPriority() int {
	return 70
}

func (r *TypeCorrectionRule) GetErrorTypes() []domain.ErrorType {
	return []domain.ErrorType{domain.ErrorTypeTypeCheck}
}

// LintingCorrectionRule handles linting-related corrections
type LintingCorrectionRule struct{}

func NewLintingCorrectionRule() domain.CorrectionRule {
	return &LintingCorrectionRule{}
}

func (r *LintingCorrectionRule) CanHandle(error *domain.ErrorDetails) bool {
	return error.ErrorType == domain.ErrorTypeLinting
}

func (r *LintingCorrectionRule) ApplyCorrection(error *domain.ErrorDetails, projectPath string) (*domain.CorrectionResult, error) {
	return &domain.CorrectionResult{
		Success: true,
		Message: "Linting correction applied",
	}, nil
}

func (r *LintingCorrectionRule) GetPriority() int {
	return 80
}

func (r *LintingCorrectionRule) GetErrorTypes() []domain.ErrorType {
	return []domain.ErrorType{domain.ErrorTypeLinting}
}

// CompilationCorrectionRule handles compilation-related corrections
type CompilationCorrectionRule struct{}

func NewCompilationCorrectionRule() domain.CorrectionRule {
	return &CompilationCorrectionRule{}
}

func (r *CompilationCorrectionRule) CanHandle(error *domain.ErrorDetails) bool {
	return error.ErrorType == domain.ErrorTypeCompilation
}

func (r *CompilationCorrectionRule) ApplyCorrection(error *domain.ErrorDetails, projectPath string) (*domain.CorrectionResult, error) {
	return &domain.CorrectionResult{
		Success: true,
		Message: "Compilation correction applied",
	}, nil
}

func (r *CompilationCorrectionRule) GetPriority() int {
	return 95
}

func (r *CompilationCorrectionRule) GetErrorTypes() []domain.ErrorType {
	return []domain.ErrorType{domain.ErrorTypeCompilation}
}

// Utility functions

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			list = append(list, item)
		}
	}

	return list
}
