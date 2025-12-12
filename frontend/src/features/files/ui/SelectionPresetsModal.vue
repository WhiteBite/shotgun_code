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
            v-if="isOpen"
            class="bg-gray-900 rounded-xl border border-gray-700 shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col"
            @click.stop
          >
            <!-- Header -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-700">
              <div class="flex items-center gap-3">
                <svg class="w-6 h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                </svg>
                <div>
                  <h3 class="text-lg font-semibold text-white">{{ t('presets.title') }}</h3>
                  <p class="text-xs text-gray-400">{{ t('presets.subtitle') }}</p>
                </div>
              </div>
              <button
                @click="close"
                class="p-2 hover:bg-gray-800 rounded transition-colors"
              >
                <svg class="w-5 h-5 text-gray-400 hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-6">
              <!-- Save New Preset -->
              <div class="mb-6 p-4 bg-gray-800/50 rounded-lg border border-gray-700">
                <h4 class="text-sm font-semibold text-white mb-3">{{ t('presets.saveNew') }}</h4>
                <div class="space-y-3">
                  <input
                    v-model="newPresetName"
                    type="text"
                    :placeholder="t('presets.name')"
                    class="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                  />
                  <input
                    v-model="newPresetDescription"
                    type="text"
                    :placeholder="t('presets.description')"
                    class="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                  />
                  <button
                    @click="handleSavePreset"
                    :disabled="!newPresetName || !fileStore.hasSelectedFiles"
                    class="btn btn-primary w-full"
                  >
                    {{ t('presets.save').replace('{count}', fileStore.selectedCount.toString()) }}
                  </button>
                </div>
              </div>

              <!-- Presets List -->
              <div class="space-y-3">
                <h4 class="text-sm font-semibold text-gray-300 mb-3">{{ t('presets.saved') }}</h4>
                
                <div v-if="projectPresets.length === 0" class="text-center py-12 text-gray-500">
                  <svg class="w-16 h-16 mx-auto mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                  <p class="text-sm">{{ t('presets.noPresets') }}</p>
                </div>

                <div
                  v-for="preset in projectPresets"
                  :key="preset.id"
                  class="p-4 bg-gray-800/50 rounded-lg border border-gray-700 hover:border-gray-600 transition-all"
                >
                  <!-- Edit Mode -->
                  <div v-if="editingPresetId === preset.id" class="space-y-3">
                    <input
                      v-model="editedName"
                      type="text"
                      class="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                    />
                    <input
                      v-model="editedDescription"
                      type="text"
                      :placeholder="t('presets.description')"
                      class="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                    />
                    <div class="flex gap-2">
                      <button @click="saveEditedPreset" class="btn btn-primary btn-xs flex-1">
                        {{ t('presets.saveEdit') }}
                      </button>
                      <button @click="cancelEdit" class="btn btn-secondary btn-xs flex-1">
                        {{ t('presets.cancelEdit') }}
                      </button>
                    </div>
                  </div>

                  <!-- View Mode -->
                  <div v-else>
                    <div class="flex items-start justify-between mb-2">
                      <div class="flex-1">
                        <div class="flex items-center gap-2">
                          <h5 class="text-sm font-semibold text-white">{{ preset.name }}</h5>
                        </div>
                        <p v-if="preset.description" class="text-xs text-gray-400 mt-1">{{ preset.description }}</p>
                        <div class="flex items-center gap-4 mt-2 text-xs text-gray-500">
                          <span class="flex items-center gap-1">
                            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                            </svg>
                            {{ preset.paths.length }} {{ t('presets.files') }}
                          </span>
                          <span>{{ formatDate(preset.createdAt) }}</span>
                        </div>
                      </div>
                      <div class="flex items-center gap-2">
                        <button
                          @click="togglePresetExpansion(preset.id)"
                          class="p-1.5 hover:bg-gray-700 rounded transition-colors"
                          :title="expandedPresetIds.includes(preset.id) ? t('presets.hideFiles') : t('presets.showFiles')"
                        >
                          <svg 
                            class="w-4 h-4 text-gray-400 transition-transform"
                            :class="{ 'rotate-180': expandedPresetIds.includes(preset.id) }"
                            fill="none" 
                            stroke="currentColor" 
                            viewBox="0 0 24 24"
                          >
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                          </svg>
                        </button>
                        <button
                          @click="startEditPreset(preset)"
                          class="p-1.5 hover:bg-gray-700 rounded transition-colors"
                          :title="t('presets.edit')"
                        >
                          <svg class="w-4 h-4 text-gray-400 hover:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button @click="handleLoadPreset(preset.id)" class="btn btn-primary btn-xs">
                          {{ t('presets.load') }}
                        </button>
                        <button @click="confirmDelete(preset.id)" class="btn btn-danger btn-xs">
                          {{ t('presets.delete') }}
                        </button>
                      </div>
                    </div>

                    <!-- Expanded Files List -->
                    <Transition name="expand">
                      <div v-if="expandedPresetIds.includes(preset.id)" class="mt-3 pt-3 border-t border-gray-700">
                        <p class="text-xs text-gray-400 mb-2">{{ t('presets.filesInPreset') }}</p>
                        <div class="max-h-48 overflow-y-auto space-y-1">
                          <div
                            v-for="(path, index) in preset.paths"
                            :key="index"
                            class="flex items-center gap-2 px-2 py-1 bg-gray-800/50 rounded text-xs text-gray-300 hover:bg-gray-800 transition-colors"
                          >
                            <svg class="w-3.5 h-3.5 text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            <span class="truncate">{{ getRelativePath(path) }}</span>
                          </div>
                        </div>
                      </div>
                    </Transition>
                  </div>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="flex items-center justify-end px-6 py-4 border-t border-gray-700">
              <button
                @click="close"
                class="px-4 py-2 text-sm bg-gray-800 hover:bg-gray-700 text-white rounded transition-colors"
              >
                {{ t('presets.close') }}
              </button>
            </div>
          </div>
        </Transition>

        <!-- Delete Confirmation Modal -->
        <Transition name="modal">
          <div
            v-if="presetToDelete"
            class="fixed inset-0 z-50 flex items-center justify-center p-4"
            @click.self="presetToDelete = null"
          >
            <div class="bg-gray-800 rounded-lg border border-gray-700 shadow-2xl w-full max-w-md p-6" @click.stop>
              <h4 class="text-lg font-semibold text-white mb-2">{{ t('presets.confirmDelete') }}</h4>
              <p class="text-sm text-gray-400 mb-4">
                {{ t('presets.confirmDeleteMessage').replace('{name}', getPresetById(presetToDelete)?.name || '') }}
              </p>
              <div class="flex gap-3">
                <button
                  @click="presetToDelete = null"
                  class="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded transition-colors"
                >
                  {{ t('presets.cancel') }}
                </button>
                <button
                  @click="handleDeletePreset"
                  class="flex-1 px-4 py-2 bg-red-600 hover:bg-red-500 text-white text-sm rounded transition-colors"
                >
                  {{ t('presets.delete') }}
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
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { ref } from 'vue'
import { useSelectionPresets, type SelectionPreset } from '../composables/useSelectionPresets'
import { getRelativePath as getRelativePathUtil } from '../lib/file-utils'
import { useFileStore } from '../model/file.store'

