<template>
  <div 
    class="responsive-layout-grid"
    :style="gridStyles"
    :class="gridClasses"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

interface LayoutConfig {
  gridTemplateRows?: string
  gridTemplateColumns?: string
  gridTemplateAreas?: string
  gap?: string | number
}

interface Props {
  layout: LayoutConfig
  gap?: string | number
  responsive?: boolean
  autoFit?: boolean
  minItemWidth?: string
  maxItemWidth?: string
}

const props = withDefaults(defineProps<Props>(), {
  gap: APP_CONFIG.ui.layout.DEFAULT_GAP,
  responsive: true,
  autoFit: false,
  minItemWidth: APP_CONFIG.ui.layout.MIN_ITEM_WIDTH,
  maxItemWidth: '1fr'
})

// Computed
const gridStyles = computed(() => {
  const styles: Record<string, string> = {}
  
  if (props.autoFit) {
    styles.gridTemplateColumns = `repeat(auto-fit, minmax(${props.minItemWidth}, ${props.maxItemWidth}))`
  } else {
    if (props.layout.gridTemplateColumns) {
      styles.gridTemplateColumns = props.layout.gridTemplateColumns
    }
  }
  
  if (props.layout.gridTemplateRows) {
    styles.gridTemplateRows = props.layout.gridTemplateRows
  }
  
  if (props.layout.gridTemplateAreas) {
    styles.gridTemplateAreas = props.layout.gridTemplateAreas
  }
  
  // Handle gap
  const gapValue = typeof props.gap === 'number' ? `${props.gap}px` : props.gap
  styles.gap = gapValue
  
  return styles
})

const gridClasses = computed(() => [
  {
    'responsive-grid': props.responsive,
    'auto-fit-grid': props.autoFit
  }
])
</script>

<style scoped>
.responsive-layout-grid {
  display: grid;
  width: 100%;
  height: 100%;
}

.responsive-grid {
  /* Responsive breakpoints using centralized config */
  @media (max-width: 768px) {
    grid-template-columns: 1fr !important;
    grid-template-rows: auto !important;
    grid-template-areas: none !important;
  }
  
  @media (max-width: 1024px) {
    gap: 8px;
  }
  
  @media (max-width: 480px) {
    gap: 4px;
  }
}

.auto-fit-grid {
  align-items: start;
}

/* Grid area utilities */
.responsive-layout-grid :deep([data-grid-area]) {
  grid-area: attr(data-grid-area);
}

/* Common responsive patterns */
@media (max-width: 768px) {
  .responsive-layout-grid :deep(.desktop-only) {
    display: none;
  }
}

@media (min-width: 769px) {
  .responsive-layout-grid :deep(.mobile-only) {
    display: none;
  }
}

/* Flex fallback for older browsers */
@supports not (display: grid) {
  .responsive-layout-grid {
    display: flex;
    flex-direction: column;
  }
  
  .responsive-layout-grid > * {
    flex: 1;
    margin-bottom: 1rem;
  }
  
  .responsive-layout-grid > *:last-child {
    margin-bottom: 0;
  }
}
</style>