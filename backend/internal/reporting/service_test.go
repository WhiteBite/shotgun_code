package reporting

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of domain.Logger for tests
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}

func TestService_CreateReport(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockLogger := new(MockLogger)
	service := &Service{
		logger:    mockLogger,
		reportDir: tempDir,
	}

	// Setup mock
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	taskID := "test-task-123"
	reportType := "context-analysis"
	title := "Test Report"
	summary := "Test summary"
	content := "Test content"

	result, err := service.CreateReport(ctx, taskID, reportType, title, summary, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Id)
	assert.Equal(t, taskID, result.TaskId)
	assert.Equal(t, reportType, result.Type)
	assert.Equal(t, title, result.Title)
	assert.Equal(t, summary, result.Summary)
	assert.Equal(t, content, result.Content)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())

	// Verify report was saved to disk
	reportFile := filepath.Join(tempDir, result.Id+".json")
	assert.FileExists(t, reportFile)

	mockLogger.AssertExpectations(t)
}

func TestService_GetReport(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockLogger := new(MockLogger)
	service := &Service{
		logger:    mockLogger,
		reportDir: tempDir,
	}

	// Create a test report
	ctx := context.Background()
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	createdReport, err := service.CreateReport(ctx, "task-123", "test-type", "Test Title", "Test Summary", "Test Content")
	assert.NoError(t, err)

	// Execute
	result, err := service.GetReport(ctx, createdReport.Id)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, createdReport.Id, result.Id)
	assert.Equal(t, createdReport.TaskId, result.TaskId)
	assert.Equal(t, createdReport.Type, result.Type)
	assert.Equal(t, createdReport.Title, result.Title)
	assert.Equal(t, createdReport.Summary, result.Summary)
	assert.Equal(t, createdReport.Content, result.Content)
}

func TestService_GetReport_NotFound(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	service := &Service{
		reportDir: tempDir,
	}

	// Execute
	ctx := context.Background()
	result, err := service.GetReport(ctx, "non-existent-id")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "report not found")
}

func TestService_ListReports(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockLogger := new(MockLogger)
	service := &Service{
		logger:    mockLogger,
		reportDir: tempDir,
	}

	// Setup mock
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Create test reports
	ctx := context.Background()
	_, err := service.CreateReport(ctx, "task-1", "type-a", "Report 1", "Summary 1", "Content 1")
	assert.NoError(t, err)

	_, err = service.CreateReport(ctx, "task-2", "type-a", "Report 2", "Summary 2", "Content 2")
	assert.NoError(t, err)

	_, err = service.CreateReport(ctx, "task-3", "type-b", "Report 3", "Summary 3", "Content 3")
	assert.NoError(t, err)

	// Test 1: List all reports
	results, err := service.ListReports(ctx, "")
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Test 2: List reports filtered by type
	filteredResults, err := service.ListReports(ctx, "type-a")
	assert.NoError(t, err)
	assert.Len(t, filteredResults, 2)

	// Verify the correct reports were returned
	for _, report := range filteredResults {
		assert.Equal(t, "type-a", report.Type)
	}

	// Test 3: List reports with non-existent type
	emptyResults, err := service.ListReports(ctx, "non-existent-type")
	assert.NoError(t, err)
	assert.Len(t, emptyResults, 0)
}

func TestService_UpdateReport(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockLogger := new(MockLogger)
	service := &Service{
		logger:    mockLogger,
		reportDir: tempDir,
	}

	// Setup mock
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Create a test report
	ctx := context.Background()
	originalReport, err := service.CreateReport(ctx, "task-123", "test-type", "Original Title", "Original Summary", "Original Content")
	assert.NoError(t, err)

	originalUpdateTime := originalReport.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(time.Millisecond * 10)

	// Execute update
	updatedTitle := "Updated Title"
	updatedSummary := "Updated Summary"
	updatedContent := "Updated Content"

	result, err := service.UpdateReport(ctx, originalReport.Id, updatedTitle, updatedSummary, updatedContent)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, originalReport.Id, result.Id)
	assert.Equal(t, originalReport.TaskId, result.TaskId)
	assert.Equal(t, originalReport.Type, result.Type)
	assert.Equal(t, updatedTitle, result.Title)
	assert.Equal(t, updatedSummary, result.Summary)
	assert.Equal(t, updatedContent, result.Content)
	assert.True(t, result.UpdatedAt.After(originalUpdateTime))

	mockLogger.AssertExpectations(t)
}

func TestService_DeleteReport(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockLogger := new(MockLogger)
	service := &Service{
		logger:    mockLogger,
		reportDir: tempDir,
	}

	// Setup mock
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Create a test report
	ctx := context.Background()
	report, err := service.CreateReport(ctx, "task-123", "test-type", "Test Title", "Test Summary", "Test Content")
	assert.NoError(t, err)

	// Verify the report file exists
	reportFile := filepath.Join(tempDir, report.Id+".json")
	assert.FileExists(t, reportFile)

	// Execute delete
	err = service.DeleteReport(ctx, report.Id)

	// Assert
	assert.NoError(t, err)
	assert.NoFileExists(t, reportFile)

	mockLogger.AssertExpectations(t)
}

func TestService_DeleteReport_NotFound(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "report_service_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	service := &Service{
		reportDir: tempDir,
	}

	// Execute
	ctx := context.Background()
	err := service.DeleteReport(ctx, "non-existent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report not found")
}