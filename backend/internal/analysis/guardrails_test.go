package analysis

import (
	"errors"
	"shotgun_code/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockGuardrailsLogger struct {
	mock.Mock
}

func (m *MockGuardrailsLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockGuardrailsLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockGuardrailsLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockGuardrailsLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockGuardrailsLogger) Fatal(msg string) {
	m.Called(msg)
}

// Mock TaskflowService for testing
type MockTaskflowService struct {
	mock.Mock
}

func (m *MockTaskflowService) ValidateTask(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

// Mock OPAService for testing
type MockOPAService struct {
	mock.Mock
}

func (m *MockOPAService) ValidatePath(path string) (*domain.OPAValidationResult, error) {
	args := m.Called(path)
	return args.Get(0).(*domain.OPAValidationResult), args.Error(1)
}

func TestGuardrailsService_NewGuardrailsService(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	// Execute
	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Assert
	assert.NotNil(t, service)
	assert.NotNil(t, service.policies)
	assert.NotNil(t, service.budgets)
	assert.NotNil(t, service.config)
	assert.True(t, service.config.FailClosed)
	assert.True(t, service.config.EnableEphemeralMode)
}

func TestGuardrailsService_ValidatePath_Allowed(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	path := "/safe/path/file.go"

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	violations, err := service.ValidatePath(path)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, violations)

	mockLogger.AssertExpectations(t)
}

func TestGuardrailsService_ValidatePath_Forbidden(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Add a forbidden path policy
	forbiddenPolicy := domain.GuardrailPolicy{
		ID:       "forbidden-test",
		Name:     "Forbidden Test Policy",
		Type:     domain.GuardrailTypeForbiddenPath,
		Severity: domain.GuardrailSeverityBlock,
		Enabled:  true,
		Rules: []domain.GuardrailRule{
			{
				ID:      "rule-1",
				Pattern: "/forbidden/*",
				Message: "Access to forbidden path denied",
			},
		},
	}

	service.mu.Lock()
	service.policies = append(service.policies, forbiddenPolicy)
	service.mu.Unlock()

	// Test data
	path := "/forbidden/sensitive/file.txt"

	// Setup mocks
	mockLogger.On("Error", mock.AnythingOfType("string")).Return()

	// Execute
	violations, err := service.ValidatePath(path)

	// Assert
	assert.Error(t, err)
	assert.NotEmpty(t, violations)
	assert.Equal(t, 1, len(violations))
	assert.Equal(t, "forbidden-test", violations[0].PolicyID)
	assert.Equal(t, "rule-1", violations[0].RuleID)
	assert.Contains(t, err.Error(), "guardrail violation")

	mockLogger.AssertExpectations(t)
}

func TestGuardrailsService_ValidatePath_WarningOnly(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Add a warning path policy
	warningPolicy := domain.GuardrailPolicy{
		ID:       "warning-test",
		Name:     "Warning Test Policy",
		Type:     domain.GuardrailTypeForbiddenPath,
		Severity: domain.GuardrailSeverityMedium,
		Enabled:  true,
		Rules: []domain.GuardrailRule{
			{
				ID:      "rule-1",
				Pattern: "/warning/*",
				Message: "Access to warning path detected",
			},
		},
	}

	service.mu.Lock()
	service.policies = append(service.policies, warningPolicy)
	service.mu.Unlock()

	// Test data
	path := "/warning/file.txt"

	// Execute
	violations, err := service.ValidatePath(path)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, violations)
	assert.Equal(t, 1, len(violations))
	assert.Equal(t, "warning-test", violations[0].PolicyID)
	assert.Equal(t, domain.GuardrailSeverityMedium, violations[0].Severity)

	mockLogger.AssertExpectations(t)
}

func TestGuardrailsService_ValidateBudget_WithinLimit(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Add a budget policy
	budgetPolicy := domain.BudgetPolicy{
		ID:      "budget-test",
		Name:    "Budget Test Policy",
		Type:    domain.BudgetTypeFiles,
		Limit:   100,
		Unit:    "files",
		Enabled: true,
	}

	service.mu.Lock()
	service.budgets = append(service.budgets, budgetPolicy)
	service.mu.Unlock()

	// Test data
	current := int64(50)

	// Execute
	violations, err := service.ValidateBudget(domain.BudgetTypeFiles, current)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, violations)
}

