package symbolgraph

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
)

// GoSymbolGraphBuilder реализует SymbolGraphBuilder для Go
type GoSymbolGraphBuilder struct {
	log domain.Logger
}

// NewGoSymbolGraphBuilder создает новый builder для Go
func NewGoSymbolGraphBuilder(log domain.Logger) *GoSymbolGraphBuilder {
	return &GoSymbolGraphBuilder{
		log: log,
	}
}

// BuildGraph строит граф символов для Go проекта
func (b *GoSymbolGraphBuilder) BuildGraph(ctx context.Context, projectRoot string) (*domain.SymbolGraph, error) {
	b.log.Info(fmt.Sprintf("Building symbol graph for Go project: %s", projectRoot))

	graph := &domain.SymbolGraph{
		Nodes: []*domain.SymbolNode{},
		Edges: []*domain.SymbolEdge{},
	}

	// Проходим по всем .go файлам в проекте
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Пропускаем vendor, node_modules и другие служебные директории
			if info.Name() == "vendor" || info.Name() == "node_modules" ||
				info.Name() == ".git" || info.Name() == "dist" ||
				strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Парсим Go файл
		nodes, edges, err := b.parseGoFile(path, projectRoot)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to parse %s: %v", path, err))
			return nil // Продолжаем с другими файлами
		}

		graph.Nodes = append(graph.Nodes, nodes...)
		graph.Edges = append(graph.Edges, edges...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to build symbol graph: %w", err)
	}

	// Обеспечиваем детерминизм: сортируем узлы и ребра
	b.sortGraphForDeterminism(graph)

	b.log.Info(fmt.Sprintf("Built symbol graph with %d nodes and %d edges", len(graph.Nodes), len(graph.Edges)))
	return graph, nil
}

// parseGoFile парсит отдельный Go файл и извлекает символы
func (b *GoSymbolGraphBuilder) parseGoFile(filePath, projectRoot string) ([]*domain.SymbolNode, []*domain.SymbolEdge, error) {
	fset := token.NewFileSet()

	// Парсим файл
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	var nodes []*domain.SymbolNode
	var edges []*domain.SymbolEdge

	// Определяем пакет
	packageName := file.Name.Name
	relPath, _ := filepath.Rel(projectRoot, filePath)

	// Добавляем узел пакета
	packageNode := &domain.SymbolNode{
		ID:         fmt.Sprintf("%s:%s", packageName, relPath),
		Name:       packageName,
		Type:       domain.SymbolTypePackage,
		Path:       relPath,
		Package:    packageName,
		Visibility: domain.VisibilityPublic,
	}
	nodes = append(nodes, packageNode)

	// Обрабатываем импорты
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		importNode := &domain.SymbolNode{
			ID:         fmt.Sprintf("import:%s:%s", relPath, importPath),
			Name:       importPath,
			Type:       domain.SymbolTypeImport,
			Path:       relPath,
			Package:    packageName,
			Visibility: domain.VisibilityPublic,
		}
		nodes = append(nodes, importNode)

		// Добавляем ребро импорта
		importEdge := &domain.SymbolEdge{
			From:   packageNode.ID,
			To:     importNode.ID,
			Type:   domain.EdgeTypeImports,
			Weight: 1.0,
		}
		edges = append(edges, importEdge)
	}

	// Обрабатываем объявления
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Функция
			funcNode := &domain.SymbolNode{
				ID:         fmt.Sprintf("func:%s:%s", relPath, x.Name.Name),
				Name:       x.Name.Name,
				Type:       domain.SymbolTypeFunction,
				Path:       relPath,
				Package:    packageName,
				Visibility: b.getVisibility(x.Name.Name),
			}
			if x.Recv != nil {
				funcNode.Type = domain.SymbolTypeMethod
				// Добавляем связь с типом-получателем
				if len(x.Recv.List) > 0 && len(x.Recv.List[0].Names) > 0 {
					receiverType := b.getReceiverType(x.Recv.List[0].Type)
					if receiverType != "" {
						receiverEdge := &domain.SymbolEdge{
							From:   funcNode.ID,
							To:     fmt.Sprintf("type:%s:%s", relPath, receiverType),
							Type:   domain.EdgeTypeReferences,
							Weight: 1.0,
						}
						edges = append(edges, receiverEdge)
					}
				}
			}
			nodes = append(nodes, funcNode)

		case *ast.GenDecl:
			switch x.Tok {
			case token.TYPE:
				// Тип
				for _, spec := range x.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						var symbolType domain.SymbolType = domain.SymbolTypeType

						// Определяем конкретный тип
						switch typeSpec.Type.(type) {
						case *ast.StructType:
							symbolType = domain.SymbolTypeStruct
							// Обрабатываем поля структуры
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								for _, field := range structType.Fields.List {
									for _, name := range field.Names {
										fieldNode := &domain.SymbolNode{
											ID:         fmt.Sprintf("field:%s:%s.%s", relPath, typeSpec.Name.Name, name.Name),
											Name:       name.Name,
											Type:       domain.SymbolTypeField,
											Path:       relPath,
											Package:    packageName,
											Visibility: b.getVisibility(name.Name),
										}
										nodes = append(nodes, fieldNode)

										// Добавляем связь с родительским типом
										fieldEdge := &domain.SymbolEdge{
											From:   fieldNode.ID,
											To:     fmt.Sprintf("type:%s:%s", relPath, typeSpec.Name.Name),
											Type:   domain.EdgeTypeReferences,
											Weight: 1.0,
										}
										edges = append(edges, fieldEdge)
									}
								}
							}
						case *ast.InterfaceType:
							symbolType = domain.SymbolTypeInterface
						}

						typeNode := &domain.SymbolNode{
							ID:         fmt.Sprintf("type:%s:%s", relPath, typeSpec.Name.Name),
							Name:       typeSpec.Name.Name,
							Type:       symbolType,
							Path:       relPath,
							Package:    packageName,
							Visibility: b.getVisibility(typeSpec.Name.Name),
						}
						nodes = append(nodes, typeNode)
					}
				}
			case token.VAR, token.CONST:
				// Переменные и константы
				for _, spec := range x.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							var symbolType domain.SymbolType
							if x.Tok == token.CONST {
								symbolType = domain.SymbolTypeConstant
							} else {
								symbolType = domain.SymbolTypeVariable
							}

							varNode := &domain.SymbolNode{
								ID:         fmt.Sprintf("%s:%s:%s", strings.ToLower(x.Tok.String()), relPath, name.Name),
								Name:       name.Name,
								Type:       symbolType,
								Path:       relPath,
								Package:    packageName,
								Visibility: b.getVisibility(name.Name),
							}
							nodes = append(nodes, varNode)
						}
					}
				}
			}
		}
		return true
	})

	return nodes, edges, nil
}

