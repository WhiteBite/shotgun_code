package analysis

import (
	"context"
	"shotgun_code/domain"
	"testing"
)

func TestNewContainer(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	if container == nil {
		t.Fatal("expected container to be created")
	}

	if container.GetRegistry() == nil {
		t.Error("expected registry to be initialized")
	}

	if container.GetProjectRoot() != "" {
		t.Error("expected empty project root initially")
	}
}

func TestContainer_SetProject(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	container.SetProject("/test/project")

	if container.GetProjectRoot() != "/test/project" {
		t.Errorf("expected project root '/test/project', got '%s'", container.GetProjectRoot())
	}

	container.SetProject("/test/project")

	container.SetProject("/test/project2")

	if container.GetProjectRoot() != "/test/project2" {
		t.Errorf("expected project root '/test/project2', got '%s'", container.GetProjectRoot())
	}
}

func TestContainer_GetSymbolIndex(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	index1 := container.GetSymbolIndex()
	if index1 == nil {
		t.Fatal("expected symbol index to be created")
	}

	index2 := container.GetSymbolIndex()
	if index1 != index2 {
		t.Error("expected same symbol index instance")
	}
}

func TestContainer_GetCallGraph(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	cg1 := container.GetCallGraph()
	if cg1 == nil {
		t.Fatal("expected call graph to be created")
	}

	cg2 := container.GetCallGraph()
	if cg1 != cg2 {
		t.Error("expected same call graph instance")
	}
}

func TestContainer_GetGitContext(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	gc := container.GetGitContext()
	if gc != nil {
		t.Error("expected nil git context without project")
	}

	container.SetProject("/test/project")
	gc = container.GetGitContext()
	if gc == nil {
		t.Error("expected git context to be created")
	}
}

func TestContainer_GetContextMemory(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	mem1 := container.GetContextMemory()
	if mem1 == nil {
		t.Fatal("expected context memory to be created")
	}

	mem2 := container.GetContextMemory()
	if mem1 != mem2 {
		t.Error("expected same context memory instance")
	}
}

func TestContainer_InvalidateCache(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	container.SetProject("/test/project")
	_ = container.GetSymbolIndex()
	_ = container.GetCallGraph()
	_ = container.GetGitContext()

	container.InvalidateCache()

	stats := container.Stats()
	if stats["symbolIndexBuilt"].(bool) {
		t.Error("expected symbolIndexBuilt to be false after invalidation")
	}
	if stats["callGraphBuilt"].(bool) {
		t.Error("expected callGraphBuilt to be false after invalidation")
	}
}

func TestContainer_SetProjectInvalidatesCache(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	container.SetProject("/project1")
	index1 := container.GetSymbolIndex()

	container.SetProject("/project2")
	index2 := container.GetSymbolIndex()

	if index1 == index2 {
		t.Error("expected different symbol index after project switch")
	}
}

func TestContainer_SemanticSearch(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	if container.GetSemanticSearch() != nil {
		t.Error("expected nil semantic search initially")
	}

	mockSS := &mockSemanticSearcher{}
	container.SetSemanticSearch(mockSS)

	if container.GetSemanticSearch() != mockSS {
		t.Error("expected semantic search to be set")
	}
}

func TestContainer_Stats(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	stats := container.Stats()

	if _, ok := stats["projectRoot"]; !ok {
		t.Error("expected projectRoot in stats")
	}
	if _, ok := stats["symbolIndexBuilt"]; !ok {
		t.Error("expected symbolIndexBuilt in stats")
	}
	if _, ok := stats["callGraphBuilt"]; !ok {
		t.Error("expected callGraphBuilt in stats")
	}
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)
	container.SetProject("/test/project")

	done := make(chan bool, 10)

	for range 10 {
		go func() {
			_ = container.GetSymbolIndex()
			_ = container.GetCallGraph()
			_ = container.GetContextMemory()
			_ = container.Stats()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

func TestContainer_EnsureSymbolIndexBuilt_NoProject(t *testing.T) {
	logger := &domain.NoopLogger{}
	container := NewContainer(logger)

	err := container.EnsureSymbolIndexBuilt(context.Background())
	if err != nil {
		t.Errorf("expected nil error without project, got %v", err)
	}
}

type mockSemanticSearcher struct{}

func (m *mockSemanticSearcher) Search(query string, limit int) ([]domain.SemanticSearchResult, error) {
	return nil, nil
}
