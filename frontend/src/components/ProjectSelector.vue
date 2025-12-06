<template>
  <div
    class="relative flex items-center justify-center min-h-screen bg-gradient-to-br from-gray-900 via-gray-850 to-gray-900 text-white overflow-hidden"
    :class="{ 'drag-over': isDragging }"
    @drop.prevent="handleDrop"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
  >
    <!-- Background decoration -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-80 h-80 bg-indigo-500/10 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500/10 rounded-full blur-3xl"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-blue-500/5 rounded-full blur-3xl"></div>
    </div>

    <!-- Language Switcher (Top Right) -->
    <div class="absolute top-4 right-4 z-10">
      <button
        @click="toggleLanguage"
        class="px-3 py-2 bg-gray-800/80 hover:bg-gray-700 text-white text-sm rounded-lg transition-all duration-200 border border-gray-700/50 flex items-center gap-2 backdrop-blur-sm hover:scale-105"
        :title="locale === 'ru' ? 'Switch to English' : 'Переключить на русский'"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
        </svg>
        <span class="font-medium">{{ locale === 'ru' ? 'RU' : 'EN' }}</span>
      </button>
    </div>

    <!-- Drag & Drop Overlay -->
    <Transition name="fade">
      <div
        v-if="isDragging"
        class="fixed inset-0 bg-indigo-600/20 border-4 border-dashed border-indigo-500 rounded-xl flex items-center justify-center z-50 m-8 pointer-events-none backdrop-blur-sm"
      >
        <div class="text-center">
          <svg class="h-20 w-20 text-indigo-400 mx-auto mb-4 animate-bounce" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <p class="text-3xl font-bold text-white drop-shadow-lg">{{ t('welcome.dropHere') }}</p>
          <p class="text-lg text-indigo-200 mt-2">{{ t('welcome.toOpenProject') }}</p>
        </div>
      </div>
    </Transition>

    <div class="relative w-full max-w-2xl px-4 py-8">
      <!-- Logo & Title -->
      <div class="text-center mb-10">
        <div class="inline-flex items-center justify-center w-20 h-20 mb-6 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl shadow-lg shadow-indigo-500/25 logo-spin">
          <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
        </div>
        <h1 class="text-5xl font-bold mb-3 bg-gradient-to-r from-white via-gray-100 to-gray-300 bg-clip-text text-transparent blur-in">
          {{ t('welcome.title') }}
        </h1>
        <p class="text-gray-400 text-lg mb-2 fade-in" style="animation-delay: 0.2s">{{ t('welcome.subtitle') }}</p>
        <p class="text-sm text-gray-500 fade-in" style="animation-delay: 0.3s">{{ t('welcome.dragDrop') }}</p>
      </div>
      
      <!-- Open Project Button -->
      <div class="flex justify-center mb-10 fade-in" style="animation-delay: 0.4s">
        <button @click="selectProject" class="action-btn-hero !px-8 !py-4 !text-base icon-bounce">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          {{ t('welcome.openProject') }}
        </button>
      </div>

      <!-- Recent Projects -->
      <div class="mb-8">
        <h2 class="text-lg font-semibold mb-4 text-gray-300 flex items-center gap-2">
          <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ t('welcome.recentProjects') }}
        </h2>
        <div class="space-y-2">
          <template v-if="recentProjects.length > 0">
            <button
              v-for="(project, index) in recentProjects"
              :key="project.path"
              @click="openRecentProject(project.path)"
              class="list-item list-item-animate group w-full text-left card-float"
              :style="{ animationDelay: `${index * 50}ms` }"
            >
              <div class="flex items-center gap-3">
                <div class="section-icon section-icon-indigo group-hover:bg-indigo-500/30 transition-colors">
                  <svg class="!w-5 !h-5 group-hover:text-indigo-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                  </svg>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="font-medium text-white group-hover:text-indigo-100 transition-colors">{{ project.name }}</div>
                  <div class="text-sm text-gray-500 truncate">{{ project.path }}</div>
                </div>
                <svg class="w-5 h-5 text-gray-600 group-hover:text-indigo-400 transition-colors opacity-0 group-hover:opacity-100" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </button>
          </template>
          <template v-else>
            <div class="empty-state py-10 bg-gray-800/30 rounded-xl border border-gray-700/30">
              <div class="empty-state-icon !w-12 !h-12 mb-3">
                <svg class="!w-6 !h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
              </div>
              <p class="empty-state-text">{{ t('welcome.noRecentProjects') }}</p>
            </div>
          </template>
        </div>
      </div>

      <!-- Settings -->
      <div class="bg-gray-800/40 border border-gray-700/50 rounded-xl p-5 backdrop-blur-sm">
        <h3 class="font-semibold text-gray-300 mb-4 flex items-center gap-2">
          <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          {{ t('welcome.settings') }}
        </h3>
        <label class="flex items-center gap-3 text-sm text-gray-300 select-none cursor-pointer group">
          <div class="relative">
            <input
              type="checkbox"
              class="sr-only peer"
              :checked="projectStore.autoOpenLast"
              @change="onToggleAutoOpen"
            />
            <div class="w-10 h-6 bg-gray-700 rounded-full peer-checked:bg-indigo-600 transition-colors"></div>
            <div class="absolute left-1 top-1 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
          </div>
          <span class="group-hover:text-white transition-colors">{{ t('welcome.autoOpen') }}</span>
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n';
import { apiService } from '@/services/api.service';
import { computed, onMounted, ref } from 'vue';
import { useProjectStore } from '../stores/project.store';
import { useUIStore } from '../stores/ui.store';

