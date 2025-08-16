package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/application"
	"shotgun_code/cmd/app"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/wailsbridge"
	"strings"
)

type App struct {
	ctx             context.Context
	projectService  *application.ProjectService
	aiService       *application.AIService
	settingsService *application.SettingsService
	contextAnalysis domain.ContextAnalyzer
	fileWatcher     domain.FileSystemWatcher
	gitRepo         domain.GitRepository
	exportService   *application.ExportService
	fileReader      domain.FileContentReader
	bridge          *wailsbridge.Bridge
}

func (a *App) startup(ctx context.Context, container *app.AppContainer) {
	a.ctx = ctx
	a.projectService = container.ProjectService
	a.aiService = container.AIService
	a.settingsService = container.SettingsService
	a.contextAnalysis = container.ContextAnalysis
	a.fileWatcher = container.Watcher
	a.gitRepo = container.GitRepo
	a.exportService = container.ExportService
	a.fileReader = container.FileReader
	a.bridge = container.Bridge
}

func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx
	a.bridge.SetWailsContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
	a.fileWatcher.Stop()
}

func (a *App) SelectDirectory() (string, error) {
	return a.bridge.OpenDirectoryDialog()
}

func (a *App) ListFiles(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	return a.projectService.ListFiles(dirPath, useGitignore, useCustomIgnore)
}

func (a *App) RequestShotgunContextGeneration(rootDir string, includedPaths []string) {
	a.projectService.GenerateContext(a.ctx, rootDir, includedPaths)
}

func (a *App) GenerateCode(systemPrompt, userPrompt string) (string, error) {
	return a.aiService.GenerateCode(a.ctx, systemPrompt, userPrompt)
}

func (a *App) SuggestContextFiles(task string, allFiles []*domain.FileNode) ([]string, error) {
	return a.contextAnalysis.SuggestFiles(a.ctx, task, allFiles)
}

func (a *App) GetSettings() (domain.SettingsDTO, error) {
	return a.settingsService.GetSettingsDTO()
}

func (a *App) SaveSettings(settingsJson string) error {
	var dto domain.SettingsDTO
	err := json.Unmarshal([]byte(settingsJson), &dto)
	if err != nil {
		return fmt.Errorf("failed to parse settings JSON: %w", err)
	}
	return a.settingsService.SaveSettingsDTO(dto)
}

func (a *App) RefreshAIModels(provider, apiKey string) error {
	return a.settingsService.RefreshModels(provider, apiKey)
}

func (a *App) StartFileWatcher(rootDirPath string) error {
	return a.fileWatcher.Start(rootDirPath)
}

func (a *App) StopFileWatcher() {
	a.fileWatcher.Stop()
}

func (a *App) IsGitAvailable() bool {
	return a.gitRepo.IsGitAvailable()
}

func (a *App) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	return a.gitRepo.GetUncommittedFiles(projectRoot)
}

func (a *App) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	return a.gitRepo.GetRichCommitHistory(projectRoot, branchName, limit)
}

func (a *App) GetFileContentAtCommit(projectRoot, filePath, commitHash string) (string, error) {
	return a.gitRepo.GetFileContentAtCommit(projectRoot, filePath, commitHash)
}

func (a *App) GetGitignoreContent(projectRoot string) (string, error) {
	return a.gitRepo.GetGitignoreContent(projectRoot)
}

func (a *App) ReadFileContent(rootDir, relPath string) (string, error) {
	contents, err := a.fileReader.ReadContents(a.ctx, []string{relPath}, rootDir, nil)
	if err != nil {
		return "", err
	}
	if content, ok := contents[relPath]; ok {
		return content, nil
	}
	return "", fmt.Errorf("file not found or could not be read: %s", relPath)
}

func (a *App) ExportContext(settingsJson string) (domain.ExportResult, error) {
	var settings domain.ExportSettings
	if err := json.Unmarshal([]byte(settingsJson), &settings); err != nil {
		return domain.ExportResult{}, fmt.Errorf("failed to parse export settings: %w", err)
	}

	result, err := a.exportService.Export(a.ctx, settings)
	if err != nil {
		return domain.ExportResult{}, err
	}

	// Если результат содержит FilePath, то это большой файл - нужно переместить в Downloads
	if result.FilePath != "" {
		// Для больших файлов возвращаем путь как есть, фронтенд сам решит что делать
		// В будущем можно добавить логику перемещения в Downloads
		return result, nil
	}

	return result, nil
}

// CleanupTempFiles - утилита для очистки временных файлов экспорта
func (a *App) CleanupTempFiles(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Проверяем что это действительно temp файл
	if !strings.Contains(filePath, "shotgun-export-") {
		return fmt.Errorf("not a temp export file")
	}

	// Удаляем весь temp каталог
	tempDir := filepath.Dir(filePath)
	return os.RemoveAll(tempDir)
}
