import { ref } from 'vue'

export interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}

export class UIStateService {
  private toasts = ref<Toast[]>([])
  private contextMenuState = ref<{
    isOpen: boolean
    x: number
    y: number
    targetPath?: string
  } | null>(null)

  constructor() {}

  // Toast management
  getToasts() {
    return this.toasts.value
  }

  addToast(message: string, type: 'success' | 'error' | 'warning' | 'info', duration: number = 3000) {
    const id = Math.random().toString(36).substr(2, 9)
    const toast: Toast = { id, message, type, duration }
    
    this.toasts.value.push(toast)
    
    // Auto remove toast after duration
    if (duration > 0) {
      setTimeout(() => {
        this.removeToast(id)
      }, duration)
    }
    
    return id
  }

  removeToast(id: string) {
    const index = this.toasts.value.findIndex(toast => toast.id === id)
    if (index !== -1) {
      this.toasts.value.splice(index, 1)
    }
  }

  clearToasts() {
    this.toasts.value = []
  }

  // Context menu management
  getContextMenuState() {
    return this.contextMenuState.value
  }

  openContextMenu(x: number, y: number, targetPath?: string) {
    this.contextMenuState.value = {
      isOpen: true,
      x,
      y,
      targetPath
    }
  }

  closeContextMenu() {
    this.contextMenuState.value = null
  }

  isContextMenuOpen() {
    return this.contextMenuState.value?.isOpen ?? false
  }
}