package application

import (
	"context"
	"regexp"
	"shotgun_code/domain"
	"strings"
)

// KeywordAnalyzer является простой реализацией ContextAnalyzer.
// Он извлекает ключевые слова из задачи и ищет их в путях к файлам.
type KeywordAnalyzer struct {
	log domain.Logger
}

// NewKeywordAnalyzer создает новый экземпляр KeywordAnalyzer.
func NewKeywordAnalyzer(log domain.Logger) *KeywordAnalyzer {
	return &KeywordAnalyzer{log: log}
}

// SuggestFiles реализует простую логику поиска по ключевым словам.
func (a *KeywordAnalyzer) SuggestFiles(ctx context.Context, task string, allFiles []*domain.FileNode) ([]string, error) {
	keywords := a.extractKeywords(task)
	if len(keywords) == 0 {
		a.log.Warning("Не удалось извлечь ключевые слова из задачи для авто-выбора.")
		return []string{}, nil
	}

	a.log.Info("Извлеченные ключевые слова для поиска: " + strings.Join(keywords, ", "))

	relevantFiles := make(map[string]struct{})
	var fileList []string

	// Рекурсивная функция для обхода дерева файлов
	var traverse func([]*domain.FileNode)
	traverse = func(nodes []*domain.FileNode) {
		for _, node := range nodes {
			// Собираем все пути в плоский список для дальнейшей обработки
			if !node.IsDir {
				fileList = append(fileList, node.RelPath)
			}
			if len(node.Children) > 0 {
				traverse(node.Children)
			}
		}
	}
	traverse(allFiles)

	for _, path := range fileList {
		for _, keyword := range keywords {
			// Используем case-insensitive поиск
			if strings.Contains(strings.ToLower(path), strings.ToLower(keyword)) {
				relevantFiles[path] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(relevantFiles))
	for path := range relevantFiles {
		result = append(result, path)
	}

	a.log.Info("Предложено " + string(rune(len(result))) + " релевантных файлов.")
	return result, nil
}

// extractKeywords извлекает потенциальные ключевые слова из строки задачи.
// Удаляет знаки препинания, короткие слова и приводит к нижнему регистру.
func (a *KeywordAnalyzer) extractKeywords(task string) []string {
	// Удаляем знаки препинания
	re := regexp.MustCompile(`[^\w\s]`)
	cleanTask := re.ReplaceAllString(task, " ")

	words := strings.Fields(strings.ToLower(cleanTask))
	var keywords []string
	for _, word := range words {
		// Игнорируем слишком короткие слова, которые вряд ли будут информативны
		if len(word) > 3 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}
