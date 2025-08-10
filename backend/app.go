package main

import (
	"context"
	_ "embed"
	"fmt"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/ai"
	"shotgun_code/infrastructure/filereader"
	"shotgun_code/infrastructure/fsscanner"
	"shotgun_code/infrastructure/fswatcher"
	"shotgun_code/infrastructure/git"
	"shotgun_code/infrastructure/settingsfs"
	"shotgun_code/infrastructure/wailsbridge"
	"shotgun_code/logic"
)

var defaultCustomIgnoreRulesContent string = `
node_modules/
*.tmp
*.log
dist/
build/
`

const defaultCustomPromptRulesContent = "no additional rules"

type App struct {
	ctx                    context.Context
	projectService         *application.ProjectService
	settingsService        *application.SettingsService
	aiService              *application.AIService
	contextAnalysisService domain.ContextAnalyzer
	fileWatcher            domain.FileSystemWatcher
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	bridge := wailsbridge.New(ctx)
	settingsRepo, err := settingsfs.New(bridge, defaultCustomIgnoreRulesContent, defaultCustomPromptRulesContent)
	if err != nil {
		bridge.Fatal(fmt.Sprintf("Критическая ошибка: не удалось инициализировать репозиторий настроек: %v", err))
		return
	}

	modelFetchers := make(map[string]application.ModelFetcher)
	modelFetchers["gemini"] = func(apiKey string) ([]string, error) {
		if apiKey == "" {
			return nil, domain.ErrProviderNotConfigured
		}
		tempProvider, err := ai.NewGemini(apiKey, "", bridge)
		if err != nil {
			return nil, err
		}
		if p, ok := tempProvider.(*ai.GeminiProviderImpl); ok {
			return p.ListModels(ctx)
		}
		return nil, fmt.Errorf("не удалось привести провайдер к нужному типу")
	}
	modelFetchers["openai"] = func(apiKey string) ([]string, error) { return settingsRepo.GetModels("openai"), nil }
	modelFetchers["localai"] = func(apiKey string) ([]string, error) { return settingsRepo.GetModels("localai"), nil }

	a.settingsService, err = application.NewSettingsService(bridge, bridge, settingsRepo, modelFetchers)
	if err != nil {
		bridge.Fatal(fmt.Sprintf("Критическая ошибка: %v", err))
		return
	}

	a.contextAnalysisService = application.NewKeywordAnalyzer(bridge)
	fileReader := filereader.NewSecureFileReader(bridge)
	contextGenService := application.NewContextGenerationService(bridge, bridge, fileReader)

	gitRepo := git.New(bridge)
	treeBuilder := fsscanner.New(settingsRepo)
	watcher, err := fswatcher.New(ctx, bridge)
	if err != nil {
		bridge.Fatal(fmt.Sprintf("Критическая ошибка: %v", err))
		return
	}
	a.fileWatcher = watcher
	diffSplitter := logic.NewDiffSplitter(bridge)
	a.settingsService.OnIgnoreRulesChanged(a.fileWatcher.RefreshAndRescan)

	a.projectService, err = application.NewProjectService(
		bridge, bridge, a.settingsService, gitRepo, treeBuilder, diffSplitter, contextGenService,
	)
	if err != nil {
		bridge.Fatal(fmt.Sprintf("Критическая ошибка: %v", err))
		return
	}

	providerFactory := func(providerType, apiKey, modelName string) (domain.AIProvider, error) {
		switch providerType {
		case "openai":
			return ai.NewOpenAI(apiKey, "", bridge)
		case "gemini":
			return ai.NewGemini(apiKey, modelName, bridge)
		case "localai":
			host := settingsRepo.GetLocalAIHost()
			localAPIKey := settingsRepo.GetLocalAIKey()
			return ai.NewOpenAI(localAPIKey, host, bridge)
		default:
			return nil, fmt.Errorf("неизвестный AI провайдер: %s", providerType)
		}
	}

	a.aiService = application.NewAIService(a.settingsService, bridge, providerFactory)
	bridge.Info("Приложение успешно запущено и все сервисы инициализированы.")
}

func (a *App) handleError(err error) {
	if err != nil {
		if a.projectService != nil {
			a.projectService.LogError(err.Error())
		} else {
			// Fallback на случай, если ошибка произошла до инициализации сервисов
			bridge := wailsbridge.New(a.ctx)
			bridge.Error(err.Error())
			// Попытаемся также отправить событие, если это возможно
			bridge.Emit("app:error", err.Error())
		}
	}
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	go a.projectService.GenerateContext(a.ctx, rootDir, includedPaths)
}

