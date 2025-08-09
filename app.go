package main

import (
	"context"
	_ "embed"
	"fmt"

	"shotgun_code/services"
	"shotgun_code/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed ignore.glob
var defaultCustomIgnoreRulesContent string

const defaultCustomPromptRulesContent = "no additional rules"

// wailsBridgeImpl предоставляет конкретную реализацию интерфейсов core,
// выступая мостом к реальному Wails runtime.
type wailsBridgeImpl struct {
	ctx context.Context
}

// Реализация core.Logger
func (b *wailsBridgeImpl) Debug(message string)   { runtime.LogDebug(b.ctx, message) }
func (b *wailsBridgeImpl) Info(message string)    { runtime.LogInfo(b.ctx, message) }
func (b *wailsBridgeImpl) Warning(message string) { runtime.LogWarning(b.ctx, message) }
func (b *wailsBridgeImpl) Error(message string)   { runtime.LogError(b.ctx, message) }
func (b *wailsBridgeImpl) Fatal(message string)   { runtime.LogFatal(b.ctx, message) }

// Реализация core.RuntimeBridge
func (b *wailsBridgeImpl) Environment() runtime.EnvironmentInfo {
	// ИСПРАВЛЕНО: runtime.Environment возвращает только одно значение.
	return runtime.Environment(b.ctx)
}
func (b *wailsBridgeImpl) ClipboardSetText(text string) error {
	return runtime.ClipboardSetText(b.ctx, text)
}
func (b *wailsBridgeImpl) EventsEmit(eventName string, data ...interface{}) {
	runtime.EventsEmit(b.ctx, eventName, data...)
}
func (b *wailsBridgeImpl) OpenDirectoryDialog(options runtime.OpenDialogOptions) (string, error) {
	return runtime.OpenDirectoryDialog(b.ctx, options)
}

// App - основная структура приложения, которая оркестрирует все компоненты.
type App struct {
	ctx         context.Context
	settingsMgr *settings.Manager
	projectSvc  *services.Service
}

func NewApp() *App {
	return &App{}
}

// startup - метод жизненного цикла Wails, вызываемый при старте.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	bridge := &wailsBridgeImpl{ctx: ctx}

	var err error
	a.settingsMgr, err = settings.NewManager(bridge, defaultCustomIgnoreRulesContent, defaultCustomPromptRulesContent)
	if err != nil {
		bridge.Fatal(fmt.Sprintf("Критическая ошибка: не удалось инициализировать менеджер настроек: %v", err))
	}

	a.projectSvc = services.NewProjectService(bridge, bridge, a.settingsMgr)
}

// handleError - централизованный обработчик ошибок для методов, связанных с Wails.
func (a *App) handleError(err error) {
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		runtime.EventsEmit(a.ctx, "app:error", err.Error())
	}
}

// --- Методы, биндящиеся в Wails ---
// Это публичные методы, доступные из фронтенда. Они являются тонкими
// обертками вокруг методов сервисов и менеджера.

func (a *App) SelectDirectory() (string, error) {
	bridge := &wailsBridgeImpl{ctx: a.ctx}
	path, err := bridge.OpenDirectoryDialog(runtime.OpenDialogOptions{})
	if err != nil {
		a.handleError(err)
		return "", err
	}
	return path, nil
}

func (a *App) ListFiles(dirPath string) ([]*services.FileNode, error) {
	nodes, err := a.projectSvc.ListFiles(dirPath)
	if err != nil {
		a.handleError(err)
		return nil, err // Важно возвращать ошибку, чтобы фронтенд знал о проблеме
	}
	return nodes, nil
}

func (a *App) SplitShotgunDiff(gitDiffText string, approxLineLimit int) ([]string, error) {
	splits, err := a.projectSvc.SplitShotgunDiff(gitDiffText, approxLineLimit)
	if err != nil {
		a.handleError(err)
		return nil, err
	}
	return splits, nil
}

func (a *App) RequestShotgunContextGeneration(rootDir string, excludedPaths []string) {
	a.projectSvc.RequestShotgunContextGeneration(rootDir, excludedPaths)
}

func (a *App) StartFileWatcher(rootDirPath string) {
	a.handleError(a.projectSvc.StartFileWatcher(rootDirPath))
}

func (a *App) StopFileWatcher() {
	a.projectSvc.StopFileWatcher()
}

func (a *App) GetCustomIgnoreRules() string {
	return a.settingsMgr.GetCustomIgnoreRules()
}

func (a *App) SetCustomIgnoreRules(rules string) {
	a.handleError(a.settingsMgr.SetCustomIgnoreRules(rules))
}

func (a *App) GetCustomPromptRules() string {
	return a.settingsMgr.GetCustomPromptRules()
}

func (a *App) SetCustomPromptRules(rules string) {
	a.handleError(a.settingsMgr.SetCustomPromptRules(rules))
}

func (a *App) SetUseGitignore(enabled bool) {
	a.handleError(a.projectSvc.SetUseGitignore(enabled))
}

func (a *App) SetUseCustomIgnore(enabled bool) {
	a.handleError(a.projectSvc.SetUseCustomIgnore(enabled))
}
