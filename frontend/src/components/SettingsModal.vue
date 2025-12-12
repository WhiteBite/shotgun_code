<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="close"></div>

        <!-- Modal Content -->
        <div class="modal-content-enhanced w-full max-w-2xl max-h-[80vh] flex flex-col relative">
          <!-- Header -->
          <div class="flex items-center justify-between p-4 border-b border-gray-700/30">
            <h2 class="text-lg font-semibold text-white">{{ t('settings.modal.title') }}</h2>
            <button
              @click="close"
              class="p-2 rounded-lg hover:bg-gray-700/50 text-gray-400 hover:text-white transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Tabs -->
          <div class="flex border-b border-gray-700/30 px-4">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              class="tab-btn"
              :class="activeTab === tab.id ? 'tab-btn-active-indigo' : 'tab-btn-inactive'"
            >
              <component :is="tab.icon" class="w-4 h-4" />
              {{ tab.label }}
            </button>
          </div>

          <!-- Tab Content -->
          <div class="flex-1 overflow-auto p-4">
            <!-- General Tab -->
            <div v-if="activeTab === 'general'" class="space-y-4">
              <div>
                <label class="text-sm text-gray-300 mb-2 block">{{ t('settings.modal.language') }}</label>
                <select v-model="selectedLanguage" class="input w-full">
                  <option value="ru">Русский</option>
                  <option value="en">English</option>
                </select>
              </div>

              <!-- Onboarding Tour -->
              <div class="pt-4 border-t border-gray-700/30">
                <button
                  @click="handleStartTour"
                  class="btn-unified btn-unified-secondary w-full"
                >
                  <HelpCircle class="w-4 h-4" />
                  {{ t('onboarding.startTour') }}
                </button>
              </div>
            </div>

            <!-- AI Tab - reuse existing AISettings -->
            <div v-if="activeTab === 'ai'">
              <AISettings />
            </div>

            <!-- Export Tab - reuse existing ExportSettings -->
            <div v-if="activeTab === 'export'">
              <ExportSettings />
            </div>

            <!-- File Explorer Tab -->
            <div v-if="activeTab === 'fileExplorer'" class="space-y-4">
              <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer hover:text-white transition-colors">
                <input
                  type="checkbox"
                  v-model="settingsStore.settings.fileExplorer.useGitignore"
                  class="rounded border-gray-600 bg-gray-700 text-indigo-500 focus:ring-indigo-500"
                />
                {{ t('settings.useGitignore') }}
              </label>
              <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer hover:text-white transition-colors">
                <input
                  type="checkbox"
                  v-model="settingsStore.settings.fileExplorer.useCustomIgnore"
                  class="rounded border-gray-600 bg-gray-700 text-indigo-500 focus:ring-indigo-500"
                />
                {{ t('settings.useCustomIgnore') }}
              </label>
              <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer hover:text-white transition-colors">
                <input
                  type="checkbox"
                  v-model="settingsStore.settings.fileExplorer.autoSaveSelection"
                  class="rounded border-gray-600 bg-gray-700 text-indigo-500 focus:ring-indigo-500"
                />
                {{ t('settings.autoSaveSelection') }}
              </label>
              <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer hover:text-white transition-colors">
                <input
                  type="checkbox"
                  v-model="settingsStore.settings.fileExplorer.compactNestedFolders"
                  class="rounded border-gray-600 bg-gray-700 text-indigo-500 focus:ring-indigo-500"
                />
                {{ t('settings.compactFolders') }}
              </label>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex justify-end gap-3 p-4 border-t border-gray-700/30">
            <button @click="close" class="btn btn-ghost">
              {{ t('settings.modal.cancel') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import AISettings from '@/components/workspace/sidebar/AISettings.vue'
import ExportSettings from '@/components/workspace/sidebar/ExportSettings.vue'
import { useI18n } from '@/composables/useI18n'
import { useOnboarding } from '@/composables/useOnboarding'
import { useSettingsStore } from '@/stores/settings.store'
import { FileText, FolderTree, HelpCircle, Lightbulb, Settings, X } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t, setLocale, locale } = useI18n()
const settingsStore = useSettingsStore()
const { startTour, resetTour } = useOnboarding()

const tabs = computed(() => [
  { id: 'general', label: t('settings.modal.general'), icon: Settings },
  { id: 'ai', label: t('settings.modal.ai'), icon: Lightbulb },
  { id: 'export', label: t('settings.modal.export'), icon: FileText },
  { id: 'fileExplorer', label: t('settings.modal.fileExplorer'), icon: FolderTree },
])

const activeTab = ref<string>('general')
const selectedLanguage = ref(locale.value)

// Sync language selection with locale
watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      selectedLanguage.value = locale.value
      activeTab.value = 'general'
    }
  }
)

// Apply language change immediately
watch(selectedLanguage, (newLang) => {
  if (newLang !== locale.value) {
    setLocale(newLang as 'ru' | 'en')
  }
})

function close() {
  emit('update:modelValue', false)
}

function handleStartTour() {
  close()
  resetTour()
  // Small delay to let modal close
  setTimeout(() => {
    startTour()
  }, 300)
}

// Handle Escape key
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue) {
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-content-enhanced,
.modal-leave-active .modal-content-enhanced {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content-enhanced,
.modal-leave-to .modal-content-enhanced {
  transform: scale(0.95) translateY(-10px);
  opacity: 0;
}
</style>
