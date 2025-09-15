import { ref, computed } from 'vue'
import type { PanelPosition, PanelState, DockPosition, ResizeHandle } from '@/presentation/components/layout/EnhancedPanel.vue'
import { APP_CONFIG } from '@/config/app-config'

export class EnhancedPanelService {
  private isCollapsed = ref(false)
  private isDragging = ref(false)
  private isResizing = ref(false)
  private resizeHandle = ref<ResizeHandle>()
  private showDockIndicator = ref(false)
  private dockPreview = ref<DockPosition>()

  constructor() {}

  // State management
  getIsCollapsed() {
    return this.isCollapsed.value
  }

  setIsCollapsed(value: boolean) {
    this.isCollapsed.value = value
  }

  toggleCollapse() {
    this.isCollapsed.value = !this.isCollapsed.value
  }

  getIsDragging() {
    return this.isDragging.value
  }

  setIsDragging(value: boolean) {
    this.isDragging.value = value
  }

  getIsResizing() {
    return this.isResizing.value
  }

  setIsResizing(value: boolean) {
    this.isResizing.value = value
  }

  getResizeHandle() {
    return this.resizeHandle.value
  }

  setResizeHandle(value: ResizeHandle | undefined) {
    this.resizeHandle.value = value
  }

  getShowDockIndicator() {
    return this.showDockIndicator.value
  }

  setShowDockIndicator(value: boolean) {
    this.showDockIndicator.value = value
  }

  getDockPreview() {
    return this.dockPreview.value
  }

  setDockPreview(value: DockPosition | undefined) {
    this.dockPreview.value = value
  }

  // Panel state calculations
  getPanelClasses(props: any, isMaximized: boolean, isMinimized: boolean, isDocked: boolean) {
    return [
      'enhanced-panel',
      `panel-${props.variant}`,
      `panel-${props.size}`,
      `panel-${props.state}`,
      {
        'dragging': this.isDragging.value,
        'resizing': this.isResizing.value,
        'collapsed': this.isCollapsed.value,
        'loading': props.loading,
        'has-error': props.error
      },
      props.className
    ]
  }

  getHeaderClasses(props: any, isMaximized: boolean) {
    return [
      'panel-header',
      `header-${props.variant}`,
      {
        'draggable': props.draggable && !isMaximized,
        'collapsed': this.isCollapsed.value
      },
      props.headerClass
    ]
  }

  getContentClasses(props: any) {
    return [
      'panel-content',
      `content-${props.variant}`,
      {
        'scrollable': !props.loading && !props.error
      },
      props.contentClass
    ]
  }

  getPanelStyles(props: any, isMaximized: boolean) {
    const styles: Record<string, string | number> = {
      zIndex: props.zIndex
    }

    if (props.modelValue && props.state === 'normal') {
      styles.transform = `translate(${props.modelValue.x}px, ${props.modelValue.y}px)`
      styles.width = `${props.modelValue.width}px`
      styles.height = `${props.modelValue.height}px`
    }

    return styles
  }

  getContentStyles() {
    if (this.isCollapsed.value) {
      return { display: 'none' }
    }
    return {}
  }

  getActiveResizeHandles(props: any, isMaximized: boolean): ResizeHandle[] {
    if (!props.resizable || isMaximized) return []
    return ['n', 's', 'e', 'w', 'ne', 'nw', 'se', 'sw']
  }

  // Size calculations
  calculateNewSize(
    startPos: PanelPosition,
    deltaX: number,
    deltaY: number,
    handle: ResizeHandle,
    props: any
  ): PanelPosition {
    let { x, y, width, height } = startPos

    switch (handle) {
      case 'n':
        y += deltaY
        height -= deltaY
        break
      case 's':
        height += deltaY
        break
      case 'e':
        width += deltaX
        break
      case 'w':
        x += deltaX
        width -= deltaX
        break
      case 'ne':
        y += deltaY
        height -= deltaY
        width += deltaX
        break
      case 'nw':
        x += deltaX
        y += deltaY
        width -= deltaX
        height -= deltaY
        break
      case 'se':
        width += deltaX
        height += deltaY
        break
      case 'sw':
        x += deltaX
        width -= deltaX
        height += deltaY
        break
    }

    // Apply constraints
    width = Math.max(props.minWidth, Math.min(width, props.maxWidth || Infinity))
    height = Math.max(props.minHeight, Math.min(height, props.maxHeight || Infinity))

    // Maintain aspect ratio if specified
    if (props.aspectRatio) {
      const currentRatio = width / height
      if (Math.abs(currentRatio - props.aspectRatio) > 0.01) {
        if (handle.includes('e') || handle.includes('w')) {
          height = width / props.aspectRatio
        } else {
          width = height * props.aspectRatio
        }
      }
    }

    return { x, y, width, height }
  }

  // Dock preview functionality
  updateDockPreview(mouseX: number, mouseY: number, windowWidth: number, windowHeight: number) {
    const margin = APP_CONFIG.ui.panels.RESIZE_INDICATOR_HEIGHT

    this.showDockIndicator.value = true

    if (mouseX < margin) {
      this.dockPreview.value = 'left'
    } else if (mouseX > windowWidth - margin) {
      this.dockPreview.value = 'right'
    } else if (mouseY < margin) {
      this.dockPreview.value = 'top'
    } else if (mouseY > windowHeight - margin) {
      this.dockPreview.value = 'bottom'
    } else {
      this.showDockIndicator.value = false
      this.dockPreview.value = undefined
    }
  }
}