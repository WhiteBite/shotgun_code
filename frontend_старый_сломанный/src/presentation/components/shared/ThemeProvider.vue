<template>
  <div class="theme-provider">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { onMounted, provide, watch } from 'vue'
import { useTheme } from '@/composables/useTheme'

// Props
interface Props {
  autoInitialize?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  autoInitialize: true
})

// Setup theme
const theme = useTheme()

// Provide theme to child components
provide('theme', theme)

// Initialize theme
onMounted(async () => {
  if (props.autoInitialize) {
    await theme.initialize()
  }
})

// Watch for theme changes and emit events
watch(
  () => theme.settings,
  (newSettings) => {
    // Emit custom events for theme changes
    window.dispatchEvent(new CustomEvent('theme-changed', {
      detail: { settings: newSettings }
    }))
  },
  { deep: true }
)

// Expose theme methods for parent components
defineExpose({
  theme,
  setColorScheme: theme.setColorScheme,
  setAccentColor: theme.setAccentColor,
  toggleAnimations: theme.toggleAnimations,
  resetToDefaults: theme.resetToDefaults
})
</script>

<style scoped>
.theme-provider {
  height: 100%;
  width: 100%;
}

/* Ensure smooth transitions when theme changes */
.theme-provider * {
  transition-property: color, background-color, border-color, box-shadow;
  transition-duration: var(--transition-fast, 150ms);
  transition-timing-function: var(--easing-standard, ease);
}

/* Disable transitions during theme initialization */
.theme-provider.theme-initializing * {
  transition: none !important;
}
</style>