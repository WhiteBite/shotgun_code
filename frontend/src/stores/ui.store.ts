import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'info' | 'warning'
  duration?: number
}

export const useUIStore = defineStore('ui', () => {
  // State
  const toasts = ref<Toast[]>([])
  const nextToastId = ref(0)

  // Actions
  function addToast(message: string, type: Toast['type'] = 'info', duration = 3000) {
    const toast: Toast = {
      id: `toast-${nextToastId.value++}`,
      message,
      type,
      duration
    }

    // Log toast to console
    const logMessage = `[Toast ${type.toUpperCase()}] ${message}`
    switch (type) {
      case 'error':
        console.error(logMessage)
        break
      case 'warning':
        console.warn(logMessage)
        break
      case 'info':
        console.info(logMessage)
        break
      case 'success':
        console.log(logMessage)
        break
    }

    toasts.value.push(toast)

    // Auto-remove after duration
    if (duration > 0) {
      setTimeout(() => {
        removeToast(toast.id)
      }, duration)
    }

    return toast.id
  }

  function removeToast(id: string) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index !== -1) {
      toasts.value.splice(index, 1)
    }
  }

  function clearToasts() {
    toasts.value = []
  }

  return {
    // State
    toasts,
    // Actions
    addToast,
    removeToast,
    clearToasts
  }
})
