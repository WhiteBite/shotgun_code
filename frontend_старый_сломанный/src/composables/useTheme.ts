import { ref, computed, watch, onMounted } from 'vue'
import { defineStore } from 'pinia'
import { createStoreWithDependencies, type StoreDependencies } from '@/stores/StoreDependencyContainer'
import type { StorageRepository } from '@/domain/repositories/StorageRepository'
import { APP_CONFIG } from '@/config/app-config'

export type ColorScheme = 'dark' | 'light' | 'auto'
export type AccentColor = 'blue' | 'purple' | 'green' | 'orange' | 'red'

export interface ThemeSettings {
  colorScheme: ColorScheme
  accentColor: AccentColor
  customColors: Record<string, string>
  enableAnimations: boolean
  enableGlassMorphism: boolean
  reducedMotion: boolean
  highContrast: boolean
}

export interface ThemeColors {
  primary: string
  secondary: string
  surface: string
  background: string
  text: string
  accent: string
  success: string
  warning: string
  error: string
  info: string
}

const THEME_STORAGE_KEY = APP_CONFIG.theme.STORAGE_KEY

// Color palettes for different accent colors
const accentColorPalettes: Record<AccentColor, Record<string, string>> = {
  blue: {
    primary: 'hsl(210, 100%, 65%)',
    primaryHover: 'hsl(210, 100%, 60%)',
    primaryLight: 'hsl(210, 100%, 80%)',
    primaryDark: 'hsl(210, 100%, 50%)',
  },
  purple: {
    primary: 'hsl(260, 85%, 70%)',
    primaryHover: 'hsl(260, 85%, 65%)',
    primaryLight: 'hsl(260, 85%, 85%)',
    primaryDark: 'hsl(260, 85%, 55%)',
  },
  green: {
    primary: 'hsl(120, 60%, 55%)',
    primaryHover: 'hsl(120, 60%, 50%)',
    primaryLight: 'hsl(120, 60%, 80%)',
    primaryDark: 'hsl(120, 60%, 45%)',
  },
  orange: {
    primary: 'hsl(35, 90%, 60%)',
    primaryHover: 'hsl(35, 90%, 55%)',
    primaryLight: 'hsl(35, 90%, 80%)',
    primaryDark: 'hsl(35, 90%, 50%)',
  },
  red: {
    primary: 'hsl(0, 75%, 65%)',
    primaryHover: 'hsl(0, 75%, 60%)',
    primaryLight: 'hsl(0, 75%, 85%)',
    primaryDark: 'hsl(0, 75%, 55%)',
  },
}

// Default theme settings
const defaultThemeSettings: ThemeSettings = {
  colorScheme: 'dark',
  accentColor: 'blue',
  customColors: {},
  enableAnimations: true,
  enableGlassMorphism: true,
  reducedMotion: false,
  highContrast: false,
}

