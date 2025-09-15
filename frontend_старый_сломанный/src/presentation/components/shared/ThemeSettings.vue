<template>
  <div class="theme-settings p-4 space-y-6">
    <div class="space-y-4">
      <h3 class="text-lg font-semibold text-gray-200">Theme Settings</h3>
      
      <!-- Color Scheme -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-gray-300">Color Scheme</label>
        <div class="flex space-x-2">
          <button
            v-for="scheme in colorSchemes"
            :key="scheme.value"
            class="px-3 py-2 text-sm rounded-lg border transition-all duration-200"
            :class="{
              'bg-blue-600 border-blue-500 text-white': theme.settings.colorScheme === scheme.value,
              'bg-gray-700 border-gray-600 text-gray-300 hover:bg-gray-600': theme.settings.colorScheme !== scheme.value
            }"
            @click="theme.setColorScheme(scheme.value)"
          >
            <component :is="scheme.icon" class="w-4 h-4 mr-1.5 inline" />
            {{ scheme.label }}
          </button>
        </div>
      </div>
      
      <!-- Accent Color -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-gray-300">Accent Color</label>
        <div class="flex space-x-2">
          <button
            v-for="color in accentColors"
            :key="color.value"
            class="w-8 h-8 rounded-lg border-2 transition-transform duration-200"
            :class="{
              'scale-110 ring-2 ring-white ring-opacity-50': theme.settings.accentColor === color.value,
              'hover:scale-105': theme.settings.accentColor !== color.value
            }"
            :style="{ backgroundColor: color.color, borderColor: color.color }"
            :title="color.label"
            @click="theme.setAccentColor(color.value)"
          />
        </div>
      </div>
      
      <!-- Animation Settings -->
      <div class="space-y-3">
        <label class="block text-sm font-medium text-gray-300">Animation & Effects</label>
        
        <div class="space-y-2">
          <label class="flex items-center space-x-3">
            <input
              type="checkbox"
              :checked="theme.settings.enableAnimations"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              @change="theme.toggleAnimations()"
            />
            <span class="text-sm text-gray-300">Enable animations</span>
          </label>
          
          <label class="flex items-center space-x-3">
            <input
              type="checkbox"
              :checked="theme.settings.enableGlassMorphism"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              @change="theme.toggleGlassMorphism()"
            />
            <span class="text-sm text-gray-300">Glass morphism effects</span>
          </label>
        </div>
      </div>
      
      <!-- Accessibility -->
      <div class="space-y-3">
        <label class="block text-sm font-medium text-gray-300">Accessibility</label>
        
        <div class="space-y-2">
          <label class="flex items-center space-x-3">
            <input
              type="checkbox"
              :checked="theme.settings.highContrast"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              @change="toggleHighContrast"
            />
            <span class="text-sm text-gray-300">High contrast mode</span>
          </label>
          
          <label class="flex items-center space-x-3">
            <input
              type="checkbox"
              :checked="theme.settings.reducedMotion"
              class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              @change="toggleReducedMotion"
            />
            <span class="text-sm text-gray-300">Reduced motion (respect system preference)</span>
          </label>
        </div>
      </div>
      
      <!-- Custom Colors -->
      <div class="space-y-3">
        <label class="block text-sm font-medium text-gray-300">Custom Colors</label>
        
        <div class="space-y-2">
          <div
            v-for="(color, key) in theme.settings.customColors"
            :key="key"
            class="flex items-center space-x-2"
          >
            <div
              class="w-6 h-6 rounded border border-gray-600"
              :style="{ backgroundColor: color }"
            />
            <span class="flex-1 text-sm text-gray-300">{{ key }}</span>
            <button
              class="text-red-400 hover:text-red-300 text-sm"
              @click="theme.removeCustomColor(key)"
            >
              Remove
            </button>
          </div>
          
          <div class="flex space-x-2">
            <input
              v-model="newColorKey"
              type="text"
              placeholder="Color name"
              class="flex-1 px-2 py-1 text-sm bg-gray-700 border border-gray-600 rounded text-gray-200"
            />
            <input
              v-model="newColorValue"
              type="color"
              class="w-10 h-8 bg-gray-700 border border-gray-600 rounded cursor-pointer"
            />
            <button
              :disabled="!newColorKey || !newColorValue"
              class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              @click="addCustomColor"
            >
              Add
            </button>
          </div>
        </div>
      </div>
      
      <!-- Preview -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-gray-300">Preview</label>
        <div class="p-4 rounded-lg border border-gray-600 space-y-2" style="background: var(--surface-secondary)">
          <div class="text-sm" style="color: var(--text-primary)">Primary text with current theme</div>
          <div class="text-sm" style="color: var(--text-secondary)">Secondary text example</div>
          <button class="px-3 py-1 text-sm rounded" style="background: var(--accent-primary); color: white;">
            Accent button
          </button>
        </div>
      </div>
      
      <!-- Actions -->
      <div class="flex space-x-2 pt-4 border-t border-gray-700">
        <button
          class="flex-1 px-4 py-2 text-sm bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors"
          @click="theme.resetToDefaults()"
        >
          Reset to Defaults
        </button>
        <button
          class="flex-1 px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          @click="exportTheme"
        >
          Export Theme
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useTheme, type ColorScheme, type AccentColor } from '@/composables/useTheme'
import {
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon
} from '@heroicons/vue/24/outline'

