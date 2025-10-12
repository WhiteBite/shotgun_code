package application

import (
	"context"
	"shotgun_code/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorAnalyzer_AnalyzeError tests error analysis functionality
func TestErrorAnalyzer_AnalyzeError(t *testing.T) {
	tests := []struct {
		name        string
		errorOutput string
		stage       domain.ProtocolStage
		expected    func(*domain.ErrorDetails) bool
	}{
		{
			name:        "go_compilation_error",
			errorOutput: "main.go:10:5: undefined: fmt",
			stage:       domain.StageBuilding,
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeCompilation &&
					details.SourceFile == "main.go" &&
					details.LineNumber == 10 &&
					details.Column == 5
			},
		},
		{
			name:        "typescript_type_error",
			errorOutput: "src/main.ts:25:10: error TS2304: Cannot find name 'SomeType'",
			stage:       domain.StageBuilding,
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeImport &&
					details.SourceFile == "src/main.ts" &&
					details.LineNumber == 25
			},
		},
		{
			name:        "linting_error",
			errorOutput: "style violation: missing semicolon",
			stage:       domain.StageLinting,
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeLinting &&
					details.Stage == domain.StageLinting
			},
		},
		{
			name:        "test_failure",
			errorOutput: "test failed: expected 5, got 3",
			stage:       domain.StageTesting,
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeTesting &&
					details.Stage == domain.StageTesting
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			analyzer := NewErrorAnalyzer(logger)

			// Execute
			result, err := analyzer.AnalyzeError(tt.errorOutput, tt.stage)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, tt.expected(result), "Error analysis result validation failed")
		})
	}
}

// TestErrorAnalyzer_SuggestCorrections tests correction suggestion functionality
func TestErrorAnalyzer_SuggestCorrections(t *testing.T) {
	tests := []struct {
		name     string
		error    *domain.ErrorDetails
		expected func([]*domain.CorrectionStep) bool
	}{
		{
			name: "import_error_corrections",
			error: &domain.ErrorDetails{
				Stage:      domain.StageBuilding,
				ErrorType:  domain.ErrorTypeImport,
				Message:    "undefined: fmt",
				SourceFile: "main.go",
			},
			expected: func(steps []*domain.CorrectionStep) bool {
				return len(steps) > 0 &&
					steps[0].Action == domain.ActionFixImport
			},
		},
		{
			name: "syntax_error_corrections",
			error: &domain.ErrorDetails{
				Stage:      domain.StageBuilding,
				ErrorType:  domain.ErrorTypeSyntax,
				Message:    "syntax error: missing )",
				SourceFile: "main.go",
			},
			expected: func(steps []*domain.CorrectionStep) bool {
				return len(steps) > 0 &&
					steps[0].Action == domain.ActionFixSyntax
			},
		},
		{
			name: "type_error_corrections",
			error: &domain.ErrorDetails{
				Stage:      domain.StageBuilding,
				ErrorType:  domain.ErrorTypeTypeCheck,
				Message:    "cannot use string as int",
				SourceFile: "main.go",
			},
			expected: func(steps []*domain.CorrectionStep) bool {
				return len(steps) > 0 &&
					steps[0].Action == domain.ActionFixType
			},
		},
		{
			name: "linting_error_corrections",
			error: &domain.ErrorDetails{
				Stage:      domain.StageLinting,
				ErrorType:  domain.ErrorTypeLinting,
				Message:    "formatting issue",
				SourceFile: "main.go",
			},
			expected: func(steps []*domain.CorrectionStep) bool {
				return len(steps) > 0 &&
					steps[0].Action == domain.ActionFormatCode
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			analyzer := NewErrorAnalyzer(logger)

			// Execute
			result, err := analyzer.SuggestCorrections(tt.error)

			// Assert
			assert.NoError(t, err)
			assert.True(t, tt.expected(result), "Correction suggestions validation failed")
		})
	}
}