const { t } = useI18n()
const fileStore = useFileStore()
const projectStore = useProjectStore()
const uiStore = useUIStore()
const { projectPresets, savePreset, loadPreset, deletePreset, updatePreset, getPresetById } = useSelectionPresets()

const isOpen = ref(false)
const newPresetName = ref('')
const newPresetDescription = ref('')
const expandedPresetIds = ref<string[]>([])
const editingPresetId = ref<string | null>(null)
const editedName = ref('')
const editedDescription = ref('')
const presetToDelete = ref<string | null>(null)

function open() {
  isOpen.value = true
  newPresetName.value = ''
  newPresetDescription.value = ''
  expandedPresetIds.value = []
  editingPresetId.value = null
  presetToDelete.value = null
}

function close() {
  isOpen.value = false
}

function handleSavePreset() {
  const preset = savePreset(newPresetName.value, newPresetDescription.value)
  if (preset) {
    uiStore.addToast(t('presets.saveSuccess').replace('{name}', preset.name), 'success')
    newPresetName.value = ''
    newPresetDescription.value = ''
  }
}

function handleLoadPreset(presetId: string) {
  if (loadPreset(presetId)) {
    uiStore.addToast(t('presets.loaded'), 'success')
    close()
  } else {
    uiStore.addToast('Failed to load preset', 'error')
  }
}

function confirmDelete(presetId: string) {
  presetToDelete.value = presetId
}

function handleDeletePreset() {
  if (presetToDelete.value) {
    deletePreset(presetToDelete.value)
    uiStore.addToast(t('presets.deleted'), 'success')
    presetToDelete.value = null
  }
}

function togglePresetExpansion(presetId: string) {
  if (expandedPresetIds.value.includes(presetId)) {
    expandedPresetIds.value = expandedPresetIds.value.filter(id => id !== presetId)
  } else {
    expandedPresetIds.value = [...expandedPresetIds.value, presetId]
  }
}

function startEditPreset(preset: SelectionPreset) {
  editingPresetId.value = preset.id
  editedName.value = preset.name
  editedDescription.value = preset.description || ''
}

function saveEditedPreset() {
  if (editingPresetId.value) {
    const success = updatePreset(editingPresetId.value, {
      name: editedName.value,
      description: editedDescription.value
    })
    
    if (success) {
      uiStore.addToast(t('presets.editSuccess'), 'success')
      editingPresetId.value = null
    } else {
      uiStore.addToast(t('presets.editFailed'), 'error')
    }
  }
}

function cancelEdit() {
  editingPresetId.value = null
  editedName.value = ''
  editedDescription.value = ''
}

function getRelativePath(absolutePath: string): string {
  if (!projectStore.currentPath) return absolutePath
  return getRelativePathUtil(absolutePath, projectStore.currentPath)
}

function formatDate(timestamp: number): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  if (diffMins < 10080) return `${Math.floor(diffMins / 1440)}d ago`
  return date.toLocaleDateString()
}

defineExpose({ open, close })
</script>

<style scoped>
.modal-backdrop-enter-active,
.modal-backdrop-leave-active {
  transition: opacity 0.2s ease;
}

.modal-backdrop-enter-from,
.modal-backdrop-leave-to {
  opacity: 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 300px;
}
</style>
