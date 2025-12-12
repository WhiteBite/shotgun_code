/**
 * Composable for creating resizable panels
 * Allows users to drag and resize panels with mouse
 * Saves panel width to localStorage for persistence
 */

import { useStorage } from '@vueuse/core'
import { onUnmounted, ref } from 'vue'

export interface ResizablePanelOptions {
  minWidth?: number
  maxWidth?: number
  defaultWidth?: number
  storageKey: string // Сделано обязательным
  invertDirection?: boolean // Для правой панели - тянем влево = увеличиваем
}

export function useResizablePanel(options: ResizablePanelOptions) {
  const {
    minWidth = 200,
    maxWidth = 800,
    defaultWidth = 300,
    storageKey,
    invertDirection = false
  } = options

  const panelRef = ref<HTMLElement>()
  // Используем useStorage для автоматического сохранения и загрузки
  const width = useStorage(storageKey, defaultWidth, localStorage, {
    mergeDefaults: true,
    serializer: {
      read: (v: string) => {
        const parsed = Number(v)
        // Проверка, чтобы значение всегда было в допустимых границах
        return !isNaN(parsed) ? Math.max(minWidth, Math.min(maxWidth, parsed)) : defaultWidth
      },
      write: (v: number) => String(v)
    }
  })
  const isResizing = ref(false)

  let startX = 0
  let startWidth = 0

  const onMouseDown = (e: MouseEvent) => {
    if (!panelRef.value) return

    isResizing.value = true
    startX = e.clientX
    startWidth = width.value

    // Prevent text selection during resize
    document.body.style.userSelect = 'none'
    document.body.style.cursor = 'col-resize'

    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
  }

  const onMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return

    const delta = e.clientX - startX
    // Для правой панели инвертируем: тянем влево = увеличиваем ширину
    const adjustedDelta = invertDirection ? -delta : delta
    const newWidth = Math.max(minWidth, Math.min(maxWidth, startWidth + adjustedDelta))
    width.value = newWidth
  }

  const onMouseUp = () => {
    if (!isResizing.value) return

    isResizing.value = false

    // Restore default cursor and text selection
    document.body.style.userSelect = ''
    document.body.style.cursor = ''

    // useStorage автоматически сохраняет изменения - ручное сохранение удалено

    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  onUnmounted(() => {
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  })

  // Reset to default width
  const resetToDefault = () => {
    width.value = defaultWidth
  }

  return {
    panelRef,
    width,
    isResizing,
    onMouseDown,
    resetToDefault,
    defaultWidth
  }
}
