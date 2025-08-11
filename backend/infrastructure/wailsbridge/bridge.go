package wailsbridge

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Bridge struct {
	ctx context.Context
}

func New(ctx context.Context) *Bridge {
	return &Bridge{ctx: ctx}
}

func (b *Bridge) Debug(message string) {
	runtime.LogDebug(b.ctx, message)
}

func (b *Bridge) Info(message string) {
	runtime.LogInfo(b.ctx, message)
}

func (b *Bridge) Warning(message string) {
	runtime.LogWarning(b.ctx, message)
}

func (b *Bridge) Error(message string) {
	runtime.LogError(b.ctx, message)
}

func (b *Bridge) Fatal(message string) {
	runtime.LogFatal(b.ctx, message)
}

func (b *Bridge) Emit(eventName string, data ...interface{}) {
	runtime.EventsEmit(b.ctx, eventName, data...)
}

func (b *Bridge) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(b.ctx, runtime.OpenDialogOptions{})
}
