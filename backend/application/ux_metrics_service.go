package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"strings"
	"sync"
	"time"
)

// Impact level constants
const (
	impactMedium = "Medium"
	impactHigh   = "High"
	impactLow    = "Low"
)

// UXMetricsServiceImpl реализует UXMetricsService
type UXMetricsServiceImpl struct {
	log     domain.Logger
	reports map[string]*domain.UXReport
	mu      sync.RWMutex
	repo    domain.UXReportRepository
}

// NewUXMetricsService создает новый сервис UX метрик
func NewUXMetricsService(log domain.Logger, repo domain.UXReportRepository) domain.UXMetricsService {
	service := &UXMetricsServiceImpl{
		log:     log,
		reports: make(map[string]*domain.UXReport),
		repo:    repo,
	}

	return service
}

// GenerateWhyViewReport генерирует отчёт "почему эти файлы"
func (s *UXMetricsServiceImpl) GenerateWhyViewReport(taskID string, files []string, taskContext map[string]interface{}) (*domain.WhyViewReport, error) {
	s.log.Info(fmt.Sprintf("Generating why view report for task: %s", taskID))

	report := &domain.WhyViewReport{
		TaskID:      taskID,
		Files:       make([]domain.FileReason, 0, len(files)),
		Context:     s.generateContextDescription(taskContext),
		Explanation: s.generateExplanation(taskContext),
		Confidence:  s.calculateConfidence(files, taskContext),
		Suggestions: s.generateSuggestions(files, taskContext),
	}

	// Анализируем каждый файл
	for _, filePath := range files {
		reason := s.analyzeFileReason(filePath, taskID, taskContext)
		report.Files = append(report.Files, reason)
	}

	// Сохраняем отчёт
	uxReport := &domain.UXReport{
		ID:          fmt.Sprintf("why-view-%s-%d", taskID, time.Now().Unix()),
		Type:        domain.UXReportTypeWhyView,
		Title:       fmt.Sprintf("Why View Report for Task %s", taskID),
		Description: "Explains why specific files were modified",
		Content:     report,
		CreatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"taskID": taskID, "fileCount": len(files)},
	}

	if err := s.SaveUXReport(uxReport); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save UX report: %v", err))
	}

	return report, nil
}

// generateContextDescription генерирует описание контекста
func (s *UXMetricsServiceImpl) generateContextDescription(taskContext map[string]interface{}) string {
	if taskContext == nil {
		return "Task execution context"
	}

	if taskName, ok := taskContext["taskName"].(string); ok {
		return fmt.Sprintf("Task execution context for: %s", taskName)
	}

	return "Task execution context"
}

// generateExplanation генерирует объяснение изменений
func (s *UXMetricsServiceImpl) generateExplanation(taskContext map[string]interface{}) string {
	if taskContext == nil {
		return "Files were modified to implement the requested changes"
	}

	if description, ok := taskContext["description"].(string); ok {
		return fmt.Sprintf("Files were modified to: %s", description)
	}

	return "Files were modified to implement the requested changes"
}

// calculateConfidence вычисляет уверенность в изменениях
func (s *UXMetricsServiceImpl) calculateConfidence(files []string, taskContext map[string]interface{}) float64 {
	baseConfidence := 0.85

	// Увеличиваем уверенность на основе количества файлов
	if len(files) <= 5 {
		baseConfidence += 0.05
	} else if len(files) <= 10 {
		baseConfidence += 0.02
	} else {
		baseConfidence -= 0.05
	}

	// Увеличиваем уверенность если есть контекст задачи
	if taskContext != nil {
		if _, ok := taskContext["taskName"]; ok {
			baseConfidence += 0.03
		}
		if _, ok := taskContext["description"]; ok {
			baseConfidence += 0.02
		}
	}

	// Ограничиваем уверенность в разумных пределах
	if baseConfidence > 0.95 {
		baseConfidence = 0.95
	} else if baseConfidence < 0.5 {
		baseConfidence = 0.5
	}

	return baseConfidence
}

