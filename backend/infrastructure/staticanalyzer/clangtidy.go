package staticanalyzer

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"time"
)

// ClangTidyAnalyzer реализует StaticAnalyzer для C/C++ с использованием ClangTidy
type ClangTidyAnalyzer struct {
	log domain.Logger
}

// NewClangTidyAnalyzer создает новый анализатор для C/C++
func NewClangTidyAnalyzer(log domain.Logger) *ClangTidyAnalyzer {
	return &ClangTidyAnalyzer{
		log: log,
	}
}

// Analyze выполняет статический анализ C/C++ кода
func (a *ClangTidyAnalyzer) Analyze(ctx context.Context, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	a.log.Info(fmt.Sprintf("Running ClangTidy analysis for project: %s", config.ProjectPath))

	startTime := time.Now()

	// Проверяем, что ClangTidy установлен
	if err := a.checkClangTidyInstalled(); err != nil {
		a.log.Warning(fmt.Sprintf("ClangTidy not available: %v", err))
		// Возвращаем успешный результат с предупреждением вместо ошибки
		return &domain.StaticAnalysisResult{
			Success:     true,
			Language:    config.Language,
			ProjectPath: config.ProjectPath,
			Analyzer:    config.Analyzer,
			Duration:    time.Since(startTime).Seconds(),
			Issues:      []*domain.StaticAnalysisIssue{},
			Summary:     &domain.StaticAnalysisSummary{},
			Error:       fmt.Sprintf("ClangTidy not available: %v", err),
		}, nil
	}

	// Строим команду для ClangTidy
	args := []string{"--format-style=json"}

	// Добавляем конфигурационный файл если указан
	if config.ConfigFilePath != "" {
		args = append(args, "--config-file="+config.ConfigFilePath)
	}

	// Добавляем правила если указаны
	if len(config.Rules) > 0 {
		args = append(args, "-checks="+strings.Join(config.Rules, ","))
	}

	// Добавляем исключения если указаны
	if len(config.ExcludeRules) > 0 {
		args = append(args, "-checks=-"+strings.Join(config.ExcludeRules, ","))
	}

	// Добавляем путь к проекту
	args = append(args, config.ProjectPath)

	// Создаем команду
	cmd := exec.CommandContext(ctx, "clang-tidy", args...)
	cmd.Dir = config.ProjectPath

	// Устанавливаем переменные окружения
	if config.EnvVars != nil {
		for key, value := range config.EnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Запускаем команду
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime).Seconds()

	result := &domain.StaticAnalysisResult{
		Success:     true,
		Language:    config.Language,
		ProjectPath: config.ProjectPath,
		Analyzer:    config.Analyzer,
		Duration:    duration,
		Issues:      []*domain.StaticAnalysisIssue{},
		Summary:     &domain.StaticAnalysisSummary{},
	}

	if err != nil {
		// ClangTidy может вернуть ошибку, если найдет проблемы
		// Это не всегда означает, что анализ не удался
		a.log.Warning(fmt.Sprintf("ClangTidy found issues: %v", err))
	}

	// Парсим JSON вывод ClangTidy
	issues, err := a.parseClangTidyOutput(output)
	if err != nil {
		a.log.Warning(fmt.Sprintf("Failed to parse ClangTidy output: %v", err))
		// Возвращаем результат с ошибкой парсинга
		result.Error = fmt.Sprintf("Failed to parse output: %v", err)
	} else {
		result.Issues = issues
	}

	// Генерируем сводку
	result.Summary = a.generateSummary(issues)

	a.log.Info(fmt.Sprintf("ClangTidy analysis completed in %.2fs, found %d issues", duration, len(issues)))
	return result, nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (a *ClangTidyAnalyzer) GetSupportedLanguages() []string {
	return []string{"c", "cpp", "cc"}
}

// GetAnalyzerType возвращает тип анализатора
func (a *ClangTidyAnalyzer) GetAnalyzerType() domain.StaticAnalyzerType {
	return domain.StaticAnalyzerTypeClangTidy
}

// ValidateConfig проверяет корректность конфигурации
func (a *ClangTidyAnalyzer) ValidateConfig(config *domain.StaticAnalyzerConfig) error {
	supportedLanguages := []string{"c", "cpp", "cc"}
	isSupported := false
	for _, lang := range supportedLanguages {
		if config.Language == lang {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("ClangTidy analyzer only supports C/C++ languages")
	}

	if config.ProjectPath == "" {
		return fmt.Errorf("project path is required")
	}

	// Проверяем, что проект содержит C/C++ файлы
	if !a.hasCppFilePaths(config.ProjectPath) {
		return fmt.Errorf("no C/C++ files found in project path")
	}

	return nil
}

// checkClangTidyInstalled проверяет, установлен ли ClangTidy
func (a *ClangTidyAnalyzer) checkClangTidyInstalled() error {
	cmd := exec.Command("clang-tidy", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ClangTidy not found: %w", err)
	}
	return nil
}

// hasCppFilePaths проверяет, есть ли C/C++ файлы в проекте
func (a *ClangTidyAnalyzer) hasCppFilePaths(projectPath string) bool {
	patterns := []string{
		filepath.Join(projectPath, "**/*.c"),
		filepath.Join(projectPath, "**/*.cpp"),
		filepath.Join(projectPath, "**/*.cc"),
		filepath.Join(projectPath, "**/*.h"),
		filepath.Join(projectPath, "**/*.hpp"),
		filepath.Join(projectPath, "src/**/*.c"),
		filepath.Join(projectPath, "src/**/*.cpp"),
		filepath.Join(projectPath, "src/**/*.cc"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// parseClangTidyOutput парсит JSON вывод ClangTidy
func (a *ClangTidyAnalyzer) parseClangTidyOutput(output []byte) ([]*domain.StaticAnalysisIssue, error) {
	var issues []*domain.StaticAnalysisIssue

	// ClangTidy выводит JSON в формате:
	// [{"DiagnosticName": "clang-diagnostic-unused-variable", "DiagnosticMessage": {"Message": "unused variable 'x'", "FilePathOffset": 123, "FilePathPath": "test.cpp", "FilePathLineNumber": 5, "FilePathColumnStart": 9}}]

	// Простой парсинг JSON (в реальной реализации нужно использовать encoding/json)
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "\"DiagnosticName\"") {
			continue
		}

		// Извлекаем имя диагностики
		nameStart := strings.Index(line, "\"DiagnosticName\": \"")
		if nameStart == -1 {
			continue
		}
		nameStart += 19
		nameEnd := strings.Index(line[nameStart:], "\"")
		if nameEnd == -1 {
			continue
		}
		diagnosticName := line[nameStart : nameStart+nameEnd]

		// Извлекаем сообщение
		messageStart := strings.Index(line, "\"Message\": \"")
		if messageStart == -1 {
			continue
		}
		messageStart += 12
		messageEnd := strings.Index(line[messageStart:], "\"")
		if messageEnd == -1 {
			continue
		}
		message := line[messageStart : messageStart+messageEnd]

		// Извлекаем файл
		fileStart := strings.Index(line, "\"FilePathPath\": \"")
		if fileStart == -1 {
			continue
		}
		fileStart += 13
		fileEnd := strings.Index(line[fileStart:], "\"")
		if fileEnd == -1 {
			continue
		}
		filePath := line[fileStart : fileStart+fileEnd]

		// Извлекаем строку
		lineStart := strings.Index(line, "\"FilePathLineNumber\": ")
		if lineStart == -1 {
			continue
		}
		lineStart += 12
		lineEnd := strings.Index(line[lineStart:], ",")
		if lineEnd == -1 {
			continue
		}
		lineNum := 0
		fmt.Sscanf(line[lineStart:lineStart+lineEnd], "%d", &lineNum)

		// Извлекаем колонку
		columnStart := strings.Index(line, "\"FilePathColumnStart\": ")
		if columnStart == -1 {
			continue
		}
		columnStart += 14
		columnEnd := strings.Index(line[columnStart:], ",")
		if columnEnd == -1 {
			continue
		}
		columnNum := 0
		fmt.Sscanf(line[columnStart:columnStart+columnEnd], "%d", &columnNum)

		// Определяем severity
		severity := "warning"
		if strings.Contains(diagnosticName, "error") {
			severity = "error"
		}

		issue := &domain.StaticAnalysisIssue{
			FilePath:    filePath,
			LineNumber:  lineNum,
			ColumnStart: columnNum,
			Severity:    severity,
			Message:     message,
			Rule:        diagnosticName,
			Category:    a.getCategory(diagnosticName),
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// getCategory определяет категорию проблемы по имени диагностики
func (a *ClangTidyAnalyzer) getCategory(diagnosticName string) string {
	diagnosticName = strings.ToLower(diagnosticName)

	if strings.Contains(diagnosticName, "unused") {
		return "unused-code"
	} else if strings.Contains(diagnosticName, "null") {
		return "null-safety"
	} else if strings.Contains(diagnosticName, "memory") {
		return "memory-management"
	} else if strings.Contains(diagnosticName, "performance") {
		return "performance"
	} else if strings.Contains(diagnosticName, "modernize") {
		return "modernization"
	} else if strings.Contains(diagnosticName, "bugprone") {
		return "bug-prone"
	}
	return "other"
}

// generateSummary генерирует сводку анализа
func (a *ClangTidyAnalyzer) generateSummary(issues []*domain.StaticAnalysisIssue) *domain.StaticAnalysisSummary {
	summary := &domain.StaticAnalysisSummary{
		TotalIssues:         len(issues),
		SeverityBreakdown:   make(map[string]int),
		CategoryBreakdown:   make(map[string]int),
		FilePathsAnalyzed:   0,
		FilePathsWithIssues: 0,
	}

	filesWithIssues := make(map[string]bool)

	for _, issue := range issues {
		// Подсчитываем по severity
		summary.SeverityBreakdown[issue.Severity]++

		// Подсчитываем по категориям
		summary.CategoryBreakdown[issue.Category]++

		// Подсчитываем файлы с проблемами
		filesWithIssues[issue.FilePath] = true

		// Подсчитываем по типам
		switch issue.Severity {
		case "error":
			summary.ErrorCount++
		case "warning":
			summary.WarningCount++
		case "info":
			summary.InfoCount++
		case "hint":
			summary.HintCount++
		}
	}

	summary.FilePathsWithIssues = len(filesWithIssues)

	return summary
}
