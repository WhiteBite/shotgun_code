package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"shotgun_code/domain"
	"testing"
)

// Unit test for the parser logic
func TestParseRichLogOutput(t *testing.T) {
	testCases := []struct {
		name        string
		gitOutput   string
		expected    []domain.CommitWithFiles
		expectError bool
	}{
		{
			name: "Standard Commit",
			gitOutput: `COMMIT 1234567 89abcde
feat: add new feature
M	README.md
A	src/feature.js
`,
			expected: []domain.CommitWithFiles{
				{Hash: "1234567", Subject: "feat: add new feature", IsMerge: false, Files: []string{"README.md", "src/feature.js"}},
			},
			expectError: false,
		},
		{
			name: "Merge Commit",
			gitOutput: `COMMIT abcdef0 1234567 fedcba9
Merge pull request #123
M	package.json
`,
			expected: []domain.CommitWithFiles{
				{Hash: "abcdef0", Subject: "Merge pull request #123", IsMerge: true, Files: []string{"package.json"}},
			},
			expectError: false,
		},
		{
			name: "Commit with no files",
			gitOutput: `COMMIT fedcba9 89abcde
docs: update contributing guide
`,
			expected: []domain.CommitWithFiles{
				{Hash: "fedcba9", Subject: "docs: update contributing guide", IsMerge: false, Files: []string{}},
			},
			expectError: false,
		},
		{
			name: "Multiple Commits",
			gitOutput: `COMMIT abcdef0 1234567 fedcba9
Merge pull request #123
M	package.json
COMMIT 1234567 89abcde
feat: add new feature
M	README.md
`,
			expected: []domain.CommitWithFiles{
				{Hash: "abcdef0", Subject: "Merge pull request #123", IsMerge: true, Files: []string{"package.json"}},
				{Hash: "1234567", Subject: "feat: add new feature", IsMerge: false, Files: []string{"README.md"}},
			},
			expectError: false,
		},
		{
			name:        "Empty Input",
			gitOutput:   "",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "Corrupted Input",
			gitOutput:   "COMMIT \nsubject",
			expected:    nil,
			expectError: false, // Parser should be resilient
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseRichLogOutput(tc.gitOutput)

			if (err != nil) != tc.expectError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectError, err)
			}

			if len(result) == 0 && len(tc.expected) == 0 {
				return // Both are nil or empty, which is a pass
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v", tc.expected, result)
			}
		})
	}
}

// Integration test for the full git command execution
func TestGetRichCommitHistory_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	dir, err := os.MkdirTemp("", "git-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	runGit := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		err := cmd.Run()
		if err != nil {
			t.Fatalf("Git command failed: %v", err)
		}
	}

	// Setup git repo
	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")

	// First commit on master
	if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}
	runGit("add", "file1.txt")
	runGit("commit", "-m", "Initial commit")

	// Create and switch to a new branch
	runGit("checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("feature"), 0644); err != nil {
		t.Fatalf("Failed to write feature file: %v", err)
	}
	runGit("add", "file2.txt")
	runGit("commit", "-m", "feat: add file2")

	// Switch back and merge
	runGit("checkout", "master")
	runGit("merge", "--no-ff", "feature", "-m", "Merge feature branch")

	repo := New(&testLogger{})
	commits, err := repo.GetRichCommitHistory(dir, "", 3)

	if err != nil {
		t.Fatalf("GetRichCommitHistory failed: %v", err)
	}

	if len(commits) != 3 {
		t.Fatalf("Expected 3 commits, got %d", len(commits))
	}

	// With --topo-order, the order is predictable: merge, then feature, then initial.
	if !commits[0].IsMerge || commits[0].Subject != "Merge feature branch" {
		t.Errorf("Expected merge commit at index 0, got %+v", commits[0])
	}
	if commits[1].IsMerge || commits[1].Subject != "feat: add file2" || len(commits[1].Files) != 1 || commits[1].Files[0] != "file2.txt" {
		t.Errorf("Expected feature commit at index 1, got %+v", commits[1])
	}
	if commits[2].IsMerge || commits[2].Subject != "Initial commit" || len(commits[2].Files) != 1 || commits[2].Files[0] != "file1.txt" {
		t.Errorf("Expected initial commit at index 2, got %+v", commits[2])
	}
}

type testLogger struct{}

func (l *testLogger) Debug(message string)   {}
func (l *testLogger) Info(message string)    {}
func (l *testLogger) Warning(message string) {}
func (l *testLogger) Error(message string)   {}
func (l *testLogger) Fatal(message string)   {}
