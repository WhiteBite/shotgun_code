import { ref, computed, reactive, watch, nextTick } from 'vue'
import { useWindowSize, useLocalStorage } from '@vueuse/core'
import { defineStore } from 'pinia'

// Types
export interface PanelConfig {
  id: string
  title: string
  component: string
  defaultPosition: PanelPosition
  minSize: { width: number; height: number }
  maxSize?: { width: number; height: number }
  resizable: boolean
  draggable: boolean
  closable: boolean
  collapsible: boolean
  persistent: boolean
  group?: string
  priority: number
}

export interface PanelPosition {
  x: number
  y: number
  width: number
  height: number
}

export interface PanelState {
  id: string
  position: PanelPosition
  state: 'normal' | 'minimized' | 'maximized' | 'docked'
  dockPosition?: 'left' | 'right' | 'top' | 'bottom'
  zIndex: number
  visible: boolean
  collapsed: boolean
}

export interface LayoutPreset {
  id: string
  name: string
  description: string
  panels: Record<string, Partial<PanelState>>
  viewport: { width: number; height: number }
}

export interface LayoutConstraints {
  minPanelSize: { width: number; height: number }
  maxPanelSize: { width: number; height: number }
  snapDistance: number
  gridSize: number
  margins: { top: number; right: number; bottom: number; left: number }
}

