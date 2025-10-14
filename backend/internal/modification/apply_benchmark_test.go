package modification

import (
	"context"
	"shotgun_code/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for benchmarking
type mockApplyLogger struct{}

func (m *mockApplyLogger) Debug(msg string)   {}
func (m *mockApplyLogger) Info(msg string)    {}
func (m *mockApplyLogger) Warning(msg string) {}
func (m *mockApplyLogger) Error(msg string)   {}
func (m *mockApplyLogger) Fatal(msg string)   {}

// Mock ApplyEngine for benchmarking
type mockApplyEngine struct {
	delayMs int
}

func (m *mockApplyEngine) ApplyOperation(ctx context.Context, op *domain.ApplyOperation) (*domain.ApplyResult, error) {
	return &domain.ApplyResult{
		OperationID:  op.ID,
		Success:      true,
		Path:         op.Path,
		AppliedLines: 1,
	}, nil
}

func (m *mockApplyEngine) ApplyOperations(ctx context.Context, ops []*domain.ApplyOperation) ([]*domain.ApplyResult, error) {
	results := make([]*domain.ApplyResult, len(ops))
	for i, op := range ops {
		results[i] = &domain.ApplyResult{
			OperationID:  op.ID,
			Success:      true,
			Path:         op.Path,
			AppliedLines: 1,
		}
	}
	return results, nil
}

func (m *mockApplyEngine) ApplyEdit(ctx context.Context, edit domain.Edit) error {
	return nil
}

func (m *mockApplyEngine) ValidateOperation(ctx context.Context, op *domain.ApplyOperation) error {
	return nil
}

func (m *mockApplyEngine) RollbackOperation(ctx context.Context, result *domain.ApplyResult) error {
	return nil
}

func (m *mockApplyEngine) RegisterFormatter(language string, formatter domain.Formatter) {
}

func (m *mockApplyEngine) RegisterImportFixer(language string, fixer domain.ImportFixer) {
}

// Mock Formatter for benchmarking
type mockFormatter struct {
	delayMs int
}

func (m *mockFormatter) FormatFile(ctx context.Context, path string) error {
	return nil
}

func (m *mockFormatter) FormatContent(ctx context.Context, content string, language string) (string, error) {
	return content, nil
}

func (m *mockFormatter) GetSupportedLanguages() []string {
	return []string{"go", "typescript", "javascript"}
}

func BenchmarkApplyService_ApplyEdits_Small(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{delayMs: 1}
	mockFormatter := &mockFormatter{delayMs: 1}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - small set of edits
	edits := []domain.Edit{
		{
			FilePath:   "/test/file1.go",
			Type:       domain.EditTypeReplace,
			OldContent: "old content 1",
			NewContent: "new content 1",
			Position:   10,
		},
		{
			FilePath:   "/test/file2.js",
			Type:       domain.EditTypeInsert,
			NewContent: "new content 2",
			Position:   20,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		err := service.ApplyEdits(ctx, edits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyService_ApplyEdits_Medium(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{delayMs: 1}
	mockFormatter := &mockFormatter{delayMs: 1}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - medium set of edits (50 edits)
	edits := make([]domain.Edit, 50)
	for i := 0; i < 50; i++ {
		edits[i] = domain.Edit{
			FilePath:   "/test/file" + string(rune(i+'0')) + ".go",
			Type:       domain.EditTypeReplace,
			OldContent: "old content " + string(rune(i+'0')),
			NewContent: "new content " + string(rune(i+'0')),
			Position:   i * 10,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		err := service.ApplyEdits(ctx, edits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyService_ApplyEdits_Large(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{delayMs: 1}
	mockFormatter := &mockFormatter{delayMs: 1}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - large set of edits (200 edits)
	edits := make([]domain.Edit, 200)
	for i := 0; i < 200; i++ {
		edits[i] = domain.Edit{
			FilePath:   "/test/module" + string(rune(i/20+'0')) + "/file" + string(rune(i%20+'0')) + ".go",
			Type:       domain.EditTypeReplace,
			OldContent: "old content " + string(rune(i+'0')),
			NewContent: "new content " + string(rune(i+'0')),
			Position:   i * 5,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		err := service.ApplyEdits(ctx, edits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyService_ValidateEdits(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{}
	mockFormatter := &mockFormatter{}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - set of edits with various validation scenarios
	edits := []domain.Edit{
		{
			FilePath:   "/test/file1.go",
			Type:       domain.EditTypeReplace,
			OldContent: "old content 1",
			NewContent: "new content 1",
			Position:   10,
		},
		{
			FilePath:   "/test/file2.js",
			Type:       domain.EditTypeInsert,
			NewContent: "new content 2",
			Position:   20,
		},
		{
			FilePath:   "/test/file3.py",
			Type:       domain.EditTypeDelete,
			OldContent: "old content 3",
			Position:   30,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		err := service.ValidateEdits(ctx, edits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyService_RollbackEdits(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{delayMs: 1}
	mockFormatter := &mockFormatter{delayMs: 1}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - set of edits to rollback
	edits := []domain.Edit{
		{
			FilePath:   "/test/file1.go",
			Type:       domain.EditTypeReplace,
			OldContent: "old content 1",
			NewContent: "new content 1",
			Position:   10,
		},
		{
			FilePath:   "/test/file2.js",
			Type:       domain.EditTypeInsert,
			NewContent: "new content 2",
			Position:   20,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		err := service.RollbackEdits(ctx, edits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyService_ShouldFormat(b *testing.B) {
	// Setup
	mockEngine := &mockApplyEngine{}
	mockFormatter := &mockFormatter{}
	mockLogger := &mockApplyLogger{}

	service := NewApplyService(mockEngine, mockFormatter, mockLogger)

	// Test data - various file paths
	testPaths := []string{
		"/test/file.go",
		"/test/file.ts",
		"/test/file.tsx",
		"/test/file.js",
		"/test/file.jsx",
		"/test/file.json",
		"/test/file.py",
		"/test/file.java",
		"/test/file.md",
		"/test/file",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			result := service.shouldFormat(path)
			assert.NotNil(b, result)
		}
	}
}

func BenchmarkDiffService_GenerateDiff(b *testing.B) {
	// Setup
	mockLogger := &mockApplyLogger{}
	mockEngine := &mockDiffEngine{delayMs: 5}

	service := &DiffService{
		log:    mockLogger,
		engine: mockEngine,
	}

	b.ResetTimer()
	b.ReportAllocs()

	beforePath := "/path/before"
	afterPath := "/path/after"
	format := domain.DiffFormatGit

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := service.GenerateDiff(ctx, beforePath, afterPath, format)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Mock DiffEngine for benchmarking
type mockDiffEngine struct {
	delayMs int
}

func (m *mockDiffEngine) GenerateDiff(ctx context.Context, beforePath, afterPath string, format domain.DiffFormat) (*domain.DiffResult, error) {
	return &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "mock diff content",
	}, nil
}

func (m *mockDiffEngine) GenerateDiffFromResults(ctx context.Context, results []*domain.ApplyResult, format domain.DiffFormat) (*domain.DiffResult, error) {
	return &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "mock diff content from results",
	}, nil
}

func (m *mockDiffEngine) GenerateDiffFromEdits(ctx context.Context, edits *domain.EditsJSON, format domain.DiffFormat) (*domain.DiffResult, error) {
	return &domain.DiffResult{
		ID:      "test-diff-id",
		Format:  format,
		Content: "mock diff content from edits",
	}, nil
}

func (m *mockDiffEngine) PublishDiff(ctx context.Context, diff *domain.DiffResult) error {
	return nil
}