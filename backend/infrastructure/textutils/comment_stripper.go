package textutils

import (
    "path/filepath"
    "shotgun_code/domain"
    "strings"
)

type commentStripperImpl struct {
    log domain.Logger
}

// NewCommentStripper creates a new implementation of domain.CommentStripper
func NewCommentStripper(log domain.Logger) domain.CommentStripper {
    return &commentStripperImpl{log: log}
}

// Strip removes comments from code content based on file extension
func (c *commentStripperImpl) Strip(content, filePath string) string {
    ext := strings.ToLower(filepath.Ext(filePath))
    switch ext {
    case ".go", ".js", ".ts", ".java", ".c", ".cpp", ".cs":
        return stripCStyleComments(content)
    case ".py", ".sh":
        return stripHashComments(content)
    case ".html", ".xml":
        return stripXMLComments(content)
    default:
        return content
    }
}

// stripCStyleComments removes C-style comments (// and /* */)
func stripCStyleComments(content string) string {
    lines := strings.Split(content, "\n")
    var result []string

    inBlockComment := false

    for _, line := range lines {
        if inBlockComment {
            if strings.Contains(line, "*/") {
                parts := strings.SplitN(line, "*/", 2)
                if len(parts) > 1 {
                    line = parts[1]
                    inBlockComment = false
                } else {
                    continue
                }
            } else {
                continue
            }
        }

        if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
            parts := strings.SplitN(line, "/*", 2)
            line = parts[0]
            inBlockComment = true
        }

        if idx := strings.Index(line, "//"); idx != -1 {
            line = line[:idx]
        }

        for strings.Contains(line, "/*") && strings.Contains(line, "*/") {
            start := strings.Index(line, "/*")
            end := strings.Index(line, "*/")
            if end > start {
                line = line[:start] + line[end+2:]
            }
        }

        if trimmed := strings.TrimSpace(line); trimmed != "" {
            result = append(result, line)
        }
    }

    return strings.Join(result, "\n")
}

// stripHashComments removes hash-style comments (#)
func stripHashComments(content string) string {
    lines := strings.Split(content, "\n")
    var result []string

    for _, line := range lines {
        if idx := strings.Index(line, "#"); idx != -1 {
            line = line[:idx]
        }
        if trimmed := strings.TrimSpace(line); trimmed != "" {
            result = append(result, line)
        }
    }

    return strings.Join(result, "\n")
}

// stripXMLComments removes XML-style comments (<!-- -->)
func stripXMLComments(content string) string {
    for strings.Contains(content, "<!--") && strings.Contains(content, "-->") {
        start := strings.Index(content, "<!--")
        end := strings.Index(content, "-->")
        if end > start {
            content = content[:start] + content[end+3:]
        }
    }
    return content
}
