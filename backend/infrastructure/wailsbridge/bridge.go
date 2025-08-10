package wailsbridge

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Bridge является конкретной реализацией интерфейсов Logger и EventBus,
// используя рантайм Wails для взаимодействия с фронтендом и ОС.
type Bridge struct {
	ctx context.Context
}

// New создает новый экземпляр Wails Bridge.
func New(ctx context.Context) *Bridge {
	return &Bridge{ctx: ctx}
}

// Debug логирует сообщение с уровнем Debug.
func (b *Bridge) Debug(message string) { runtime.LogDebug(b.ctx, message) }

// Info логирует сообщение с уровнем Info.
func (b *Bridge) Info(message string) { runtime.LogInfo(b.ctx, message) }

// Warning логирует сообщение с уровнем Warning.
func (b *Bridge) Warning(message string) { runtime.LogWarning(b.ctx, message) }

// Error логирует сообщение с уровнем Error.
func (b *Bridge) Error(message string) { runtime.LogError(b.ctx, message) }

// Fatal логирует сообщение с уровнем Fatal и завершает приложение.
func (b *Bridge) Fatal(message string) { runtime.LogFatal(b.ctx, message) }

// Emit отправляет событие на фронтенд.
func (b *Bridge) Emit(eventName string, data ...interface{}) {
	runtime.EventsEmit(b.ctx, eventName, data...)
}

// OpenDirectoryDialog открывает системный диалог выбора директории.
func (b *Bridge) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(b.ctx, runtime.OpenDialogOptions{})
}