// Store definition
export const useLayoutManagerStore = defineStore('layoutManager', () => {
  // Window dimensions
  const { width: windowWidth, height: windowHeight } = useWindowSize()
  
  // State
  const panels = ref<Map<string, PanelState>>(new Map())
  const panelConfigs = ref<Map<string, PanelConfig>>(new Map())
  const currentPreset = ref<string>('default')
  const presets = useLocalStorage<Record<string, LayoutPreset>>('shotgun-layout-presets', {})
  const nextZIndex = ref(1000)
  
  // Layout constraints
  const constraints = ref<LayoutConstraints>({
    minPanelSize: { width: 200, height: 150 },
    maxPanelSize: { width: 1200, height: 800 },
    snapDistance: 15,
    gridSize: 10,
    margins: { top: 60, right: 20, bottom: 40, left: 20 }
  })
  
  // Grid and snapping
  const gridEnabled = ref(true)
  const snapEnabled = ref(true)
  const magneticEdges = ref(true)
  
  // Collision detection
  const collisionDetection = ref(true)
  
  // Computed
  const visiblePanels = computed(() => {
    return Array.from(panels.value.values()).filter(panel => panel.visible)
  })
  
  const dockedPanels = computed(() => {
    return visiblePanels.value.filter(panel => panel.state === 'docked')
  })
  
  const floatingPanels = computed(() => {
    return visiblePanels.value.filter(panel => panel.state === 'normal')
  })
  
  const maximizedPanel = computed(() => {
    return visiblePanels.value.find(panel => panel.state === 'maximized')
  })
  
  const availableSpace = computed(() => {
    const docked = dockedPanels.value
    let left = constraints.value.margins.left
    let right = constraints.value.margins.right
    let top = constraints.value.margins.top
    let bottom = constraints.value.margins.bottom
    
    docked.forEach(panel => {
      switch (panel.dockPosition) {
        case 'left':
          left += panel.position.width
          break
        case 'right':
          right += panel.position.width
          break
        case 'top':
          top += panel.position.height
          break
        case 'bottom':
          bottom += panel.position.height
          break
      }
    })
    
    return {
      x: left,
      y: top,
      width: windowWidth.value - left - right,
      height: windowHeight.value - top - bottom
    }
  })

  // Actions
  const registerPanel = (config: PanelConfig): void => {
    panelConfigs.value.set(config.id, config)
    
    if (!panels.value.has(config.id)) {
      const position = calculateInitialPosition(config)
      const state: PanelState = {
        id: config.id,
        position,
        state: 'normal',
        zIndex: nextZIndex.value++,
        visible: !config.persistent, // Persistent panels start hidden
        collapsed: false
      }
      
      panels.value.set(config.id, state)
    }
  }
  
  const unregisterPanel = (panelId: string): void => {
    panels.value.delete(panelId)
    panelConfigs.value.delete(panelId)
  }
  
  const showPanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.visible = true
      bringToFront(panelId)
      
      // Ensure panel is within viewport
      ensurePanelInViewport(panelId)
    }
  }
  
  const hidePanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.visible = false
    }
  }
  
  const togglePanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      if (panel.visible) {
        hidePanel(panelId)
      } else {
        showPanel(panelId)
      }
    }
  }
  
  const bringToFront = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.zIndex = nextZIndex.value++
    }
  }
  
  const updatePanelPosition = (panelId: string, position: Partial<PanelPosition>): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      const newPosition = { ...panel.position, ...position }
      
      // Apply constraints
      const constrainedPosition = applyConstraints(panelId, newPosition)
      
      // Check for collisions
      if (collisionDetection.value) {
        const adjustedPosition = resolveCollisions(panelId, constrainedPosition)
        panel.position = adjustedPosition
      } else {
        panel.position = constrainedPosition
      }
      
      // Trigger save
      saveCurrentLayout()
    }
  }
  
  const updatePanelState = (panelId: string, state: Partial<PanelState>): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      Object.assign(panel, state)
      saveCurrentLayout()
    }
  }
  
  const maximizePanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      // Hide other maximized panels
      panels.value.forEach(p => {
        if (p.id !== panelId && p.state === 'maximized') {
          p.state = 'normal'
        }
      })
      
      panel.state = 'maximized'
      panel.zIndex = nextZIndex.value++
    }
  }
  
  const restorePanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.state = 'normal'
    }
  }
  
  const dockPanel = (panelId: string, position: 'left' | 'right' | 'top' | 'bottom'): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.state = 'docked'
      panel.dockPosition = position
      
      // Calculate docked size and position
      const dockedPosition = calculateDockedPosition(position)
      panel.position = dockedPosition
    }
  }
  
  const undockPanel = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.state = 'normal'
      delete panel.dockPosition
      
      // Position in available space
      ensurePanelInViewport(panelId)
    }
  }
  
  // Smart positioning algorithms
  const calculateInitialPosition = (config: PanelConfig): PanelPosition => {
    const space = availableSpace.value
    const existing = visiblePanels.value
    
    let position = { ...config.defaultPosition }
    
    // If default position is occupied, find a better spot
    if (isPositionOccupied(position, config.id)) {
      position = findOptimalPosition(config.minSize, existing)
    }
    
    // Ensure within viewport
    return constrainToViewport(position)
  }
  
  const findOptimalPosition = (
    size: { width: number; height: number },
    existingPanels: PanelState[]
  ): PanelPosition => {
    const space = availableSpace.value
    const step = 30 // Offset step for positioning
    
    // Try positions in a spiral pattern
    for (let offset = 0; offset < 300; offset += step) {
      const candidates = [
        { x: space.x + offset, y: space.y + offset },
        { x: space.x + space.width - size.width - offset, y: space.y + offset },
        { x: space.x + offset, y: space.y + space.height - size.height - offset },
        { 
          x: space.x + space.width - size.width - offset, 
          y: space.y + space.height - size.height - offset 
        }
      ]
      
      for (const candidate of candidates) {
        const position = { ...candidate, ...size }
        if (!isPositionOccupied(position)) {
          return position
        }
      }
    }
    
    // Fallback to center
    return {
      x: space.x + (space.width - size.width) / 2,
      y: space.y + (space.height - size.height) / 2,
      width: size.width,
      height: size.height
    }
  }
  
  const isPositionOccupied = (position: PanelPosition, excludeId?: string): boolean => {
    return visiblePanels.value.some(panel => {
      if (panel.id === excludeId || panel.state !== 'normal') return false
      return isOverlapping(position, panel.position)
    })
  }
  
  const isOverlapping = (a: PanelPosition, b: PanelPosition): boolean => {
    return !(
      a.x + a.width <= b.x ||
      b.x + b.width <= a.x ||
      a.y + a.height <= b.y ||
      b.y + b.height <= a.y
    )
  }
  
  const applyConstraints = (panelId: string, position: PanelPosition): PanelPosition => {
    const config = panelConfigs.value.get(panelId)
    if (!config) return position
    
    let { x, y, width, height } = position
    
    // Size constraints
    width = Math.max(config.minSize.width, width)
    height = Math.max(config.minSize.height, height)
    
    if (config.maxSize) {
      width = Math.min(config.maxSize.width, width)
      height = Math.min(config.maxSize.height, height)
    }
    
    // Grid snapping
    if (gridEnabled.value) {
      x = Math.round(x / constraints.value.gridSize) * constraints.value.gridSize
      y = Math.round(y / constraints.value.gridSize) * constraints.value.gridSize
    }
    
    // Viewport constraints
    return constrainToViewport({ x, y, width, height })
  }
  
  const constrainToViewport = (position: PanelPosition): PanelPosition => {
    const space = availableSpace.value
    let { x, y, width, height } = position
    
    // Constrain to available space
    x = Math.max(space.x, Math.min(x, space.x + space.width - width))
    y = Math.max(space.y, Math.min(y, space.y + space.height - height))
    
    return { x, y, width, height }
  }
  
  const ensurePanelInViewport = (panelId: string): void => {
    const panel = panels.value.get(panelId)
    if (panel) {
      panel.position = constrainToViewport(panel.position)
    }
  }
  
  const resolveCollisions = (panelId: string, position: PanelPosition): PanelPosition => {
    const overlapping = visiblePanels.value.filter(panel => {
      if (panel.id === panelId || panel.state !== 'normal') return false
      return isOverlapping(position, panel.position)
    })
    
    if (overlapping.length === 0) return position
    
    // Try to push overlapping panels
    const offset = 20
    let attempts = 0
    let newPosition = { ...position }
    
    while (overlapping.length > 0 && attempts < 10) {
      newPosition.x += offset
      newPosition.y += offset
      
      // Check if new position still overlaps
      const stillOverlapping = overlapping.filter(panel => 
        isOverlapping(newPosition, panel.position)
      )
      
      if (stillOverlapping.length === 0) break
      attempts++
    }
    
    return constrainToViewport(newPosition)
  }
  
  const calculateDockedPosition = (dock: 'left' | 'right' | 'top' | 'bottom'): PanelPosition => {
    const margins = constraints.value.margins
    
    switch (dock) {
      case 'left':
        return {
          x: margins.left,
          y: margins.top,
          width: 300,
          height: windowHeight.value - margins.top - margins.bottom
        }
      case 'right':
        return {
          x: windowWidth.value - 300 - margins.right,
          y: margins.top,
          width: 300,
          height: windowHeight.value - margins.top - margins.bottom
        }
      case 'top':
        return {
          x: margins.left,
          y: margins.top,
          width: windowWidth.value - margins.left - margins.right,
          height: 200
        }
      case 'bottom':
        return {
          x: margins.left,
          y: windowHeight.value - 200 - margins.bottom,
          width: windowWidth.value - margins.left - margins.right,
          height: 200
        }
    }
  }
  
  // Layout presets
  const savePreset = (name: string, description: string = ''): void => {
    const preset: LayoutPreset = {
      id: name,
      name,
      description,
      panels: {},
      viewport: {
        width: windowWidth.value,
        height: windowHeight.value
      }
    }
    
    panels.value.forEach(panel => {
      preset.panels[panel.id] = {
        position: { ...panel.position },
        state: panel.state,
        dockPosition: panel.dockPosition,
        visible: panel.visible,
        collapsed: panel.collapsed
      }
    })
    
    presets.value[name] = preset
  }
  
  const loadPreset = (presetId: string): void => {
    const preset = presets.value[presetId]
    if (!preset) return
    
    panels.value.forEach(panel => {
      const presetPanel = preset.panels[panel.id]
      if (presetPanel) {
        Object.assign(panel, presetPanel)
      }
    })
    
    currentPreset.value = presetId
  }
  
  const deletePreset = (presetId: string): void => {
    delete presets.value[presetId]
  }
  
  const saveCurrentLayout = (): void => {
    // Auto-save current layout
    savePreset('current', 'Auto-saved current layout')
  }
  
  // Layout algorithms
  const cascadeWindows = (): void => {
    const floating = floatingPanels.value
    const offset = 30
    let currentOffset = 0
    
    floating.forEach(panel => {
      panel.position.x = availableSpace.value.x + currentOffset
      panel.position.y = availableSpace.value.y + currentOffset
      currentOffset += offset
    })
  }
  
  const tileWindows = (arrangement: 'horizontal' | 'vertical' | 'grid' = 'grid'): void => {
    const floating = floatingPanels.value
    if (floating.length === 0) return
    
    const space = availableSpace.value
    
    switch (arrangement) {
      case 'horizontal':
        tileHorizontally(floating, space)
        break
      case 'vertical':
        tileVertically(floating, space)
        break
      case 'grid':
        tileInGrid(floating, space)
        break
    }
  }
  
  const tileHorizontally = (panels: PanelState[], space: any): void => {
    const width = space.width / panels.length
    panels.forEach((panel, index) => {
      panel.position = {
        x: space.x + index * width,
        y: space.y,
        width: width,
        height: space.height
      }
    })
  }
  
  const tileVertically = (panels: PanelState[], space: any): void => {
    const height = space.height / panels.length
    panels.forEach((panel, index) => {
      panel.position = {
        x: space.x,
        y: space.y + index * height,
        width: space.width,
        height: height
      }
    })
  }
  
  const tileInGrid = (panels: PanelState[], space: any): void => {
    const count = panels.length
    const cols = Math.ceil(Math.sqrt(count))
    const rows = Math.ceil(count / cols)
    const cellWidth = space.width / cols
    const cellHeight = space.height / rows
    
    panels.forEach((panel, index) => {
      const col = index % cols
      const row = Math.floor(index / cols)
      
      panel.position = {
        x: space.x + col * cellWidth,
        y: space.y + row * cellHeight,
        width: cellWidth,
        height: cellHeight
      }
    })
  }
  
  const minimizeAll = (): void => {
    panels.value.forEach(panel => {
      if (panel.visible && panel.state === 'normal') {
        panel.state = 'minimized'
      }
    })
  }
  
  const restoreAll = (): void => {
    panels.value.forEach(panel => {
      if (panel.state === 'minimized') {
        panel.state = 'normal'
      }
    })
  }
  
  // Window resize handler
  watch([windowWidth, windowHeight], () => {
    // Adjust panel positions on window resize
    panels.value.forEach(panel => {
      if (panel.state === 'docked' && panel.dockPosition) {
        panel.position = calculateDockedPosition(panel.dockPosition)
      } else if (panel.state === 'normal') {
        ensurePanelInViewport(panel.id)
      }
    })
  })
  
  // Auto-save on changes
  watch(panels, () => {
    nextTick(() => {
      saveCurrentLayout()
    })
  }, { deep: true })
  
  return {
    // State
    panels: computed(() => panels.value),
    panelConfigs: computed(() => panelConfigs.value),
    visiblePanels,
    dockedPanels,
    floatingPanels,
    maximizedPanel,
    availableSpace,
    constraints: computed(() => constraints.value),
    presets: computed(() => presets.value),
    currentPreset: computed(() => currentPreset.value),
    
    // Settings
    gridEnabled: computed(() => gridEnabled.value),
    snapEnabled: computed(() => snapEnabled.value),
    magneticEdges: computed(() => magneticEdges.value),
    collisionDetection: computed(() => collisionDetection.value),
    
    // Actions
    registerPanel,
    unregisterPanel,
    showPanel,
    hidePanel,
    togglePanel,
    bringToFront,
    updatePanelPosition,
    updatePanelState,
    maximizePanel,
    restorePanel,
    dockPanel,
    undockPanel,
    
    // Layout management
    cascadeWindows,
    tileWindows,
    minimizeAll,
    restoreAll,
    
    // Presets
    savePreset,
    loadPreset,
    deletePreset,
    
    // Utilities
    ensurePanelInViewport,
    isPositionOccupied,
    findOptimalPosition
  }
})

// Composable hook
export function useLayoutManager() {
  return useLayoutManagerStore()
}