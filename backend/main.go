package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/ai"
	"shotgun_code/infrastructure/filereader"
	"shotgun_code/infrastructure/fsscanner"
	"shotgun_code/infrastructure/fswatcher"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/settingsfs"
	"shotgun_code/infrastructure/textutils"
	"shotgun_code/infrastructure/wailsbridge"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/ignore.glob
var embeddedIgnoreGlob string

func init() {
	defaultCustomIgnoreRulesContent = embeddedIgnoreGlob
}

var defaultCustomIgnoreRulesContent = embeddedIgnoreGlob

const defaultCustomPromptRulesContent = "no additional rules"
const openRouterHost = "https://openrouter.ai/api/v1"

type AppContainer struct {
	Log             domain.Logger
	Bus             domain.EventBus
	SettingsRepo    domain.SettingsRepository
	FileReader      domain.FileContentReader
	GitRepo         domain.GitRepository
	TreeBuilder     domain.FileTreeBuilder
	DiffSplitter    domain.DiffSplitter
	Watcher         domain.FileSystemWatcher
	SettingsService *application.SettingsService
	ProjectService  *application.ProjectService
	AIService       *application.AIService
	ContextAnalysis domain.ContextAnalyzer
	ExportService   *application.ExportService
}

func NewContainer(ctx context.Context) (*AppContainer, error) {
	c := &AppContainer{}
	var err error

	c.Log = wailsbridge.New(ctx)
	c.Bus = c.Log.(*wailsbridge.Bridge)

	c.SettingsRepo, err = settingsfs.New(c.Log, defaultCustomIgnoreRulesContent, defaultCustomPromptRulesContent)
	if err != nil {
		return nil, err
	}
	c.FileReader = filereader.NewSecureFileReader(c.Log)
	c.GitRepo = git.New(c.Log)
	c.TreeBuilder = fsscanner.New(c.SettingsRepo)
	c.DiffSplitter = textutils.NewDiffSplitter(c.Log)
	c.Watcher, err = fswatcher.New(ctx, c.Bus)
	if err != nil {
		return nil, err
	}

	modelFetchers := createModelFetchers(ctx, c.Log, c.SettingsRepo)
	c.SettingsService, err = application.NewSettingsService(c.Log, c.Bus, c.SettingsRepo, modelFetchers)
	if err != nil {
		return nil, err
	}
	c.SettingsService.OnIgnoreRulesChanged(c.Watcher.RefreshAndRescan)

	contextGenService := application.NewContextGenerationService(c.Log, c.Bus, c.FileReader)
	c.ProjectService, err = application.NewProjectService(c.Log, c.Bus, c.SettingsService, c.GitRepo, c.TreeBuilder, c.DiffSplitter, contextGenService)
	if err != nil {
		return nil, err
	}

	providerFactory := createProviderFactory(c.Log, c.SettingsService)
	c.AIService = application.NewAIService(c.SettingsService, c.Log, providerFactory)
	c.ContextAnalysis = application.NewKeywordAnalyzer(c.Log)
	c.ExportService = application.NewExportService()

	return c, nil
}

func main() {
	app := &App{}

	err := wails.Run(&options.App{
		Title:       "Shotgun App",
		Width:       1280,
		Height:      800,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup: func(ctx context.Context) {
			container, err := NewContainer(ctx)
			if err != nil {
				log.Fatalf("Failed to create DI container: %v", err)
			}
			app.projectService = container.ProjectService
			app.aiService = container.AIService
			app.settingsService = container.SettingsService
			app.contextAnalysis = container.ContextAnalysis
			app.fileWatcher = container.Watcher
			app.gitRepo = container.GitRepo
			app.exportService = container.ExportService
			app.startup(ctx)
		},
		Bind: []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createModelFetchers(ctx context.Context, log domain.Logger, repo domain.SettingsRepository) map[string]application.ModelFetcher {
	return map[string]application.ModelFetcher{
		"gemini": func(apiKey string) ([]string, error) {
			p, err := ai.NewGemini(apiKey, "", log)
			if err != nil {
				return nil, err
			}
			return p.(*ai.GeminiProviderImpl).ListModels(ctx)
		},
		"openai":     func(_ string) ([]string, error) { return repo.GetModels("openai"), nil },
		"openrouter": func(_ string) ([]string, error) { return repo.GetModels("openrouter"), nil },
		"localai":    func(_ string) ([]string, error) { return repo.GetModels("localai"), nil },
	}
}

func createProviderFactory(log domain.Logger, settingsService *application.SettingsService) application.ProviderFactory {
	return func(providerType, apiKey, modelName string) (domain.AIProvider, error) {
		switch providerType {
		case "openai":
			return ai.NewOpenAI(apiKey, "", log)
		case "gemini":
			return ai.NewGemini(apiKey, modelName, log)
		case "openrouter":
			return ai.NewOpenAI(apiKey, openRouterHost, log)
		case "localai":
			dto, err := settingsService.GetSettingsDTO()
			if err != nil {
				return nil, err
			}
			return ai.NewOpenAI(dto.LocalAIAPIKey, dto.LocalAIHost, log)
		default:
			return nil, fmt.Errorf("unknown AI provider: %s", providerType)
		}
	}
}