// generateSuggestions генерирует предложения
func (s *UXMetricsServiceImpl) generateSuggestions(files []string, taskContext map[string]interface{}) []string {
	suggestions := []string{
		"Review changes before committing",
		"Run tests to ensure functionality",
		"Check for any unintended side effects",
	}

	// Добавляем специфичные предложения на основе контекста
	if taskContext != nil {
		if taskType, ok := taskContext["taskType"].(string); ok {
			switch taskType {
			case "refactor":
				suggestions = append(suggestions, "Verify that refactoring maintains existing functionality")
			case "bugfix":
				suggestions = append(suggestions, "Test the specific bug scenario")
			case "feature":
				suggestions = append(suggestions, "Add tests for new functionality")
			}
		}
	}

	// Добавляем предложения на основе количества файлов
	if len(files) > 10 {
		suggestions = append(suggestions, "Consider breaking changes into smaller commits")
	}

	return suggestions
}

// GenerateTimeToGreenMetrics генерирует метрики time_to_green
func (s *UXMetricsServiceImpl) GenerateTimeToGreenMetrics(taskID string) (*domain.TimeToGreenMetrics, error) {
	s.log.Info(fmt.Sprintf("Generating time to green metrics for task: %s", taskID))

	// Симуляция метрик
	startTime := time.Now().Add(-5 * time.Minute)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	metrics := &domain.TimeToGreenMetrics{
		TaskID:             taskID,
		StartTime:          startTime,
		EndTime:            endTime,
		Duration:           duration,
		Attempts:           1,
		RepairAttempts:     0,
		BuildTime:          30 * time.Second,
		TestTime:           45 * time.Second,
		StaticAnalysisTime: 15 * time.Second,
		TotalTime:          duration,
		Success:            true,
		Bottlenecks: []domain.Bottleneck{
			{
				Type:        "build",
				Description: "Initial compilation took longer than expected",
				Duration:    30 * time.Second,
				Impact:      impactMedium,
				Suggestions: []string{"Optimize build configuration", "Use incremental builds"},
			},
		},
	}

	// Сохраняем отчёт
	uxReport := &domain.UXReport{
		ID:          fmt.Sprintf("time-to-green-%s-%d", taskID, time.Now().Unix()),
		Type:        domain.UXReportTypeTimeToGreen,
		Title:       fmt.Sprintf("Time to Green Metrics for Task %s", taskID),
		Description: "Metrics showing time to achieve green status",
		Content:     metrics,
		CreatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"taskID": taskID, "success": metrics.Success},
	}

	if err := s.SaveUXReport(uxReport); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save UX report: %v", err))
	}

	return metrics, nil
}

// GenerateDerivedDiffReport генерирует отчёт о derived diff
func (s *UXMetricsServiceImpl) GenerateDerivedDiffReport(taskID string, originalDiff, derivedDiff string) (*domain.DerivedDiffReport, error) {
	s.log.Info(fmt.Sprintf("Generating derived diff report for task: %s", taskID))

	// Анализируем diff
	changes := s.analyzeDiffChanges(originalDiff, derivedDiff)
	summary := s.calculateDiffSummary(originalDiff, derivedDiff)
	impact := s.assessDiffImpact(changes, summary)

	report := &domain.DerivedDiffReport{
		TaskID:       taskID,
		OriginalDiff: originalDiff,
		DerivedDiff:  derivedDiff,
		Changes:      changes,
		Summary:      summary,
		Impact:       impact,
	}

	// Сохраняем отчёт
	uxReport := &domain.UXReport{
		ID:          fmt.Sprintf("derived-diff-%s-%d", taskID, time.Now().Unix()),
		Type:        domain.UXReportTypeDerivedDiff,
		Title:       fmt.Sprintf("Derived Diff Report for Task %s", taskID),
		Description: "Analysis of derived diff changes",
		Content:     report,
		CreatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"taskID": taskID, "totalChanges": len(changes)},
	}

	if err := s.SaveUXReport(uxReport); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save UX report: %v", err))
	}

	return report, nil
}

