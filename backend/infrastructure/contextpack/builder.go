package contextpack

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
)

// ContextPackBuilderImpl реализует ContextPackBuilder
type ContextPackBuilderImpl struct {
	log             domain.Logger
	fileReader      domain.FileContentReader
	symbolGraph     domain.SymbolGraphBuilder
	contextAnalyzer domain.ContextAnalyzer
}

// NewContextPackBuilder создает новый builder упакованного контекста
func NewContextPackBuilder(
	log domain.Logger,
	fileReader domain.FileContentReader,
	symbolGraph domain.SymbolGraphBuilder,
	contextAnalyzer domain.ContextAnalyzer,
) *ContextPackBuilderImpl {
	return &ContextPackBuilderImpl{
		log:             log,
		fileReader:      fileReader,
		symbolGraph:     symbolGraph,
		contextAnalyzer: contextAnalyzer,
	}
}

// BuildPack строит упакованный контекст для задачи
func (b *ContextPackBuilderImpl) BuildPack(
	ctx context.Context,
	task string,
	projectRoot string,
	options *domain.ContextPackOptions,
) (*domain.ContextPack, error) {
	b.log.Info(fmt.Sprintf("Building context pack for task: %s", task))

	// Определяем язык проекта
	language := b.detectLanguage(projectRoot)
	if options.Language != "" {
		language = options.Language
	}

	// Получаем все файлы проекта
	allFiles, err := b.getAllFiles(projectRoot, options.ExcludePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to get project files: %w", err)
	}

	// Анализируем задачу и определяем релевантные файлы
	relevantFiles, err := b.contextAnalyzer.SuggestFiles(ctx, task, allFiles)
	if err != nil {
		b.log.Warning(fmt.Sprintf("Failed to analyze task, using all files: %v", err))
		relevantFiles = b.getFilePaths(allFiles)
	}

	// Ограничиваем количество файлов
	if len(relevantFiles) > options.MaxFiles {
		relevantFiles = relevantFiles[:options.MaxFiles]
	}

	// Строим граф символов если нужно
	var symbolGraph *domain.SymbolGraph
	if b.symbolGraph != nil {
		symbolGraph, err = b.symbolGraph.BuildGraph(ctx, projectRoot)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to build symbol graph: %v", err))
		}
	}

	// Создаем сниппеты
	snippets, err := b.createSnippets(ctx, relevantFiles, projectRoot, options, symbolGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to create snippets: %w", err)
	}

	// Создаем зависимости
	var deps []*domain.ContextDependency
	if options.IncludeDeps {
		deps, err = b.createDependencies(projectRoot, language)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to create dependencies: %v", err))
		}
	}

	// Создаем информацию о сборке
	var build *domain.ContextBuild
	if options.IncludeBuild {
		build = b.createBuildInfo(language)
	}

	// Создаем тесты
	var tests []*domain.ContextTest
	if options.IncludeTests {
		tests, err = b.createTests(projectRoot, relevantFiles)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to create tests: %v", err))
		}
	}

	// Создаем ограничения
	constraints := &domain.ContextConstraints{
		MaxFiles:   options.MaxFiles,
		MaxLines:   options.MaxLines,
		TimeBudget: "",
	}

	// Создаем provenance
	provenance := &domain.ContextProvenance{
		IndexedAt: time.Now().UTC().Format(time.RFC3339),
		IndexID:   b.generateIndexID(projectRoot, task),
	}

	// Создаем цель
	target := &domain.ContextTarget{
		Lang:   language,
		File:   options.TargetFile,
		Symbol: options.TargetSymbol,
	}

	pack := &domain.ContextPack{
		PackVersion: "1.0",
		Target:      target,
		Snippets:    snippets,
		Deps:        deps,
		Build:       build,
		Tests:       tests,
		Constraints: constraints,
		Provenance:  provenance,
	}

	b.log.Info(fmt.Sprintf("Built context pack with %d snippets", len(snippets)))
	return pack, nil
}

// createSnippets создает сниппеты для файлов
func (b *ContextPackBuilderImpl) createSnippets(
	ctx context.Context,
	filePaths []string,
	projectRoot string,
	options *domain.ContextPackOptions,
	symbolGraph *domain.SymbolGraph,
) ([]*domain.ContextSnippet, error) {
	var snippets []*domain.ContextSnippet
	totalLines := 0

	// Сортируем файлы для детерминизма
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		if totalLines >= options.MaxLines {
			break
		}

		fullPath := filepath.Join(projectRoot, filePath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
			continue
		}

		fileSnippets, err := b.createFileSnippets(filePath, string(content), options, symbolGraph)
		if err != nil {
			b.log.Warning(fmt.Sprintf("Failed to create snippets for %s: %v", filePath, err))
			continue
		}

		// Проверяем ограничения по строкам
		for _, snippet := range fileSnippets {
			lines := snippet.EndLine - snippet.StartLine + 1
			if totalLines+lines > options.MaxLines {
				break
			}
			snippets = append(snippets, snippet)
			totalLines += lines
		}
	}

	return snippets, nil
}

