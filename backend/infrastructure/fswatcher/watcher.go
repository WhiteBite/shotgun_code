package fswatcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Watcher struct {
	log       domain.Logger
	bus       domain.EventBus
	fsWatcher *fsnotify.Watcher
	mu        sync.Mutex
	cancel    context.CancelFunc
	rootDir   string
	appCtx    context.Context
}

func New(ctx context.Context, bus domain.EventBus) (*Watcher, error) {
	return &Watcher{
		appCtx: ctx,
		log:    wailsLogger{ctx: ctx},
		bus:    bus,
	}, nil
}

func (w *Watcher) shouldSkipDir(name string) bool {
	// Общий набор шумных директорий
	switch name {
	case ".git", "node_modules", ".idea", "dist", "build", ".cache", ".vite", ".wails", "out", "target", "bin", "obj", "coverage":
		return true
	default:
		return false
	}
}

func (w *Watcher) Start(path string) error {
	w.Stop()

	w.mu.Lock()
	defer w.mu.Unlock()
	w.rootDir = path

	var err error
	w.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	// Рекурсивно подписываемся на директории, пропуская шумные
	err = filepath.WalkDir(w.rootDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if w.shouldSkipDir(d.Name()) {
				w.log.Info("Watcher: skip dir " + p)
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(p)
		}
		return nil
	})
	if err != nil {
		return err
	}

	go w.run(ctx)
	w.log.Info("Наблюдатель запущен для: " + path)
	return nil
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	if w.fsWatcher != nil {
		w.fsWatcher.Close()
		w.fsWatcher = nil
		w.log.Info("Наблюдатель остановлен.")
	}
}

func (w *Watcher) run(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		if w.fsWatcher != nil {
			w.fsWatcher.Close()
			w.fsWatcher = nil
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			// Логируем только дебагом, чтобы не засорять
			w.log.Debug(fmt.Sprintf("FS event: %s", event.String()))
			w.bus.Emit("projectFilesChanged", w.rootDir)
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			w.log.Error(fmt.Sprintf("Ошибка наблюдателя: %v", err))
		}
	}
}

func (w *Watcher) RefreshAndRescan() error {
	w.mu.Lock()
	path := w.rootDir
	w.mu.Unlock()
	if path == "" {
		return nil
	}
	w.log.Info("Обновление наблюдателя из-за смены правил...")
	return w.Start(path)
}

type wailsLogger struct {
	ctx context.Context
}

func (l wailsLogger) Debug(message string)   { runtime.LogDebug(l.ctx, message) }
func (l wailsLogger) Info(message string)    { runtime.LogInfo(l.ctx, message) }
func (l wailsLogger) Warning(message string) { runtime.LogWarning(l.ctx, message) }
func (l wailsLogger) Error(message string)   { runtime.LogError(l.ctx, message) }
func (l wailsLogger) Fatal(message string)   { runtime.LogFatal(l.ctx, message) }