func TestGuardrailsService_ValidateBudget_Exceeded(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Add a budget policy
	budgetPolicy := domain.BudgetPolicy{
		ID:      "budget-test",
		Name:    "Budget Test Policy",
		Type:    domain.BudgetTypeFiles,
		Limit:   100,
		Unit:    "files",
		Enabled: true,
	}

	service.mu.Lock()
	service.budgets = append(service.budgets, budgetPolicy)
	service.mu.Unlock()

	// Test data
	current := int64(150)

	// Setup mocks
	mockLogger.On("Error", mock.AnythingOfType("string")).Return()

	// Execute
	violations, err := service.ValidateBudget(domain.BudgetTypeFiles, current)

	// Assert
	assert.Error(t, err)
	assert.NotEmpty(t, violations)
	assert.Equal(t, 1, len(violations))
	assert.Equal(t, "budget-test", violations[0].PolicyID)
	assert.Contains(t, err.Error(), "budget violation")

	mockLogger.AssertExpectations(t)
}

func TestGuardrailsService_ValidateBudget_WarningOnly(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	// Disable fail-closed for this test
	service := NewGuardrailsService(mockLogger, mockTaskflow)
	service.config.FailClosed = false

	// Add a budget policy
	budgetPolicy := domain.BudgetPolicy{
		ID:      "budget-test",
		Name:    "Budget Test Policy",
		Type:    domain.BudgetTypeFiles,
		Limit:   100,
		Unit:    "files",
		Enabled: true,
	}

	service.mu.Lock()
	service.budgets = append(service.budgets, budgetPolicy)
	service.mu.Unlock()

	// Test data
	current := int64(150)

	// Execute
	violations, err := service.ValidateBudget(domain.BudgetTypeFiles, current)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, violations)
	assert.Equal(t, 1, len(violations))
	assert.Equal(t, "budget-test", violations[0].PolicyID)
}

func TestGuardrailsService_ValidateTask_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	taskID := "task-123"
	files := []string{"/safe/file1.go", "/safe/file2.js"}
	linesChanged := int64(100)

	// Setup mocks
	mockTaskflow.On("ValidateTask", taskID).Return(nil)

	// Execute
	result, err := service.ValidateTask(taskID, files, linesChanged)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskID, result.TaskID)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Violations)
	assert.Empty(t, result.BudgetViolations)

	mockTaskflow.AssertExpectations(t)
}

func TestGuardrailsService_ValidateTask_PathViolation(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Add a forbidden path policy
	forbiddenPolicy := domain.GuardrailPolicy{
		ID:       "forbidden-test",
		Name:     "Forbidden Test Policy",
		Type:     domain.GuardrailTypeForbiddenPath,
		Severity: domain.GuardrailSeverityMedium,
		Enabled:  true,
		Rules: []domain.GuardrailRule{
			{
				ID:      "rule-1",
				Pattern: "/forbidden/*",
				Message: "Access to forbidden path denied",
			},
		},
	}

	service.mu.Lock()
	service.policies = append(service.policies, forbiddenPolicy)
	service.mu.Unlock()

	// Test data
	taskID := "task-123"
	files := []string{"/safe/file1.go", "/forbidden/sensitive.txt"}
	linesChanged := int64(100)

	// Setup mocks
	mockTaskflow.On("ValidateTask", taskID).Return(nil)

	// Execute
	result, err := service.ValidateTask(taskID, files, linesChanged)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskID, result.TaskID)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Violations)
	assert.Equal(t, 1, len(result.Violations))
	assert.Equal(t, "forbidden-test", result.Violations[0].PolicyID)

	mockTaskflow.AssertExpectations(t)
}