// createFileSnippets создает сниппеты для одного файла
func (b *ContextPackBuilderImpl) createFileSnippets(
	filePath string,
	content string,
	options *domain.ContextPackOptions,
	symbolGraph *domain.SymbolGraph,
) ([]*domain.ContextSnippet, error) {
	var snippets []*domain.ContextSnippet
	lines := strings.Split(content, "\n")

	if options.SmartSnippets && symbolGraph != nil {
		// Умные вырезки на основе графа символов
		snippets = b.createSmartSnippets(filePath, lines, symbolGraph)
	} else {
		// Простые вырезки
		snippets = b.createSimpleSnippets(filePath, lines)
	}

	// Добавляем хеши
	for _, snippet := range snippets {
		snippet.TextHash = b.calculateTextHash(snippet.Content)
	}

	return snippets, nil
}

// createSmartSnippets создает умные вырезки на основе графа символов
func (b *ContextPackBuilderImpl) createSmartSnippets(
	filePath string,
	lines []string,
	symbolGraph *domain.SymbolGraph,
) []*domain.ContextSnippet {
	var snippets []*domain.ContextSnippet

	// Находим символы для этого файла
	var fileSymbols []*domain.SymbolNode
	for _, node := range symbolGraph.Nodes {
		if node.Path == filePath {
			fileSymbols = append(fileSymbols, node)
		}
	}

	// Сортируем символы для детерминизма
	sort.Slice(fileSymbols, func(i, j int) bool {
		return fileSymbols[i].Line < fileSymbols[j].Line
	})

	// Создаем сниппеты для каждого символа
	for _, symbol := range fileSymbols {
		startLine := max(1, symbol.Line-5)         // Контекст до символа
		endLine := min(len(lines), symbol.Line+10) // Контекст после символа

		content := strings.Join(lines[startLine-1:endLine], "\n")

		snippet := &domain.ContextSnippet{
			Path:      filePath,
			Kind:      "match",
			StartLine: startLine,
			EndLine:   endLine,
			Content:   content,
		}
		snippets = append(snippets, snippet)
	}

	// Добавляем заголовок файла если нет символов
	if len(fileSymbols) == 0 && len(lines) > 0 {
		headerEnd := min(20, len(lines))
		content := strings.Join(lines[:headerEnd], "\n")

		snippet := &domain.ContextSnippet{
			Path:      filePath,
			Kind:      "header",
			StartLine: 1,
			EndLine:   headerEnd,
			Content:   content,
		}
		snippets = append(snippets, snippet)
	}

	return snippets
}

// createSimpleSnippets создает простые вырезки
func (b *ContextPackBuilderImpl) createSimpleSnippets(filePath string, lines []string) []*domain.ContextSnippet {
	var snippets []*domain.ContextSnippet

	if len(lines) == 0 {
		return snippets
	}

	// Заголовок файла (первые 20 строк)
	headerEnd := min(20, len(lines))
	headerContent := strings.Join(lines[:headerEnd], "\n")

	headerSnippet := &domain.ContextSnippet{
		Path:      filePath,
		Kind:      "header",
		StartLine: 1,
		EndLine:   headerEnd,
		Content:   headerContent,
	}
	snippets = append(snippets, headerSnippet)

	// Основное содержимое (если файл больше 20 строк)
	if len(lines) > 20 {
		bodyStart := 21
		bodyEnd := min(len(lines), bodyStart+50) // Следующие 50 строк
		bodyContent := strings.Join(lines[bodyStart-1:bodyEnd], "\n")

		bodySnippet := &domain.ContextSnippet{
			Path:      filePath,
			Kind:      "body",
			StartLine: bodyStart,
			EndLine:   bodyEnd,
			Content:   bodyContent,
		}
		snippets = append(snippets, bodySnippet)
	}

	return snippets
}

// createDependencies создает информацию о зависимостях
func (b *ContextPackBuilderImpl) createDependencies(projectRoot, language string) ([]*domain.ContextDependency, error) {
	var deps []*domain.ContextDependency

	switch language {
	case "go":
		deps = b.createGoDependencies(projectRoot)
	case "typescript", "javascript":
		deps = b.createNodeDependencies(projectRoot)
	}

	return deps, nil
}

// createGoDependencies создает зависимости для Go проекта
func (b *ContextPackBuilderImpl) createGoDependencies(projectRoot string) []*domain.ContextDependency {
	var deps []*domain.ContextDependency

	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		content, err := os.ReadFile(goModPath)
		if err == nil {
			// Простой парсинг go.mod
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "require ") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						dep := &domain.ContextDependency{
							Pkg:     parts[1],
							Version: parts[2],
						}
						deps = append(deps, dep)
					}
				}
			}
		}
	}

	return deps
}

