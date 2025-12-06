package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"shotgun_code/application"
	"shotgun_code/domain"
	"sync"
	"sync/atomic"
	"time"
)

// AIHandler handles all AI-related operations with rate limiting and caching
type AIHandler struct {
	log             domain.Logger
	aiService       *application.AIService
	contextAnalysis domain.ContextAnalyzer

	// Rate limiting
	requestCount int64
	lastReset    time.Time
	rateMu       sync.Mutex

	// Graceful shutdown
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

const (
	maxRequestsPerMin = 60
)

// NewAIHandler creates a new AI handler
func NewAIHandler(
	log domain.Logger,
	aiService *application.AIService,
	contextAnalysis domain.ContextAnalyzer,
) *AIHandler {
	h := &AIHandler{
		log:             log,
		aiService:       aiService,
		contextAnalysis: contextAnalysis,
		lastReset:       time.Now(),
		stopCh:          make(chan struct{}),
	}

	return h
}

// Shutdown gracefully stops the AI handler
func (h *AIHandler) Shutdown(ctx context.Context) error {
	h.stopOnce.Do(func() {
		close(h.stopCh)
	})

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		h.log.Info("AIHandler shutdown complete")
	case <-ctx.Done():
		h.log.Warning("AIHandler shutdown timed out")
		return ctx.Err()
	}

	return nil
}

// GenerateCode generates code using AI
func (h *AIHandler) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if err := h.checkRateLimit(); err != nil {
		return "", err
	}

	atomic.AddInt64(&h.requestCount, 1)
	return h.aiService.GenerateCode(ctx, systemPrompt, userPrompt)
}

// GenerateIntelligentCode generates code with intelligent options
func (h *AIHandler) GenerateIntelligentCode(ctx context.Context, task, contextStr, optionsJson string) (string, error) {
	if err := h.checkRateLimit(); err != nil {
		return "", err
	}

	var options application.IntelligentGenerationOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		return "", fmt.Errorf("failed to parse options JSON: %w", err)
	}

	atomic.AddInt64(&h.requestCount, 1)

	result, err := h.aiService.GenerateIntelligentCode(ctx, task, contextStr, options)
	if err != nil {
		return "", err
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJson), nil
}

// GenerateCodeWithOptions generates code with additional options
func (h *AIHandler) GenerateCodeWithOptions(ctx context.Context, systemPrompt, userPrompt, optionsJson string) (string, error) {
	if err := h.checkRateLimit(); err != nil {
		return "", err
	}

	var options application.CodeGenerationOptions
	if err := json.Unmarshal([]byte(optionsJson), &options); err != nil {
		return "", fmt.Errorf("failed to parse options JSON: %w", err)
	}

	atomic.AddInt64(&h.requestCount, 1)
	return h.aiService.GenerateCodeWithOptions(ctx, systemPrompt, userPrompt, options)
}

// GetProviderInfo returns current AI provider info
func (h *AIHandler) GetProviderInfo(ctx context.Context) (string, error) {
	info, err := h.aiService.GetProviderInfo(ctx)
	if err != nil {
		return "", err
	}

	infoJson, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("failed to marshal provider info: %w", err)
	}

	return string(infoJson), nil
}

// ListAvailableModels returns available AI models
func (h *AIHandler) ListAvailableModels(ctx context.Context) ([]string, error) {
	return h.aiService.ListAvailableModels(ctx)
}

// SuggestContextFiles suggests relevant files for a task
func (h *AIHandler) SuggestContextFiles(ctx context.Context, task string, allFiles []*domain.FileNode) ([]string, error) {
	if h.contextAnalysis == nil {
		return nil, fmt.Errorf("context analysis service not available")
	}

	return h.contextAnalysis.SuggestFiles(ctx, task, allFiles)
}

// AnalyzeTaskAndCollectContext analyzes task and collects relevant context
func (h *AIHandler) AnalyzeTaskAndCollectContext(ctx context.Context, task string, allFilesJson string, rootDir string) (string, error) {
	var allFiles []*domain.FileNode
	if err := json.Unmarshal([]byte(allFilesJson), &allFiles); err != nil {
		return "", fmt.Errorf("invalid allFiles JSON: %w", err)
	}

	if h.contextAnalysis == nil {
		return "", fmt.Errorf("context analysis service not available")
	}

	// Check if contextAnalysis implements the extended interface
	type extendedAnalyzer interface {
		domain.ContextAnalyzer
		AnalyzeTaskAndCollectContext(ctx context.Context, task string, allFiles []*domain.FileNode, rootDir string) (*domain.ContextAnalysisResult, error)
	}

	analyzer, ok := h.contextAnalysis.(extendedAnalyzer)
	if !ok {
		return "", fmt.Errorf("context analysis service does not support AnalyzeTaskAndCollectContext")
	}

	result, err := analyzer.AnalyzeTaskAndCollectContext(ctx, task, allFiles, rootDir)
	if err != nil {
		return "", err
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal analysis result: %w", err)
	}

	return string(resultJson), nil
}

// checkRateLimit checks if rate limit is exceeded
func (h *AIHandler) checkRateLimit() error {
	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	now := time.Now()
	if now.Sub(h.lastReset) > time.Minute {
		h.requestCount = 0
		h.lastReset = now
	}

	if h.requestCount >= maxRequestsPerMin {
		return fmt.Errorf("rate limit exceeded: %d requests per minute", maxRequestsPerMin)
	}

	return nil
}

// GetMetrics returns handler metrics
func (h *AIHandler) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"request_count": atomic.LoadInt64(&h.requestCount),
		"last_reset":    h.lastReset,
	}
}

// GenerateCodeStream generates code with streaming response
func (h *AIHandler) GenerateCodeStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(chunk domain.StreamChunk)) error {
	if err := h.checkRateLimit(); err != nil {
		onChunk(domain.StreamChunk{Done: true, Error: err.Error()})
		return err
	}

	atomic.AddInt64(&h.requestCount, 1)
	return h.aiService.GenerateCodeStream(ctx, systemPrompt, userPrompt, onChunk)
}