// getVisibility определяет видимость символа по его имени
func (b *GoSymbolGraphBuilder) getVisibility(name string) domain.Visibility {
	if name == "" {
		return domain.VisibilityPrivate
	}

	// В Go: заглавная буква = публичный, строчная = приватный
	if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
		return domain.VisibilityPublic
	}
	return domain.VisibilityPrivate
}

// getReceiverType извлекает тип получателя из AST
func (b *GoSymbolGraphBuilder) getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// UpdateGraph обновляет граф для измененных файлов
func (b *GoSymbolGraphBuilder) UpdateGraph(ctx context.Context, projectRoot string, changedFiles []string) (*domain.SymbolGraph, error) {
	// Для простоты перестраиваем весь граф
	// В будущем можно оптимизировать для инкрементальных обновлений
	return b.BuildGraph(ctx, projectRoot)
}

// GetSuggestions возвращает предложения символов на основе запроса
func (b *GoSymbolGraphBuilder) GetSuggestions(ctx context.Context, query string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	var suggestions []*domain.SymbolNode

	query = strings.ToLower(query)
	for _, node := range graph.Nodes {
		if strings.Contains(strings.ToLower(node.Name), query) ||
			strings.Contains(strings.ToLower(node.Package), query) {
			suggestions = append(suggestions, node)
		}
	}

	return suggestions, nil
}

// GetDependencies возвращает зависимости для указанного символа
func (b *GoSymbolGraphBuilder) GetDependencies(ctx context.Context, symbolID string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	var dependencies []*domain.SymbolNode
	dependencyIDs := make(map[string]bool)

	// Находим все ребра, где To = symbolID
	for _, edge := range graph.Edges {
		if edge.To == symbolID {
			dependencyIDs[edge.From] = true
		}
	}

	// Находим узлы по ID
	for _, node := range graph.Nodes {
		if dependencyIDs[node.ID] {
			dependencies = append(dependencies, node)
		}
	}

	return dependencies, nil
}

// GetDependents возвращает символы, которые зависят от указанного
func (b *GoSymbolGraphBuilder) GetDependents(ctx context.Context, symbolID string, graph *domain.SymbolGraph) ([]*domain.SymbolNode, error) {
	var dependents []*domain.SymbolNode
	dependentIDs := make(map[string]bool)

	// Находим все ребра, где From = symbolID
	for _, edge := range graph.Edges {
		if edge.From == symbolID {
			dependentIDs[edge.To] = true
		}
	}

	// Находим узлы по ID
	for _, node := range graph.Nodes {
		if dependentIDs[node.ID] {
			dependents = append(dependents, node)
		}
	}

	return dependents, nil
}

// sortGraphForDeterminism сортирует узлы и ребра для обеспечения детерминизма
func (b *GoSymbolGraphBuilder) sortGraphForDeterminism(graph *domain.SymbolGraph) {
	// Сортируем узлы по ID для детерминизма
	sort.Slice(graph.Nodes, func(i, j int) bool {
		return graph.Nodes[i].ID < graph.Nodes[j].ID
	})

	// Сортируем ребра по From, затем по To для детерминизма
	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].From != graph.Edges[j].From {
			return graph.Edges[i].From < graph.Edges[j].From
		}
		return graph.Edges[i].To < graph.Edges[j].To
	})
}
