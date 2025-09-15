package analysis

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/staticanalyzer"
)

// StaticService provides high-level API for static analysis within the analysis bounded context
type StaticService struct {
	log    domain.Logger
	engine domain.StaticAnalyzerEngine
}

// NewStaticService creates a new static analysis service
func NewStaticService(log domain.Logger) *StaticService {
	engine := staticanalyzer.NewStaticAnalyzerEngine(log)

	// Register analyzers for supported languages
	engine.RegisterAnalyzer(staticanalyzer.NewStaticcheckAnalyzer(log))
	engine.RegisterAnalyzer(staticanalyzer.NewESLintAnalyzer(log))
	engine.RegisterAnalyzer(staticanalyzer.NewErrorProneAnalyzer(log))
	engine.RegisterAnalyzer(staticanalyzer.NewRuffAnalyzer(log))
	engine.RegisterAnalyzer(staticanalyzer.NewClangTidyAnalyzer(log))

	return &StaticService{
		log:    log,
		engine: engine,
	}
}

// AnalyzeProject performs analysis of a project
func (s *StaticService) AnalyzeProject(ctx context.Context, projectPath string, languages []string) (*domain.StaticAnalysisReport, error) {
	s.log.Info(fmt.Sprintf("Analyzing project: %s for languages: %v", projectPath, languages))

	// Perform analysis for each language
	results, err := s.engine.AnalyzeProject(ctx, projectPath, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Generate report
	report := s.engine.GenerateReport(results, projectPath)

	s.log.Info(fmt.Sprintf("Analysis completed for %d languages", len(results)))
	return report, nil
}

// AnalyzeFile performs analysis of a single file
func (s *StaticService) AnalyzeFile(ctx context.Context, filePath, language string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing file: %s", filePath))

	config := &domain.StaticAnalyzerConfig{
		Language:     language,
		ProjectPath:  filePath,
		Timeout:      60, // 1 minute timeout for single file
		OutputFormat: "json",
	}

	return s.engine.AnalyzeFile(ctx, filePath, config)
}

// GetSupportedAnalyzers returns supported analyzers
func (s *StaticService) GetSupportedAnalyzers() []domain.StaticAnalyzerType {
	return s.engine.GetSupportedAnalyzers()
}

// GetAnalyzerForLanguage returns analyzer for a language
func (s *StaticService) GetAnalyzerForLanguage(language string) (domain.StaticAnalyzer, error) {
	return s.engine.GetAnalyzerForLanguage(language)
}

// AnalyzeGoProject performs analysis of a Go project
func (s *StaticService) AnalyzeGoProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Go project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "go",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeStaticcheck,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("go")
	if err != nil {
		return nil, fmt.Errorf("no Go analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeTypeScriptProject performs analysis of a TypeScript project
func (s *StaticService) AnalyzeTypeScriptProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing TypeScript project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "typescript",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeESLint,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("typescript")
	if err != nil {
		return nil, fmt.Errorf("no TypeScript analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeJavaScriptProject performs analysis of a JavaScript project
func (s *StaticService) AnalyzeJavaScriptProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing JavaScript project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "javascript",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeESLint,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("javascript")
	if err != nil {
		return nil, fmt.Errorf("no JavaScript analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeJavaProject performs analysis of a Java project
func (s *StaticService) AnalyzeJavaProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Java project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "java",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeErrorProne,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("java")
	if err != nil {
		return nil, fmt.Errorf("no Java analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzePythonProject performs analysis of a Python project
func (s *StaticService) AnalyzePythonProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing Python project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "python",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeRuff,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("python")
	if err != nil {
		return nil, fmt.Errorf("no Python analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// AnalyzeCppProject performs analysis of a C/C++ project
func (s *StaticService) AnalyzeCppProject(ctx context.Context, projectPath string) (*domain.StaticAnalysisResult, error) {
	s.log.Info(fmt.Sprintf("Analyzing C/C++ project: %s", projectPath))

	config := &domain.StaticAnalyzerConfig{
		Language:     "cpp",
		ProjectPath:  projectPath,
		Analyzer:     domain.StaticAnalyzerTypeClangTidy,
		Timeout:      300, // 5 minutes timeout
		OutputFormat: "json",
	}

	analyzer, err := s.engine.GetAnalyzerForLanguage("cpp")
	if err != nil {
		return nil, fmt.Errorf("no C/C++ analyzer available: %w", err)
	}

	return analyzer.Analyze(ctx, config)
}

// ValidateAnalysisResults validates analysis results
func (s *StaticService) ValidateAnalysisResults(results map[string]*domain.StaticAnalysisResult) *domain.StaticAnalysisValidationResult {
	validation := &domain.StaticAnalysisValidationResult{
		TotalLanguages: len(results),
		SuccessCount:   0,
		FailureCount:   0,
		TotalIssues:    0,
		Languages:      make(map[string]*domain.LanguageAnalysisValidation),
	}

	for language, result := range results {
		langValidation := &domain.LanguageAnalysisValidation{
			Language: language,
		}

		if result != nil && result.Success {
			validation.SuccessCount++
			langValidation.Success = true
			langValidation.IssueCount = len(result.Issues)
			validation.TotalIssues += len(result.Issues)
		} else {
			validation.FailureCount++
			langValidation.Success = false
			if result != nil {
				langValidation.Error = result.Error
			}
		}

		validation.Languages[language] = langValidation
	}

	return validation
}

// GetAnalysisMetrics returns analysis metrics
func (s *StaticService) GetAnalysisMetrics(results map[string]*domain.StaticAnalysisResult) *domain.StaticAnalysisMetrics {
	metrics := &domain.StaticAnalysisMetrics{
		TotalLanguages: len(results),
		IssuesByType:   make(map[string]int),
		IssuesByLevel:  make(map[string]int),
	}

	for _, result := range results {
		if result == nil || !result.Success {
			continue
		}

		for _, issue := range result.Issues {
			metrics.IssuesByType[issue.Category]++
			metrics.IssuesByLevel[issue.Severity]++
			metrics.TotalIssues++
		}
	}

	return metrics
}

// FilterIssuesByLevel filters issues by severity level
func (s *StaticService) FilterIssuesByLevel(results map[string]*domain.StaticAnalysisResult, level string) map[string][]*domain.StaticAnalysisIssue {
	filtered := make(map[string][]*domain.StaticAnalysisIssue)

	for language, result := range results {
		if result == nil || !result.Success {
			continue
		}

		var issues []*domain.StaticAnalysisIssue
		for _, issue := range result.Issues {
			if issue.Severity == level {
				staticIssue := &domain.StaticAnalysisIssue{
					ID:       fmt.Sprintf("%s:%d:%d", issue.File, issue.Line, issue.Column),
					Rule:     "static-analysis",
					Severity: issue.Severity,
					Category: issue.Category,
					Message:  issue.Message,
					FilePath: issue.File,
					LineNumber: issue.Line,
				}
				issues = append(issues, staticIssue)
			}
		}

		if len(issues) > 0 {
			filtered[language] = issues
		}
	}

	return filtered
}

// GetCriticalIssues returns critical issues across all languages
func (s *StaticService) GetCriticalIssues(results map[string]*domain.StaticAnalysisResult) []*domain.StaticAnalysisIssue {
	var critical []*domain.StaticAnalysisIssue

	for _, result := range results {
		if result == nil || !result.Success {
			continue
		}

		for _, issue := range result.Issues {
			if issue.Severity == "critical" || issue.Severity == "error" {
				staticIssue := &domain.StaticAnalysisIssue{
					ID:       fmt.Sprintf("%s:%d:%d", issue.File, issue.Line, issue.Column),
					Rule:     "static-analysis",
					Severity: issue.Severity,
					Category: issue.Category,
					Message:  issue.Message,
					FilePath: issue.File,
					LineNumber: issue.Line,
				}
				critical = append(critical, staticIssue)
			}
		}
	}

	return critical
}