package modification

import (
	"context"
	"errors"
	"shotgun_code/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockApplyLogger struct {
	mock.Mock
}

func (m *MockApplyLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockApplyLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockApplyLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockApplyLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockApplyLogger) Fatal(msg string) {
	m.Called(msg)
}

// Mock ApplyEngine for testing
type MockApplyEngine struct {
	mock.Mock
}

func (m *MockApplyEngine) ApplyEdit(ctx context.Context, edit domain.Edit) error {
	args := m.Called(ctx, edit)
	return args.Error(0)
}

// Mock Formatter for testing
type MockFormatter struct {
	mock.Mock
}

func (m *MockFormatter) FormatFile(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func TestApplyService_NewApplyService(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)

	// Execute
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockEngine, service.applyEngine)
	assert.Equal(t, mockFormatter, service.formatter)
	assert.Equal(t, mockLogger, service.log)
}

func TestApplyService_ApplyEdits_Success(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file1.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content 1",
			NewContent:  "new content 1",
			Position:    10,
		},
		{
			FilePath:    "/test/file2.js",
			Type:        domain.EditTypeInsert,
			NewContent:  "new content 2",
			Position:    20,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("ApplyEdit", mock.Anything, edits[0]).Return(nil)
	mockEngine.On("ApplyEdit", mock.Anything, edits[1]).Return(nil)
	mockFormatter.On("FormatFile", "/test/file1.go").Return(nil)
	mockFormatter.On("FormatFile", "/test/file2.js").Return(nil)

	// Execute
	ctx := context.Background()
	err := service.ApplyEdits(ctx, edits)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
}

func TestApplyService_ApplyEdits_ApplyError(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("ApplyEdit", mock.Anything, edits[0]).Return(errors.New("apply failed"))

	// Execute
	ctx := context.Background()
	err := service.ApplyEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to apply edit")
	assert.Contains(t, err.Error(), "apply failed")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestApplyService_ApplyEdits_FormatError(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockLogger.On("Warning", mock.AnythingOfType("string")).Return()
	mockEngine.On("ApplyEdit", mock.Anything, edits[0]).Return(nil)
	mockFormatter.On("FormatFile", "/test/file.go").Return(errors.New("format failed"))

	// Execute
	ctx := context.Background()
	err := service.ApplyEdits(ctx, edits)

	// Assert
	assert.NoError(t, err) // Should not fail the entire operation due to formatting error

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockFormatter.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_Success(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file1.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
		{
			FilePath:    "/test/file2.js",
			Type:        domain.EditTypeInsert,
			NewContent:  "new content",
			Position:    20,
		},
		{
			FilePath:    "/test/file3.py",
			Type:        domain.EditTypeDelete,
			OldContent:  "old content",
			Position:    30,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_EmptyFilePath(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "", // Empty file path
			Type:        domain.EditTypeReplace,
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "edit has empty file path")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_EmptyType(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        "", // Empty type
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "edit has empty type")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_ReplaceWithEmptyOldContent(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "", // Empty old content
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace edit requires non-empty old content")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_ReplaceWithIdenticalContent(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "same content",
			NewContent:  "same content", // Identical content
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace edit has identical old and new content")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_InsertWithEmptyNewContent(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeInsert,
			NewContent:  "", // Empty new content
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert edit requires non-empty new content")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_InsertWithNegativePosition(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeInsert,
			NewContent:  "new content",
			Position:    -1, // Negative position
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert edit requires valid position")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_DeleteWithEmptyOldContent(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeDelete,
			OldContent:  "", // Empty old content
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete edit requires non-empty old content")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_ValidateEdits_UnsupportedType(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        "unsupported_type", // Unsupported type
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.ValidateEdits(ctx, edits)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported edit type")

	mockLogger.AssertExpectations(t)
}

func TestApplyService_RollbackEdits_Success(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file1.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content 1",
			NewContent:  "new content 1",
			Position:    10,
		},
		{
			FilePath:    "/test/file2.js",
			Type:        domain.EditTypeInsert,
			NewContent:  "new content 2",
			Position:    20,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	// For rollback of replace edit (swap old and new content)
	mockEngine.On("ApplyEdit", mock.Anything, domain.Edit{
		FilePath:    "/test/file1.go",
		Type:        domain.EditTypeReplace,
		OldContent:  "new content 1",
		NewContent:  "old content 1",
		Position:    10,
	}).Return(nil)
	// For rollback of insert edit (convert to delete)
	mockEngine.On("ApplyEdit", mock.Anything, domain.Edit{
		FilePath:    "/test/file2.js",
		Type:        domain.EditTypeDelete,
		OldContent:  "new content 2",
		Position:    20,
	}).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.RollbackEdits(ctx, edits)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestApplyService_RollbackEdits_RollbackError(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        domain.EditTypeReplace,
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockLogger.On("Error", mock.AnythingOfType("string")).Return()
	// For rollback of replace edit (swap old and new content)
	mockEngine.On("ApplyEdit", mock.Anything, domain.Edit{
		FilePath:    "/test/file.go",
		Type:        domain.EditTypeReplace,
		OldContent:  "new content",
		NewContent:  "old content",
		Position:    10,
	}).Return(errors.New("rollback failed"))

	// Execute
	ctx := context.Background()
	err := service.RollbackEdits(ctx, edits)

	// Assert
	assert.NoError(t, err) // Should not fail the entire operation due to rollback error

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestApplyService_RollbackEdits_UnsupportedType(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data
	edits := []domain.Edit{
		{
			FilePath:    "/test/file.go",
			Type:        "unsupported_type", // Unsupported type
			OldContent:  "old content",
			NewContent:  "new content",
			Position:    10,
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockLogger.On("Error", mock.AnythingOfType("string")).Return()

	// Execute
	ctx := context.Background()
	err := service.RollbackEdits(ctx, edits)

	// Assert
	assert.NoError(t, err) // Should not fail the entire operation due to unsupported type

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestApplyService_ShouldFormat(t *testing.T) {
	// Setup
	mockEngine := new(MockApplyEngine)
	mockFormatter := new(MockFormatter)
	mockLogger := new(MockApplyLogger)
	
	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test cases
	testCases := []struct {
		filePath   string
		shouldFormat bool
	}{
		{"/test/file.go", true},
		{"/test/file.ts", true},
		{"/test/file.tsx", true},
		{"/test/file.js", true},
		{"/test/file.jsx", true},
		{"/test/file.json", true},
		{"/test/file.py", false},
		{"/test/file.java", false},
		{"/test/file.md", false},
		{"/test/file", false},
		{"", false},
	}

	// Execute and assert
	for _, tc := range testCases {
		result := service.shouldFormat(tc.filePath)
		assert.Equal(t, tc.shouldFormat, result, "FilePath: %s", tc.filePath)
	}
}