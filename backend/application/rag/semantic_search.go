// Package rag provides backward compatibility aliases for semantic search.
// The actual implementation is in shotgun_code/application/semantic.
package rag

import (
	"shotgun_code/application/semantic"
	"shotgun_code/domain"
	"shotgun_code/domain/analysis"
)

// SemanticSearchService is an alias for semantic.ServiceImpl
type SemanticSearchService = semantic.ServiceImpl

// IndexingState is an alias for semantic.IndexingState
type IndexingState = semantic.IndexingState

// SymbolInfoForChunking is an alias for semantic.SymbolInfoForChunking
type SymbolInfoForChunking = semantic.SymbolInfoForChunking

// NewSemanticSearchService creates a new semantic search service.
func NewSemanticSearchService(
	embeddingProvider domain.EmbeddingProvider,
	vectorStore domain.VectorStore,
	symbolIndex analysis.SymbolIndex,
	log domain.Logger,
) *SemanticSearchService {
	return semantic.NewService(embeddingProvider, vectorStore, symbolIndex, log)
}