func TestGuardrailsService_ValidateTask_TaskflowError(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	taskID := "task-123"
	files := []string{"/safe/file1.go"}
	linesChanged := int64(100)

	// Setup mocks
	mockTaskflow.On("ValidateTask", taskID).Return(errors.New("task validation failed"))

	// Execute
	result, err := service.ValidateTask(taskID, files, linesChanged)

	// Assert
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskID, result.TaskID)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error, "task validation failed")

	mockTaskflow.AssertExpectations(t)
}

func TestGuardrailsService_EnableEphemeralMode(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Execute
	err := service.EnableEphemeralMode(5 * time.Minute)

	// Assert
	assert.NoError(t, err)
	assert.True(t, service.ephemeralMode)
	assert.False(t, service.ephemeralEnd.IsZero())
	assert.True(t, service.ephemeralEnd.After(time.Now()))
}

func TestGuardrailsService_DisableEphemeralMode(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)
	err := service.EnableEphemeralMode(5 * time.Minute)
	assert.NoError(t, err)

	// Execute
	service.disableEphemeralMode()

	// Assert
	assert.False(t, service.ephemeralMode)
}

func TestGuardrailsService_IsEphemeralExpired(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test when not in ephemeral mode
	assert.False(t, service.isEphemeralExpired())

	// Enable ephemeral mode
	err := service.EnableEphemeralMode(5 * time.Minute)
	assert.NoError(t, err)
	assert.False(t, service.isEphemeralExpired())

	// Manually set expiration time to past
	service.ephemeralEnd = time.Now().Add(-time.Minute)
	assert.True(t, service.isEphemeralExpired())
}

func TestGuardrailsService_AddPolicy(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	policy := domain.GuardrailPolicy{
		ID:       "test-policy",
		Name:     "Test Policy",
		Type:     domain.GuardrailTypeForbiddenPath,
		Severity: domain.GuardrailSeverityBlock,
		Enabled:  true,
		Rules: []domain.GuardrailRule{
			{
				ID:      "rule-1",
				Pattern: "/test/*",
				Message: "Test rule violation",
			},
		},
	}

	// Execute
	service.AddPolicy(policy)

	// Assert
	service.mu.RLock()
	defer service.mu.RUnlock()
	assert.Equal(t, 4, len(service.policies)) // 3 default + 1 added
	assert.Equal(t, "test-policy", service.policies[3].ID)
}

func TestGuardrailsService_AddBudget(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	budget := domain.BudgetPolicy{
		ID:      "test-budget",
		Name:    "Test Budget",
		Type:    domain.BudgetTypeTokens,
		Limit:   10000,
		Unit:    "tokens",
		Enabled: true,
	}

	// Execute
	service.AddBudgetPolicy(budget)

	// Assert
	service.mu.RLock()
	defer service.mu.RUnlock()
	assert.Equal(t, 1, len(service.budgets)) // 0 default + 1 added
	assert.Equal(t, "test-budget", service.budgets[0].ID)
}

func TestGuardrailsService_GetConfig(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Execute
	config := service.GetConfig()

	// Assert
	assert.NotNil(t, config)
	assert.True(t, config.FailClosed)
	assert.True(t, config.EnableEphemeralMode)
	assert.True(t, config.EnableTaskValidation)
	assert.True(t, config.EnableBudgetTracking)
	assert.True(t, config.EnablePathValidation)
}

func TestGuardrailsService_SetConfig(t *testing.T) {
	// Setup
	mockLogger := new(MockGuardrailsLogger)
	mockTaskflow := new(MockTaskflowService)

	service := NewGuardrailsService(mockLogger, mockTaskflow)

	// Test data
	newConfig := domain.GuardrailConfig{
		FailClosed:           false,
		EnableEphemeralMode:  false,
		EphemeralTimeout:     10 * time.Minute,
		EnableTaskValidation: false,
		EnableBudgetTracking: false,
		EnablePathValidation: false,
	}

	// Execute
	service.UpdateConfig(newConfig)

	// Assert
	config := service.GetConfig()
	assert.Equal(t, newConfig, config)
	assert.False(t, config.FailClosed)
	assert.False(t, config.EnableEphemeralMode)
}