// TestErrorAnalyzer_ClassifyErrorType tests error type classification
func TestErrorAnalyzer_ClassifyErrorType(t *testing.T) {
	tests := []struct {
		name        string
		errorOutput string
		expected    domain.ErrorType
	}{
		{
			name:        "compilation_error",
			errorOutput: "compile error: syntax error",
			expected:    domain.ErrorTypeCompilation,
		},
		{
			name:        "type_check_error",
			errorOutput: "type error: cannot use string as int",
			expected:    domain.ErrorTypeTypeCheck,
		},
		{
			name:        "import_error",
			errorOutput: "import error: module not found",
			expected:    domain.ErrorTypeImport,
		},
		{
			name:        "linting_error",
			errorOutput: "lint error: style violation",
			expected:    domain.ErrorTypeLinting,
		},
		{
			name:        "test_error",
			errorOutput: "test failed: assertion error",
			expected:    domain.ErrorTypeTesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			analyzer := NewErrorAnalyzer(logger)

			// Execute
			result := analyzer.ClassifyErrorType(tt.errorOutput)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCorrectionEngine_ApplyCorrection tests single correction application
func TestCorrectionEngine_ApplyCorrection(t *testing.T) {
	tests := []struct {
		name      string
		step      *domain.CorrectionStep
		mockSetup func(*MockFileSystemProvider)
		expected  func(*domain.CorrectionResult) bool
	}{
		{
			name: "format_code_correction",
			step: &domain.CorrectionStep{
				Action:      domain.ActionFormatCode,
				Target:      "main.go",
				Description: "Format Go file",
			},
			mockSetup: func(mock *MockFileSystemProvider) {
				// Mock file system interactions
			},
			expected: func(result *domain.CorrectionResult) bool {
				return result.Success &&
					result.Message != ""
			},
		},
		{
			name: "import_fix_correction",
			step: &domain.CorrectionStep{
				Action:      domain.ActionFixImport,
				Target:      "main.go",
				Description: "Fix import statement",
			},
			mockSetup: func(mock *MockFileSystemProvider) {
				mock.readFileContent = "package main\n\nfunc main() {}"
			},
			expected: func(result *domain.CorrectionResult) bool {
				return result.Success
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			fileSystemProvider := &MockFileSystemProvider{}
			tt.mockSetup(fileSystemProvider)

			engine := NewCorrectionEngine(logger, fileSystemProvider)

			// Execute
			result, err := engine.ApplyCorrection(context.Background(), tt.step, "/test/project")

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, tt.expected(result), "Correction result validation failed")
		})
	}
}

// TestCorrectionEngine_ApplyCorrections tests multiple correction application
func TestCorrectionEngine_ApplyCorrections(t *testing.T) {
	tests := []struct {
		name      string
		steps     []*domain.CorrectionStep
		mockSetup func(*MockFileSystemProvider)
		expected  func(*domain.CorrectionResult) bool
	}{
		{
			name: "multiple_corrections_success",
			steps: []*domain.CorrectionStep{
				{
					Action:      domain.ActionFixImport,
					Target:      "main.go",
					Description: "Fix import",
				},
				{
					Action:      domain.ActionFormatCode,
					Target:      "main.go",
					Description: "Format code",
				},
			},
			mockSetup: func(mock *MockFileSystemProvider) {
				mock.readFileContent = "package main\n\nfunc main() {}"
			},
			expected: func(result *domain.CorrectionResult) bool {
				return result.Success &&
					len(result.FilesChanged) > 0 &&
					result.Message != ""
			},
		},
		{
			name: "mixed_success_failure",
			steps: []*domain.CorrectionStep{
				{
					Action:      domain.ActionFixImport,
					Target:      "main.go",
					Description: "Fix import",
				},
				{
					Action:      domain.CorrectionAction("unsupported"),
					Target:      "main.go",
					Description: "Unsupported action",
				},
			},
			mockSetup: func(mock *MockFileSystemProvider) {
				mock.readFileContent = "package main\n\nfunc main() {}"
			},
			expected: func(result *domain.CorrectionResult) bool {
				return !result.Success // Should fail due to unsupported action
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			fileSystemProvider := &MockFileSystemProvider{}
			tt.mockSetup(fileSystemProvider)

			engine := NewCorrectionEngine(logger, fileSystemProvider)

			// Execute
			result, err := engine.ApplyCorrections(context.Background(), tt.steps, "/test/project")

			// Assert
			if tt.expected(result) {
				assert.NoError(t, err)
			}
			assert.NotNil(t, result)
			assert.True(t, tt.expected(result), "Multiple corrections result validation failed")
		})
	}
}

// TestCorrectionEngine_CanHandle tests error handling capability check
func TestCorrectionEngine_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		error    *domain.ErrorDetails
		expected bool
	}{
		{
			name: "import_error_can_handle",
			error: &domain.ErrorDetails{
				ErrorType: domain.ErrorTypeImport,
			},
			expected: true,
		},
		{
			name: "syntax_error_can_handle",
			error: &domain.ErrorDetails{
				ErrorType: domain.ErrorTypeSyntax,
			},
			expected: true,
		},
		{
			name: "type_error_can_handle",
			error: &domain.ErrorDetails{
				ErrorType: domain.ErrorTypeTypeCheck,
			},
			expected: true,
		},
		{
			name: "linting_error_can_handle",
			error: &domain.ErrorDetails{
				ErrorType: domain.ErrorTypeLinting,
			},
			expected: true,
		},
		{
			name: "unsupported_error_cannot_handle",
			error: &domain.ErrorDetails{
				ErrorType: domain.ErrorType("unsupported"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &TestLogger{}
			fileSystemProvider := &MockFileSystemProvider{}
			engine := NewCorrectionEngine(logger, fileSystemProvider)

			// Execute
			result := engine.CanHandle(tt.error)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Language-specific analyzer tests

func TestGoErrorAnalyzer(t *testing.T) {
	analyzer := NewGoErrorAnalyzer()

	tests := []struct {
		name        string
		errorOutput string
		expected    func(*domain.ErrorDetails) bool
	}{
		{
			name:        "go_undefined_error",
			errorOutput: "undefined: fmt",
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeCompilation &&
					details.Tool == "go" &&
					len(details.Suggestions) > 0
			},
		},
		{
			name:        "go_type_error",
			errorOutput: "cannot use string as int in assignment",
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeTypeCheck &&
					details.Tool == "go" &&
					len(details.Suggestions) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := analyzer.AnalyzeError(tt.errorOutput)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, tt.expected(result), "Go error analysis validation failed")
			assert.Equal(t, "go", analyzer.GetLanguage())
		})
	}
}

func TestTypeScriptErrorAnalyzer(t *testing.T) {
	analyzer := NewTypeScriptErrorAnalyzer()

	tests := []struct {
		name        string
		errorOutput string
		expected    func(*domain.ErrorDetails) bool
	}{
		{
			name:        "typescript_cannot_find_name",
			errorOutput: "error TS2304: Cannot find name 'SomeType'",
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeImport &&
					details.Tool == "tsc" &&
					len(details.Suggestions) > 0
			},
		},
		{
			name:        "typescript_type_assignment",
			errorOutput: "error TS2322: Type 'string' is not assignable to type 'number'",
			expected: func(details *domain.ErrorDetails) bool {
				return details.ErrorType == domain.ErrorTypeTypeCheck &&
					details.Tool == "tsc" &&
					len(details.Suggestions) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := analyzer.AnalyzeError(tt.errorOutput)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, tt.expected(result), "TypeScript error analysis validation failed")
			assert.Equal(t, "typescript", analyzer.GetLanguage())
		})
	}
}

