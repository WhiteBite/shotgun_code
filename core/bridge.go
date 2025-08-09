package core

import "github.com/wailsapp/wails/v2/pkg/runtime"

// RuntimeBridge определяет интерфейс для взаимодействия со средой выполнения Wails.
// Эта абстракция позволяет сервисам использовать функции runtime, не будучи
// напрямую связанными с пакетом Wails, что облегчает тестирование с мок-реализациями.
type RuntimeBridge interface {
	// Environment возвращает информацию об операционной системе и среде сборки.
	// ВАЖНО: Эта функция в Wails не возвращает ошибку.
	Environment() runtime.EnvironmentInfo

	// ClipboardSetText устанавливает текст в системный буфер обмена.
	ClipboardSetText(text string) error

	// EventsEmit отправляет событие на фронтенд с необязательными данными.
	EventsEmit(eventName string, data ...interface{})

	// OpenDirectoryDialog открывает системный диалог для выбора директории.
	OpenDirectoryDialog(options runtime.OpenDialogOptions) (string, error)
}
