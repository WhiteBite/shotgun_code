package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"shotgun_code/domain"
	"strconv"
	"time"
)

// SemanticTools provides semantic search tools for AI agents
type SemanticTools struct {
	semanticSearch domain.SemanticSearchService
	ragService     domain.RAGService
	log            domain.Logger
}

// NewSemanticTools creates new semantic tools
func NewSemanticTools(
	semanticSearch domain.SemanticSearchService,
	ragService domain.RAGService,
	log domain.Logger,
) *SemanticTools {
	return &SemanticTools{
		semanticSearch: semanticSearch,
		ragService:     ragService,
		log:            log,
	}
}

// GetTools returns all semantic search tools
func (t *SemanticTools) GetTools() []domain.Tool {
	return []domain.Tool{
		{
			Name:        "semantic_search",
			Description: "Search code by meaning/intent. Use for queries like 'find code that handles user authentication' or 'where are errors processed'. Returns semantically similar code chunks.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"query": {
						Type:        "string",
						Description: "Natural language description of what you're looking for",
					},
					"top_k": {
						Type:        "integer",
						Description: "Number of results to return (default: 10)",
						Default:     10,
					},
					"min_score": {
						Type:        "number",
						Description: "Minimum similarity score 0-1 (default: 0.5)",
						Default:     0.5,
					},
					"languages": {
						Type:        "array",
						Description: "Filter by programming languages (e.g., ['go', 'typescript'])",
					},
					"search_type": {
						Type:        "string",
						Description: "Search type: 'semantic', 'keyword', or 'hybrid' (default: hybrid)",
						Enum:        []string{"semantic", "keyword", "hybrid"},
						Default:     "hybrid",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "find_similar_code",
			Description: "Find code similar to a given code snippet. Useful for finding duplicate code, similar implementations, or related functionality.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"file_path": {
						Type:        "string",
						Description: "Path to the file containing the source code",
					},
					"start_line": {
						Type:        "integer",
						Description: "Starting line number of the code snippet",
					},
					"end_line": {
						Type:        "integer",
						Description: "Ending line number of the code snippet",
					},
					"top_k": {
						Type:        "integer",
						Description: "Number of similar results to return (default: 5)",
						Default:     5,
					},
				},
				Required: []string{"file_path", "start_line", "end_line"},
			},
		},
		{
			Name:        "get_relevant_context",
			Description: "Get relevant code context for a task or question. Uses RAG to retrieve the most relevant code chunks that fit within a token budget.",
			Parameters: domain.ToolParameters{
				Type: "object",
				Properties: map[string]domain.ToolProperty{
					"query": {
						Type:        "string",
						Description: "Description of the task or question",
					},
					"max_tokens": {
						Type:        "integer",
						Description: "Maximum tokens for context (default: 4000)",
						Default:     4000,
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "index_project",
			Description: "Index or re-index the project for semantic search. Run this if semantic search returns no results or after major code changes.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
		{
			Name:        "semantic_search_stats",
			Description: "Get statistics about the semantic search index for the current project.",
			Parameters: domain.ToolParameters{
				Type:       "object",
				Properties: map[string]domain.ToolProperty{},
			},
		},
	}
}

// ExecuteTool executes a semantic search tool
func (t *SemanticTools) ExecuteTool(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	switch call.Name {
	case "semantic_search":
		return t.executeSemanticSearch(ctx, call, projectRoot)
	case "find_similar_code":
		return t.executeFindSimilar(ctx, call, projectRoot)
	case "get_relevant_context":
		return t.executeGetContext(ctx, call, projectRoot)
	case "index_project":
		return t.executeIndexProject(ctx, call, projectRoot)
	case "semantic_search_stats":
		return t.executeGetStats(ctx, call, projectRoot)
	default:
		return domain.ToolResult{
			ToolCallID: call.ID,
			Error:      fmt.Sprintf("unknown tool: %s", call.Name),
		}
	}
}

func (t *SemanticTools) executeSemanticSearch(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	query, _ := call.Arguments["query"].(string)
	if query == "" {
		return domain.ToolResult{ToolCallID: call.ID, Error: "query is required"}
	}
	
	topK := 10
	if v, ok := call.Arguments["top_k"]; ok {
		topK = toInt(v, 10)
	}
	
	minScore := float32(0.5)
	if v, ok := call.Arguments["min_score"]; ok {
		minScore = float32(toFloat(v, 0.5))
	}
	
	searchType := domain.SearchTypeHybrid
	if v, ok := call.Arguments["search_type"].(string); ok {
		switch v {
		case "semantic":
			searchType = domain.SearchTypeSemantic
		case "keyword":
			searchType = domain.SearchTypeKeyword
		}
	}
	
	var languages []string
	if v, ok := call.Arguments["languages"].([]interface{}); ok {
		for _, lang := range v {
			if s, ok := lang.(string); ok {
				languages = append(languages, s)
			}
		}
	}
	
	req := domain.SemanticSearchRequest{
		Query:       query,
		ProjectRoot: projectRoot,
		TopK:        topK,
		MinScore:    minScore,
		SearchType:  searchType,
	}
	
	if len(languages) > 0 {
		req.Filters = &domain.SearchFilters{Languages: languages}
	}
	
	results, err := t.semanticSearch.Search(ctx, req)
	if err != nil {
		return domain.ToolResult{ToolCallID: call.ID, Error: err.Error()}
	}
	
	// Format results
	output := formatSearchResults(results)
	
	return domain.ToolResult{
		ToolCallID: call.ID,
		Content:    output,
	}
}

func (t *SemanticTools) executeFindSimilar(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	filePath, _ := call.Arguments["file_path"].(string)
	startLine := toInt(call.Arguments["start_line"], 0)
	endLine := toInt(call.Arguments["end_line"], 0)
	topK := toInt(call.Arguments["top_k"], 5)
	
	if filePath == "" || startLine == 0 || endLine == 0 {
		return domain.ToolResult{ToolCallID: call.ID, Error: "file_path, start_line, and end_line are required"}
	}
	
	req := domain.SimilarCodeRequest{
		FilePath:    filePath,
		StartLine:   startLine,
		EndLine:     endLine,
		TopK:        topK,
		MinScore:    0.5,
		ExcludeSelf: true,
	}
	
	results, err := t.semanticSearch.FindSimilar(ctx, req)
	if err != nil {
		return domain.ToolResult{ToolCallID: call.ID, Error: err.Error()}
	}
	
	output := formatSearchResults(results)
	
	return domain.ToolResult{
		ToolCallID: call.ID,
		Content:    output,
	}
}

func (t *SemanticTools) executeGetContext(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	query, _ := call.Arguments["query"].(string)
	if query == "" {
		return domain.ToolResult{ToolCallID: call.ID, Error: "query is required"}
	}
	
	maxTokens := toInt(call.Arguments["max_tokens"], 4000)
	
	chunks, err := t.ragService.RetrieveContext(ctx, query, projectRoot, maxTokens)
	if err != nil {
		return domain.ToolResult{ToolCallID: call.ID, Error: err.Error()}
	}
	
	// Format as context
	output := formatContextChunks(chunks)
	
	return domain.ToolResult{
		ToolCallID: call.ID,
		Content:    output,
	}
}

func (t *SemanticTools) executeIndexProject(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	t.log.Info("Starting project indexing via tool call")
	
	// Run indexing in background
	go func() {
		indexCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		
		if err := t.semanticSearch.IndexProject(indexCtx, projectRoot); err != nil {
			t.log.Error(fmt.Sprintf("Indexing failed: %v", err))
		} else {
			t.log.Info("Indexing completed successfully")
		}
	}()
	
	return domain.ToolResult{
		ToolCallID: call.ID,
		Content:    "Project indexing started in background. This may take a few minutes depending on project size.",
	}
}

func (t *SemanticTools) executeGetStats(ctx context.Context, call domain.ToolCall, projectRoot string) domain.ToolResult {
	stats, err := t.semanticSearch.GetStats(ctx, projectRoot)
	if err != nil {
		return domain.ToolResult{ToolCallID: call.ID, Error: err.Error()}
	}
	
	output := fmt.Sprintf(`Semantic Search Index Statistics:
- Total chunks: %d
- Total files: %d
- Total tokens: %d
- Index size: %.2f MB
- Dimensions: %d
- Last updated: %s
- Indexed: %v`,
		stats.TotalChunks,
		stats.TotalFiles,
		stats.TotalTokens,
		float64(stats.IndexSize)/(1024*1024),
		stats.Dimensions,
		stats.LastUpdated.Format(time.RFC3339),
		stats.TotalChunks > 0,
	)
	
	return domain.ToolResult{
		ToolCallID: call.ID,
		Content:    output,
	}
}

// Helper functions

func formatSearchResults(results *domain.SemanticSearchResponse) string {
	if results == nil || len(results.Results) == 0 {
		return "No results found. The project may not be indexed yet. Use 'index_project' tool to index it."
	}
	
	var sb stringBuilder
	sb.WriteString(fmt.Sprintf("Found %d results (query time: %v):\n\n", results.TotalResults, results.QueryTime))
	
	for i, r := range results.Results {
		sb.WriteString(fmt.Sprintf("### Result %d (score: %.3f)\n", i+1, r.Score))
		sb.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n", r.Chunk.FilePath, r.Chunk.StartLine, r.Chunk.EndLine))
		
		if r.Chunk.SymbolName != "" {
			sb.WriteString(fmt.Sprintf("Symbol: %s %s\n", r.Chunk.SymbolKind, r.Chunk.SymbolName))
		}
		
		sb.WriteString(fmt.Sprintf("Language: %s\n", r.Chunk.Language))
		sb.WriteString("```")
		sb.WriteString(r.Chunk.Language)
		sb.WriteString("\n")
		
		// Truncate very long content
		content := r.Chunk.Content
		if len(content) > 1000 {
			content = content[:1000] + "\n... (truncated)"
		}
		sb.WriteString(content)
		sb.WriteString("\n```\n\n")
	}
	
	return sb.String()
}

func formatContextChunks(chunks []domain.CodeChunk) string {
	if len(chunks) == 0 {
		return "No relevant context found."
	}
	
	var sb stringBuilder
	sb.WriteString(fmt.Sprintf("Retrieved %d relevant code chunks:\n\n", len(chunks)))
	
	currentFile := ""
	for _, chunk := range chunks {
		if chunk.FilePath != currentFile {
			sb.WriteString(fmt.Sprintf("\n## File: %s\n", chunk.FilePath))
			currentFile = chunk.FilePath
		}
		
		if chunk.SymbolName != "" {
			sb.WriteString(fmt.Sprintf("\n### %s %s (lines %d-%d)\n", 
				chunk.SymbolKind, chunk.SymbolName, chunk.StartLine, chunk.EndLine))
		} else {
			sb.WriteString(fmt.Sprintf("\n### Lines %d-%d\n", chunk.StartLine, chunk.EndLine))
		}
		
		sb.WriteString("```")
		sb.WriteString(chunk.Language)
		sb.WriteString("\n")
		sb.WriteString(chunk.Content)
		sb.WriteString("\n```\n")
	}
	
	return sb.String()
}

type stringBuilder struct {
	data []byte
}

func (sb *stringBuilder) WriteString(s string) {
	sb.data = append(sb.data, s...)
}

func (sb *stringBuilder) String() string {
	return string(sb.data)
}

func toInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i)
		}
	}
	return defaultVal
}

func toFloat(v interface{}, defaultVal float64) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f
		}
	}
	return defaultVal
}