export const useThemeStore = defineStore('theme', () => {
  return createStoreWithDependencies('theme', (dependencies: StoreDependencies) => {
    // Inject StorageRepository
    const { localStorageService } = dependencies;

    // State
    const settings = ref<ThemeSettings>({ ...defaultThemeSettings })
    const isInitialized = ref(false)

    // Computed
    const isDark = computed(() => {
      if (settings.value.colorScheme === 'auto') {
        return window.matchMedia('(prefers-color-scheme: dark)').matches
      }
      return settings.value.colorScheme === 'dark'
    })

    const currentAccentColors = computed(() => {
      return accentColorPalettes[settings.value.accentColor]
    })

    const themeColors = computed((): ThemeColors => {
      const base = isDark.value ? {
        primary: 'hsl(222, 15%, 12%)',
        secondary: 'hsl(222, 15%, 16%)',
        surface: 'hsl(222, 15%, 20%)',
        background: 'hsl(222, 15%, 8%)',
        text: 'hsl(220, 15%, 97%)',
      } : {
        primary: 'hsl(0, 0%, 100%)',
        secondary: 'hsl(0, 0%, 98%)',
        surface: 'hsl(0, 0%, 95%)',
        background: 'hsl(0, 0%, 100%)',
        text: 'hsl(222, 15%, 12%)',
      }

      return {
        ...base,
        accent: currentAccentColors.value.primary,
        success: 'hsl(120, 60%, 55%)',
        warning: 'hsl(35, 90%, 60%)',
        error: 'hsl(0, 75%, 65%)',
        info: 'hsl(195, 85%, 65%)',
        ...settings.value.customColors,
      }
    })

    const cssVariables = computed(() => {
      const colors = themeColors.value
      const accent = currentAccentColors.value
      
      const variables: Record<string, string> = {
        // Base colors
        '--surface-primary': colors.primary,
        '--surface-secondary': colors.secondary,
        '--surface-tertiary': colors.surface,
        '--surface-background': colors.background,
        '--text-primary': colors.text,
        
        // Accent colors
        '--accent-primary': accent.primary,
        '--accent-primary-hover': accent.primaryHover,
        '--accent-primary-light': accent.primaryLight,
        '--accent-primary-dark': accent.primaryDark,
        
        // Semantic colors
        '--accent-success': colors.success,
        '--accent-warning': colors.warning,
        '--accent-error': colors.error,
        '--accent-info': colors.info,
      }

      // Add high contrast adjustments
      if (settings.value.highContrast) {
        variables['--border-primary'] = isDark.value ? 'hsl(220, 13%, 40%)' : 'hsl(220, 13%, 20%)'
        variables['--text-secondary'] = isDark.value ? 'hsl(220, 12%, 90%)' : 'hsl(220, 12%, 30%)'
      }

      // Add animation controls
      if (settings.value.reducedMotion || !settings.value.enableAnimations) {
        variables['--transition-fast'] = '0ms'
        variables['--transition-normal'] = '0ms'
        variables['--transition-slow'] = '0ms'
      }

      return variables
    })

    // Actions
    const updateSettings = (newSettings: Partial<ThemeSettings>) => {
      settings.value = { ...settings.value, ...newSettings }
      saveSettings()
      applyTheme()
    }

    const setColorScheme = (scheme: ColorScheme) => {
      updateSettings({ colorScheme: scheme })
    }

    const setAccentColor = (color: AccentColor) => {
      updateSettings({ accentColor: color })
    }

    const setCustomColor = (key: string, value: string) => {
      const customColors = { ...settings.value.customColors }
      customColors[key] = value
      updateSettings({ customColors })
    }

    const removeCustomColor = (key: string) => {
      const customColors = { ...settings.value.customColors }
      delete customColors[key]
      updateSettings({ customColors })
    }

    const toggleAnimations = () => {
      updateSettings({ enableAnimations: !settings.value.enableAnimations })
    }

    const toggleGlassMorphism = () => {
      updateSettings({ enableGlassMorphism: !settings.value.enableGlassMorphism })
    }

    const resetToDefaults = () => {
      settings.value = { ...defaultThemeSettings }
      saveSettings()
      applyTheme()
    }

    const applyTheme = () => {
      const root = document.documentElement
      const vars = cssVariables.value

      // Apply CSS variables
      Object.entries(vars).forEach(([key, value]) => {
        root.style.setProperty(key, value)
      })

      // Apply theme classes
      root.classList.toggle('theme-dark', isDark.value)
      root.classList.toggle('theme-light', !isDark.value)
      root.classList.toggle('theme-high-contrast', settings.value.highContrast)
      root.classList.toggle('theme-reduced-motion', settings.value.reducedMotion)
      root.classList.toggle('theme-no-animations', !settings.value.enableAnimations)
      root.classList.toggle('theme-glass-disabled', !settings.value.enableGlassMorphism)

      // Set color scheme meta tag for browser UI
      const metaThemeColor = document.querySelector('meta[name="theme-color"]')
      if (metaThemeColor) {
        metaThemeColor.setAttribute('content', themeColors.value.primary)
      }
    }

    const saveSettings = async () => {
      try {
        localStorageService.set(THEME_STORAGE_KEY, settings.value)
      } catch (error) {
        console.warn('Failed to save theme settings:', error)
      }
    }

    const loadSettings = async () => {
      try {
        const saved = localStorageService.get<ThemeSettings>(THEME_STORAGE_KEY)
        if (saved) {
          settings.value = { ...defaultThemeSettings, ...saved }
        }

        // Also check for system preferences
        if (settings.value.colorScheme === 'auto') {
          // Listen for system theme changes
          const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
          const handleChange = () => applyTheme()
          mediaQuery.addEventListener('change', handleChange)
        }

        // Check for reduced motion preference
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
        if (prefersReducedMotion && !settings.value.reducedMotion) {
          settings.value.reducedMotion = true
        }

        // Check for high contrast preference
        const prefersHighContrast = window.matchMedia('(prefers-contrast: high)').matches
        if (prefersHighContrast && !settings.value.highContrast) {
          settings.value.highContrast = true
        }

      } catch (error) {
        console.warn('Failed to load theme settings:', error)
        settings.value = { ...defaultThemeSettings }
      }
    }

    const initialize = async () => {
      if (isInitialized.value) return

      await loadSettings()
      applyTheme()
      isInitialized.value = true

      // Watch for settings changes
      watch(
        () => settings.value,
        () => {
          applyTheme()
        },
        { deep: true }
      )
    }

    return {
      // State
      settings: computed(() => settings.value),
      isInitialized: computed(() => isInitialized.value),
      
      // Computed
      isDark,
      themeColors,
      currentAccentColors,
      cssVariables,
      
      // Actions
      updateSettings,
      setColorScheme,
      setAccentColor,
      setCustomColor,
      removeCustomColor,
      toggleAnimations,
      toggleGlassMorphism,
      resetToDefaults,
      initialize,
    }
  })
})

// Composable hook
export function useTheme() {
  const store = useThemeStore()
  
  // Auto-initialize on first use
  onMounted(() => {
    if (!store.isInitialized) {
      store.initialize()
    }
  })
  
  return store
}