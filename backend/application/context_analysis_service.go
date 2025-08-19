package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
	"time"
)

// ContextAnalysisService предоставляет интеллектуальные возможности для анализа и сбора контекста
type ContextAnalysisService struct {
	aiService       *AIService
	fileReader      domain.FileContentReader
	log             domain.Logger
	settingsService *SettingsService
}

// NewContextAnalysisService создает новый сервис анализа контекста
func NewContextAnalysisService(
	aiService *AIService,
	fileReader domain.FileContentReader,
	log domain.Logger,
	settingsService *SettingsService,
) *ContextAnalysisService {
	return &ContextAnalysisService{
		aiService:       aiService,
		fileReader:      fileReader,
		log:             log,
		settingsService: settingsService,
	}
}

// AnalyzeTaskAndCollectContext анализирует задачу и автоматически собирает релевантный контекст
func (s *ContextAnalysisService) AnalyzeTaskAndCollectContext(
	ctx context.Context,
	task string,
	allFiles []*domain.FileNode,
	rootDir string,
) (*ContextAnalysisResult, error) {
	startTime := time.Now()
	s.log.Info(fmt.Sprintf("Начинаем анализ задачи: %s", task))

	// 1. Анализируем задачу и определяем тип
	taskAnalysis, err := s.analyzeTaskType(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("ошибка анализа типа задачи: %w", err)
	}

	// 2. Определяем приоритетные файлы на основе анализа
	priorityFiles, err := s.determinePriorityFiles(ctx, task, taskAnalysis, allFiles)
	if err != nil {
		return nil, fmt.Errorf("ошибка определения приоритетных файлов: %w", err)
	}

	// 3. Собираем контекст из приоритетных файлов
	context, err := s.collectContextFromFiles(ctx, priorityFiles, rootDir, taskAnalysis)
	if err != nil {
		return nil, fmt.Errorf("ошибка сбора контекста: %w", err)
	}

	// 4. Анализируем зависимости и добавляем связанные файлы
	dependencyFiles, err := s.analyzeDependencies(ctx, priorityFiles, allFiles, rootDir)
	if err != nil {
		s.log.Warning(fmt.Sprintf("Ошибка анализа зависимостей: %v", err))
	}

	// 5. Формируем финальный результат
	result := &ContextAnalysisResult{
		Task:            task,
		TaskType:        taskAnalysis.Type,
		Priority:        taskAnalysis.Priority,
		SelectedFiles:   priorityFiles,
		DependencyFiles: dependencyFiles,
		Context:         context,
		AnalysisTime:    time.Since(startTime),
		Recommendations: s.generateRecommendations(taskAnalysis, priorityFiles),
		EstimatedTokens: s.estimateTokens(context),
		Confidence:      s.calculateConfidence(taskAnalysis, priorityFiles),
	}

	s.log.Info(fmt.Sprintf("Анализ завершен за %v, выбрано %d файлов", result.AnalysisTime, len(priorityFiles)))
	return result, nil
}

// analyzeTaskType анализирует тип задачи и определяет стратегию сбора контекста
func (s *ContextAnalysisService) analyzeTaskType(ctx context.Context, task string) (*TaskAnalysis, error) {
	systemPrompt := `Ты - эксперт по анализу задач разработки. Проанализируй задачу и определи:
1. Тип задачи (bug_fix, feature, refactor, test, documentation, optimization)
2. Приоритет (low, normal, high, critical)
3. Ключевые технологии/фреймворки
4. Типы файлов, которые нужно искать
5. Ключевые слова для поиска

Ответь в формате JSON:
{
  "type": "тип_задачи",
  "priority": "приоритет",
  "technologies": ["tech1", "tech2"],
  "fileTypes": [".go", ".ts", ".vue"],
  "keywords": ["keyword1", "keyword2"],
  "reasoning": "объяснение выбора"
}`

	response, err := s.aiService.GenerateCode(ctx, systemPrompt, task)
	if err != nil {
		return nil, fmt.Errorf("ошибка анализа задачи: %w", err)
	}

	// Парсим ответ (упрощенная версия)
	analysis := &TaskAnalysis{
		Type:         s.extractTaskType(task, response),
		Priority:     s.extractPriority(task, response),
		Technologies: s.extractTechnologies(task, response),
		FileTypes:    s.extractFileTypes(task, response),
		Keywords:     s.extractKeywords(task, response),
		Reasoning:    response,
	}

	return analysis, nil
}

