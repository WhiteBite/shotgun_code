package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
	"sort"
	"strings"
	"sync"
	"time"
)

// SemanticSearchServiceImpl implements SemanticSearchService
type SemanticSearchServiceImpl struct {
	embeddingProvider domain.EmbeddingProvider
	vectorStore       domain.VectorStore
	symbolIndex       analysis.SymbolIndex
	log               domain.Logger
	
	// Indexing state
	indexingMu    sync.RWMutex
	indexingState map[string]*IndexingState
}

// IndexingState tracks the state of project indexing
type IndexingState struct {
	ProjectID   string
	InProgress  bool
	Progress    float64
	TotalFiles  int
	IndexedFiles int
	StartedAt   time.Time
	Error       error
}

// NewSemanticSearchService creates a new semantic search service
func NewSemanticSearchService(
	embeddingProvider domain.EmbeddingProvider,
	vectorStore domain.VectorStore,
	symbolIndex analysis.SymbolIndex,
	log domain.Logger,
) *SemanticSearchServiceImpl {
	return &SemanticSearchServiceImpl{
		embeddingProvider: embeddingProvider,
		vectorStore:       vectorStore,
		symbolIndex:       symbolIndex,
		log:               log,
		indexingState:     make(map[string]*IndexingState),
	}
}

// IndexProject indexes all files in a project
func (s *SemanticSearchServiceImpl) IndexProject(ctx context.Context, projectRoot string) error {
	projectID := generateProjectID(projectRoot)
	
	// Check if already indexing
	s.indexingMu.Lock()
	if state, exists := s.indexingState[projectID]; exists && state.InProgress {
		s.indexingMu.Unlock()
		return fmt.Errorf("indexing already in progress for project")
	}
	
	state := &IndexingState{
		ProjectID:  projectID,
		InProgress: true,
		StartedAt:  time.Now(),
	}
	s.indexingState[projectID] = state
	s.indexingMu.Unlock()
	
	defer func() {
		s.indexingMu.Lock()
		state.InProgress = false
		s.indexingMu.Unlock()
	}()
	
	s.log.Info(fmt.Sprintf("Starting semantic indexing for project: %s", projectRoot))
	
	// Ensure symbol index is up-to-date before chunking
	// For CachedSymbolIndex this will use SQLite cache for incremental indexing
	if s.symbolIndex != nil {
		s.log.Info("Indexing symbols for project...")
		if err := s.symbolIndex.IndexProject(ctx, projectRoot); err != nil {
			s.log.Warning(fmt.Sprintf("Symbol indexing failed (non-critical): %v", err))
		}
	}
	
	// Collect files to index
	var files []string
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// Skip directories
		if info.IsDir() {
			name := info.Name()
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Only index code files
		if isCodeFile(path) {
			relPath, _ := filepath.Rel(projectRoot, path)
			files = append(files, relPath)
		}
		
		return nil
	})
	if err != nil {
		state.Error = err
		return fmt.Errorf("failed to walk project: %w", err)
	}
	
	state.TotalFiles = len(files)
	s.log.Info(fmt.Sprintf("Found %d files to index", len(files)))
	
	// Process files in batches
	batchSize := 10
	for i := 0; i < len(files); i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		
		batch := files[i:end]
		if err := s.indexFileBatch(ctx, projectRoot, projectID, batch); err != nil {
			s.log.Warning(fmt.Sprintf("Failed to index batch: %v", err))
		}
		
		state.IndexedFiles = end
		state.Progress = float64(end) / float64(len(files))
	}
	
	s.log.Info(fmt.Sprintf("Completed semantic indexing: %d files indexed", state.IndexedFiles))
	return nil
}

