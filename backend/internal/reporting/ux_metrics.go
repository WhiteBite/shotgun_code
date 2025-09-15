package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sync"
	"time"
)

// UXMetricsService implements domain.UXMetricsService for UX metrics operations
type UXMetricsService struct {
	log        domain.Logger
	reports    map[string]*domain.UXReport
	mu         sync.RWMutex
	reportsDir string
}

// NewUXMetricsService creates a new UX metrics service
func NewUXMetricsService(log domain.Logger) *UXMetricsService {
	service := &UXMetricsService{
		log:        log,
		reports:    make(map[string]*domain.UXReport),
		reportsDir: "reports/ux",
	}

	// Create reports directory
	if err := os.MkdirAll(service.reportsDir, 0755); err != nil {
		log.Warning(fmt.Sprintf("Failed to create reports directory: %v", err))
	}

	return service
}

// GenerateWhyViewReport generates "why these files" report
func (s *UXMetricsService) GenerateWhyViewReport(taskID string, files []string, taskContext map[string]interface{}) (*domain.WhyViewReport, error) {
	s.log.Info(fmt.Sprintf("Generating why view report for task: %s", taskID))

	report := &domain.WhyViewReport{
		TaskID:      taskID,
		Files:       make([]domain.FileReason, 0, len(files)),
		Context:     s.generateContextDescription(taskContext),
		Explanation: s.generateExplanation(taskContext),
		Confidence:  s.calculateConfidence(files, taskContext),
		Suggestions: s.generateSuggestions(files, taskContext),
	}

	// Analyze each file
	for _, filePath := range files {
		reason := s.analyzeFileReason(filePath, taskID, taskContext)
		report.Files = append(report.Files, reason)
	}

	// Save report
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

// GenerateTimeToGreenMetrics generates time_to_green metrics
func (s *UXMetricsService) GenerateTimeToGreenMetrics(taskID string) (*domain.TimeToGreenMetrics, error) {
	s.log.Info(fmt.Sprintf("Generating time to green metrics for task: %s", taskID))

	// Simulate metrics
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
				Impact:      "Medium",
				Suggestions: []string{"Optimize build configuration", "Use incremental builds"},
			},
		},
	}

	// Save report
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

// GetUXReport returns UX report
func (s *UXMetricsService) GetUXReport(reportID string) (*domain.UXReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("UX report not found: %s", reportID)
	}

	return report, nil
}

// SaveUXReport saves UX report
func (s *UXMetricsService) SaveUXReport(report *domain.UXReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reports[report.ID] = report

	// Save to file
	reportPath := filepath.Join(s.reportsDir, fmt.Sprintf("%s.json", report.ID))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	s.log.Info(fmt.Sprintf("Saved UX report: %s", report.ID))
	return nil
}

// Helper methods (simplified for brevity)
func (s *UXMetricsService) generateContextDescription(taskContext map[string]interface{}) string {
	if taskContext == nil {
		return "Task execution context"
	}
	if taskName, ok := taskContext["taskName"].(string); ok {
		return fmt.Sprintf("Task execution context for: %s", taskName)
	}
	return "Task execution context"
}

func (s *UXMetricsService) generateExplanation(taskContext map[string]interface{}) string {
	if taskContext != nil {
		if description, ok := taskContext["description"].(string); ok {
			return fmt.Sprintf("Files were modified to: %s", description)
		}
	}
	return "Files were modified to implement the requested changes"
}

func (s *UXMetricsService) calculateConfidence(files []string, taskContext map[string]interface{}) float64 {
	baseConfidence := 0.85
	if len(files) <= 5 {
		baseConfidence += 0.05
	}
	if taskContext != nil {
		baseConfidence += 0.03
	}
	if baseConfidence > 0.95 {
		baseConfidence = 0.95
	}
	return baseConfidence
}

func (s *UXMetricsService) generateSuggestions(files []string, taskContext map[string]interface{}) []string {
	suggestions := []string{
		"Review changes before committing",
		"Run tests to ensure functionality",
		"Check for any unintended side effects",
	}
	if len(files) > 10 {
		suggestions = append(suggestions, "Consider breaking changes into smaller commits")
	}
	return suggestions
}

func (s *UXMetricsService) analyzeFileReason(filePath, taskID string, taskContext map[string]interface{}) domain.FileReason {
	return domain.FileReason{
		FilePath:    filePath,
		Reason:      "Modified to implement task requirements",
		Confidence:  0.85,
		Category:    "implementation",
		Importance:  "high",
		Suggestions: []string{"Review changes carefully"},
	}
}