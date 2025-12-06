package textutils

import (
	"fmt"
	"strings"
	"testing"
)

// Uses mockLogger from context_splitter_test.go

func TestNewDiffSplitter(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})
	if ds == nil {
		t.Fatal("NewDiffSplitter returned nil")
	}
}

func TestDiffSplitter_Split_Empty(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	result, err := ds.Split("", 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 splits for empty input, got %d", len(result))
	}
}

func TestDiffSplitter_Split_SingleFile(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
`

	result, err := ds.Split(diff, 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected at least 1 split")
	}
	if !strings.Contains(result[0], "diff --git") {
		t.Error("result should contain diff header")
	}
}

func TestDiffSplitter_Split_MultipleFiles(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	diff := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1 +1 @@
-old
+new

diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1 +1 @@
-old2
+new2
`

	result, err := ds.Split(diff, 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected at least 1 split")
	}
}

func TestDiffSplitter_Split_LargeFile(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	// Create a large diff with multiple hunks to trigger splitting
	var sb strings.Builder
	sb.WriteString("diff --git a/large.go b/large.go\n")
	sb.WriteString("--- a/large.go\n")
	sb.WriteString("+++ b/large.go\n")
	// Add multiple hunks so the splitter can split by hunk
	for h := 0; h < 5; h++ {
		sb.WriteString(fmt.Sprintf("@@ -%d,10 +%d,10 @@\n", h*10+1, h*10+1))
		for i := 0; i < 10; i++ {
			sb.WriteString("+line " + string(rune('a'+i%26)) + "\n")
		}
	}

	result, err := ds.Split(sb.String(), 20)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	// With 5 hunks of ~11 lines each (header + 10 lines) and limit 20, we should get multiple splits
	if len(result) < 2 {
		t.Errorf("expected multiple splits for large diff with multiple hunks, got %d", len(result))
	}
}

func TestDiffSplitter_Split_NoDiffHeader(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	// Content without diff --git header
	content := `--- a/file.go
+++ b/file.go
@@ -1 +1 @@
-old
+new
`

	result, err := ds.Split(content, 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	// Should still return the content
	if len(result) == 0 {
		t.Error("expected at least 1 split even without diff header")
	}
}

func TestDiffSplitter_Split_WhitespaceOnly(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	result, err := ds.Split("   \n\t\n  ", 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 splits for whitespace-only input, got %d", len(result))
	}
}

func TestDiffSplitter_Split_MultipleHunks(t *testing.T) {
	ds := NewDiffSplitter(&mockLogger{})

	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 line1
-old1
+new1
@@ -10,3 +10,3 @@
 line10
-old10
+new10
`

	result, err := ds.Split(diff, 100)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected at least 1 split")
	}
	// Should contain both hunks
	combined := strings.Join(result, "\n")
	if !strings.Contains(combined, "@@ -1,3") {
		t.Error("should contain first hunk")
	}
	if !strings.Contains(combined, "@@ -10,3") {
		t.Error("should contain second hunk")
	}
}
