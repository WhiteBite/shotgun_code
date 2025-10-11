package analysis

import (
	"context"
	"errors"
	"shotgun_code/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockStaticLogger struct {
	mock.Mock
}

func (m *MockStaticLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockStaticLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockStaticLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockStaticLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockStaticLogger) Fatal(msg string) {
	m.Called(msg)
}

// Mock StaticAnalyzerEngine for testing
type MockStaticAnalyzerEngine struct {
	mock.Mock
}

func (m *MockStaticAnalyzerEngine) RegisterAnalyzer(analyzer domain.StaticAnalyzer) {}

func (m *MockStaticAnalyzerEngine) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (map[string]*domain.StaticAnalysisResult, error) {
	args := m.Called(ctx, projectPath, languages)
	return args.Get(0).(map[string]*domain.StaticAnalysisResult), args.Error(1)
}

func (m *MockStaticAnalyzerEngine) AnalyzeFile(ctx context.Context, filePath string, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	args := m.Called(ctx, filePath, config)
	return args.Get(0).(*domain.StaticAnalysisResult), args.Error(1)
}

func (m *MockStaticAnalyzerEngine) GenerateReport(results map[string]*domain.StaticAnalysisResult, projectPath string) *domain.StaticAnalysisReport {
	args := m.Called(results, projectPath)
	return args.Get(0).(*domain.StaticAnalysisReport)
}

func (m *MockStaticAnalyzerEngine) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	args := m.Called()
	return args.Get(0).([]domain.StaticAnalyzerType)
}

func (m *MockStaticAnalyzerEngine) GetAnalyzerForLanguage(language string) (domain.StaticAnalyzer, error) {
    args := m.Called(language)
    var analyzer domain.StaticAnalyzer
    if v := args.Get(0); v != nil {
        analyzer = v.(domain.StaticAnalyzer)
    }
    return analyzer, args.Error(1)
}

// Mock StaticAnalyzer for testing
type MockStaticAnalyzer struct {
	mock.Mock
}

func (m *MockStaticAnalyzer) Analyze(ctx context.Context, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(*domain.StaticAnalysisResult), args.Error(1)
}

func (m *MockStaticAnalyzer) GetAnalyzerType() domain.StaticAnalyzerType {
	args := m.Called()
	return args.Get(0).(domain.StaticAnalyzerType)
}

func (m *MockStaticAnalyzer) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockStaticAnalyzer) GetSupportedLanguages() []string {
    args := m.Called()
    return args.Get(0).([]string)
}

func (m *MockStaticAnalyzer) ValidateConfig(config *domain.StaticAnalyzerConfig) error {
    args := m.Called(config)
    return args.Error(0)
}

func TestStaticService_NewStaticService(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)

	// Execute
	service := NewStaticService(mockLogger)

	// Assert
	assert.NotNil(t, service)
	assert.NotNil(t, service.engine)
}

func TestStaticService_AnalyzeProject_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	projectPath := "/test/project"
	languages := []string{"go", "javascript"}
	
    analysisResults := map[string]*domain.StaticAnalysisResult{
        "go": {
            Language: "go",
            Issues:   []*domain.StaticIssue{},
            Success:  true,
        },
        "javascript": {
            Language: "javascript",
            Issues:   []*domain.StaticIssue{},
            Success:  true,
        },
    }
	
    report := &domain.StaticAnalysisReport{
        ProjectPath: projectPath,
        Results:     analysisResults,
        Summary: &domain.StaticAnalysisReportSummary{
            TotalIssues:   0,
            TotalErrors:   0,
            TotalWarnings: 0,
            LanguagesAnalyzed: []string{"go", "javascript"},
            AnalyzersUsed:     []string{"staticcheck", "eslint"},
            Success:           true,
        },
    }

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("AnalyzeProject", mock.Anything, projectPath, languages).Return(analysisResults, nil)
	mockEngine.On("GenerateReport", analysisResults, projectPath).Return(report)

	// Execute
	ctx := context.Background()
	result, err := service.AnalyzeProject(ctx, projectPath, languages)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectPath, result.ProjectPath)
	assert.Equal(t, 2, len(result.Results))

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestStaticService_AnalyzeProject_Error(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	projectPath := "/test/project"
	languages := []string{"go"}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("AnalyzeProject", mock.Anything, projectPath, languages).Return((map[string]*domain.StaticAnalysisResult)(nil), errors.New("analysis failed"))

	// Execute
	ctx := context.Background()
	result, err := service.AnalyzeProject(ctx, projectPath, languages)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to analyze project")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestStaticService_AnalyzeFile_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	filePath := "/test/file.go"
	language := "go"
	
    result := &domain.StaticAnalysisResult{
        Language: language,
        Issues:   []*domain.StaticIssue{},
        Success:  true,
    }

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("AnalyzeFile", mock.Anything, filePath, mock.AnythingOfType("*domain.StaticAnalyzerConfig")).Return(result, nil)

	// Execute
	ctx := context.Background()
	analysisResult, err := service.AnalyzeFile(ctx, filePath, language)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, analysisResult)
	assert.Equal(t, language, analysisResult.Language)
	assert.True(t, analysisResult.Success)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestStaticService_GetSupportedAnalyzers(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	supportedAnalyzers := []domain.StaticAnalyzerType{
		domain.StaticAnalyzerTypeStaticcheck,
		domain.StaticAnalyzerTypeESLint,
		domain.StaticAnalyzerTypeErrorProne,
	}

	// Setup mocks
	mockEngine.On("GetSupportedAnalyzers").Return(supportedAnalyzers)

	// Execute
	result := service.GetSupportedAnalyzers()

	// Assert
	assert.Equal(t, supportedAnalyzers, result)
	assert.Len(t, result, 3)

	mockEngine.AssertExpectations(t)
}

