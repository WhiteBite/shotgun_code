import { ref, computed } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface ContextMenuState {
  isOpen: boolean
  x: number
  y: number
  targetPath?: string
}

export class ContextMenuService {
  private contextMenuState = ref<ContextMenuState | null>(null)

  constructor() {}

  // Get current context menu state
  getContextMenuState() {
    return this.contextMenuState.value
  }

  // Set context menu state
  setContextMenuState(state: ContextMenuState | null) {
    this.contextMenuState.value = state
  }

  // Open context menu
  openContextMenu(x: number, y: number, targetPath?: string) {
    // Clamp position to viewport
    const margin = 8
    const width = 224
    const height = 180
    const left = Math.min(Math.max(x, margin), window.innerWidth - width - margin)
    const top = Math.min(Math.max(y, margin), window.innerHeight - height - margin)
    
    this.contextMenuState.value = {
      isOpen: true,
      x: left,
      y: top,
      targetPath
    }
  }

  // Close context menu
  closeContextMenu() {
    this.contextMenuState.value = null
  }

  // Calculate menu style
  calculateMenuStyle() {
    const cm = this.contextMenuState.value
    if (!cm) return {}
    
    const margin = 8
    const width = 224
    const height = 180
    const left = Math.min(Math.max(cm.x, margin), window.innerWidth - width - margin)
    const top = Math.min(Math.max(cm.y, margin), window.innerHeight - height - margin)
    
    return { left: left + 'px', top: top + 'px' }
  }

  // Computed property for menu style
  getMenuStyle() {
    return computed(() => this.calculateMenuStyle())
  }
}