// GeneratePerformanceMetrics генерирует метрики производительности
func (s *UXMetricsServiceImpl) GeneratePerformanceMetrics(taskID string) (*domain.PerformanceMetrics, error) {
	s.log.Info(fmt.Sprintf("Generating performance metrics for task: %s", taskID))

	// Симуляция метрик производительности
	metrics := &domain.PerformanceMetrics{
		TaskID:         taskID,
		MemoryUsage:    1024 * 1024 * 50, // 50MB
		CPUUsage:       25.5,
		DiskIO:         1024 * 1024 * 10, // 10MB
		NetworkIO:      1024 * 512,       // 512KB
		FileOperations: 150,
		APIRequests:    5,
		CacheHits:      80,
		CacheMisses:    20,
		Timestamps:     []time.Time{time.Now().Add(-5 * time.Minute), time.Now()},
		Values:         []float64{0.0, 100.0},
	}

	// Сохраняем отчёт
	uxReport := &domain.UXReport{
		ID:          fmt.Sprintf("performance-%s-%d", taskID, time.Now().Unix()),
		Type:        domain.UXReportTypePerformance,
		Title:       fmt.Sprintf("Performance Metrics for Task %s", taskID),
		Description: "Performance metrics during task execution",
		Content:     metrics,
		CreatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"taskID": taskID, "memoryUsageMB": metrics.MemoryUsage / 1024 / 1024},
	}

	if err := s.SaveUXReport(uxReport); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to save UX report: %v", err))
	}

	return metrics, nil
}

// GetUXReport возвращает UX отчёт
func (s *UXMetricsServiceImpl) GetUXReport(reportID string) (*domain.UXReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[reportID]
	if !exists {
		// Пытаемся загрузить через репозиторий
		report, err := s.repo.LoadReport(reportID)
		if err != nil {
			return nil, fmt.Errorf("UX report not found: %s", reportID)
		}

		// Кэшируем загруженный отчет
		s.mu.Lock()
		s.reports[reportID] = report
		s.mu.Unlock()

		return report, nil
	}

	return report, nil
}

// SaveUXReport сохраняет UX отчёт
func (s *UXMetricsServiceImpl) SaveUXReport(report *domain.UXReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reports[report.ID] = report

	// Сохраняем через репозиторий
	if err := s.repo.SaveReport(report); err != nil {
		return fmt.Errorf("failed to save report through repository: %w", err)
	}

	s.log.Info(fmt.Sprintf("Saved UX report: %s", report.ID))
	return nil
}

// GetUXReports возвращает все UX отчёты
func (s *UXMetricsServiceImpl) GetUXReports(reportType domain.UXReportType) ([]*domain.UXReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем, есть ли отчеты в памяти
	if len(s.reports) > 0 {
		var reports []*domain.UXReport
		for _, report := range s.reports {
			if reportType == "" || report.Type == reportType {
				reports = append(reports, report)
			}
		}
		return reports, nil
	}

	// Загружаем через репозиторий
	reports, err := s.repo.ListReports(reportType)
	if err != nil {
		return nil, err
	}

	// Кэшируем загруженные отчеты
	s.mu.Lock()
	for _, report := range reports {
		s.reports[report.ID] = report
	}
	s.mu.Unlock()

	return reports, nil
}

// DeleteUXReport удаляет UX отчёт
func (s *UXMetricsServiceImpl) DeleteUXReport(reportID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Удаляем через репозиторий
	if err := s.repo.DeleteReport(reportID); err != nil {
		return fmt.Errorf("failed to delete report through repository: %w", err)
	}

	// Удаляем из памяти
	delete(s.reports, reportID)

	s.log.Info(fmt.Sprintf("Deleted UX report: %s", reportID))
	return nil
}

// GetMetricsSummary возвращает сводку метрик
func (s *UXMetricsServiceImpl) GetMetricsSummary() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := map[string]interface{}{
		"totalReports":      len(s.reports),
		"reportsByType":     make(map[string]int),
		"lastReportTime":    nil,
		"averageConfidence": 0.0,
	}

	// Подсчитываем отчеты по типам
	typeCounts := make(map[string]int)
	var totalConfidence float64
	var confidenceCount int
	var lastReportTime *time.Time

	for _, report := range s.reports {
		typeCounts[string(report.Type)]++

		// Подсчитываем среднюю уверенность для why-view отчетов
		if report.Type == domain.UXReportTypeWhyView {
			if whyView, ok := report.Content.(*domain.WhyViewReport); ok {
				totalConfidence += whyView.Confidence
				confidenceCount++
			}
		}

		// Находим время последнего отчета
		if lastReportTime == nil || report.CreatedAt.After(*lastReportTime) {
			lastReportTime = &report.CreatedAt
		}
	}

	summary["reportsByType"] = typeCounts
	summary["lastReportTime"] = lastReportTime

	if confidenceCount > 0 {
		summary["averageConfidence"] = totalConfidence / float64(confidenceCount)
	}

	return summary, nil
}

