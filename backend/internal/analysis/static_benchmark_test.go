package analysis

import (
	"context"
	"shotgun_code/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for benchmarking
type mockStaticLogger struct{}

func (m *mockStaticLogger) Debug(msg string)   {}
func (m *mockStaticLogger) Info(msg string)    {}
func (m *mockStaticLogger) Warning(msg string) {}
func (m *mockStaticLogger) Error(msg string)   {}
func (m *mockStaticLogger) Fatal(msg string)   {}

// Mock StaticAnalyzerEngine for benchmarking
type mockStaticAnalyzerEngine struct {
	delayMs int
}

func (m *mockStaticAnalyzerEngine) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (map[string]*domain.StaticAnalysisResult, error) {
	// Simulate analysis delay
	results := make(map[string]*domain.StaticAnalysisResult)
	for _, lang := range languages {
		results[lang] = &domain.StaticAnalysisResult{
			Language: lang,
			Issues: []*domain.StaticIssue{
				{
					FilePath:    "test.go",
					LineNumber:  10,
					ColumnStart: 5,
					Severity:    domain.GuardrailSeverityMedium,
					Message:     "Test issue",
					Rule:        "test-code",
				},
			},
			Success: true,
		}
	}
	return results, nil
}

func (m *mockStaticAnalyzerEngine) AnalyzeFile(ctx context.Context, filePath string, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	return &domain.StaticAnalysisResult{
		Language: config.Language,
		Issues: []*domain.StaticIssue{
			{
				FilePath:    filePath,
				LineNumber:  15,
				ColumnStart: 10,
				Severity:    domain.GuardrailSeverityLow,
				Message:     "File issue",
				Rule:        "file-code",
			},
		},
		Success: true,
	}, nil
}

func (m *mockStaticAnalyzerEngine) GenerateReport(results map[string]*domain.StaticAnalysisResult, projectPath string) *domain.StaticAnalysisReport {
	return &domain.StaticAnalysisReport{
		ProjectPath: projectPath,
		Results:     results,
		Summary: &domain.StaticAnalysisReportSummary{
			TotalIssues:   len(results),
			TotalErrors:   0,
			TotalWarnings: len(results),
		},
	}
}

func (m *mockStaticAnalyzerEngine) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	return []domain.StaticAnalyzerType{
		domain.StaticAnalyzerTypeStaticcheck,
		domain.StaticAnalyzerTypeESLint,
		domain.StaticAnalyzerTypeErrorProne,
	}
}

func (m *mockStaticAnalyzerEngine) GetAnalyzerForLanguage(language string) (domain.StaticAnalyzer, error) {
	return &mockStaticAnalyzer{language: language}, nil
}

// Mock StaticAnalyzer for benchmarking
type mockStaticAnalyzer struct {
	language string
}

func (m *mockStaticAnalyzer) Analyze(ctx context.Context, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	return &domain.StaticAnalysisResult{
		Language: m.language,
		Issues:   []*domain.StaticIssue{},
		Success:  true,
	}, nil
}

func (m *mockStaticAnalyzer) GetAnalyzerType() domain.StaticAnalyzerType {
	return domain.StaticAnalyzerTypeStaticcheck
}

func (m *mockStaticAnalyzer) GetSupportedLanguages() []string {
	return []string{m.language}
}

func (m *mockStaticAnalyzer) ValidateConfig(config *domain.StaticAnalyzerConfig) error {
	return nil
}

func (m *mockStaticAnalyzerEngine) RegisterAnalyzer(analyzer domain.StaticAnalyzer) {
	// No-op for mock
}

func BenchmarkStaticService_AnalyzeProject(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	mockEngine := &mockStaticAnalyzerEngine{delayMs: 10}

	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	projectPath := "/test/project"
	languages := []string{"go", "javascript", "python"}

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := service.AnalyzeProject(ctx, projectPath, languages)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStaticService_AnalyzeFile(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	mockEngine := &mockStaticAnalyzerEngine{delayMs: 5}

	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	filePath := "/test/file.go"
	language := "go"

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := service.AnalyzeFile(ctx, filePath, language)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStaticService_AnalyzeGoProject(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	mockEngine := &mockStaticAnalyzerEngine{delayMs: 15}

	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	projectPath := "/test/go-project"

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := service.AnalyzeGoProject(ctx, projectPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStaticService_AnalyzeTypeScriptProject(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	mockEngine := &mockStaticAnalyzerEngine{delayMs: 15}

	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	projectPath := "/test/ts-project"

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := service.AnalyzeTypeScriptProject(ctx, projectPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStaticService_GetSupportedAnalyzers(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	mockEngine := &mockStaticAnalyzerEngine{}

	service := &StaticService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := service.GetSupportedAnalyzers()
		assert.NotNil(b, result)
	}
}

func BenchmarkStaticService_ValidateAnalysisResults(b *testing.B) {
	// Setup
	mockLogger := &mockStaticLogger{}
	service := NewStaticService(mockLogger)

	// Test data
	results := map[string]*domain.StaticAnalysisResult{
		"go": {
			Language: "go",
			Issues: []*domain.StaticIssue{
				{Severity: domain.GuardrailSeverityHigh},
				{Severity: domain.GuardrailSeverityMedium},
				{Severity: domain.GuardrailSeverityLow},
			},
			Success: true,
		},
		"javascript": {
			Language: "javascript",
			Issues: []*domain.StaticIssue{
				{Severity: domain.GuardrailSeverityLow},
				{Severity: domain.GuardrailSeverityMedium},
			},
			Success: false,
		},
		"python": {
			Language: "python",
			Issues: []*domain.StaticIssue{
				{Severity: domain.GuardrailSeverityHigh},
				{Severity: domain.GuardrailSeverityMedium},
				{Severity: domain.GuardrailSeverityLow},
			},
			Success: true,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validation := service.ValidateAnalysisResults(results)
		assert.NotNil(b, validation)
	}
}
