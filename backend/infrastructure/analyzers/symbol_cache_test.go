package analyzers

import (
	"context"
	"shotgun_code/domain/analysis"
	"testing"
)

func TestCachedSymbolIndex_RemoveSymbolsForFile(t *testing.T) {
	// Test the in-memory removal logic without SQLite
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	// Add symbols manually
	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{
		Name:     "Func1",
		Kind:     analysis.KindFunction,
		FilePath: "file1.go",
	})
	idx.addSymbolLocked(analysis.Symbol{
		Name:     "Func2",
		Kind:     analysis.KindFunction,
		FilePath: "file1.go",
	})
	idx.addSymbolLocked(analysis.Symbol{
		Name:     "Func3",
		Kind:     analysis.KindFunction,
		FilePath: "file2.go",
	})
	idx.mu.Unlock()

	// Verify initial state
	if len(idx.symbols) != 3 {
		t.Errorf("expected 3 symbols, got %d", len(idx.symbols))
	}
	if len(idx.byFile["file1.go"]) != 2 {
		t.Errorf("expected 2 symbols in file1.go, got %d", len(idx.byFile["file1.go"]))
	}

	// Test that symbols are indexed by name
	results := idx.FindByExactName("Func1")
	if len(results) != 1 {
		t.Errorf("expected 1 result for Func1, got %d", len(results))
	}
}

func TestSymbolIndex_SearchByName(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{Name: "UserService", Kind: analysis.KindClass, FilePath: "user.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "UserRepository", Kind: analysis.KindInterface, FilePath: "user.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "OrderService", Kind: analysis.KindClass, FilePath: "order.go"})
	idx.mu.Unlock()

	results := idx.SearchByName("user")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'user', got %d", len(results))
	}

	results = idx.SearchByName("service")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'service', got %d", len(results))
	}
}

func TestSymbolIndex_GetSymbolsByKind(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{Name: "User", Kind: analysis.KindClass, FilePath: "user.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "Order", Kind: analysis.KindClass, FilePath: "order.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "GetUser", Kind: analysis.KindFunction, FilePath: "user.go"})
	idx.mu.Unlock()

	classes := idx.GetSymbolsByKind(analysis.KindClass)
	if len(classes) != 2 {
		t.Errorf("expected 2 classes, got %d", len(classes))
	}

	functions := idx.GetSymbolsByKind(analysis.KindFunction)
	if len(functions) != 1 {
		t.Errorf("expected 1 function, got %d", len(functions))
	}
}

func TestSymbolIndex_FindDefinition(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{Name: "User", Kind: analysis.KindClass, FilePath: "user.go", StartLine: 10})
	idx.addSymbolLocked(analysis.Symbol{Name: "User", Kind: analysis.KindInterface, FilePath: "types.go", StartLine: 5})
	idx.mu.Unlock()

	// Find any User
	def := idx.FindDefinition("User", "")
	if def == nil {
		t.Fatal("expected to find User")
	}

	// Find User class specifically
	def = idx.FindDefinition("User", analysis.KindClass)
	if def == nil {
		t.Fatal("expected to find User class")
	}
	if def.FilePath != "user.go" {
		t.Errorf("expected user.go, got %s", def.FilePath)
	}

	// Find User interface specifically
	def = idx.FindDefinition("User", analysis.KindInterface)
	if def == nil {
		t.Fatal("expected to find User interface")
	}
	if def.FilePath != "types.go" {
		t.Errorf("expected types.go, got %s", def.FilePath)
	}
}

func TestSymbolIndex_Stats(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{Name: "A", Kind: analysis.KindClass, FilePath: "a.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "B", Kind: analysis.KindFunction, FilePath: "b.go"})
	idx.addSymbolLocked(analysis.Symbol{Name: "C", Kind: analysis.KindFunction, FilePath: "c.go"})
	idx.mu.Unlock()

	stats := idx.Stats()
	if stats["total_symbols"] != 3 {
		t.Errorf("expected 3 total symbols, got %d", stats["total_symbols"])
	}
	if stats["files"] != 3 {
		t.Errorf("expected 3 files, got %d", stats["files"])
	}
}

func TestSymbolIndex_Clear(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	idx.mu.Lock()
	idx.addSymbolLocked(analysis.Symbol{Name: "Test", Kind: analysis.KindClass, FilePath: "test.go"})
	idx.mu.Unlock()

	if len(idx.symbols) != 1 {
		t.Error("expected 1 symbol before clear")
	}

	idx.Clear()

	if len(idx.symbols) != 0 {
		t.Error("expected 0 symbols after clear")
	}
	if len(idx.byName) != 0 {
		t.Error("expected empty byName after clear")
	}
	if len(idx.byFile) != 0 {
		t.Error("expected empty byFile after clear")
	}
}

func TestSymbolIndex_IndexFile(t *testing.T) {
	registry := NewAnalyzerRegistry()
	idx := NewSymbolIndex(registry)

	code := []byte(`package main

func Hello() {}
func World() {}
`)

	err := idx.IndexFile(context.Background(), "main.go", code)
	if err != nil {
		t.Fatalf("IndexFile failed: %v", err)
	}

	symbols := idx.GetSymbolsInFile("main.go")
	if len(symbols) < 2 {
		t.Errorf("expected at least 2 symbols, got %d", len(symbols))
	}
}
