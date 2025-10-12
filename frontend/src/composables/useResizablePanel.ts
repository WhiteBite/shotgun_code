/**
 * Composable for creating resizable panels
 * Allows users to drag and resize panels with mouse
 * Saves panel width to localStorage for persistence
 */

import { ref, onMounted, onUnmounted } from 'vue'

export interface ResizablePanelOptions {
  minWidth?: number
  maxWidth?: number
  defaultWidth?: number
  storageKey?: string
}

export function useResizablePanel(options: ResizablePanelOptions = {}) {
  const {
    minWidth = 200,
    maxWidth = 800,
    defaultWidth = 300,
    storageKey = 'panel-width'
  } = options

  const panelRef = ref<HTMLElement>()
  const width = ref(defaultWidth)
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
    const newWidth = Math.max(minWidth, Math.min(maxWidth, startWidth + delta))
    width.value = newWidth
  }

  const onMouseUp = () => {
    if (!isResizing.value) return
    
    isResizing.value = false
    
    // Restore default cursor and text selection
    document.body.style.userSelect = ''
    document.body.style.cursor = ''
    
    // Save to localStorage
    try {
      localStorage.setItem(storageKey, String(width.value))
    } catch (err) {
      console.warn('Failed to save panel width to localStorage:', err)
    }
    
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  onMounted(() => {
    // Restore saved width from localStorage
    try {
      const saved = localStorage.getItem(storageKey)
      if (saved) {
        const savedWidth = Number(saved)
        if (!isNaN(savedWidth) && savedWidth >= minWidth && savedWidth <= maxWidth) {
          width.value = savedWidth
        }
      }
    } catch (err) {
      console.warn('Failed to load panel width from localStorage:', err)
    }
  })

  onUnmounted(() => {
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  })

  return {
    panelRef,
    width,
