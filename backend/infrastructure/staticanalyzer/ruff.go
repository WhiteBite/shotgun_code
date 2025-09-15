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

// RuffAnalyzer реализует StaticAnalyzer для Python с использованием Ruff
type RuffAnalyzer struct {
	log domain.Logger
}

// NewRuffAnalyzer создает новый анализатор для Python
func NewRuffAnalyzer(log domain.Logger) *RuffAnalyzer {
	return &RuffAnalyzer{
		log: log,
	}
}

// Analyze выполняет статический анализ Python кода
func (a *RuffAnalyzer) Analyze(ctx context.Context, config *domain.StaticAnalyzerConfig) (*domain.StaticAnalysisResult, error) {
	a.log.Info(fmt.Sprintf("Running Ruff analysis for project: %s", config.ProjectPath))

	startTime := time.Now()

	// Проверяем, что Ruff установлен
	if err := a.checkRuffInstalled(); err != nil {
		a.log.Warning(fmt.Sprintf("Ruff not available: %v", err))
		// Возвращаем успешный результат с предупреждением вместо ошибки
		return &domain.StaticAnalysisResult{
			Success:     true,
			Language:    config.Language,
			ProjectPath: config.ProjectPath,
			Analyzer:    config.Analyzer,
			Duration:    time.Since(startTime).Seconds(),
			Issues:      []*domain.StaticIssue{},
			Summary:     &domain.StaticAnalysisSummary{},
			Error:       fmt.Sprintf("Ruff not available: %v", err),
		}, nil
	}

	// Строим команду для Ruff
	args := []string{"check", "--output-format", "json"}

	// Добавляем конфигурационный файл если указан
	if config.ConfigFile != "" {
		args = append(args, "--config", config.ConfigFile)
	}

	// Добавляем правила если указаны
	if len(config.Rules) > 0 {
		args = append(args, "--select", strings.Join(config.Rules, ","))
	}

	// Добавляем исключения если указаны
	if len(config.ExcludeRules) > 0 {
		args = append(args, "--ignore", strings.Join(config.ExcludeRules, ","))
	}

	// Добавляем путь к проекту
	args = append(args, config.ProjectPath)

	// Создаем команду
	cmd := exec.CommandContext(ctx, "ruff", args...)
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
		Issues:      []*domain.StaticIssue{},
		Summary:     &domain.StaticAnalysisSummary{},
	}

	if err != nil {
		// Ruff может вернуть ошибку, если найдет проблемы
		// Это не всегда означает, что анализ не удался
		a.log.Warning(fmt.Sprintf("Ruff found issues: %v", err))
	}

	// Парсим JSON вывод Ruff
	issues, err := a.parseRuffOutput(output)
	if err != nil {
		a.log.Warning(fmt.Sprintf("Failed to parse Ruff output: %v", err))
		// Возвращаем результат с ошибкой парсинга
		result.Error = fmt.Sprintf("Failed to parse output: %v", err)
	} else {
		result.Issues = issues
	}

	// Генерируем сводку
	result.Summary = a.generateSummary(issues)

	a.log.Info(fmt.Sprintf("Ruff analysis completed in %.2fs, found %d issues", duration, len(issues)))
	return result, nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (a *RuffAnalyzer) GetSupportedLanguages() []string {
	return []string{"python", "py"}
}

// GetAnalyzerType возвращает тип анализатора
func (a *RuffAnalyzer) GetAnalyzerType() domain.StaticAnalyzerType {
	return domain.StaticAnalyzerTypeRuff
}