const emit = defineEmits<{
  (e: 'opened', path: string): void
}>()

const projectStore = useProjectStore()
const uiStore = useUIStore()
const { t, locale, setLocale } = useI18n()
const isDragging = ref(false)

const recentProjects = computed(() => projectStore.recentProjects)

// Принудительно загружаем список недавних проектов при монтировании компонента
onMounted(async () => {
  try {
    await projectStore.fetchRecentProjects()
    console.log('Recent projects loaded from backend:', recentProjects.value.length)
  } catch (error) {
    console.error('Failed to load recent projects:', error)
  }
})

function toggleLanguage() {
  const newLocale = locale.value === 'ru' ? 'en' : 'ru'
  setLocale(newLocale)
  uiStore.addToast(
    newLocale === 'ru' ? 'Язык изменён на русский' : 'Language changed to English',
    'success'
  )
}

function onToggleAutoOpen(e: Event) {
  const target = e.target as HTMLInputElement
  projectStore.setAutoOpenLast(target.checked)
}

// Drag & Drop handlers
async function handleDrop(e: DragEvent) {
  isDragging.value = false
  
  const items = e.dataTransfer?.items
  if (!items) return
  
  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (item.kind === 'file') {
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        // @ts-ignore
        const path = entry.fullPath
        console.log('Dropped folder:', path)
        
        const success = await projectStore.openProjectByPath(path)
        if (success) {
          emit('opened', path)
        }
        return
      }
    }
  }
  
  uiStore.addToast('Please drop a folder, not a file', 'warning')
}

async function selectProject() {
  try {
    const dirPath = await apiService.selectDirectory()
    
    if (!dirPath || dirPath === '') {
      return
    }
    
    const success = await projectStore.openProjectByPath(dirPath)
    if (success && projectStore.currentPath) {
      emit('opened', projectStore.currentPath)
    }
  } catch (error) {
    console.error('Failed to select project:', error)
    const errorMessage = error instanceof Error ? error.message : 'Unknown error'
    uiStore.addToast(`Failed to select directory: ${errorMessage}`, 'error')
  }
}

async function openRecentProject(path: string) {
  try {
    await projectStore.openProjectByPath(path)
    emit('opened', path)
  } catch (error) {
    console.error('Failed to open recent project:', error)
    uiStore.addToast('Failed to open project', 'error')
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
