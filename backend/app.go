package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/wailsbridge"
	"strings"
)

type App struct {
	ctx             context.Context
	projectService  *application.ProjectService
	aiService       *application.AIService
	settingsService *application.SettingsService
	fileWatcher     domain.FileSystemWatcher
	contextAnalysis domain.ContextAnalyzer
	gitRepo         domain.GitRepository
	exportService   *application.ExportService
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) handleError(err error) {
	if err != nil {
		bridge := wailsbridge.New(a.ctx)
		bridge.Error(err.Error())
		bridge.Emit("app:error", err.Error())
	}
}

func (a *App) validateProjectPath(path string) error {
	if path == "" {
		return fmt.Errorf("project path cannot be empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", path)
	}
	return nil
}

func (a *App) ListFiles(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	if err := a.validateProjectPath(dirPath); err != nil {
		a.handleError(err)
		return nil, err
	}
	nodes, err := a.projectService.ListFiles(dirPath, useGitignore, useCustomIgnore)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return nodes, nil
}

func (a *App) ReadFileContent(rootDir, relPath string) (string, error) {
	if err := a.validateProjectPath(rootDir); err != nil {
		return "", err
	}
	cleanRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for root: %w", err)
	}
	rootEval, _ := filepath.EvalSymlinks(cleanRootDir)
	absPath := filepath.Join(cleanRootDir, relPath)
	absEval, _ := filepath.EvalSymlinks(absPath)
	rel, err := filepath.Rel(rootEval, absEval)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path traversal attempt detected: %s", relPath)
	}
	data, err := os.ReadFile(absEval)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", relPath, err)
	}
	return string(data), nil
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	if err := a.validateProjectPath(rootDir); err != nil {
		a.handleError(err)
		return
	}
	go a.projectService.GenerateContext(a.ctx, rootDir, includedPaths)
}

func (a *App) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	if err := a.validateProjectPath(projectRoot); err != nil {
		a.handleError(err)
		return nil, err
	}
	files, err := a.projectService.GetUncommittedFiles(projectRoot)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return files, nil
}

func (a *App) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	if err := a.validateProjectPath(projectRoot); err != nil {
		a.handleError(err)
		return nil, err
	}
	commits, err := a.projectService.GetRichCommitHistory(projectRoot, branchName, limit)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return commits, nil
}

func (a *App) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	if err := a.validateProjectPath(projectRoot); err != nil {
		a.handleError(err)
		return "", err
	}
	content, err := a.gitRepo.GetFileContentAtCommit(projectRoot, filePath, commitHash)
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return content, nil
}

func (a *App) GetGitignoreContent(projectRoot string) (string, error) {
	if err := a.validateProjectPath(projectRoot); err != nil {
		a.handleError(err)
		return "", err
	}
	content, err := a.gitRepo.GetGitignoreContent(projectRoot)
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return content, nil
}

func (a *App) IsGitAvailable() bool { return a.projectService.IsGitAvailable() }

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

func (a *App) StartFileWatcher(rootDirPath string) {
	if err := a.validateProjectPath(rootDirPath); err != nil {
		a.handleError(err)
		return
	}
	a.handleError(a.fileWatcher.Start(rootDirPath))
}

func (a *App) StopFileWatcher() { a.fileWatcher.Stop() }

func (a *App) SelectDirectory() (string, error) {
	bridge := wailsbridge.New(a.ctx)
	path, err := bridge.OpenDirectoryDialog()
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return path, nil
}

// NEW: Export
func (a *App) ExportContext(settingsJSON string) (domain.ExportResult, error) {
	var s domain.ExportSettings
	if err := json.Unmarshal([]byte(settingsJSON), &s); err != nil {
		a.handleError(err)
		return domain.ExportResult{}, err
	}
	res, err := a.exportService.Export(a.ctx, s)
	if err != nil {
		a.handleError(err)
		return domain.ExportResult{}, err
	}
	return res, nil
}
