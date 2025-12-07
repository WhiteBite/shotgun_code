package domain

import (
	"fmt"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	ErrCodeInvalidAPIKey      ErrorCode = "INVALID_API_KEY" //nolint:gosec // This is an error code, not a credential
	ErrCodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeTaskNotFound       ErrorCode = "TASK_NOT_FOUND"
	ErrCodeInvalidTaskState   ErrorCode = "INVALID_TASK_STATE"
	ErrCodeConfigurationError ErrorCode = "CONFIGURATION_ERROR"
	ErrCodeFileSystemError    ErrorCode = "FILE_SYSTEM_ERROR"
	ErrCodeNetworkError       ErrorCode = "NETWORK_ERROR"
	ErrCodeValidationError    ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
)

// DomainError represents a structured error with context and recovery information
type DomainError struct {
	Code        ErrorCode
	Message     string
	Context     map[string]interface{}
	Cause       error
	Recoverable bool
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %s)", e.Code, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Constructor functions for common domain errors
func NewTaskNotFoundError(taskID string) *DomainError {
	return &DomainError{
		Code:        ErrCodeTaskNotFound,
		Message:     fmt.Sprintf("Task not found: %s", taskID),
		Context:     map[string]interface{}{"taskId": taskID},
		Recoverable: false,
	}
}

func NewInvalidTaskStateError(taskID string, currentState, expectedState string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidTaskState,
		Message: fmt.Sprintf("Invalid task state: %s (current: %s, expected: %s)", taskID, currentState, expectedState),
		Context: map[string]interface{}{
			"taskId":        taskID,
			"currentState":  currentState,
			"expectedState": expectedState,
		},
		Recoverable: false,
	}
}

func NewValidationError(message string, details map[string]interface{}) *DomainError {
	return &DomainError{
		Code:        ErrCodeValidationError,
		Message:     message,
		Context:     details,
		Recoverable: false,
	}
}

func NewConfigurationError(message string, cause error) *DomainError {
	return &DomainError{
		Code:        ErrCodeConfigurationError,
		Message:     message,
		Cause:       cause,
		Recoverable: false,
	}
}

func NewNetworkError(message string, cause error) *DomainError {
	return &DomainError{
		Code:        ErrCodeNetworkError,
		Message:     message,
		Cause:       cause,
		Recoverable: true,
	}
}

func NewInternalError(message string, cause error) *DomainError {
	return &DomainError{
		Code:        ErrCodeInternalError,
		Message:     message,
		Cause:       cause,
		Recoverable: false,
	}
}

// Sentinel errors used across the application domain.
var (
	// ErrInvalidAPIKey is returned when an AI provider rejects the API key.
	ErrInvalidAPIKey = &DomainError{
		Code:        ErrCodeInvalidAPIKey,
		Message:     "Invalid API key provided",
		Recoverable: false,
	}

	// ErrRateLimitExceeded is returned when an AI provider indicates a rate limit has been hit.
	ErrRateLimitExceeded = &DomainError{
		Code:        ErrCodeRateLimitExceeded,
		Message:     "Rate limit exceeded, please try again later",
		Recoverable: true,
	}
)
