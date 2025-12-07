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
	"shotgun_code/infrastructure/embeddings"
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
	chunker           *CodeChunker

	// Indexing state
	indexingMu    sync.RWMutex
	indexingState map[string]*IndexingState
}

// CodeChunker is imported from embeddings package
type CodeChunker = embeddings.CodeChunker

// IndexingState tracks the state of project indexing
type IndexingState struct {
	ProjectID    string
	InProgress   bool
	Progress     float64
	TotalFiles   int
	IndexedFiles int
	StartedAt    time.Time
	Error        error
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
		chunker:           embeddings.NewCodeChunker(embeddings.DefaultChunkerConfig()),
		indexingState:     make(map[string]*IndexingState),
	}
}

// startIndexingState initializes indexing state, returns error if already indexing
func (s *SemanticSearchServiceImpl) startIndexingState(projectID string) (*IndexingState, error) {
	s.indexingMu.Lock()
	defer s.indexingMu.Unlock()

	if state, exists := s.indexingState[projectID]; exists && state.InProgress {
		return nil, fmt.Errorf("indexing already in progress for project")
	}

	state := &IndexingState{ProjectID: projectID, InProgress: true, StartedAt: time.Now()}
	s.indexingState[projectID] = state
	return state, nil
}

// collectCodeFiles collects all code files from project
func (s *SemanticSearchServiceImpl) collectCodeFiles(projectRoot string) ([]string, error) {
	var files []string
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if isCodeFile(path) {
			relPath, _ := filepath.Rel(projectRoot, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// IndexProject indexes all files in a project
func (s *SemanticSearchServiceImpl) IndexProject(ctx context.Context, projectRoot string) error {
	projectID := generateProjectID(projectRoot)

	state, err := s.startIndexingState(projectID)
	if err != nil {
		return err
	}

	defer func() {
		s.indexingMu.Lock()
		state.InProgress = false
		s.indexingMu.Unlock()
	}()

	s.log.Info(fmt.Sprintf("Starting semantic indexing for project: %s", projectRoot))

	if s.symbolIndex != nil {
		s.log.Info("Indexing symbols for project...")
		if err := s.symbolIndex.IndexProject(ctx, projectRoot); err != nil {
			s.log.Warning(fmt.Sprintf("Symbol indexing failed (non-critical): %v", err))
		}
	}

	files, err := s.collectCodeFiles(projectRoot)
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

	results := make([]domain.SemanticSearchResult, 0, min(len(symbols), req.TopK))
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
	results := make([]domain.SemanticSearchResult, 0, len(resultMap))
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

// extractSourceCode reads file and extracts lines in range
func extractSourceCode(filePath string, startLine, endLine int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	if startLine < 1 || endLine > len(lines) {
		return "", fmt.Errorf("invalid line range")
	}
	return strings.Join(lines[startLine-1:endLine], "\n"), nil
}

// findProjectRoot walks up to find .git directory
func findProjectRoot(filePath string) string {
	projectRoot := filepath.Dir(filePath)
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			return projectRoot
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			return filepath.Dir(filePath)
		}
		projectRoot = parent
	}
}

// filterSelfFromResults removes the source chunk from results
func filterSelfFromResults(results []domain.SemanticSearchResult, filePath string, startLine, endLine int) []domain.SemanticSearchResult {
	filtered := make([]domain.SemanticSearchResult, 0, len(results))
	for _, r := range results {
		if r.Chunk.FilePath == filePath && r.Chunk.StartLine == startLine && r.Chunk.EndLine == endLine {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// FindSimilar finds similar code
func (s *SemanticSearchServiceImpl) FindSimilar(ctx context.Context, req domain.SimilarCodeRequest) (*domain.SemanticSearchResponse, error) {
	startTime := time.Now()

	sourceCode, err := extractSourceCode(req.FilePath, req.StartLine, req.EndLine)
	if err != nil {
		return nil, err
	}

	resp, err := s.generateEmbeddingsWithRetry(ctx, []string{sourceCode})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	projectID := generateProjectID(findProjectRoot(req.FilePath))
	topK := req.TopK
	if req.ExcludeSelf {
		topK++
	}

	results, err := s.vectorStore.Search(ctx, projectID, resp.Embeddings[0], topK, req.MinScore)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if req.ExcludeSelf {
		results = filterSelfFromResults(results, req.FilePath, req.StartLine, req.EndLine)
	}
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return &domain.SemanticSearchResponse{
		Results: results, TotalResults: len(results),
		QueryTime: time.Since(startTime), SearchType: domain.SearchTypeSemantic,
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
	// Convert symbols to embeddings.SymbolInfo format
	symbolInfos := make([]embeddings.SymbolInfo, 0, len(symbols))
	for _, sym := range symbols {
		symbolInfos = append(symbolInfos, embeddings.SymbolInfo{
			Name:      sym.Name,
			Kind:      sym.Kind,
			StartLine: sym.StartLine,
			EndLine:   sym.EndLine,
		})
	}

	// Use CodeChunker for intelligent chunking with overlap and symbol boundaries
	return s.chunker.ChunkFile(filePath, content, symbolInfos)
}

// Note: chunkBySymbols and chunkBySize methods removed - now using CodeChunker from embeddings package
// which provides better chunking with overlap, symbol boundaries, and configurable token limits

// matchesLanguageFilter checks if chunk matches language filter
func matchesLanguageFilter(chunk *domain.CodeChunk, languages []string) bool {
	if len(languages) == 0 {
		return true
	}
	for _, lang := range languages {
		if chunk.Language == lang {
			return true
		}
	}
	return false
}

// matchesChunkTypeFilter checks if chunk matches chunk type filter
func matchesChunkTypeFilter(chunk *domain.CodeChunk, chunkTypes []domain.ChunkType) bool {
	if len(chunkTypes) == 0 {
		return true
	}
	for _, ct := range chunkTypes {
		if chunk.ChunkType == ct {
			return true
		}
	}
	return false
}

// matchesFilePathFilter checks if chunk matches file path filter
func matchesFilePathFilter(chunk *domain.CodeChunk, filePaths []string) bool {
	if len(filePaths) == 0 {
		return true
	}
	for _, fp := range filePaths {
		if strings.HasPrefix(chunk.FilePath, fp) {
			return true
		}
	}
	return false
}

// isExcludedDir checks if chunk is in an excluded directory
func isExcludedDir(chunk *domain.CodeChunk, excludeDirs []string) bool {
	for _, dir := range excludeDirs {
		if strings.Contains(chunk.FilePath, dir) {
			return true
		}
	}
	return false
}

func (s *SemanticSearchServiceImpl) applyFilters(results []domain.SemanticSearchResult, filters *domain.SearchFilters) []domain.SemanticSearchResult {
	if filters == nil {
		return results
	}

	filtered := make([]domain.SemanticSearchResult, 0, len(results))
	for _, r := range results {
		chunk := &r.Chunk
		if !matchesLanguageFilter(chunk, filters.Languages) ||
			!matchesChunkTypeFilter(chunk, filters.ChunkTypes) ||
			!matchesFilePathFilter(chunk, filters.FilePaths) ||
			isExcludedDir(chunk, filters.ExcludeDirs) {
			continue
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
