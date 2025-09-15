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
type MockDiffLogger struct {
	mock.Mock
}

func (m *MockDiffLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockDiffLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockDiffLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockDiffLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockDiffLogger) Fatal(msg string) {
	m.Called(msg)
}

// Mock DiffEngine for testing
type MockDiffEngine struct {
	mock.Mock
}

func (m *MockDiffEngine) GenerateDiff(ctx context.Context, beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	args := m.Called(ctx, beforePath, afterPath, format)
	return args.Get(0).(*domain.DiffResult), args.Error(1)
}

func (m *MockDiffEngine) GenerateDiffFromResults(ctx context.Context, results []*domain.ApplyResult, format domain.DiffFormat) (*domain.DiffResult, error) {
	args := m.Called(ctx, results, format)
	return args.Get(0).(*domain.DiffResult), args.Error(1)
}

func (m *MockDiffEngine) GenerateDiffFromEdits(ctx context.Context, edits *domain.EditsJSON, format domain.DiffFormat) (*domain.DiffResult, error) {
	args := m.Called(ctx, edits, format)
	return args.Get(0).(*domain.DiffResult), args.Error(1)
}

func (m *MockDiffEngine) PublishDiff(ctx context.Context, diff *domain.DiffResult) error {
	args := m.Called(ctx, diff)
	return args.Error(0)
}

func TestDiffService_NewDiffService(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)

	// Execute
	service := NewDiffService(mockLogger)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, mockLogger, service.log)
	assert.NotNil(t, service.engine)
}

func TestDiffService_GenerateDiff_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit
	
	diffResult := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "diff content",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiff", mock.Anything, beforePath, afterPath, format).Return(diffResult, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GenerateDiff(ctx, beforePath, afterPath, format)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, diffResult, result)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateDiff_Error(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiff", mock.Anything, beforePath, afterPath, format).Return((*domain.DiffResult)(nil), errors.New("generation failed"))

	// Execute
	ctx := context.Background()
	result, err := service.GenerateDiff(ctx, beforePath, afterPath, format)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "generation failed")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateDiffFromResults_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	results := []*domain.ApplyResult{
		{
			FilePath: "/test/file1.go",
			Success:  true,
		},
		{
			FilePath: "/test/file2.js",
			Success:  true,
		},
	}
	format := domain.DiffFormatJSON
	
	diffResult := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "diff content from results",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiffFromResults", mock.Anything, results, format).Return(diffResult, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GenerateDiffFromResults(ctx, results, format)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, diffResult, result)
	assert.Equal(t, 2, len(results))

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateDiffFromEdits_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	edits := &domain.EditsJSON{
		Edits: []domain.Edit{
			{
				FilePath:    "/test/file.go",
				Type:        domain.EditTypeReplace,
				OldContent:  "old content",
				NewContent:  "new content",
				Position:    10,
			},
		},
	}
	format := domain.DiffFormatUnified
	
	diffResult := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "diff content from edits",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiffFromEdits", mock.Anything, edits, format).Return(diffResult, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GenerateDiffFromEdits(ctx, edits, format)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, diffResult, result)
	assert.Equal(t, 1, len(edits.Edits))

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_PublishDiff_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	diff := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  domain.DiffFormatGit,
		Content: "diff content",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("PublishDiff", mock.Anything, diff).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.PublishDiff(ctx, diff)

	// Assert
	assert.NoError(t, err)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_PublishDiff_Error(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	diff := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  domain.DiffFormatGit,
		Content: "diff content",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("PublishDiff", mock.Anything, diff).Return(errors.New("publish failed"))

	// Execute
	ctx := context.Background()
	err := service.PublishDiff(ctx, diff)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish failed")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateAndPublishDiff_Success(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit
	
	diffResult := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "diff content",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiff", mock.Anything, beforePath, afterPath, format).Return(diffResult, nil)
	mockEngine.On("PublishDiff", mock.Anything, diffResult).Return(nil)

	// Execute
	ctx := context.Background()
	result, err := service.GenerateAndPublishDiff(ctx, beforePath, afterPath, format)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, diffResult, result)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateAndPublishDiff_GenerateError(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiff", mock.Anything, beforePath, afterPath, format).Return((*domain.DiffResult)(nil), errors.New("generation failed"))

	// Execute
	ctx := context.Background()
	result, err := service.GenerateAndPublishDiff(ctx, beforePath, afterPath, format)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate diff")

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GenerateAndPublishDiff_PublishError(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	mockEngine := new(MockDiffEngine)
	
	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	// Test data
	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit
	
	diffResult := &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "diff content",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string")).Return()
	mockLogger.On("Warning", mock.AnythingOfType("string")).Return()
	mockEngine.On("GenerateDiff", mock.Anything, beforePath, afterPath, format).Return(diffResult, nil)
	mockEngine.On("PublishDiff", mock.Anything, diffResult).Return(errors.New("publish failed"))

	// Execute
	ctx := context.Background()
	result, err := service.GenerateAndPublishDiff(ctx, beforePath, afterPath, format)

	// Assert
	assert.NoError(t, err) // Should not fail the entire operation due to publish error
	assert.NotNil(t, result)
	assert.Equal(t, diffResult, result)

	mockLogger.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestDiffService_GetSupportedFormats(t *testing.T) {
	// Setup
	mockLogger := new(MockDiffLogger)
	service := NewDiffService(mockLogger)

	// Execute
	formats := service.GetSupportedFormats()

	// Assert
	assert.NotNil(t, formats)
	assert.Equal(t, 4, len(formats))
	assert.Contains(t, formats, domain.DiffFormatGit)
	assert.Contains(t, formats, domain.DiffFormatUnified)
	assert.Contains(t, formats, domain.DiffFormatJSON)
	assert.Contains(t, formats, domain.DiffFormatHTML)
}