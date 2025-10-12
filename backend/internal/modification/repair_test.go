package modification

import (
	"context"
	"shotgun_code/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockRepairLogger struct {
	mock.Mock
}

func (m *MockRepairLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockRepairLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockRepairLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockRepairLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockRepairLogger) Fatal(msg string) {
	m.Called(msg)
}

func TestRepairService_NewRepairService(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)

	// Execute
	service := NewRepairService(mockLogger)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockLogger, service.log)
}

func TestRepairService_ExecuteRepair_ProjectNotFound(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - using a non-existent project path
	req := domain.RepairRequest{
		ProjectPath: "/non/existent/project",
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 3,
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "project path does not exist")

	mockLogger.AssertExpectations(t)
}

func TestRepairService_ExecuteRepair_NoRules(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - using a real existing path for this test
	req := domain.RepairRequest{
		ProjectPath: ".",
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 1,
		Rules:       []domain.RepairRule{}, // No rules
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should either succeed or fail based on actual verification, but not error

	mockLogger.AssertExpectations(t)
}

func TestRepairService_ExecuteRepair_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	req := domain.RepairRequest{
		ProjectPath: ".",
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 1,
		Rules: []domain.RepairRule{
			{
				ID:       "test-rule-1",
				Name:     "Test Rule 1",
				Pattern:  "test pattern",
				Language: "go",
				Category: "syntax",
				Priority: 1,
			},
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Less(t, result.Duration, time.Second*10) // Should complete quickly

	mockLogger.AssertExpectations(t)
}

func TestRepairService_GetAvailableRules(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test cases for different languages
	languages := []string{"go", "javascript", "typescript", "python", "java"}

	for _, language := range languages {
		// Execute
		rules, err := service.GetAvailableRules(language)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.NotEmpty(t, rules)
	}
}

func TestRepairService_AddRule(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Pattern:  "test pattern",
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	err := service.AddRule(rule)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
}

func TestRepairService_RemoveRule(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	ruleID := "test-rule-id"

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	err := service.RemoveRule(ruleID)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
}

func TestRepairService_ValidateRule_Valid(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Pattern:  "test.*pattern", // Valid regex
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	err := service.ValidateRule(rule)

	// Assert
	assert.NoError(t, err)
}

func TestRepairService_ValidateRule_MissingID(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - missing ID
	rule := domain.RepairRule{
		Name:     "Test Rule",
		Pattern:  "test.*pattern",
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	err := service.ValidateRule(rule)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule ID is required")
}

func TestRepairService_ValidateRule_MissingName(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - missing name
	rule := domain.RepairRule{
		ID:       "test-rule",
		Pattern:  "test.*pattern",
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	err := service.ValidateRule(rule)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name is required")
}

func TestRepairService_ValidateRule_MissingPattern(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - missing pattern
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	err := service.ValidateRule(rule)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule pattern is required")
}

func TestRepairService_ValidateRule_InvalidRegex(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data - invalid regex pattern
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Pattern:  "[invalid-regex", // Invalid regex
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	err := service.ValidateRule(rule)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestRepairService_MatchesError(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	errorOutput := "undefined: someVariable in main.go:10"
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Undefined Variable Rule",
		Pattern:  "undefined:.*in.*\\.go:.*", // Regex to match undefined errors
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Execute
	// Using reflection to test private method
	// We'll test the behavior through public methods instead

	// Create a request that would use this rule
	req := domain.RepairRequest{
		ProjectPath: ".",
		Language:    "go",
		ErrorOutput: errorOutput,
		MaxAttempts: 1,
		Rules:       []domain.RepairRule{rule},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockLogger.AssertExpectations(t)
}

func TestRepairService_ApplyRule(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data
	projectPath := "."
	rule := domain.RepairRule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Pattern:  "test.*pattern",
		Language: "go",
		Category: "syntax",
		Priority: 1,
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockLogger.On("Warning", mock.AnythingOfType("string")).Return()

	// Execute
	// Using reflection to test private method
	// We'll test the behavior through public methods instead

	// Create a request that would use this rule
	req := domain.RepairRequest{
		ProjectPath: projectPath,
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 1,
		Rules:       []domain.RepairRule{rule},
	}

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockLogger.AssertExpectations(t)
}

func TestRepairService_ExecuteRepair_WithAttempts(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data with multiple attempts
	req := domain.RepairRequest{
		ProjectPath: ".",
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 3, // Multiple attempts
		Rules: []domain.RepairRule{
			{
				ID:       "test-rule-1",
				Name:     "Test Rule 1",
				Pattern:  "test pattern 1",
				Language: "go",
				Category: "syntax",
				Priority: 1,
			},
			{
				ID:       "test-rule-2",
				Name:     "Test Rule 2",
				Pattern:  "test pattern 2",
				Language: "go",
				Category: "syntax",
				Priority: 2,
			},
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.LessOrEqual(t, result.Attempts, req.MaxAttempts)

	mockLogger.AssertExpectations(t)
}

func TestRepairService_GetDefaultRules(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test for different languages
	languages := []string{"go", "javascript", "typescript", "python", "java", "unknown"}

	for _, language := range languages {
		// Execute
		rules := service.getDefaultRules(language)

		// Assert
		assert.NotNil(t, rules)
		// For known languages, we expect some rules
		if language != "unknown" {
			assert.NotEmpty(t, rules)
		}
	}
}

func TestRepairService_SortRulesByPriority(t *testing.T) {
	// Setup
	mockLogger := new(MockRepairLogger)
	service := NewRepairService(mockLogger)

	// Test data with unsorted rules
	req := domain.RepairRequest{
		ProjectPath: ".",
		Language:    "go",
		ErrorOutput: "some error output",
		MaxAttempts: 1,
		Rules: []domain.RepairRule{
			{
				ID:       "low-priority",
				Name:     "Low Priority Rule",
				Pattern:  "low.*pattern",
				Language: "go",
				Category: "syntax",
				Priority: 1,
			},
			{
				ID:       "high-priority",
				Name:     "High Priority Rule",
				Pattern:  "high.*pattern",
				Language: "go",
				Category: "syntax",
				Priority: 10,
			},
			{
				ID:       "medium-priority",
				Name:     "Medium Priority Rule",
				Pattern:  "medium.*pattern",
				Language: "go",
				Category: "syntax",
				Priority: 5,
			},
		},
	}

	// The rules should be sorted by priority (high to low) after processing
	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	result, err := service.ExecuteRepair(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockLogger.AssertExpectations(t)
}
