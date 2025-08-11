package domain

type FileNode struct {
	Name            string      `json:"name"`
	Path            string      `json:"path"`
	RelPath         string      `json:"relPath"`
	IsDir           bool        `json:"isDir"`
	Children        []*FileNode `json:"children,omitempty"`
	IsGitignored    bool        `json:"isGitignored"`
	IsCustomIgnored bool        `json:"isCustomIgnored"`
}

// FileStatus represents the status of a file in Git.
type FileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"` // e.g., "M", "A", "D", "R", "C", "U" for Untracked as '??' -> 'U'
}

type Commit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

// CommitWithFiles extends Commit with a list of files changed in it and a merge flag.
type CommitWithFiles struct {
	Hash    string   `json:"hash"`
	Subject string   `json:"subject"`
	Files   []string `json:"files"`
	IsMerge bool     `json:"isMerge"`
}

type ParsedDiff struct {
	FileDiffs []FileDiff `json:"fileDiffs"`
}

type FileDiff struct {
	FilePath string `json:"filePath"`
	Hunks    []Hunk `json:"hunks"`
}

type Hunk struct {
	Header string   `json:"header"`
	Lines  []string `json:"lines"`
}