// ValidateConfig проверяет корректность конфигурации
func (a *RuffAnalyzer) ValidateConfig(config *domain.StaticAnalyzerConfig) error {
	supportedLanguages := []string{"python", "py"}
	isSupported := false
	for _, lang := range supportedLanguages {
		if config.Language == lang {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("Ruff analyzer only supports Python language")
	}

	if config.ProjectPath == "" {
		return fmt.Errorf("project path is required")
	}

	// Проверяем, что проект содержит Python файлы
	if !a.hasPythonFiles(config.ProjectPath) {
		return fmt.Errorf("no Python files found in project path")
	}

	return nil
}

// checkRuffInstalled проверяет, установлен ли Ruff
func (a *RuffAnalyzer) checkRuffInstalled() error {
	cmd := exec.Command("ruff", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Ruff not found: %w", err)
	}
	return nil
}

// hasPythonFiles проверяет, есть ли Python файлы в проекте
func (a *RuffAnalyzer) hasPythonFiles(projectPath string) bool {
	patterns := []string{
		filepath.Join(projectPath, "**/*.py"),
		filepath.Join(projectPath, "src/**/*.py"),
		filepath.Join(projectPath, "main/**/*.py"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// parseRuffOutput парсит JSON вывод Ruff
func (a *RuffAnalyzer) parseRuffOutput(output []byte) ([]*domain.StaticIssue, error) {
	var issues []*domain.StaticIssue

	// Ruff выводит JSON в формате:
	// [{"code": "E501", "message": "line too long", "location": {"row": 1, "column": 80}, "end_location": {"row": 1, "column": 120}, "filename": "test.py"}]

	// Простой парсинг JSON (в реальной реализации нужно использовать encoding/json)
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "\"code\"") {
			continue
		}

		// Извлекаем код ошибки
		codeStart := strings.Index(line, "\"code\": \"")
		if codeStart == -1 {
			continue
		}
		codeStart += 9
		codeEnd := strings.Index(line[codeStart:], "\"")
		if codeEnd == -1 {
			continue
		}
		code := line[codeStart : codeStart+codeEnd]

		// Извлекаем сообщение
		messageStart := strings.Index(line, "\"message\": \"")
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
		filenameStart := strings.Index(line, "\"filename\": \"")
		if filenameStart == -1 {
			continue
		}
		filenameStart += 13
		filenameEnd := strings.Index(line[filenameStart:], "\"")
		if filenameEnd == -1 {
			continue
		}
		filename := line[filenameStart : filenameStart+filenameEnd]

		// Извлекаем строку
		rowStart := strings.Index(line, "\"row\": ")
		if rowStart == -1 {
			continue
		}
		rowStart += 7
		rowEnd := strings.Index(line[rowStart:], ",")
		if rowEnd == -1 {
			continue
		}
		row := 0
		fmt.Sscanf(line[rowStart:rowStart+rowEnd], "%d", &row)

		// Извлекаем колонку
		columnStart := strings.Index(line, "\"column\": ")
		if columnStart == -1 {
			continue
		}
		columnStart += 11
		columnEnd := strings.Index(line[columnStart:], ",")
		if columnEnd == -1 {
			continue
		}
		column := 0
		fmt.Sscanf(line[columnStart:columnStart+columnEnd], "%d", &column)

		// Определяем severity
		severity := "warning"
		if strings.HasPrefix(code, "E") {
			severity = "error"
		}

		issue := &domain.StaticIssue{
			File:     filename,
			Line:     row,
			Column:   column,
			Severity: severity,
			Message:  message,
			Code:     code,
			Category: a.getCategory(code),
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// getCategory определяет категорию проблемы по коду
func (a *RuffAnalyzer) getCategory(code string) string {
	if strings.HasPrefix(code, "E") {
		return "error"
	} else if strings.HasPrefix(code, "W") {
		return "warning"
	} else if strings.HasPrefix(code, "F") {
		return "flake8"
	} else if strings.HasPrefix(code, "I") {
		return "imports"
	} else if strings.HasPrefix(code, "N") {
		return "naming"
	} else if strings.HasPrefix(code, "UP") {
		return "pyupgrade"
	}
	return "other"
}

// generateSummary генерирует сводку анализа
func (a *RuffAnalyzer) generateSummary(issues []*domain.StaticIssue) *domain.StaticAnalysisSummary {
	summary := &domain.StaticAnalysisSummary{
		TotalIssues:       len(issues),
		SeverityBreakdown: make(map[string]int),
		CategoryBreakdown: make(map[string]int),
		FilesAnalyzed:     0,
		FilesWithIssues:   0,
	}

	filesWithIssues := make(map[string]bool)

	for _, issue := range issues {
		// Подсчитываем по severity
		summary.SeverityBreakdown[issue.Severity]++

		// Подсчитываем по категориям
		summary.CategoryBreakdown[issue.Category]++

		// Подсчитываем файлы с проблемами
		filesWithIssues[issue.File] = true

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

	summary.FilesWithIssues = len(filesWithIssues)

	return summary
}