// determinePriorityFiles определяет приоритетные файлы на основе анализа задачи
func (s *ContextAnalysisService) determinePriorityFiles(
	ctx context.Context,
	task string,
	analysis *TaskAnalysis,
	allFiles []*domain.FileNode,
) ([]*domain.FileNode, error) {
	var priorityFiles []*domain.FileNode

	// 1. Фильтруем по типам файлов
	filteredFiles := s.filterByFileTypes(allFiles, analysis.FileTypes)

	// 2. Сортируем по релевантности
	scoredFiles := s.scoreFilesByRelevance(filteredFiles, task, analysis)

	// 3. Выбираем топ файлы
	priorityFiles = s.selectTopFiles(scoredFiles, analysis.Priority)

	// 4. Если файлов мало, расширяем поиск
	if len(priorityFiles) < 3 {
		additionalFiles := s.expandSearch(allFiles, task, analysis)
		priorityFiles = append(priorityFiles, additionalFiles...)
	}

	return priorityFiles, nil
}

// collectContextFromFiles собирает контекст из выбранных файлов
func (s *ContextAnalysisService) collectContextFromFiles(
	ctx context.Context,
	files []*domain.FileNode,
	rootDir string,
	analysis *TaskAnalysis,
) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("нет файлов для сбора контекста")
	}

	// Получаем пути файлов
	var filePaths []string
	for _, file := range files {
		filePaths = append(filePaths, file.RelPath)
	}

	// Читаем содержимое файлов
	contents, err := s.fileReader.ReadContents(ctx, filePaths, rootDir, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения файлов: %w", err)
	}

	// Формируем контекст
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("// Анализ задачи: %s\n", analysis.Type))
	contextBuilder.WriteString(fmt.Sprintf("// Приоритет: %s\n", analysis.Priority))
	contextBuilder.WriteString(fmt.Sprintf("// Технологии: %s\n", strings.Join(analysis.Technologies, ", ")))
	contextBuilder.WriteString("// Контекст проекта:\n\n")

	for _, file := range files {
		if content, exists := contents[file.RelPath]; exists {
			contextBuilder.WriteString(fmt.Sprintf("// Файл: %s\n", file.RelPath))
			contextBuilder.WriteString(content)
			contextBuilder.WriteString("\n\n")
		}
	}

	return contextBuilder.String(), nil
}

// analyzeDependencies анализирует зависимости и добавляет связанные файлы
func (s *ContextAnalysisService) analyzeDependencies(
	ctx context.Context,
	priorityFiles []*domain.FileNode,
	allFiles []*domain.FileNode,
	rootDir string,
) ([]*domain.FileNode, error) {
	var dependencyFiles []*domain.FileNode

	// Простая логика анализа зависимостей
	for _, file := range priorityFiles {
		// Ищем файлы с похожими именами или в тех же папках
		relatedFiles := s.findRelatedFiles(file, allFiles)
		dependencyFiles = append(dependencyFiles, relatedFiles...)
	}

	// Убираем дубликаты
	dependencyFiles = s.removeDuplicates(dependencyFiles)

	return dependencyFiles, nil
}

// Вспомогательные методы для анализа
func (s *ContextAnalysisService) extractTaskType(task, response string) string {
	// Простая логика извлечения типа задачи
	task = strings.ToLower(task)

	if strings.Contains(task, "bug") || strings.Contains(task, "ошибк") || strings.Contains(task, "исправ") {
		return "bug_fix"
	}
	if strings.Contains(task, "тест") || strings.Contains(task, "test") {
		return "test"
	}
	if strings.Contains(task, "документ") || strings.Contains(task, "readme") {
		return "documentation"
	}
	if strings.Contains(task, "оптимиз") || strings.Contains(task, "улучш") {
		return "optimization"
	}
	if strings.Contains(task, "рефактор") || strings.Contains(task, "refactor") {
		return "refactor"
	}

	return "feature"
}

func (s *ContextAnalysisService) extractPriority(task, response string) string {
	task = strings.ToLower(task)

	if strings.Contains(task, "критич") || strings.Contains(task, "сроч") {
		return "critical"
	}
	if strings.Contains(task, "важн") || strings.Contains(task, "приоритет") {
		return "high"
	}
	if strings.Contains(task, "неважн") || strings.Contains(task, "позже") {
		return "low"
	}

	return "normal"
}

func (s *ContextAnalysisService) extractTechnologies(task, response string) []string {
	var technologies []string
	task = strings.ToLower(task)

	if strings.Contains(task, "go") || strings.Contains(task, "golang") {
		technologies = append(technologies, "Go")
	}
	if strings.Contains(task, "vue") || strings.Contains(task, "frontend") {
		technologies = append(technologies, "Vue.js")
	}
	if strings.Contains(task, "typescript") || strings.Contains(task, "ts") {
		technologies = append(technologies, "TypeScript")
	}
	if strings.Contains(task, "react") {
		technologies = append(technologies, "React")
	}
	if strings.Contains(task, "python") {
		technologies = append(technologies, "Python")
	}

	return technologies
}

