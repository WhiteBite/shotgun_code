package uxreports

import (
	"shotgun_code/domain"
	"sync"
)

// InMemoryUXReportRepository implements UXReportRepository using in-memory storage
type InMemoryUXReportRepository struct {
	reports map[string]*domain.UXReport
	mu      sync.RWMutex
}

// NewInMemoryUXReportRepository creates a new in-memory UX report repository
func NewInMemoryUXReportRepository() *InMemoryUXReportRepository {
	return &InMemoryUXReportRepository{
		reports: make(map[string]*domain.UXReport),
	}
}

// LoadReport loads a UX report by ID from memory
func (r *InMemoryUXReportRepository) LoadReport(reportID string) (*domain.UXReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, exists := r.reports[reportID]
	if !exists {
		return nil, nil // Return nil if not found, consistent with filesystem implementation
	}

	return report, nil
}

// SaveReport saves a UX report to memory
func (r *InMemoryUXReportRepository) SaveReport(report *domain.UXReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reports[report.ID] = report
	return nil
}

// ListReports returns all UX reports of a specific type from memory
func (r *InMemoryUXReportRepository) ListReports(reportType domain.UXReportType) ([]*domain.UXReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var reports []*domain.UXReport
	for _, report := range r.reports {
		// Filter by report type if specified
		if reportType != "" && report.Type != reportType {
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// DeleteReport deletes a UX report from memory
func (r *InMemoryUXReportRepository) DeleteReport(reportID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.reports, reportID)
	return nil
}