package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service handles report management operations
type Service struct {
	logger    domain.Logger
	reportDir string
}

// NewService creates a new report service
func NewService(logger domain.Logger) *Service {
	homeDir, _ := os.UserHomeDir()
	reportDir := filepath.Join(homeDir, ".shotgun-code", "reports")
	os.MkdirAll(reportDir, 0755)

	return &Service{
		logger:    logger,
		reportDir: reportDir,
	}
}

// GetReport retrieves a report by ID
func (s *Service) GetReport(ctx context.Context, reportID string) (*domain.GenericReport, error) {
	reportPath := filepath.Join(s.reportDir, reportID+".json")

	data, err := os.ReadFile(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("report not found: %s", reportID)
		}
		return nil, fmt.Errorf("failed to read report file: %w", err)
	}

	var report domain.GenericReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// ListReports lists all reports, optionally filtered by type
func (s *Service) ListReports(ctx context.Context, reportType string) ([]*domain.GenericReport, error) {
	entries, err := os.ReadDir(s.reportDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read report directory: %w", err)
	}

	var reports []*domain.GenericReport

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		reportID := strings.TrimSuffix(entry.Name(), ".json")
		report, err := s.GetReport(ctx, reportID)
		if err != nil {
			s.logger.Warning(fmt.Sprintf("Failed to load report %s: %v", reportID, err))
			continue
		}

		if reportType != "" && report.Type != reportType {
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// CreateReport creates a new report
func (s *Service) CreateReport(ctx context.Context, taskID, reportType, title, summary, content string) (*domain.GenericReport, error) {
	reportID := uuid.New().String()
	now := time.Now()

	report := &domain.GenericReport{
		Id:        reportID,
		TaskId:    taskID,
		Type:      reportType,
		Title:     title,
		Summary:   summary,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Created report %s of type %s", reportID, reportType))
	return report, nil
}

// UpdateReport updates an existing report
func (s *Service) UpdateReport(ctx context.Context, reportID, title, summary, content string) (*domain.GenericReport, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	if title != "" {
		report.Title = title
	}
	if summary != "" {
		report.Summary = summary
	}
	if content != "" {
		report.Content = content
	}
	report.UpdatedAt = time.Now()

	if err := s.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save updated report: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Updated report %s", reportID))
	return report, nil
}

// DeleteReport deletes a report by ID
func (s *Service) DeleteReport(ctx context.Context, reportID string) error {
	reportPath := filepath.Join(s.reportDir, reportID+".json")

	if err := os.Remove(reportPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("report not found: %s", reportID)
		}
		return fmt.Errorf("failed to delete report file: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Deleted report %s", reportID))
	return nil
}

// GetReportsByTask retrieves all reports for a specific task
func (s *Service) GetReportsByTask(ctx context.Context, taskID string) ([]*domain.GenericReport, error) {
	entries, err := os.ReadDir(s.reportDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read report directory: %w", err)
	}

	var reports []*domain.GenericReport

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		reportID := strings.TrimSuffix(entry.Name(), ".json")
		report, err := s.GetReport(ctx, reportID)
		if err != nil {
			s.logger.Warning(fmt.Sprintf("Failed to load report %s: %v", reportID, err))
			continue
		}

		if report.TaskId == taskID {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// saveReport saves a report to disk
func (s *Service) saveReport(report *domain.GenericReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	reportPath := filepath.Join(s.reportDir, report.Id+".json")
	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}