func TestStaticService_GetAnalyzerForLanguage_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	language := "go"
	mockAnalyzer := new(MockStaticAnalyzer)

	// Setup mocks
	mockEngine.On("GetAnalyzerForLanguage", language).Return(mockAnalyzer, nil)

	// Execute
	result, err := service.GetAnalyzerForLanguage(language)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockAnalyzer, result)

	mockEngine.AssertExpectations(t)
}

func TestStaticService_GetAnalyzerForLanguage_Error(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	language := "unknown"

	// Setup mocks
	mockEngine.On("GetAnalyzerForLanguage", language).Return((domain.StaticAnalyzer)(nil), errors.New("analyzer not found"))

	// Execute
	result, err := service.GetAnalyzerForLanguage(language)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "analyzer not found")

	mockEngine.AssertExpectations(t)
}

func TestStaticService_AnalyzeGoProject_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	mockAnalyzer := new(MockStaticAnalyzer)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	projectPath := "/test/go-project"
    result := &domain.StaticAnalysisResult{
        Language: "go",
        Issues:   []*domain.StaticIssue{},
        Success:  true,
    }

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GetAnalyzerForLanguage", "go").Return(mockAnalyzer, nil)
	mockAnalyzer.On("Analyze", mock.Anything, mock.AnythingOfType("*domain.StaticAnalyzerConfig")).Return(result, nil)

	// Execute
	ctx := context.Background()
	analysisResult, err := service.AnalyzeGoProject(ctx, projectPath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, analysisResult)
	assert.Equal(t, "go", analysisResult.Language)
	assert.True(t, analysisResult.Success)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockAnalyzer.AssertExpectations(t)
}

func TestStaticService_AnalyzeGoProject_NoAnalyzer(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	projectPath := "/test/go-project"

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GetAnalyzerForLanguage", "go").Return((domain.StaticAnalyzer)(nil), errors.New("no Go analyzer available"))

	// Execute
	ctx := context.Background()
	result, err := service.AnalyzeGoProject(ctx, projectPath)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no Go analyzer available")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestStaticService_AnalyzeTypeScriptProject_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	mockEngine := new(MockStaticAnalyzerEngine)
	mockAnalyzer := new(MockStaticAnalyzer)
	
	// Create service with mock engine
	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	projectPath := "/test/ts-project"
    result := &domain.StaticAnalysisResult{
        Language: "typescript",
        Issues:   []*domain.StaticIssue{},
        Success:  true,
    }

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GetAnalyzerForLanguage", "typescript").Return(mockAnalyzer, nil)
	mockAnalyzer.On("Analyze", mock.Anything, mock.AnythingOfType("*domain.StaticAnalyzerConfig")).Return(result, nil)

	// Execute
	ctx := context.Background()
	analysisResult, err := service.AnalyzeTypeScriptProject(ctx, projectPath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, analysisResult)
	assert.Equal(t, "typescript", analysisResult.Language)
	assert.True(t, analysisResult.Success)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockAnalyzer.AssertExpectations(t)
}

func TestStaticService_ValidateAnalysisResults(t *testing.T) {
	// Setup
	mockLogger := new(MockStaticLogger)
	service := NewStaticService(mockLogger)

	// Test data
    results := map[string]*domain.StaticAnalysisResult{
        "go": {
            Language: "go",
            Issues: []*domain.StaticIssue{
                {Severity: "critical"},
                {Severity: "high"},
                {Severity: "medium"},
            },
            Success: true,
        },
        "javascript": {
            Language: "javascript",
            Issues: []*domain.StaticIssue{
                {Severity: "low"},
                {Severity: "medium"},
            },
            Success: false,
        },
    }

	// Execute
	validation := service.ValidateAnalysisResults(results)

	// Assert
	assert.NotNil(t, validation)
	assert.Equal(t, 2, validation.TotalLanguages)
	assert.Equal(t, 1, validation.SuccessCount)
	assert.Equal(t, 1, validation.FailureCount)
    assert.Equal(t, 5, validation.TotalIssues)
}