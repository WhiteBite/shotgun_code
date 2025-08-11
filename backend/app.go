package main

import (
	"context"
	"fmt"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/wailsbridge"
)

// App теперь выступает в роли фасада, скрывая сервисы.
type App struct {
	ctx             context.Context
	projectService  *application.ProjectService
	aiService       *application.AIService
	settingsService *application.SettingsService
	fileWatcher     domain.FileSystemWatcher
	contextAnalysis domain.ContextAnalyzer
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) handleError(err error) {
	if err != nil {
		if a.projectService != nil {
			a.projectService.LogError(err.Error())
		} else {
			bridge := wailsbridge.New(a.ctx)
			bridge.Error(err.Error())
			bridge.Emit("app:error", err.Error())
		}
	}
}

// --- Project & Context Methods ---
func (a *App) ListFiles(dirPath string) ([]*domain.FileNode, error) {
	nodes, err := a.projectService.ListFiles(dirPath)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return nodes, nil
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.handleError(fmt.Errorf("PANIC recovered in GenerateContext: %v", r))
			}
		}()
		a.projectService.GenerateContext(a.ctx, rootDir, includedPaths)
	}()
}

// --- Git Methods ---
func (a *App) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	files, err := a.projectService.GetUncommittedFiles(projectRoot)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return files, nil
}

func (a *App) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	commits, err := a.projectService.GetRichCommitHistory(projectRoot, branchName, limit)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return commits, nil
}

func (a *App) IsGitAvailable() bool {
	return a.projectService.IsGitAvailable()
}

// --- AI Methods ---
func (a *App) GenerateCode(systemPrompt, userPrompt string) (string, error) {
	content, err := a.aiService.GenerateCode(a.ctx, systemPrompt, userPrompt)
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return content, nil
}

func (a *App) SuggestContextFiles(task string, allFiles []*domain.FileNode) ([]string, error) {
	files, err := a.contextAnalysis.SuggestFiles(a.ctx, task, allFiles)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return files, nil
}

// --- Settings Methods ---
func (a *App) GetSettings() (domain.SettingsDTO, error) {
	dto, err := a.settingsService.GetSettingsDTO()
	if err != nil {
		a.handleError(err)
		return domain.SettingsDTO{}, err
	}
	return dto, nil
}

func (a *App) SaveSettings(dto domain.SettingsDTO) error {
	err := a.settingsService.SaveSettingsDTO(dto)
	if err != nil {
		a.handleError(err)
	}
	return err
}

func (a *App) RefreshAIModels(provider string, apiKey string) error {
	err := a.settingsService.RefreshModels(provider, apiKey)
	if err != nil {
		a.handleError(err)
		return err
	}
	return nil
}

// --- FS Watcher & System Dialogs ---
func (a *App) StartFileWatcher(rootDirPath string) {
	a.handleError(a.fileWatcher.Start(rootDirPath))
}

func (a *App) StopFileWatcher() {
	a.fileWatcher.Stop()
}

func (a *App) SelectDirectory() (string, error) {
	bridge := wailsbridge.New(a.ctx)
	path, err := bridge.OpenDirectoryDialog()
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return path, nil
}
