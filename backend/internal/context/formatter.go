package context

import (
	"fmt"
	"path/filepath"
	"strings"
)

// OutputFormat defines the format for context output
type OutputFormat string

const (
	FormatMarkdown OutputFormat = "markdown" // Default: ## File: path\n```lang\ncontent\n```
	FormatXML      OutputFormat = "xml"      // <file path="..."><content>...</content></file>
	FormatJSON     OutputFormat = "json"     // {"files": [{"path": "...", "content": "..."}]}
	FormatPlain    OutputFormat = "plain"    // --- File: path ---\ncontent
)

// formatFileHeader returns the header for a file based on output format
func formatFileHeader(filePath string, format OutputFormat) string {
	switch format {
	case FormatXML:
		return fmt.Sprintf("<file path=\"%s\">\n<content>\n", filePath)
	case FormatJSON:
		return "" // JSON is handled separately
	case FormatPlain:
		return fmt.Sprintf("--- File: %s ---\n", filePath)
	default: // FormatMarkdown
		ext := filepath.Ext(filePath)
		lang := ""
		if len(ext) > 1 {
			lang = ext[1:]
		}
		return fmt.Sprintf("## File: %s\n\n```%s\n", filePath, lang)
	}
}

// formatFileFooter returns the footer for a file based on output format
func formatFileFooter(format OutputFormat) string {
	switch format {
	case FormatXML:
		return "\n</content>\n</file>\n\n"
	case FormatJSON:
		return "" // JSON is handled separately
	case FormatPlain:
		return "\n\n"
	default: // FormatMarkdown
		return "\n```\n\n"
	}
}

// escapeForFormat escapes content based on output format
func escapeForFormat(content string, format OutputFormat) string {
	switch format {
	case FormatXML:
		// Escape XML special characters
		content = strings.ReplaceAll(content, "&", "&amp;")
		content = strings.ReplaceAll(content, "<", "&lt;")
		content = strings.ReplaceAll(content, ">", "&gt;")
		return content
	case FormatJSON:
		// JSON escaping handled by json.Marshal
		return content
	default:
		return content
	}
}

// Service method wrappers for backward compatibility
func (s *Service) formatFileHeader(filePath string, format OutputFormat) string {
	return formatFileHeader(filePath, format)
}

func (s *Service) formatFileFooter(format OutputFormat) string {
	return formatFileFooter(format)
}

func (s *Service) escapeForFormat(content string, format OutputFormat) string {
	return escapeForFormat(content, format)
}