// ListReports получает список отчетов
func (s *UXMetricsServiceImpl) ListReports(_ context.Context, reportType string) ([]*domain.GenericReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make([]*domain.GenericReport, 0, len(s.reports))

	for _, report := range s.reports {
		// Фильтруем по типу если указан
		if reportType != "" && string(report.Type) != reportType {
			continue
		}

		// Конвертируем UXReport в GenericReport
		genericReport := &domain.GenericReport{
			Id:        report.ID,
			TaskId:    s.extractTaskID(report),
			Title:     report.Title,
			Type:      string(report.Type),
			Summary:   report.Description,
			Content:   s.serializeContent(report.Content),
			CreatedAt: report.CreatedAt,
			UpdatedAt: report.CreatedAt, // UXReport не имеет UpdatedAt
		}

		reports = append(reports, genericReport)
	}

	return reports, nil
}

// GetReport получает конкретный отчет
func (s *UXMetricsServiceImpl) GetReport(_ context.Context, reportID string) (*domain.GenericReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	// Конвертируем UXReport в GenericReport
	genericReport := &domain.GenericReport{
		Id:        report.ID,
		TaskId:    s.extractTaskID(report),
		Title:     report.Title,
		Type:      string(report.Type),
		Summary:   report.Description,
		Content:   s.serializeContent(report.Content),
		CreatedAt: report.CreatedAt,
		UpdatedAt: report.CreatedAt, // UXReport не имеет UpdatedAt
	}

	return genericReport, nil
}

// extractTaskID извлекает TaskID из отчета
func (s *UXMetricsServiceImpl) extractTaskID(report *domain.UXReport) string {
	if taskID, ok := report.Metadata["taskID"].(string); ok {
		return taskID
	}
	return ""
}

// serializeContent сериализует содержимое отчета в JSON
func (s *UXMetricsServiceImpl) serializeContent(content interface{}) string {
	if content == nil {
		return ""
	}

	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Sprintf("Error serializing content: %v", err)
	}

	return string(data)
}

// analyzeFileReason анализирует причину изменения файла
func (s *UXMetricsServiceImpl) analyzeFileReason(filePath, _ string, taskContext map[string]interface{}) domain.FileReason {
	reason := domain.FileReason{
		FilePath:     filePath,
		Reason:       "File was modified as part of task execution",
		Impact:       "Medium",
		Confidence:   0.8,
		RelatedFiles: []string{},
		Context:      make(map[string]interface{}),
	}

	// Определяем тип файла и его роль
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		reason.Reason = "Go source file modified for task implementation"
		reason.Impact = impactHigh
		reason.Confidence = 0.9
	case ".ts", ".js":
		reason.Reason = "TypeScript/JavaScript file modified for frontend changes"
		reason.Impact = impactMedium
		reason.Confidence = 0.85
	case ".vue":
		reason.Reason = "Vue component modified for UI changes"
		reason.Impact = impactMedium
		reason.Confidence = 0.85
	case ".yaml", ".yml":
		reason.Reason = "Configuration file modified for task setup"
		reason.Impact = impactLow
		reason.Confidence = 0.95
	case ".md":
		reason.Reason = "Documentation file updated"
		reason.Impact = impactLow
		reason.Confidence = 0.9
	default:
		reason.Reason = "File modified for task implementation"
		reason.Impact = impactMedium
		reason.Confidence = 0.8
	}

	// Добавляем контекст задачи
	if taskContext != nil {
		if taskName, ok := taskContext["taskName"].(string); ok {
			reason.Context["taskName"] = taskName
		}
		if taskType, ok := taskContext["taskType"].(string); ok {
			reason.Context["taskType"] = taskType
			// Адаптируем объяснение на основе типа задачи
			if taskName, ok := taskContext["taskName"].(string); ok {
				switch taskType {
				case "refactor":
					reason.Reason = fmt.Sprintf("File refactored as part of %s", taskName)
				case "bugfix":
					reason.Reason = fmt.Sprintf("File modified to fix bug in %s", taskName)
				case "feature":
					reason.Reason = fmt.Sprintf("File modified to implement feature in %s", taskName)
				}
			}
		}
		if description, ok := taskContext["description"].(string); ok {
			reason.Context["description"] = description
		}
	}

	// Определяем связанные файлы на основе расширения
	reason.RelatedFiles = s.findRelatedFiles(filePath, ext)

	return reason
}

