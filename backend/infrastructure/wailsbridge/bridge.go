package wailsbridge

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Bridge implements domain.Logger and domain.EventBus using the Wails runtime.
// It acts as a bridge between the backend application logic and the Wails frontend/runtime.
type Bridge struct {
	ctx context.Context
}

// New creates a new Bridge instance.
func New(ctx context.Context) *Bridge {
	return &Bridge{ctx: ctx}
}

// SetWailsContext allows updating the context, which is necessary because the
// Wails context changes during the application lifecycle (e.g., OnStartup vs. OnDomReady).
func (b *Bridge) SetWailsContext(ctx context.Context) {
	b.ctx = ctx
}

// --- domain.Logger implementation ---

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

// --- domain.EventBus implementation ---

func (b *Bridge) Emit(eventName string, data ...interface{}) {
	// MEMORY OPTIMIZATION: Removed all logging from Emit to prevent log accumulation
	// Logs were causing memory leaks by accumulating in Wails runtime
	if len(data) > 0 {
		runtime.EventsEmit(b.ctx, eventName, data[0])
	} else {
		runtime.EventsEmit(b.ctx, eventName)
	}
}

// --- Wails Dialogs ---

// OpenDirectoryDialog opens a native directory selection dialog.
func (b *Bridge) OpenDirectoryDialog() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(b.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	return dir, nil
}
