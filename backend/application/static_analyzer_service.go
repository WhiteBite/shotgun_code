package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
)

// StaticAnalyzerService предоставляет высокоуровневый API для статического анализа
type StaticAnalyzerService struct {
	log    domain.Logger
	engine domain.StaticAnalyzerEngine
}

// NewStaticAnalyzerService создает новый сервис статического анализа
func NewStaticAnalyzerService(log domain.Logger, engine domain.StaticAnalyzerEngine) *StaticAnalyzerService {
	return &StaticAnalyzerService{
		log:    log,
		engine: engine,
	}
}

// AnalyzeProject выполняет анализ проекта
func (s *StaticAnalyzerService) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (*domain.StaticAnalysisReport, error) {
	s.log.Info(fmt.Sprintf("Analyzing project: %s for languages: %v", projectPath, languages))

	// Выполняем анализ для каждого языка
	results, err := s.engine.AnalyzeProject(ctx, projectPath, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Генерируем отчет
	report := s.engine.GenerateReport(results, projectPath)

	s.log.Info(fmt.Sprintf("Analysis completed for %d languages", len(results)))
	return report, nil
}

// AnalyzeFile выполняет анализ одного файла
func (s *StaticAnalyzerService) AnalyzeFile(ctx context.Context, filePath, language string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing file: %s", filePath))

	config := &domain.StaticAnalyzerConfig{
		Language:     language,
		ProjectPath:  filePath,
		Timeout:      60, // 1 минута таймаут для одного файла
		OutputFormat: "json",
	}

	return s.engine.AnalyzeFile(ctx, filePath, config)
}

// GetSupportedAnalyzers возвращает поддерживаемые анализаторы
func (s *StaticAnalyzerService) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	return s.engine.GetSupportedAnalyzers()
}

// GetAnalyzerForLanguage возвращает анализатор для языка
func (s *StaticAnalyzerService) GetAnalyzerForLanguage(language string) (domain.StaticAnalyzer, error) {
	return s.engine.GetAnalyzerForLanguage(language)
}

// AnalyzeGoProject выполняет анализ Go проекта
func (s *StaticAnalyzerService) AnalyzeGoProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Go project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "go",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeStaticcheck,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("go")
	if err != nil {
		return nil, fmt.Errorf("no Go analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeTypeScriptProject выполняет анализ TypeScript проекта
func (s *StaticAnalyzerService) AnalyzeTypeScriptProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing TypeScript project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "typescript",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeESLint,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("typescript")
	if err != nil {
		return nil, fmt.Errorf("no TypeScript analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeJavaScriptProject выполняет анализ JavaScript проекта
func (s *StaticAnalyzerService) AnalyzeJavaScriptProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing JavaScript project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "javascript",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeESLint,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("javascript")
	if err != nil {
		return nil, fmt.Errorf("no JavaScript analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeJavaProject выполняет анализ Java проекта
func (s *StaticAnalyzerService) AnalyzeJavaProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Java project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "java",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeErrorProne,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("java")
	if err != nil {
		return nil, fmt.Errorf("no Java analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzePythonProject выполняет анализ Python проекта
func (s *StaticAnalyzerService) AnalyzePythonProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Python project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "python",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeRuff,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("python")
	if err != nil {
		return nil, fmt.Errorf("no Python analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeCppProject выполняет анализ C/C++ проекта
func (s *StaticAnalyzerService) AnalyzeCppProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing C/C++ project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "cpp",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeClangTidy,
		Timeout:      300, // 5 минут таймаут
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("cpp")
	if err != nil {
		return nil, fmt.Errorf("no C/C++ analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// ValidateAnalysisResults валидирует результаты анализа
func (s *StaticAnalyzerService) ValidateAnalysisResults(results map[string]*domain.StaticAnalysisResult) *domain.StaticAnalysisValidationResult {
	validation := &domain.StaticAnalysisValidationResult{
		TotalLanguages:  len(results),
		SuccessCount:    0,
		FailureCount:    0,
		TotalIssues:     0,
		TotalErrors:     0,
		TotalWarnings:   0,
		FailedLanguages: []string{},
		CriticalIssues:  []*domain.StaticIssue{},
	}

	for language, result := range results {
		if result.Success {
			validation.SuccessCount++
			if result.Summary != nil {
				validation.TotalIssues += result.Summary.TotalIssues
				validation.TotalErrors += result.Summary.ErrorCount
				validation.TotalWarnings += result.Summary.WarningCount

				// Добавляем критические проблемы
				for _, issue := range result.Issues {
					if issue.Severity == "error" {
						validation.CriticalIssues = append(validation.CriticalIssues, issue)
					}
				}
			}
		} else {
			validation.FailureCount++
			validation.FailedLanguages = append(validation.FailedLanguages, language)
		}
	}

	validation.Success = validation.FailureCount == 0 && validation.TotalErrors == 0
	validation.SuccessRate = float64(validation.SuccessCount) / float64(validation.TotalLanguages) * 100

	return validation
}