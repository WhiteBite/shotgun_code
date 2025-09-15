<template>
  <div class="tooltip-provider">
    <slot />
    
    <!-- Tooltip Portal -->
    <Teleport to="body">
      <div
        v-for="tooltip in visibleTooltips"
        :key="tooltip.id"
        ref="tooltipElements"
        class="tooltip-container"
        :class="[
          `tooltip-${tooltip.config.size}`,
          `tooltip-${tooltip.computedPosition}`,
          tooltip.config.className
        ]"
        :style="{
          position: 'fixed',
          top: `${tooltip.position.y}px`,
          left: `${tooltip.position.x}px`,
          zIndex: tooltip.config.zIndex,
          maxWidth: `${tooltip.config.maxWidth}px`,
          transform: getTransformStyle(tooltip.computedPosition),
          pointerEvents: tooltip.config.interactive ? 'auto' : 'none'
        }"
        @mouseenter="handleTooltipMouseEnter(tooltip.id)"
        @mouseleave="handleTooltipMouseLeave(tooltip.id)"
      >
        <!-- Tooltip Content -->
        <div
          class="tooltip-content"
          :class="{
            'tooltip-with-arrow': tooltip.config.arrow,
            'tooltip-interactive': tooltip.config.interactive
          }"
        >
          <!-- Arrow -->
          <div
            v-if="tooltip.config.arrow"
            class="tooltip-arrow"
            :class="`tooltip-arrow-${tooltip.computedPosition}`"
          />
          
          <!-- Content -->
          <div class="tooltip-text">
            <div v-if="tooltip.config.allowHTML" v-html="tooltip.content"></div>
            <template v-else>
              {{ tooltip.content }}
            </template>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useTooltip } from '@/composables/useTooltip'

const tooltipStore = useTooltip()

// Refs
const tooltipElements = ref<HTMLElement[]>([])

// Computed
const visibleTooltips = computed(() => tooltipStore.visibleTooltips)

// Methods
const getTransformStyle = (position: string): string => {
  switch (position) {
    case 'top':
      return 'translateX(-50%) translateY(-100%)'
    case 'bottom':
      return 'translateX(-50%)'
    case 'left':
      return 'translateY(-50%) translateX(-100%)'
    case 'right':
      return 'translateY(-50%)'
    default:
      return 'translateX(-50%)'
  }
}

const handleTooltipMouseEnter = (id: string) => {
  const tooltip = tooltipStore.tooltips.get(id)
  if (tooltip?.config.interactive) {
    // Keep tooltip visible when hovering over it
    tooltipStore.showTooltip(id, true)
  }
}

const handleTooltipMouseLeave = (id: string) => {
  const tooltip = tooltipStore.tooltips.get(id)
  if (tooltip?.config.interactive) {
    tooltipStore.hideTooltip(id)
  }
}

// Watch for tooltip changes and update positions
watch(
  visibleTooltips,
  async () => {
    await nextTick()
    // Update positions for any tooltips that might have been repositioned
    visibleTooltips.value.forEach(tooltip => {
      if (tooltip.config.position === 'auto') {
        // Recalculate position if needed
        const rect = tooltip.element.getBoundingClientRect()
        // Position update logic would go here
      }
    })
  },
  { flush: 'post' }
)
</script>

<style scoped>
.tooltip-provider {
  position: relative;
}

.tooltip-container {
  position: fixed;
  z-index: var(--z-tooltip, 1020);
  pointer-events: none;
}

.tooltip-content {
  background: var(--surface-elevated, hsl(222, 15%, 22%));
  border: 1px solid var(--border-secondary, hsl(220, 13%, 26%));
  border-radius: var(--radius-lg, 0.5rem);
  box-shadow: var(--shadow-xl, 0 20px 25px -5px rgba(0, 0, 0, 0.1));
  backdrop-filter: blur(12px);
  position: relative;
  overflow: hidden;
}

.tooltip-text {
  padding: var(--space-2, 0.5rem) var(--space-3, 0.75rem);
  color: var(--text-primary, hsl(220, 15%, 97%));
  font-size: var(--text-sm, 0.875rem);
  line-height: var(--leading-normal, 1.5);
  word-wrap: break-word;
  max-width: 100%;
}

/* Size variants */
.tooltip-sm .tooltip-text {
  padding: var(--space-1, 0.25rem) var(--space-2, 0.5rem);
  font-size: var(--text-xs, 0.75rem);
}

.tooltip-lg .tooltip-text {
  padding: var(--space-3, 0.75rem) var(--space-4, 1rem);
  font-size: var(--text-base, 1rem);
}

/* Interactive tooltips */
.tooltip-interactive {
  pointer-events: auto;
  cursor: default;
}

.tooltip-interactive:hover {
  border-color: var(--border-primary, hsl(220, 13%, 30%));
}

/* Arrow styles */
.tooltip-arrow {
  position: absolute;
  width: 0;
  height: 0;
  border-style: solid;
}

.tooltip-arrow-top {
  bottom: -8px;
  left: 50%;
  transform: translateX(-50%);
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-bottom {
  top: -8px;
  left: 50%;
  transform: translateX(-50%);
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 8px solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-left {
  right: -8px;
  top: 50%;
  transform: translateY(-50%);
  border-top: 8px solid transparent;
  border-bottom: 8px solid transparent;
  border-left: 8px solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-right {
  left: -8px;
  top: 50%;
  transform: translateY(-50%);
  border-top: 8px solid transparent;
  border-bottom: 8px solid transparent;
  border-right: 8px solid var(--surface-elevated, hsl(222, 15%, 22%));
}

/* Animation */
.tooltip-container {
  animation: tooltipFadeIn var(--transition-fast, 150ms) var(--easing-standard, ease);
}

@keyframes tooltipFadeIn {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-100%) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(-100%) scale(1);
  }
}

.tooltip-bottom {
  animation-name: tooltipFadeInBottom;
}

@keyframes tooltipFadeInBottom {
  from {
    opacity: 0;
    transform: translateX(-50%) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) scale(1);
  }
}

.tooltip-left {
  animation-name: tooltipFadeInLeft;
}

@keyframes tooltipFadeInLeft {
  from {
    opacity: 0;
    transform: translateY(-50%) translateX(-100%) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(-50%) translateX(-100%) scale(1);
  }
}

.tooltip-right {
  animation-name: tooltipFadeInRight;
}

@keyframes tooltipFadeInRight {
  from {
    opacity: 0;
    transform: translateY(-50%) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(-50%) scale(1);
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .tooltip-container {
    animation: none;
  }
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .tooltip-content {
    border-width: 2px;
    border-color: var(--text-primary, hsl(220, 15%, 97%));
    background: var(--surface-primary, hsl(222, 15%, 12%));
  }
}
</style>