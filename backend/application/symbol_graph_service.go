package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/symbolgraph"
)

// SymbolGraphService предоставляет высокоуровневый API для работы с графом символов
type SymbolGraphService struct {
	log                 domain.Logger
	symbolGraphBuilders map[string]domain.SymbolGraphBuilder
	importGraphBuilders map[string]domain.ImportGraphBuilder
}

// NewSymbolGraphService создает новый сервис графа символов
func NewSymbolGraphService(log domain.Logger) *SymbolGraphService {
	service := &SymbolGraphService{
		log:                 log,
		symbolGraphBuilders: make(map[string]domain.SymbolGraphBuilder),
		importGraphBuilders: make(map[string]domain.ImportGraphBuilder),
	}

	// Регистрируем builders для поддерживаемых языков
	service.RegisterSymbolGraphBuilder("go", symbolgraph.NewGoSymbolGraphBuilder(log))

	return service
}

// RegisterSymbolGraphBuilder регистрирует builder для языка
func (s *SymbolGraphService) RegisterSymbolGraphBuilder(language string, builder domain.SymbolGraphBuilder) {
	s.symbolGraphBuilders[language] = builder
	s.log.Info(fmt.Sprintf("Registered symbol graph builder for language: %s", language))
}

// RegisterImportGraphBuilder регистрирует builder импортов для языка
func (s *SymbolGraphService) RegisterImportGraphBuilder(language string, builder domain.ImportGraphBuilder) {
	s.importGraphBuilders[language] = builder
	s.log.Info(fmt.Sprintf("Registered import graph builder for language: %s", language))
}

// BuildSymbolGraph строит граф символов для проекта
func (s *SymbolGraphService) BuildSymbolGraph(ctx context.Context, projectRoot, language string) (*domain.SymbolGraph, error) {
	builder, exists := s.symbolGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no symbol graph builder registered for language: %s", language)
	}

	s.log.Info(fmt.Sprintf("Building symbol graph for project: %s (language: %s)", projectRoot, language))
	return builder.BuildGraph(ctx, projectRoot)
}

// BuildImportGraph строит граф импортов для проекта
func (s *SymbolGraphService) BuildImportGraph(ctx context.Context, projectRoot, language string) (*domain.ImportGraph, error) {
	builder, exists := s.importGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no import graph builder registered for language: %s", language)
	}

	s.log.Info(fmt.Sprintf("Building import graph for project: %s (language: %s)", projectRoot, language))
	return builder.BuildImportGraph(ctx, projectRoot)
}

// GetSuggestions возвращает предложения символов
func (s *SymbolGraphService) GetSuggestions(ctx context.Context, query, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	builder, exists := s.symbolGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no symbol graph builder registered for language: %s", language)
	}

	return builder.GetSuggestions(ctx, query, graph)
}

// GetDependencies возвращает зависимости символа
func (s *SymbolGraphService) GetDependencies(ctx context.Context, symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	builder, exists := s.symbolGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no symbol graph builder registered for language: %s", language)
	}

	return builder.GetDependencies(ctx, symbolID, graph)
}

// GetDependents возвращает символы, зависящие от указанного
func (s *SymbolGraphService) GetDependents(ctx context.Context, symbolID, language string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	builder, exists := s.symbolGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no symbol graph builder registered for language: %s", language)
	}

	return builder.GetDependents(ctx, symbolID, graph)
}

// GetCircularImports возвращает циклические импорты
func (s *SymbolGraphService) GetCircularImports(ctx context.Context, language string, graph *domain.ImportGraph) ([][]string, error) {
	builder, exists := s.importGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no import graph builder registered for language: %s", language)
	}

	return builder.GetCircularImports(ctx, graph)
}

// GetImportPath возвращает путь импорта между пакетами
func (s *SymbolGraphService) GetImportPath(ctx context.Context, from, to, language string, graph *domain.ImportGraph) ([]string, error) {
	builder, exists := s.importGraphBuilders[language]
	if !exists {
		return nil, fmt.Errorf("no import graph builder registered for language: %s", language)
	}

	return builder.GetImportPath(ctx, from, to, graph)
}

// GetSupportedLanguages возвращает список поддерживаемых языков
func (s *SymbolGraphService) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(s.symbolGraphBuilders))
	for lang := range s.symbolGraphBuilders {
		languages = append(languages, lang)
	}
	return languages
}