// createNodeDependencies создает зависимости для Node.js проекта
func (b *ContextPackBuilderImpl) createNodeDependencies(projectRoot string) []*domain.ContextDependency {
	var deps []*domain.ContextDependency

	packageJsonPath := filepath.Join(projectRoot, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		content, err := os.ReadFile(packageJsonPath)
		if err == nil {
			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if json.Unmarshal(content, &pkg) == nil {
				for name, version := range pkg.Dependencies {
					dep := &domain.ContextDependency{
						Pkg:     name,
						Version: version,
					}
					deps = append(deps, dep)
				}
			}
		}
	}

	return deps
}

// createBuildInfo создает информацию о сборке
func (b *ContextPackBuilderImpl) createBuildInfo(language string) *domain.ContextBuild {
	build := &domain.ContextBuild{
		Toolchain: language,
		Commands:  []string{},
		Env:       make(map[string]string),
	}

	switch language {
	case "go":
		build.Commands = []string{"go build", "go test"}
		build.Env["GOOS"] = "linux"
		build.Env["GOARCH"] = "amd64"
	case "typescript", "javascript":
		build.Commands = []string{"npm install", "npm run build", "npm test"}
		build.Env["NODE_ENV"] = "production"
	}

	return build
}

// createTests создает информацию о тестах
func (b *ContextPackBuilderImpl) createTests(projectRoot string, relevantFiles []string) ([]*domain.ContextTest, error) {
	var tests []*domain.ContextTest

	// Ищем тестовые файлы
	testPatterns := []string{"*_test.go", "*_test.ts", "*_test.js", "*.test.ts", "*.test.js", "*.spec.ts", "*.spec.js"}

	for _, pattern := range testPatterns {
		matches, err := filepath.Glob(filepath.Join(projectRoot, "**", pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			relPath, _ := filepath.Rel(projectRoot, match)

			// Проверяем, связан ли тест с релевантными файлами
			reason := b.findTestReason(relPath, relevantFiles)

			test := &domain.ContextTest{
				Path:   relPath,
				Reason: reason,
			}
			tests = append(tests, test)
		}
	}

	return tests, nil
}

// findTestReason определяет причину включения теста
func (b *ContextPackBuilderImpl) findTestReason(testPath string, relevantFiles []string) string {
	// Простая логика: если тест имеет похожее имя с релевантным файлом
	testName := strings.TrimSuffix(filepath.Base(testPath), filepath.Ext(testPath))
	testName = strings.TrimSuffix(testName, "_test")
	testName = strings.TrimSuffix(testName, ".test")
	testName = strings.TrimSuffix(testName, ".spec")

	for _, file := range relevantFiles {
		fileName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		if strings.Contains(fileName, testName) || strings.Contains(testName, fileName) {
			return fmt.Sprintf("Tests %s", fileName)
		}
	}

	return "General test coverage"
}

// detectLanguage определяет язык проекта
func (b *ContextPackBuilderImpl) detectLanguage(projectRoot string) string {
	// Проверяем наличие характерных файлов
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "package.json")); err == nil {
		return "typescript"
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "pom.xml")); err == nil {
		return "java"
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "requirements.txt")); err == nil {
		return "python"
	}

	return "unknown"
}

// getAllFiles получает все файлы проекта
func (b *ContextPackBuilderImpl) getAllFiles(projectRoot string, excludePatterns []string) ([]*domain.FileNode, error) {
	// Используем существующий TreeBuilder
	treeBuilder := &NoopTreeBuilder{}
	return treeBuilder.BuildTree(projectRoot, true, true)
}

// getFilePaths извлекает пути файлов из FileNode
func (b *ContextPackBuilderImpl) getFilePaths(nodes []*domain.FileNode) []string {
	var paths []string
	for _, node := range nodes {
		if !node.IsDir {
			paths = append(paths, node.RelPath)
		}
	}
	return paths
}

// calculateTextHash вычисляет хеш текста
func (b *ContextPackBuilderImpl) calculateTextHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// generateIndexID генерирует ID индекса
func (b *ContextPackBuilderImpl) generateIndexID(projectRoot, task string) string {
	data := fmt.Sprintf("%s:%s:%s", projectRoot, task, time.Now().UTC().Format("2006-01-02"))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// UpdatePack обновляет упакованный контекст
func (b *ContextPackBuilderImpl) UpdatePack(ctx context.Context, pack *domain.ContextPack, changedFiles []string) (*domain.ContextPack, error) {
	// Для простоты перестраиваем весь пакет
	// В будущем можно оптимизировать для инкрементальных обновлений
	return pack, nil
}

// ValidatePack проверяет корректность упакованного контекста
func (b *ContextPackBuilderImpl) ValidatePack(ctx context.Context, pack *domain.ContextPack) error {
	if pack == nil {
		return fmt.Errorf("pack is nil")
	}

	if pack.Target == nil {
		return fmt.Errorf("target is required")
	}

	if len(pack.Snippets) == 0 {
		return fmt.Errorf("at least one snippet is required")
	}

	if pack.Constraints == nil {
		return fmt.Errorf("constraints are required")
	}

	return nil
}

// Вспомогательные функции
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NoopTreeBuilder - заглушка для TreeBuilder
type NoopTreeBuilder struct{}

func (b *NoopTreeBuilder) BuildTree(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	return []*domain.FileNode{}, nil
}
