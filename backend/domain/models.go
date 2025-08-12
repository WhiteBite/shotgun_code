package domain

type FileNode struct {
	Name            string      `json:"name"`
	Path            string      `json:"path"`
	RelPath         string      `json:"relPath"`
	IsDir           bool        `json:"isDir"`
	Size            int64       `json:"size"` // Added file size
	Children        []*FileNode `json:"children,omitempty"`
	IsGitignored    bool        `json:"isGitignored"`
	IsCustomIgnored bool        `json:"isCustomIgnored"`
}

type FileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type Commit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

type CommitWithFiles struct {
	Hash    string   `json:"hash"`
	Subject string   `json:"subject"`
	Author  string   `json:"author"`
	Date    string   `json:"date"`
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
