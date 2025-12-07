package contextbuilder

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

type BuildOptions struct {
	StripComments   bool
	IncludeManifest bool
}

// entry — одна запись контекста
type entry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// --- File: path --- парсер
var headerRe = regexp.MustCompile(`(?m)^--- File:\s+(.*?)\s+---\s*`)

func parseContext(ctx string) []entry {
	res := []entry{}
	idxs := headerRe.FindAllStringIndex(ctx, -1)
	if len(idxs) == 0 {
		return res
	}
	for i := 0; i < len(idxs); i++ {
		start := idxs[i][0]
		end := len(ctx)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		block := ctx[start:end]
		m := headerRe.FindStringSubmatch(block)
		if len(m) < 2 {
			continue
		}
		path := strings.TrimSpace(m[1])
		content := strings.TrimSpace(block[len(m[0]):])
		res = append(res, entry{Path: path, Content: content})
	}

	// Обеспечиваем детерминированный порядок
	sort.Slice(res, func(i, j int) bool {
		return res[i].Path < res[j].Path
	})

	return res
}

// Грубая очистка комментариев
func stripComments(text string) string {
	lineRE := regexp.MustCompile(`(?m)^\s*(//|#).*$`)
	blockRE := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	out := lineRE.ReplaceAllString(text, "")
	out = blockRE.ReplaceAllString(out, "")
	return out
}

// Построение ASCII-дерева из путей
func buildTree(paths []string) string {
	type node struct {
		name     string
		children map[string]*node
	}
	root := &node{name: ".", children: map[string]*node{}}

	// Сортируем входные пути для детерминизма
	sortedPaths := make([]string, len(paths))
	copy(sortedPaths, paths)
	sort.Strings(sortedPaths)

	for _, p := range sortedPaths {
		// нормализуем слеши
		pp := strings.ReplaceAll(p, "\\", "/")
		parts := strings.Split(pp, "/")
		cur := root
		for _, part := range parts {
			if part == "" || part == "." {
				continue
			}
			if cur.children[part] == nil {
				cur.children[part] = &node{name: part, children: map[string]*node{}}
			}
			cur = cur.children[part]
		}
	}

	var b strings.Builder
	var walk func(n *node, prefix string, isLast bool)
	walk = func(n *node, prefix string, isLast bool) {
		if n != root {
			if isLast {
				b.WriteString(prefix + "└─ " + n.name + "\n")
				prefix += "   "
			} else {
				b.WriteString(prefix + "├─ " + n.name + "\n")
				prefix += "│  "
			}
		}
		// детерминированный порядок
		keys := make([]string, 0, len(n.children))
		for k := range n.children {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			walk(n.children[k], prefix, i == len(keys)-1)
		}
	}
	walk(root, "", true)
	return b.String()
}

// buildPlainFormat builds plain text format output
func buildPlainFormat(entries []entry, opts BuildOptions) string {
	var b strings.Builder
	for _, e := range entries {
		content := e.Content
		if opts.StripComments {
			content = stripComments(content)
		}
		b.WriteString("--- File: " + e.Path + " ---\n")
		b.WriteString(content)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// buildManifestFormat builds manifest format output with tree
func buildManifestFormat(entries []entry, opts BuildOptions) string {
	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		paths = append(paths, e.Path)
	}
	var b strings.Builder
	b.WriteString("Manifest:\n")
	b.WriteString(buildTree(paths))
	b.WriteString("\n")
	for _, e := range entries {
		content := e.Content
		if opts.StripComments {
			content = stripComments(content)
		}
		b.WriteString("--- File: " + e.Path + " ---\n")
		b.WriteString(content)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// buildJSONFormat builds JSON format output
func buildJSONFormat(entries []entry, opts BuildOptions) (string, error) {
	j := make([]entry, 0, len(entries))
	for _, e := range entries {
		c := e.Content
		if opts.StripComments {
			c = stripComments(c)
		}
		j = append(j, entry{Path: e.Path, Content: c})
	}
	raw, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// BuildFromContext — собирает строку по формату: "plain" | "manifest" | "json"
func BuildFromContext(format string, ctx string, opts BuildOptions) (string, error) {
	entries := parseContext(ctx)

	switch strings.ToLower(format) {
	case "plain":
		return buildPlainFormat(entries, opts), nil
	case "manifest", "manifest+text":
		return buildManifestFormat(entries, opts), nil
	case "json":
		return buildJSONFormat(entries, opts)
	default:
		return BuildFromContext("manifest", ctx, opts)
	}
}
