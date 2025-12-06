<template>
  <Teleport to="body">
    <Transition name="modal-backdrop">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 flex items-center justify-center p-4"
        @click.self="close"
      >
        <Transition name="modal">
          <div
            class="bg-gray-800/95 backdrop-blur-md rounded-2xl border border-gray-700/50 shadow-2xl w-full max-w-3xl max-h-[80vh] flex flex-col"
            @click.stop
          >
            <!-- Header -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-700/50">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-purple-500/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                  </svg>
                </div>
                <div>
                  <h3 class="text-lg font-semibold text-white">{{ t('ignoreModal.title') }}</h3>
                  <p class="text-xs text-gray-400">{{ t('ignoreModal.subtitle') }}</p>
                </div>
              </div>
              <button
                @click="close"
                class="p-2 hover:bg-gray-700/50 rounded-xl transition-colors"
              >
                <svg class="w-5 h-5 text-gray-400 hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Tabs -->
            <div class="flex gap-1 p-2 border-b border-gray-700/50">
              <button
                @click="currentTab = 'gitignore'"
                :class="['tab-btn', currentTab === 'gitignore' ? 'tab-btn-active tab-btn-active-purple' : 'tab-btn-inactive']"
              >
                .gitignore
              </button>
              <button
                @click="currentTab = 'custom'"
                :class="['tab-btn', currentTab === 'custom' ? 'tab-btn-active tab-btn-active-purple' : 'tab-btn-inactive']"
              >
                {{ t('ignoreModal.customRules') }}
              </button>
            </div>

            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-6">
              <!-- .gitignore Tab -->
              <div v-if="currentTab === 'gitignore'" class="space-y-4">
                <div class="p-4 bg-indigo-500/10 border border-indigo-500/30 rounded-xl">
                  <p class="text-sm text-indigo-300">
                    {{ t('ignoreModal.gitignoreInfo') }}
                  </p>
                </div>

                <textarea
                  v-model="gitignoreContent"
                  readonly
                  class="w-full h-64 px-4 py-3 bg-gray-800/50 border border-gray-700/50 rounded-xl text-white font-mono text-sm resize-none focus:outline-none"
                  :placeholder="t('ignoreModal.gitignorePlaceholder')"
                ></textarea>
              </div>

              <!-- Custom Rules Tab -->
              <div v-if="currentTab === 'custom'" class="space-y-4">
                <div class="flex items-center justify-between">
                  <div class="text-sm text-gray-400">
                    {{ t('ignoreModal.rulesCount') }}: <span class="text-white font-semibold">{{ rulesCount }}</span>
                  </div>
                  <button
                    @click="testRules"
                    :disabled="isTesting"
                    class="btn btn-primary btn-sm flex items-center gap-2"
                  >
                    <svg v-if="!isTesting" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                    <svg v-else class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                    {{ t('ignoreModal.testRules') }}
                  </button>
                </div>

                <textarea
                  v-model="customRules"
                  class="w-full h-64 px-4 py-3 bg-gray-800/50 border border-gray-700/50 rounded-xl text-white font-mono text-sm resize-none focus:outline-none focus:border-indigo-500/50 transition-colors"
                  :placeholder="t('ignoreModal.customPlaceholder')"
                ></textarea>

                <!-- Preview -->
                <div v-if="previewFiles && previewFiles.length > 0" class="p-4 bg-gray-800/50 rounded-xl border border-gray-700/30">
                  <p class="text-sm font-semibold text-gray-300 mb-2">
                    {{ t('ignoreModal.preview') }}: {{ previewFiles.length }} {{ t('ignoreModal.filesWillBeIgnored') }}
                  </p>
                  <div class="max-h-32 overflow-y-auto space-y-1">
                    <div
                      v-for="file in previewFiles.slice(0, 20)"
                      :key="file"
                      class="text-xs text-gray-400 font-mono"
                    >
                      {{ file }}
                    </div>
                    <div v-if="previewFiles.length > 20" class="text-xs text-gray-500">
                      ... {{ t('ignoreModal.andMore').replace('{count}', (previewFiles.length - 20).toString()) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="flex items-center justify-between px-6 py-4 border-t border-gray-700/50">
              <div class="flex gap-2">
                <button v-if="currentTab === 'custom'" @click="resetToDefaults" class="btn btn-ghost btn-sm">
                  {{ t('ignoreModal.reset') }}
                </button>
                <button v-if="currentTab === 'custom'" @click="clearAll" class="btn btn-ghost btn-sm">
                  {{ t('ignoreModal.clearAll') }}
                </button>
              </div>
              <div class="flex gap-3">
                <button @click="close" class="btn btn-secondary btn-sm">
                  {{ t('ignoreModal.cancel') }}
                </button>
                <button
                  v-if="currentTab === 'custom'"
                  @click="save"
                  :disabled="isSaving"
                  class="btn btn-primary btn-sm flex items-center gap-2"
                >
                  <svg v-if="!isSaving" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg v-else class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                  {{ t('ignoreModal.save') }}
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { useSettingsStore } from '@/stores/settings.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, ref } from 'vue'
import { parseIgnoreRules } from '../lib/file-utils'

const { t } = useI18n()
const settingsStore = useSettingsStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()

const isOpen = ref(false)
const currentTab = ref<'gitignore' | 'custom'>('custom')
const gitignoreContent = ref('')
const customRules = ref('')
const previewFiles = ref<string[]>([])
const isTesting = ref(false)
const isSaving = ref(false)

const rulesCount = computed(() => parseIgnoreRules(customRules.value).length)

async function open() {
  isOpen.value = true
  
  // Load custom rules from settings
  customRules.value = settingsStore.getCustomIgnoreRules()
  
  // Load .gitignore content from backend
  if (projectStore.currentPath) {
    try {
      gitignoreContent.value = await apiService.getGitignoreContent(projectStore.currentPath)
    } catch (err) {
      console.warn('Failed to load .gitignore:', err)
      gitignoreContent.value = '# No .gitignore file found or failed to load'
    }
  }
}

function close() {
  isOpen.value = false
  previewFiles.value = []
}

async function testRules() {
  if (!projectStore.currentPath) {
    uiStore.addToast('No project selected', 'warning')
    return
  }

  isTesting.value = true
  previewFiles.value = []
  try {
    const ignoredFiles = await apiService.testIgnoreRules(projectStore.currentPath, customRules.value)
    previewFiles.value = ignoredFiles || []
    uiStore.addToast(`Found ${previewFiles.value.length} files that will be ignored`, 'success')
  } catch (err) {
    console.error('Failed to test rules:', err)
    previewFiles.value = []
    uiStore.addToast('Failed to test ignore rules', 'error')
  } finally {
    isTesting.value = false
  }
}

async function save() {
  isSaving.value = true
  try {
    await apiService.updateCustomIgnoreRules(customRules.value)
    settingsStore.setCustomIgnoreRules(customRules.value)
    uiStore.addToast('Ignore rules saved successfully', 'success')
    close()
  } catch (err) {
    console.error('Failed to save rules:', err)
    uiStore.addToast('Failed to save ignore rules', 'error')
  } finally {
    isSaving.value = false
  }
}

function resetToDefaults() {
  customRules.value = ''
  previewFiles.value = []
}

function clearAll() {
  customRules.value = ''
  previewFiles.value = []
}

defineExpose({ open, close })
</script>