// findRelatedFiles находит связанные файлы
func (s *UXMetricsServiceImpl) findRelatedFiles(filePath, ext string) []string {
	var relatedFiles []string

	// Базовые связанные файлы на основе расширения
	switch ext {
	case ".go":
		// Для Go файлов ищем тесты и связанные файлы
		baseName := strings.TrimSuffix(filePath, ".go")
		relatedFiles = append(relatedFiles, baseName+"_test.go")
	case ".ts", ".js":
		// Для TypeScript/JavaScript ищем связанные файлы
		baseName := strings.TrimSuffix(filePath, ext)
		relatedFiles = append(relatedFiles, baseName+".d.ts")
	case ".vue":
		// Для Vue компонентов ищем связанные файлы
		baseName := strings.TrimSuffix(filePath, ".vue")
		relatedFiles = append(relatedFiles, baseName+".spec.ts")
		relatedFiles = append(relatedFiles, baseName+".test.ts")
	}

	return relatedFiles
}

// analyzeDiffChanges анализирует изменения в diff
func (s *UXMetricsServiceImpl) analyzeDiffChanges(originalDiff, derivedDiff string) []domain.DiffChange {
	var changes []domain.DiffChange

	// Простой анализ diff (в реальной реализации был бы более сложный парсинг)
	lines := strings.Split(derivedDiff, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			change := domain.DiffChange{
				Type:       "modification",
				LineNumber: i + 1,
				OldContent: "",
				NewContent: line,
				Reason:     "Code modification",
				Confidence: 0.9,
			}

			if strings.HasPrefix(line, "+") {
				change.Type = "addition"
				change.NewContent = strings.TrimPrefix(line, "+")
			} else if strings.HasPrefix(line, "-") {
				change.Type = "deletion"
				change.OldContent = strings.TrimPrefix(line, "-")
			}

			changes = append(changes, change)
		}
	}

	return changes
}

// calculateDiffSummary вычисляет сводку по diff
func (s *UXMetricsServiceImpl) calculateDiffSummary(originalDiff, derivedDiff string) *domain.DiffSummary {
	summary := &domain.DiffSummary{}

	// Подсчитываем строки
	lines := strings.Split(derivedDiff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			summary.AddedLines++
		} else if strings.HasPrefix(line, "-") {
			summary.RemovedLines++
		}
	}

	summary.TotalLines = summary.AddedLines + summary.RemovedLines

	// Подсчитываем файлы (упрощенно)
	filePattern := regexp.MustCompile(`^--- a/(.+)$`)
	matches := filePattern.FindAllStringSubmatch(derivedDiff, -1)
	summary.TotalFiles = len(matches)
	summary.ModifiedFiles = len(matches)

	return summary
}

// assessDiffImpact оценивает влияние изменений
func (s *UXMetricsServiceImpl) assessDiffImpact(changes []domain.DiffChange, summary *domain.DiffSummary) domain.DiffImpact {
	impact := domain.DiffImpact{
		RiskLevel:         "Low",
		AffectedTests:     []string{},
		BreakingChanges:   []string{},
		PerformanceImpact: "None",
		SecurityImpact:    "None",
	}

	// Оцениваем риск на основе количества изменений
	if summary.TotalLines > 100 {
		impact.RiskLevel = impactHigh
	} else if summary.TotalLines > 50 {
		impact.RiskLevel = impactMedium
	}

	// Проверяем на потенциальные breaking changes
	for _, change := range changes {
		if strings.Contains(change.NewContent, "func") || strings.Contains(change.NewContent, "interface") {
			impact.BreakingChanges = append(impact.BreakingChanges, "API changes detected")
		}
	}

	return impact
}