func (s *ContextAnalysisService) extractFileTypes(task, response string) []string {
	var fileTypes []string
	task = strings.ToLower(task)

	if strings.Contains(task, "go") || strings.Contains(task, "golang") {
		fileTypes = append(fileTypes, ".go")
	}
	if strings.Contains(task, "vue") || strings.Contains(task, "frontend") {
		fileTypes = append(fileTypes, ".vue", ".ts", ".js")
	}
	if strings.Contains(task, "typescript") || strings.Contains(task, "ts") {
		fileTypes = append(fileTypes, ".ts", ".tsx")
	}
	if strings.Contains(task, "test") || strings.Contains(task, "тест") {
		fileTypes = append(fileTypes, "_test.go", ".test.ts", ".spec.ts")
	}

	return fileTypes
}

func (s *ContextAnalysisService) extractKeywords(task, response string) []string {
	var keywords []string
	task = strings.ToLower(task)

	// Извлекаем ключевые слова из задачи
	words := strings.Fields(task)
	for _, word := range words {
		if len(word) > 3 && !s.isCommonWord(word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func (s *ContextAnalysisService) isCommonWord(word string) bool {
	commonWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "is", "are", "was", "were", "be", "been", "have", "has", "had", "do", "does", "did", "will", "would", "could", "should", "may", "might", "can", "this", "that", "these", "those", "a", "an", "as", "so", "than", "too", "very", "just", "now", "then", "here", "there", "when", "where", "why", "how", "all", "any", "both", "each", "few", "more", "most", "other", "some", "such", "no", "nor", "not", "only", "own", "same", "such", "too", "very", "you", "your", "yours", "yourself", "yourselves", "i", "me", "my", "myself", "we", "our", "ours", "ourselves", "what", "which", "who", "whom", "this", "that", "these", "those", "am", "is", "are", "was", "were", "be", "been", "being", "have", "has", "had", "having", "do", "does", "did", "doing", "will", "would", "could", "should", "may", "might", "must", "shall", "can"}

	for _, common := range commonWords {
		if word == common {
			return true
		}
	}
	return false
}

func (s *ContextAnalysisService) filterByFileTypes(files []*domain.FileNode, fileTypes []string) []*domain.FileNode {
	if len(fileTypes) == 0 {
		return files
	}

	var filtered []*domain.FileNode
	for _, file := range files {
		for _, fileType := range fileTypes {
			if strings.HasSuffix(file.RelPath, fileType) {
				filtered = append(filtered, file)
				break
			}
		}
	}

	return filtered
}

func (s *ContextAnalysisService) scoreFilesByRelevance(files []*domain.FileNode, task string, analysis *TaskAnalysis) []ScoredFile {
	var scoredFiles []ScoredFile

	for _, file := range files {
		score := s.calculateFileScore(file, task, analysis)
		scoredFiles = append(scoredFiles, ScoredFile{
			File:  file,
			Score: score,
		})
	}

	// Сортируем по убыванию релевантности
	for i := 0; i < len(scoredFiles); i++ {
		for j := i + 1; j < len(scoredFiles); j++ {
			if scoredFiles[i].Score < scoredFiles[j].Score {
				scoredFiles[i], scoredFiles[j] = scoredFiles[j], scoredFiles[i]
			}
		}
	}

	return scoredFiles
}

func (s *ContextAnalysisService) calculateFileScore(file *domain.FileNode, task string, analysis *TaskAnalysis) float64 {
	score := 0.0
	task = strings.ToLower(task)
	fileName := strings.ToLower(file.RelPath)

	// Базовый скор за размер файла (предпочитаем файлы среднего размера)
	if file.Size > 100 && file.Size < 10000 {
		score += 0.3
	}

	// Скор за ключевые слова в имени файла
	for _, keyword := range analysis.Keywords {
		if strings.Contains(fileName, strings.ToLower(keyword)) {
			score += 0.5
		}
	}

	// Скор за тип файла
	for _, fileType := range analysis.FileTypes {
		if strings.HasSuffix(fileName, strings.ToLower(fileType)) {
			score += 0.4
		}
	}

	// Скор за расположение файла (предпочитаем файлы в корне или src)
	if strings.Count(fileName, "/") <= 1 {
		score += 0.2
	}

	return score
}

func (s *ContextAnalysisService) selectTopFiles(scoredFiles []ScoredFile, priority string) []*domain.FileNode {
	var selectedFiles []*domain.FileNode

	// Выбираем количество файлов в зависимости от приоритета
	maxFiles := 5
	switch priority {
	case "critical":
		maxFiles = 10
	case "high":
		maxFiles = 8
	case "normal":
		maxFiles = 5
	case "low":
		maxFiles = 3
	}

	// Выбираем топ файлы
	for i := 0; i < len(scoredFiles) && i < maxFiles; i++ {
		if scoredFiles[i].Score > 0.1 { // Минимальный порог релевантности
			selectedFiles = append(selectedFiles, scoredFiles[i].File)
		}
	}

	return selectedFiles
}

func (s *ContextAnalysisService) expandSearch(allFiles []*domain.FileNode, task string, analysis *TaskAnalysis) []*domain.FileNode {
	var additionalFiles []*domain.FileNode

	// Ищем файлы с похожими именами или в тех же папках
	for _, file := range allFiles {
		fileName := strings.ToLower(file.RelPath)

		// Проверяем, содержит ли имя файла ключевые слова
		for _, keyword := range analysis.Keywords {
			if strings.Contains(fileName, strings.ToLower(keyword)) {
				additionalFiles = append(additionalFiles, file)
				break
			}
		}
	}

	return additionalFiles
}

func (s *ContextAnalysisService) findRelatedFiles(file *domain.FileNode, allFiles []*domain.FileNode) []*domain.FileNode {
	var relatedFiles []*domain.FileNode

	// Ищем файлы в той же папке
	dir := s.getDirectory(file.RelPath)
	for _, f := range allFiles {
		if s.getDirectory(f.RelPath) == dir && f.RelPath != file.RelPath {
			relatedFiles = append(relatedFiles, f)
		}
	}

	return relatedFiles
}

func (s *ContextAnalysisService) getDirectory(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return ""
	}
	return path[:lastSlash]
}

func (s *ContextAnalysisService) removeDuplicates(files []*domain.FileNode) []*domain.FileNode {
	seen := make(map[string]bool)
	var unique []*domain.FileNode

	for _, file := range files {
		if !seen[file.RelPath] {
			seen[file.RelPath] = true
			unique = append(unique, file)
		}
	}

	return unique
}

func (s *ContextAnalysisService) generateRecommendations(analysis *TaskAnalysis, files []*domain.FileNode) []string {
	var recommendations []string

	if len(files) < 3 {
		recommendations = append(recommendations, "Рекомендуется добавить больше файлов для лучшего понимания контекста")
	}

	if analysis.Type == "bug_fix" {
		recommendations = append(recommendations, "Для исправления багов рекомендуется также включить файлы с тестами")
	}

	if analysis.Type == "test" {
		recommendations = append(recommendations, "Убедитесь, что включены все необходимые файлы для тестирования")
	}

	return recommendations
}

func (s *ContextAnalysisService) estimateTokens(context string) int {
	// Простая оценка токенов (примерно 4 символа на токен)
	return len(context) / 4
}

func (s *ContextAnalysisService) calculateConfidence(analysis *TaskAnalysis, files []*domain.FileNode) float64 {
	confidence := 0.5 // Базовая уверенность

	// Увеличиваем уверенность за количество файлов
	if len(files) >= 5 {
		confidence += 0.2
	} else if len(files) >= 3 {
		confidence += 0.1
	}

	// Увеличиваем уверенность за четкий тип задачи
	if analysis.Type != "feature" {
		confidence += 0.1
	}

	// Увеличиваем уверенность за высокий приоритет
	if analysis.Priority == "critical" || analysis.Priority == "high" {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// Структуры данных
type TaskAnalysis struct {
	Type         string   `json:"type"`
	Priority     string   `json:"priority"`
	Technologies []string `json:"technologies"`
	FileTypes    []string `json:"fileTypes"`
	Keywords     []string `json:"keywords"`
	Reasoning    string   `json:"reasoning"`
}

type ScoredFile struct {
	File  *domain.FileNode
	Score float64
}

type ContextAnalysisResult struct {
	Task            string             `json:"task"`
	TaskType        string             `json:"taskType"`
	Priority        string             `json:"priority"`
	SelectedFiles   []*domain.FileNode `json:"selectedFiles"`
	DependencyFiles []*domain.FileNode `json:"dependencyFiles"`
	Context         string             `json:"context"`
	AnalysisTime    time.Duration      `json:"analysisTime"`
	Recommendations []string           `json:"recommendations"`
	EstimatedTokens int                `json:"estimatedTokens"`
	Confidence      float64            `json:"confidence"`
}
