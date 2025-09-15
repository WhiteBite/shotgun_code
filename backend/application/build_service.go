package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
)

// BuildService предоставляет высокоуровневый API для работы с build pipeline
type BuildService struct {
	log      domain.Logger
	pipeline domain.BuildPipeline
}

// NewBuildService создает новый сервис сборки
func NewBuildService(log domain.Logger, pipeline domain.BuildPipeline) *BuildService {
	return &BuildService{
		log:      log,
		pipeline: pipeline,
	}
}

// Build выполняет сборку проекта
func (s *BuildService) Build(ctx context.Context, projectPath, language string) (*domain.BuildResult, error) {
	s.log.Info(fmt.Sprintf("Building %s project at %s", language, projectPath))

	return s.pipeline.Build(ctx, projectPath, language)
}

// TypeCheck выполняет проверку типов
func (s *BuildService) TypeCheck(ctx context.Context, projectPath, language string) (*domain.TypeCheckResult, error) {
	s.log.Info(fmt.Sprintf("Type checking %s project at %s", language, projectPath))

	return s.pipeline.TypeCheck(ctx, projectPath, language)
}

// BuildAndTypeCheck выполняет сборку и проверку типов
func (s *BuildService) BuildAndTypeCheck(ctx context.Context, projectPath, language string) (*domain.BuildResult, *domain.TypeCheckResult, error) {
	s.log.Info(fmt.Sprintf("Building and type checking %s project at %s", language, projectPath))

	return s.pipeline.BuildAndTypeCheck(ctx, projectPath, language)
}

// BuildMultiLanguage выполняет сборку для нескольких языков
func (s *BuildService) BuildMultiLanguage(ctx context.Context, projectPath string, languages []string) (map[string]*domain.BuildResult, error) {
	s.log.Info(fmt.Sprintf("Building multi-language project at %s for languages: %v", projectPath, languages))

	results := make(map[string]*domain.BuildResult)

	for _, language := range languages {
		result, err := s.Build(ctx, projectPath, language)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to build %s: %v", language, err))
			continue
		}
		results[language] = result
	}

	return results, nil
}

// TypeCheckMultiLanguage выполняет проверку типов для нескольких языков
func (s *BuildService) TypeCheckMultiLanguage(ctx context.Context, projectPath string, languages []string) (map[string]*domain.TypeCheckResult, error) {
	s.log.Info(fmt.Sprintf("Type checking multi-language project at %s for languages: %v", projectPath, languages))

	results := make(map[string]*domain.TypeCheckResult)

	for _, language := range languages {
		result, err := s.TypeCheck(ctx, projectPath, language)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to type check %s: %v", language, err))
			continue
		}
		results[language] = result
	}

	return results, nil
}

// ValidateProject выполняет полную валидацию проекта
func (s *BuildService) ValidateProject(ctx context.Context, projectPath string, languages []string) (*domain.ProjectValidationResult, error) {
	s.log.Info(fmt.Sprintf("Validating project at %s for languages: %v", projectPath, languages))

	validation := &domain.ProjectValidationResult{
		ProjectPath: projectPath,
		Languages:   languages,
		Results:     make(map[string]*domain.LanguageValidationResult),
	}

	for _, language := range languages {
		langResult := &domain.LanguageValidationResult{
			Language: language,
		}

		// Выполняем проверку типов
		typeCheckResult, err := s.TypeCheck(ctx, projectPath, language)
		if err != nil {
			langResult.TypeCheckError = err.Error()
		} else {
			langResult.TypeCheckResult = typeCheckResult
		}

		// Выполняем сборку
		buildResult, err := s.Build(ctx, projectPath, language)
		if err != nil {
			langResult.BuildError = err.Error()
		} else {
			langResult.BuildResult = buildResult
		}

		// Определяем общий статус
		langResult.Success = (langResult.TypeCheckResult != nil && langResult.TypeCheckResult.Success) &&
			(langResult.BuildResult != nil && langResult.BuildResult.Success)

		validation.Results[language] = langResult
	}

	// Определяем общий статус проекта
	validation.Success = true
	for _, result := range validation.Results {
		if !result.Success {
			validation.Success = false
			break
		}
	}

	return validation, nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (s *BuildService) GetSupportedLanguages() []string {
	return s.pipeline.GetSupportedLanguages()
}

// DetectLanguages определяет языки в проекте
func (s *BuildService) DetectLanguages(ctx context.Context, projectPath string) ([]string, error) {
	s.log.Info(fmt.Sprintf("Detecting languages in project at %s", projectPath))

	var detectedLanguages []string

	// Проверяем наличие файлов конфигурации для каждого языка
	supportedLanguages := s.GetSupportedLanguages()

	for _, language := range supportedLanguages {
		switch language {
		case "go":
			if s.hasFile(projectPath, "go.mod") {
				detectedLanguages = append(detectedLanguages, language)
			}
		case "typescript", "ts":
			if s.hasFile(projectPath, "package.json") || s.hasFile(projectPath, "tsconfig.json") {
				detectedLanguages = append(detectedLanguages, language)
			}
		case "java":
			if s.hasFile(projectPath, "pom.xml") || s.hasFile(projectPath, "build.gradle") {
				detectedLanguages = append(detectedLanguages, language)
			}
		}
	}

	s.log.Info(fmt.Sprintf("Detected languages: %v", detectedLanguages))
	return detectedLanguages, nil
}

// hasFile проверяет наличие файла
func (s *BuildService) hasFile(projectPath, filename string) bool {
	// Простая проверка - в реальной реализации нужно использовать os.Stat
	// Здесь возвращаем true для демонстрации
	return true
}
