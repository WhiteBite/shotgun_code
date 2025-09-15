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
          
          <!-- Content with proper HTML sanitization -->
          <div class="tooltip-text">
            <div v-if="tooltip.config.allowHTML" v-html="sanitizeHTML(tooltip.content)"></div>
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
import { APP_CONFIG } from '@/config/app-config'

// HTML Sanitization function using centralized configuration
function sanitizeHTML(html: string): string {
  // Create a temporary DOM element to parse and sanitize HTML
  const tempDiv = document.createElement('div')
  tempDiv.innerHTML = html
  
  // Remove forbidden elements and attributes using centralized configuration
  const allowedTags = APP_CONFIG.security.sanitization.ALLOWED_HTML_TAGS
  const allowedAttributes = APP_CONFIG.security.sanitization.ALLOWED_ATTRIBUTES
  
  // Helper function to sanitize a single element
  function sanitizeElement(element: Element): Element | null {
    const tagName = element.tagName.toLowerCase()
    
    // Check if tag is allowed
    if (!allowedTags.includes(tagName)) {
      return null // Remove this element
    }
    
    // Remove non-allowed attributes
    const attributes = Array.from(element.attributes)
    attributes.forEach(attr => {
      if (!allowedAttributes.includes(attr.name)) {
        element.removeAttribute(attr.name)
      }
    })
    
    return element
  }
  
  // Recursively sanitize all elements
  function sanitizeRecursive(node: Node): Node | null {
    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element
      const sanitizedElement = sanitizeElement(element)
      
      if (!sanitizedElement) {
        return null // Element was removed
      }
      
      // Sanitize child nodes
      const children = Array.from(sanitizedElement.childNodes)
      children.forEach(child => {
        const sanitizedChild = sanitizeRecursive(child)
        if (!sanitizedChild) {
          sanitizedElement.removeChild(child)
        }
      })
      
      return sanitizedElement
    } else if (node.nodeType === Node.TEXT_NODE) {
      // Text nodes are safe, just return them
      return node
    }
    
    return null // Remove other types of nodes
  }
  
  // Sanitize all child nodes
  const children = Array.from(tempDiv.childNodes)
  children.forEach(child => {
    const sanitizedChild = sanitizeRecursive(child)
    if (!sanitizedChild) {
      tempDiv.removeChild(child)
    }
  })
  
  // Apply additional security policies from centralized configuration
  if (APP_CONFIG.security.sanitization.STRIP_SCRIPTS) {
    // Remove any script-like content
    const scripts = tempDiv.querySelectorAll('script')
    scripts.forEach(script => script.remove())
    
    // Remove event handlers
    const elements = tempDiv.querySelectorAll('*')
    elements.forEach(el => {
      const attributes = Array.from(el.attributes)
      attributes.forEach(attr => {
        if (attr.name.startsWith('on')) {
          el.removeAttribute(attr.name)
        }
      })
    })
  }
  
  return tempDiv.innerHTML
}

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
        // const rect = tooltip.element.getBoundingClientRect()
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
  bottom: v-bind('(-APP_CONFIG.ui.tooltips.ARROW_SIZE) + "px"');
  left: 50%;
  transform: translateX(-50%);
  border-left: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-right: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-top: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-bottom {
  top: v-bind('(-APP_CONFIG.ui.tooltips.ARROW_SIZE) + "px"');
  left: 50%;
  transform: translateX(-50%);
  border-left: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-right: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-bottom: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-left {
  right: v-bind('(-APP_CONFIG.ui.tooltips.ARROW_SIZE) + "px"');
  top: 50%;
  transform: translateY(-50%);
  border-top: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-bottom: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-left: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid var(--surface-elevated, hsl(222, 15%, 22%));
}

.tooltip-arrow-right {
  left: v-bind('(-APP_CONFIG.ui.tooltips.ARROW_SIZE) + "px"');
  top: 50%;
  transform: translateY(-50%);
  border-top: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-bottom: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid transparent;
  border-right: v-bind('APP_CONFIG.ui.tooltips.ARROW_SIZE + "px"') solid var(--surface-elevated, hsl(222, 15%, 22%));
}

/* Animation */
.tooltip-container {
  animation: tooltipFadeIn var(--transition-fast, 150ms) var(--easing-standard, ease);
}

@keyframes tooltipFadeIn {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-100%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_INITIAL'));
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(-100%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_FINAL'));
  }
}

.tooltip-bottom {
  animation-name: tooltipFadeInBottom;
}

@keyframes tooltipFadeInBottom {
  from {
    opacity: 0;
    transform: translateX(-50%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_INITIAL'));
  }
  to {
    opacity: 1;
    transform: translateX(-50%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_FINAL'));
  }
}

.tooltip-left {
  animation-name: tooltipFadeInLeft;
}

@keyframes tooltipFadeInLeft {
  from {
    opacity: 0;
    transform: translateY(-50%) translateX(-100%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_INITIAL'));
  }
  to {
    opacity: 1;
    transform: translateY(-50%) translateX(-100%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_FINAL'));
  }
}

.tooltip-right {
  animation-name: tooltipFadeInRight;
}

@keyframes tooltipFadeInRight {
  from {
    opacity: 0;
    transform: translateY(-50%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_INITIAL'));
  }
  to {
    opacity: 1;
    transform: translateY(-50%) scale(v-bind('APP_CONFIG.ui.tooltips.SCALE_FINAL'));
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
    border-width: v-bind('APP_CONFIG.ui.tooltips.HIGH_CONTRAST_BORDER_WIDTH + "px"');
    border-color: var(--text-primary, hsl(220, 15%, 97%));
    background: var(--surface-primary, hsl(222, 15%, 12%));
  }
}
</style>