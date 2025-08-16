package app

import (
	"context"
	"fmt"
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

	// new wiring
	archiverinfra "shotgun_code/infrastructure/archiver"
	"shotgun_code/infrastructure/pdfgen"
)

const openRouterHost = "https://openrouter.ai/api/v1"

// AppContainer holds all the services and repositories for the application.
type AppContainer struct {
	Log             domain.Logger
	Bus             domain.EventBus
	SettingsRepo    domain.SettingsRepository
	FileReader      domain.FileContentReader
	GitRepo         domain.GitRepository
	TreeBuilder     domain.TreeBuilder
	ContextSplitter domain.ContextSplitter
	Watcher         domain.FileSystemWatcher
	SettingsService *application.SettingsService
	ProjectService  *application.ProjectService
	AIService       *application.AIService
	ContextAnalysis domain.ContextAnalyzer
	ExportService   *application.ExportService
	Bridge          *wailsbridge.Bridge
}

// NewContainer creates and wires up all the application dependencies.
func NewContainer(ctx context.Context, embeddedIgnoreGlob, defaultCustomPrompt string) (*AppContainer, error) {
	c := &AppContainer{}
	var err error

	// Bridge for Wails (Logger and EventBus)
	bridge := wailsbridge.New(ctx)
	c.Bridge = bridge
	c.Log = bridge
	c.Bus = bridge

	// Repositories and Infrastructure
	c.SettingsRepo, err = settingsfs.New(c.Log, embeddedIgnoreGlob, defaultCustomPrompt)
	if err != nil {
		return nil, err
	}
	c.FileReader = filereader.NewSecureFileReader(c.Log)
	c.GitRepo = git.New(c.Log)
	c.TreeBuilder = fsscanner.New(c.SettingsRepo, c.Log)
	c.ContextSplitter = textutils.NewContextSplitter(c.Log)
	c.Watcher, err = fswatcher.New(ctx, c.Bus)
	if err != nil {
		return nil, err
	}

	// Application Services
	modelFetchers := createModelFetchers(ctx, c.Log, c.SettingsRepo)
	c.SettingsService, err = application.NewSettingsService(c.Log, c.Bus, c.SettingsRepo, modelFetchers)
	if err != nil {
		return nil, err
	}
	// Connect watcher to settings changes
	c.SettingsService.OnIgnoreRulesChanged(c.Watcher.RefreshAndRescan)

	contextGenService := application.NewContextGenerationService(c.Log, c.Bus, c.FileReader)
	c.ProjectService = application.NewProjectService(c.Log, c.Bus, c.TreeBuilder, c.GitRepo, contextGenService)

	providerFactory := createProviderFactory(c.Log, c.SettingsService)
	c.AIService = application.NewAIService(c.SettingsService, c.Log, providerFactory)
	c.ContextAnalysis = application.NewKeywordAnalyzer(c.Log)

	// new: wire PDF and ZIP implementations
	pdfGen := pdfgen.NewGofpdfGenerator(c.Log)
	arch := archiverinfra.NewZipArchiver(c.Log)
	c.ExportService = application.NewExportService(c.Log, c.ContextSplitter, pdfGen, arch)

	return c, nil
}

func createModelFetchers(ctx context.Context, log domain.Logger, repo domain.SettingsRepository) map[string]application.ModelFetcher {
	return map[string]application.ModelFetcher{
		"gemini": func(apiKey string) ([]string, error) {
			p, err := ai.NewGemini(apiKey, "", log)
			if err != nil {
				log.Warning("Failed to create Gemini client for model listing: " + err.Error())
				return nil, err
			}
			return p.(*ai.GeminiProviderImpl).ListModels(ctx)
		},
		"openai": func(apiKey string) ([]string, error) {
			p, err := ai.NewOpenAI(apiKey, "", log)
			if err != nil {
				log.Warning("Failed to create OpenAI client for model listing: " + err.Error())
				return nil, err
			}
			return p.(*ai.OpenAIProviderImpl).ListModels(ctx)
		},
		"openrouter": func(apiKey string) ([]string, error) {
			p, err := ai.NewOpenAI(apiKey, openRouterHost, log)
			if err != nil {
				log.Warning("Failed to create OpenRouter client for model listing: " + err.Error())
				return nil, err
			}
			return p.(*ai.OpenAIProviderImpl).ListModels(ctx)
		},
		"localai": func(apiKey string) ([]string, error) {
			localAIHost := repo.GetLocalAIHost()
			p, err := ai.NewOpenAI(apiKey, localAIHost, log)
			if err != nil {
				log.Warning("Failed to create LocalAI client for model listing: " + err.Error())
				return nil, err
			}
			return p.(*ai.OpenAIProviderImpl).ListModels(ctx)
		},
	}
}

func createProviderFactory(log domain.Logger, settingsService *application.SettingsService) application.ProviderFactory {
	return func(providerType, apiKey string) (domain.AIProvider, error) {
		switch providerType {
		case "openai":
			return ai.NewOpenAI(apiKey, "", log)
		case "gemini":
			return ai.NewGemini(apiKey, "", log)
		case "openrouter":
			return ai.NewOpenAI(apiKey, openRouterHost, log)
		case "localai":
			dto, err := settingsService.GetSettingsDTO()
			if err != nil {
				return nil, err
			}
			return ai.NewOpenAI(apiKey, dto.LocalAIHost, log)
		default:
			return nil, fmt.Errorf("unknown AI provider: %s", providerType)
		}
	}
}