// indexFileBatch indexes a batch of files
func (s *SemanticSearchServiceImpl) indexFileBatch(ctx context.Context, projectRoot, projectID string, files []string) error {
	var allChunks []domain.CodeChunk
	
	for _, relPath := range files {
		fullPath := filepath.Join(projectRoot, relPath)
		
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		
		// Get symbols for better chunking
		var symbols []SymbolInfoForChunking
		if s.symbolIndex != nil {
			syms := s.symbolIndex.GetSymbolsInFile(relPath)
			for _, sym := range syms {
				symbols = append(symbols, SymbolInfoForChunking{
					Name:      sym.Name,
					Kind:      string(sym.Kind),
					StartLine: sym.StartLine,
					EndLine:   sym.EndLine,
				})
			}
		}
		
		// Chunk the file
		chunks := s.chunkFile(relPath, content, symbols)
		allChunks = append(allChunks, chunks...)
	}
	
	if len(allChunks) == 0 {
		return nil
	}
	
	// Generate embeddings for all chunks
	texts := make([]string, len(allChunks))
	for i, chunk := range allChunks {
		// Create embedding text with context
		texts[i] = createEmbeddingText(chunk)
	}
	
	// Generate embeddings with retry logic for API resilience
	resp, err := s.generateEmbeddingsWithRetry(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	// Create embedded chunks
	now := time.Now()
	embeddedChunks := make([]domain.EmbeddedChunk, len(allChunks))
	for i, chunk := range allChunks {
		embeddedChunks[i] = domain.EmbeddedChunk{
			Chunk:     chunk,
			Embedding: resp.Embeddings[i],
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	
	// Store in vector store
	return s.vectorStore.StoreBatch(ctx, projectID, embeddedChunks)
}

// IndexFile indexes a single file
func (s *SemanticSearchServiceImpl) IndexFile(ctx context.Context, projectRoot string, filePath string) error {
	projectID := generateProjectID(projectRoot)
	fullPath := filepath.Join(projectRoot, filePath)
	
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Delete existing chunks for this file
	if err := s.vectorStore.Delete(ctx, projectID, filePath); err != nil {
		s.log.Warning(fmt.Sprintf("Failed to delete existing chunks: %v", err))
	}
	
	// Get symbols
	var symbols []SymbolInfoForChunking
	if s.symbolIndex != nil {
		syms := s.symbolIndex.GetSymbolsInFile(filePath)
		for _, sym := range syms {
			symbols = append(symbols, SymbolInfoForChunking{
				Name:      sym.Name,
				Kind:      string(sym.Kind),
				StartLine: sym.StartLine,
				EndLine:   sym.EndLine,
			})
		}
	}
	
	// Chunk the file
	chunks := s.chunkFile(filePath, content, symbols)
	if len(chunks) == 0 {
		return nil
	}
	
	// Generate embeddings with retry logic for API resilience
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = createEmbeddingText(chunk)
	}
	
	resp, err := s.generateEmbeddingsWithRetry(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	// Store
	now := time.Now()
	embeddedChunks := make([]domain.EmbeddedChunk, len(chunks))
	for i, chunk := range chunks {
		embeddedChunks[i] = domain.EmbeddedChunk{
			Chunk:     chunk,
			Embedding: resp.Embeddings[i],
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	
	return s.vectorStore.StoreBatch(ctx, projectID, embeddedChunks)
}

// Search performs semantic search
func (s *SemanticSearchServiceImpl) Search(ctx context.Context, req domain.SemanticSearchRequest) (*domain.SemanticSearchResponse, error) {
	startTime := time.Now()
	projectID := generateProjectID(req.ProjectRoot)
	
	// Set defaults
	if req.TopK == 0 {
		req.TopK = 10
	}
	if req.MinScore == 0 {
		req.MinScore = 0.5
	}
	
	s.log.Info(fmt.Sprintf("Semantic search: query='%s', topK=%d", truncateStringForSearch(req.Query, 50), req.TopK))
	
	// Handle different search types
	switch req.SearchType {
	case domain.SearchTypeKeyword:
		return s.keywordSearch(ctx, req, startTime)
	case domain.SearchTypeHybrid:
		return s.hybridSearch(ctx, req, startTime)
	default:
		// Semantic search
		return s.semanticSearch(ctx, projectID, req, startTime)
	}
}

// semanticSearch performs pure semantic search
func (s *SemanticSearchServiceImpl) semanticSearch(ctx context.Context, projectID string, req domain.SemanticSearchRequest, startTime time.Time) (*domain.SemanticSearchResponse, error) {
	// Generate embedding for query with retry logic for API resilience
	resp, err := s.generateEmbeddingsWithRetry(ctx, []string{req.Query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	queryEmbedding := resp.Embeddings[0]
	
	// Search vector store
	results, err := s.vectorStore.Search(ctx, projectID, queryEmbedding, req.TopK*2, req.MinScore)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}
	
	// Apply filters
	results = s.applyFilters(results, req.Filters)
	
	// Limit to topK
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}
	
	return &domain.SemanticSearchResponse{
		Results:      results,
		TotalResults: len(results),
		QueryTime:    time.Since(startTime),
		SearchType:   domain.SearchTypeSemantic,
	}, nil
}

// keywordSearch performs keyword-based search
func (s *SemanticSearchServiceImpl) keywordSearch(ctx context.Context, req domain.SemanticSearchRequest, startTime time.Time) (*domain.SemanticSearchResponse, error) {
	// Use symbol index for keyword search
	if s.symbolIndex == nil {
		return &domain.SemanticSearchResponse{
			Results:    []domain.SemanticSearchResult{},
			QueryTime:  time.Since(startTime),
			SearchType: domain.SearchTypeKeyword,
		}, nil
	}
	
	symbols := s.symbolIndex.SearchByName(req.Query)
	
	var results []domain.SemanticSearchResult
	for _, sym := range symbols {
		if len(results) >= req.TopK {
			break
		}
		
		// Read file content to populate chunk Content
		content := ""
		fullPath := filepath.Join(req.ProjectRoot, sym.FilePath)
		if fileContent, err := os.ReadFile(fullPath); err == nil {
			lines := strings.Split(string(fileContent), "\n")
			startIdx := sym.StartLine - 1
			endIdx := sym.EndLine
			
			// Bounds checking
			if startIdx < 0 {
				startIdx = 0
			}
			if endIdx > len(lines) {
				endIdx = len(lines)
			}
			if startIdx < endIdx && startIdx < len(lines) {
				content = strings.Join(lines[startIdx:endIdx], "\n")
			}
		}
		
		// Use StartLine consistently for ID and chunk
		chunk := domain.CodeChunk{
			ID:         fmt.Sprintf("%s:%d", sym.FilePath, sym.StartLine),
			FilePath:   sym.FilePath,
			Content:    content,
			StartLine:  sym.StartLine,
			EndLine:    sym.EndLine,
			SymbolName: sym.Name,
			SymbolKind: string(sym.Kind),
			Language:   sym.Language,
		}
		
		results = append(results, domain.SemanticSearchResult{
			Chunk: chunk,
			Score: 1.0, // Exact match
		})
	}
	
	return &domain.SemanticSearchResponse{
		Results:      results,
		TotalResults: len(results),
		QueryTime:    time.Since(startTime),
		SearchType:   domain.SearchTypeKeyword,
	}, nil
}

// hybridSearch combines semantic and keyword search
func (s *SemanticSearchServiceImpl) hybridSearch(ctx context.Context, req domain.SemanticSearchRequest, startTime time.Time) (*domain.SemanticSearchResponse, error) {
	projectID := generateProjectID(req.ProjectRoot)
	
	// Get semantic results
	semanticReq := req
	semanticReq.TopK = req.TopK * 2
	semanticResults, err := s.semanticSearch(ctx, projectID, semanticReq, startTime)
	if err != nil {
		return nil, err
	}
	
	// Get keyword results
	keywordReq := req
	keywordReq.TopK = req.TopK
	keywordResults, err := s.keywordSearch(ctx, keywordReq, startTime)
	if err != nil {
		return nil, err
	}
	
	// Merge and re-rank results
	resultMap := make(map[string]domain.SemanticSearchResult)
	
	// Add semantic results with weight 0.7
	for _, r := range semanticResults.Results {
		key := fmt.Sprintf("%s:%d", r.Chunk.FilePath, r.Chunk.StartLine)
		r.Score *= 0.7
		resultMap[key] = r
	}
	
	// Add keyword results with weight 0.3, boost if already present
	for _, r := range keywordResults.Results {
		key := fmt.Sprintf("%s:%d", r.Chunk.FilePath, r.Chunk.StartLine)
		if existing, ok := resultMap[key]; ok {
			existing.Score += r.Score * 0.3
			resultMap[key] = existing
		} else {
			r.Score *= 0.3
			resultMap[key] = r
		}
	}
	
	// Convert to slice and sort
	var results []domain.SemanticSearchResult
	for _, r := range resultMap {
		results = append(results, r)
	}
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}
	
	return &domain.SemanticSearchResponse{
		Results:      results,
		TotalResults: len(results),
		QueryTime:    time.Since(startTime),
		SearchType:   domain.SearchTypeHybrid,
	}, nil
}

// FindSimilar finds similar code
func (s *SemanticSearchServiceImpl) FindSimilar(ctx context.Context, req domain.SimilarCodeRequest) (*domain.SemanticSearchResponse, error) {
	startTime := time.Now()
	
	// Read the source code
	content, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	lines := strings.Split(string(content), "\n")
	if req.StartLine < 1 || req.EndLine > len(lines) {
		return nil, fmt.Errorf("invalid line range")
	}
	
	sourceCode := strings.Join(lines[req.StartLine-1:req.EndLine], "\n")
	
	// Generate embedding for source code with retry logic for API resilience
	resp, err := s.generateEmbeddingsWithRetry(ctx, []string{sourceCode})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	
	// Get project root from file path
	projectRoot := filepath.Dir(req.FilePath)
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			projectRoot = filepath.Dir(req.FilePath)
			break
		}
		projectRoot = parent
	}
	
	projectID := generateProjectID(projectRoot)
	
	// Search for similar code
	topK := req.TopK
	if req.ExcludeSelf {
		topK++ // Get one extra to potentially exclude self
	}
	
	results, err := s.vectorStore.Search(ctx, projectID, resp.Embeddings[0], topK, req.MinScore)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	
	// Filter out self if requested
	if req.ExcludeSelf {
		var filtered []domain.SemanticSearchResult
		for _, r := range results {
			if r.Chunk.FilePath == req.FilePath && 
			   r.Chunk.StartLine == req.StartLine && 
			   r.Chunk.EndLine == req.EndLine {
				continue
			}
			filtered = append(filtered, r)
		}
		results = filtered
	}
	
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}
	
	return &domain.SemanticSearchResponse{
		Results:      results,
		TotalResults: len(results),
		QueryTime:    time.Since(startTime),
		SearchType:   domain.SearchTypeSemantic,
	}, nil
}

// GetClusters returns code clusters (simplified implementation)
func (s *SemanticSearchServiceImpl) GetClusters(ctx context.Context, projectRoot string, numClusters int) ([]domain.ClusterInfo, error) {
	// This is a simplified implementation
	// A full implementation would use k-means or similar clustering
	s.log.Info("GetClusters: clustering not fully implemented yet")
	return []domain.ClusterInfo{}, nil
}

// GetStats returns indexing statistics
func (s *SemanticSearchServiceImpl) GetStats(ctx context.Context, projectRoot string) (*domain.VectorStoreStats, error) {
	projectID := generateProjectID(projectRoot)
	return s.vectorStore.GetStats(ctx, projectID)
}

// IsIndexed checks if a project is indexed
func (s *SemanticSearchServiceImpl) IsIndexed(ctx context.Context, projectRoot string) bool {
	projectID := generateProjectID(projectRoot)
	stats, err := s.vectorStore.GetStats(ctx, projectID)
	if err != nil {
		return false
	}
	return stats.TotalChunks > 0
}

// InvalidateFile marks a file for re-indexing
func (s *SemanticSearchServiceImpl) InvalidateFile(ctx context.Context, projectRoot string, filePath string) error {
	projectID := generateProjectID(projectRoot)
	return s.vectorStore.Delete(ctx, projectID, filePath)
}

// GetIndexingState returns the current indexing state
func (s *SemanticSearchServiceImpl) GetIndexingState(projectRoot string) *IndexingState {
	projectID := generateProjectID(projectRoot)
	s.indexingMu.RLock()
	defer s.indexingMu.RUnlock()
	return s.indexingState[projectID]
}

// Helper methods

func (s *SemanticSearchServiceImpl) chunkFile(filePath string, content []byte, symbols []SymbolInfoForChunking) []domain.CodeChunk {
	language := detectLanguageFromPath(filePath)
	lines := strings.Split(string(content), "\n")
	
	// Simple chunking - prefer symbol-based if available
	if len(symbols) > 0 {
		return s.chunkBySymbols(filePath, lines, symbols, language)
	}
	
	return s.chunkBySize(filePath, lines, language)
}

func (s *SemanticSearchServiceImpl) chunkBySymbols(filePath string, lines []string, symbols []SymbolInfoForChunking, language string) []domain.CodeChunk {
	var chunks []domain.CodeChunk
	
	for _, sym := range symbols {
		if sym.StartLine < 1 || sym.EndLine > len(lines) || sym.EndLine < sym.StartLine {
			continue
		}
		
		symbolLines := lines[sym.StartLine-1 : sym.EndLine]
		content := strings.Join(symbolLines, "\n")
		tokenCount := len(content) / 4
		
		if tokenCount < 20 {
			continue
		}
		
		chunk := domain.CodeChunk{
			ID:         generateChunkID(filePath, sym.StartLine, sym.EndLine),
			FilePath:   filePath,
			Content:    content,
			StartLine:  sym.StartLine,
			EndLine:    sym.EndLine,
			ChunkType:  mapKindToChunkType(sym.Kind),
			SymbolName: sym.Name,
			SymbolKind: sym.Kind,
			Language:   language,
			TokenCount: tokenCount,
			Hash:       hashContent(content),
		}
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

func (s *SemanticSearchServiceImpl) chunkBySize(filePath string, lines []string, language string) []domain.CodeChunk {
	var chunks []domain.CodeChunk
	maxTokens := 512
	
	currentStart := 1
	var currentLines []string
	currentTokens := 0
	
	for i, line := range lines {
		lineTokens := len(line) / 4
		
		if currentTokens+lineTokens > maxTokens && len(currentLines) > 0 {
			content := strings.Join(currentLines, "\n")
			chunk := domain.CodeChunk{
				ID:         generateChunkID(filePath, currentStart, i),
				FilePath:   filePath,
				Content:    content,
				StartLine:  currentStart,
				EndLine:    i,
				ChunkType:  domain.ChunkTypeBlock,
				Language:   language,
				TokenCount: currentTokens,
				Hash:       hashContent(content),
			}
			chunks = append(chunks, chunk)
			
			currentLines = nil
			currentStart = i + 1
			currentTokens = 0
		}
		
		currentLines = append(currentLines, line)
		currentTokens += lineTokens
	}
	
	if len(currentLines) > 0 && currentTokens >= 20 {
		content := strings.Join(currentLines, "\n")
		chunk := domain.CodeChunk{
			ID:         generateChunkID(filePath, currentStart, len(lines)),
			FilePath:   filePath,
			Content:    content,
			StartLine:  currentStart,
			EndLine:    len(lines),
			ChunkType:  domain.ChunkTypeBlock,
			Language:   language,
			TokenCount: currentTokens,
			Hash:       hashContent(content),
		}
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

func (s *SemanticSearchServiceImpl) applyFilters(results []domain.SemanticSearchResult, filters *domain.SearchFilters) []domain.SemanticSearchResult {
	if filters == nil {
		return results
	}
	
	var filtered []domain.SemanticSearchResult
	
	for _, r := range results {
		// Filter by language
		if len(filters.Languages) > 0 {
			found := false
			for _, lang := range filters.Languages {
				if r.Chunk.Language == lang {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Filter by chunk type
		if len(filters.ChunkTypes) > 0 {
			found := false
			for _, ct := range filters.ChunkTypes {
				if r.Chunk.ChunkType == ct {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Filter by file path
		if len(filters.FilePaths) > 0 {
			found := false
			for _, fp := range filters.FilePaths {
				if strings.HasPrefix(r.Chunk.FilePath, fp) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Exclude directories
		if len(filters.ExcludeDirs) > 0 {
			excluded := false
			for _, dir := range filters.ExcludeDirs {
				if strings.Contains(r.Chunk.FilePath, dir) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		filtered = append(filtered, r)
	}
	
	return filtered
}

// SymbolInfoForChunking represents symbol info for chunking
type SymbolInfoForChunking struct {
	Name      string
	Kind      string
	StartLine int
	EndLine   int
}

// Helper functions

func generateProjectID(projectRoot string) string {
	hash := sha256.Sum256([]byte(projectRoot))
	return hex.EncodeToString(hash[:8])
}

func generateChunkID(filePath string, startLine, endLine int) string {
	data := fmt.Sprintf("%s:%d:%d", filePath, startLine, endLine)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:16])
}

func shouldSkipDir(name string) bool {
	skipDirs := []string{
		".git", ".svn", ".hg",
		"node_modules", "vendor", "venv", ".venv",
		"build", "dist", "target", "out",
		".idea", ".vscode", ".vs",
		"__pycache__", ".pytest_cache",
		"coverage", ".nyc_output",
	}
	
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return strings.HasPrefix(name, ".")
}

func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := map[string]bool{
		".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
		".vue": true, ".py": true, ".java": true, ".kt": true, ".kts": true,
		".dart": true, ".rs": true, ".cs": true, ".cpp": true, ".c": true,
		".h": true, ".hpp": true, ".rb": true, ".php": true, ".swift": true,
		".scala": true, ".clj": true, ".ex": true, ".exs": true,
	}
	return codeExts[ext]
}

func detectLanguageFromPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	langMap := map[string]string{
		".go": "go", ".ts": "typescript", ".tsx": "typescript",
		".js": "javascript", ".jsx": "javascript", ".mjs": "javascript",
		".vue": "vue", ".py": "python", ".java": "java",
		".kt": "kotlin", ".kts": "kotlin", ".dart": "dart",
		".rs": "rust", ".cs": "csharp", ".cpp": "cpp", ".c": "c",
		".h": "c", ".hpp": "cpp", ".rb": "ruby", ".php": "php",
		".swift": "swift", ".scala": "scala",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

func mapKindToChunkType(kind string) domain.ChunkType {
	switch strings.ToLower(kind) {
	case "function", "func":
		return domain.ChunkTypeFunction
	case "class":
		return domain.ChunkTypeClass
	case "method":
		return domain.ChunkTypeMethod
	default:
		return domain.ChunkTypeBlock
	}
}

func createEmbeddingText(chunk domain.CodeChunk) string {
	var sb strings.Builder
	
	// Add metadata as prefix for better semantic understanding
	if chunk.SymbolName != "" {
		sb.WriteString(fmt.Sprintf("// %s %s in %s\n", chunk.SymbolKind, chunk.SymbolName, chunk.FilePath))
	} else {
		sb.WriteString(fmt.Sprintf("// Code from %s (lines %d-%d)\n", chunk.FilePath, chunk.StartLine, chunk.EndLine))
	}
	
	sb.WriteString(chunk.Content)
	return sb.String()
}

func truncateStringForSearch(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// generateEmbeddingsWithRetry generates embeddings with exponential backoff retry logic
// for handling transient API errors (rate limits, network issues, etc.)
func (s *SemanticSearchServiceImpl) generateEmbeddingsWithRetry(ctx context.Context, texts []string) (*domain.EmbeddingResponse, error) {
	maxRetries := 3
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second
	
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		resp, err := s.embeddingProvider.GenerateEmbeddings(ctx, domain.EmbeddingRequest{
			Texts: texts,
		})
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// Check if error is retryable (rate limit, temporary network issues)
		errStr := err.Error()
		isRetryable := strings.Contains(errStr, "rate limit") ||
			strings.Contains(errStr, "429") ||
			strings.Contains(errStr, "503") ||
			strings.Contains(errStr, "502") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection") ||
			strings.Contains(errStr, "temporary")
		
		if !isRetryable {
			// Non-retryable error, fail immediately
			return nil, err
		}
		
		// Calculate delay with exponential backoff and jitter
		delay := baseDelay * time.Duration(1<<uint(attempt))
		if delay > maxDelay {
			delay = maxDelay
		}
		// Add jitter (±25%)
		jitter := time.Duration(float64(delay) * 0.25 * (0.5 - float64(time.Now().UnixNano()%100)/100))
		delay += jitter
		
		s.log.Warning(fmt.Sprintf("Embedding API error (attempt %d/%d): %v. Retrying in %v...", 
			attempt+1, maxRetries, err, delay))
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}
	
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