func (a *App) RefreshAIModels(provider string, apiKey string) error {
	err := a.settingsService.RefreshModels(provider, apiKey)
	if err != nil {
		a.handleError(err)
		return err
	}
	return nil
}

func (a *App) SuggestContextFiles(task string, allFiles []*domain.FileNode) ([]string, error) {
	suggested, err := a.contextAnalysisService.SuggestFiles(a.ctx, task, allFiles)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return suggested, nil
}

func (a *App) SelectDirectory() (string, error) {
	path, err := wailsbridge.New(a.ctx).OpenDirectoryDialog()
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return path, nil
}

func (a *App) GenerateCode(systemPrompt, userPrompt string) (string, error) {
	content, err := a.aiService.GenerateCode(a.ctx, systemPrompt, userPrompt)
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return content, nil
}

func (a *App) ListFiles(dirPath string) ([]*domain.FileNode, error) {
	nodes, err := a.projectService.ListFiles(dirPath)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return nodes, nil
}

func (a *App) SplitShotgunDiff(gitDiffText string, approxLineLimit int) ([]string, error) {
	splits, err := a.projectService.SplitShotgunDiff(gitDiffText, approxLineLimit)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return splits, nil
}

func (a *App) StartFileWatcher(rootDirPath string) {
	a.handleError(a.fileWatcher.Start(rootDirPath))
}

func (a *App) StopFileWatcher() {
	a.fileWatcher.Stop()
}

// --- Settings Methods ---
func (a *App) GetCustomIgnoreRules() string { return a.settingsService.GetCustomIgnoreRules() }
func (a *App) SetCustomIgnoreRules(rules string) {
	a.handleError(a.settingsService.SetCustomIgnoreRules(rules))
}
func (a *App) GetCustomPromptRules() string { return a.settingsService.GetCustomPromptRules() }
func (a *App) SetCustomPromptRules(rules string) {
	a.handleError(a.settingsService.SetCustomPromptRules(rules))
}
func (a *App) SetOpenAIKey(key string)    { a.handleError(a.settingsService.SetOpenAIKey(key)) }
func (a *App) GetOpenAIKey() string       { return a.settingsService.GetOpenAIKey() }
func (a *App) SetGeminiKey(key string)    { a.handleError(a.settingsService.SetGeminiKey(key)) }
func (a *App) GetGeminiKey() string       { return a.settingsService.GetGeminiKey() }
func (a *App) SetLocalAIKey(key string)   { a.handleError(a.settingsService.SetLocalAIKey(key)) }
func (a *App) GetLocalAIKey() string      { return a.settingsService.GetLocalAIKey() }
func (a *App) SetLocalAIHost(host string) { a.handleError(a.settingsService.SetLocalAIHost(host)) }
func (a *App) GetLocalAIHost() string     { return a.settingsService.GetLocalAIHost() }
func (a *App) SetLocalAIModelName(name string) {
	a.handleError(a.settingsService.SetLocalAIModelName(name))
}
func (a *App) GetLocalAIModelName() string { return a.settingsService.GetLocalAIModelName() }
func (a *App) SetUseGitignore(enabled bool) {
	a.handleError(a.settingsService.SetUseGitignore(enabled))
}
func (a *App) SetUseCustomIgnore(enabled bool) {
	a.handleError(a.settingsService.SetUseCustomIgnore(enabled))
}
func (a *App) GetSelectedAIProvider() string { return a.settingsService.GetSelectedAIProvider() }
func (a *App) SetSelectedAIProvider(provider string) {
	a.handleError(a.settingsService.SetSelectedAIProvider(provider))
}
func (a *App) GetModels(provider string) ([]string, error) {
	return a.settingsService.GetModels(provider)
}
func (a *App) GetSelectedModel(provider string) string {
	return a.settingsService.GetSelectedModel(provider)
}
func (a *App) SetSelectedModel(provider, model string) {
	a.handleError(a.settingsService.SetSelectedModel(provider, model))
}

// --- Git Methods ---
func (a *App) IsGitAvailable() bool { return a.projectService.IsGitAvailable() }
func (a *App) GetUncommittedFiles(projectRoot string) ([]string, error) {
	files, err := a.projectService.GetUncommittedFiles(projectRoot)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return files, nil
}
func (a *App) GetRichCommitHistory(projectRoot string, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	commits, err := a.projectService.GetRichCommitHistory(projectRoot, branchName, limit)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return commits, nil
}