const theme = useTheme()

// Form state
const newColorKey = ref('')
const newColorValue = ref('#3b82f6')

// Options
const colorSchemes = [
  { value: 'light' as ColorScheme, label: 'Light', icon: SunIcon },
  { value: 'dark' as ColorScheme, label: 'Dark', icon: MoonIcon },
  { value: 'auto' as ColorScheme, label: 'Auto', icon: ComputerDesktopIcon },
]

const accentColors = [
  { value: 'blue' as AccentColor, label: 'Blue', color: 'hsl(210, 100%, 65%)' },
  { value: 'purple' as AccentColor, label: 'Purple', color: 'hsl(260, 85%, 70%)' },
  { value: 'green' as AccentColor, label: 'Green', color: 'hsl(120, 60%, 55%)' },
  { value: 'orange' as AccentColor, label: 'Orange', color: 'hsl(35, 90%, 60%)' },
  { value: 'red' as AccentColor, label: 'Red', color: 'hsl(0, 75%, 65%)' },
]

// Methods
const toggleHighContrast = () => {
  theme.updateSettings({ highContrast: !theme.settings.highContrast })
}

const toggleReducedMotion = () => {
  theme.updateSettings({ reducedMotion: !theme.settings.reducedMotion })
}

const addCustomColor = () => {
  if (newColorKey.value && newColorValue.value) {
    theme.setCustomColor(newColorKey.value, newColorValue.value)
    newColorKey.value = ''
    newColorValue.value = '#3b82f6'
  }
}

const exportTheme = () => {
  const themeData = {
    settings: theme.settings,
    timestamp: new Date().toISOString(),
    version: '1.0'
  }
  
  const blob = new Blob([JSON.stringify(themeData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `shotgun-theme-${Date.now()}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
/* Custom scrollbar for the settings panel */
.theme-settings {
  scrollbar-width: thin;
  scrollbar-color: var(--accent-primary) var(--surface-secondary);
}

.theme-settings::-webkit-scrollbar {
  width: 6px;
}

.theme-settings::-webkit-scrollbar-track {
  background: var(--surface-secondary);
  border-radius: 3px;
}

.theme-settings::-webkit-scrollbar-thumb {
  background: var(--accent-primary);
  border-radius: 3px;
}

.theme-settings::-webkit-scrollbar-thumb:hover {
  background: var(--accent-primary-hover);
}

/* Color input styling */
input[type="color"] {
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
  cursor: pointer;
}

input[type="color"]::-webkit-color-swatch {
  border: none;
  border-radius: 4px;
}

input[type="color"]::-moz-color-swatch {
  border: none;
  border-radius: 4px;
}
</style>