// Integration tests

func TestErrorAnalyzerAndCorrectionEngine_Integration(t *testing.T) {
	// Setup
	logger := &TestLogger{}
	analyzer := NewErrorAnalyzer(logger)
	fileSystemProvider := &MockFileSystemProvider{}
	fileSystemProvider.readFileContent = "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"
	engine := NewCorrectionEngine(logger, fileSystemProvider)

	// Simulate a compilation error
	errorOutput := "main.go:4:2: undefined: fmt"
	stage := domain.StageBuilding

	// Analyze error
	errorDetails, err := analyzer.AnalyzeError(errorOutput, stage)
	assert.NoError(t, err)
	assert.NotNil(t, errorDetails)

	// Get correction suggestions
	corrections, err := analyzer.SuggestCorrections(errorDetails)
	assert.NoError(t, err)
	assert.True(t, len(corrections) > 0)

	// Check if engine can handle the error
	canHandle := engine.CanHandle(errorDetails)
	assert.True(t, canHandle)

	// Apply corrections
	result, err := engine.ApplyCorrections(context.Background(), corrections, "/test/project")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	t.Logf("Integration test completed: error analyzed, corrections suggested, and applied")
}

// Enhanced MockFileSystemProvider with configurable behavior
type MockFileSystemProvider struct {
	readFileContent string
	writeError      error
	mkdirError      error
}

func (m *MockFileSystemProvider) ReadFile(filename string) ([]byte, error) {
	if m.readFileContent != "" {
		return []byte(m.readFileContent), nil
	}
	return []byte("mock content"), nil
}

func (m *MockFileSystemProvider) WriteFile(filename string, data []byte, perm int) error {
	return m.writeError
}

func (m *MockFileSystemProvider) MkdirAll(path string, perm int) error {
	return m.mkdirError
}
