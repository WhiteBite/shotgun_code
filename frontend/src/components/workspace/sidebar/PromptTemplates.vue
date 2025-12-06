<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <p class="text-xs font-semibold text-gray-400">{{ t('prompts.title') }}</p>
      <button
        @click="showAddForm = !showAddForm"
        class="icon-btn-sm icon-btn-primary"
        :title="t('prompts.add')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </div>

    <!-- Add Form -->
    <Transition name="slide-fade">
      <div v-if="showAddForm" class="context-stats space-y-3">
        <input
          v-model="newTemplate.name"
          type="text"
          class="input text-xs !py-2"
          :placeholder="t('prompts.name')"
        />
        <textarea
          v-model="newTemplate.content"
          class="textarea text-xs !py-2 h-24"
          :placeholder="t('prompts.placeholder')"
        ></textarea>
        <p class="text-[10px] text-gray-500">{{ t('prompts.variables') }}</p>
        <div class="flex gap-2">
          <button
            @click="showAddForm = false"
            class="btn-unified btn-unified-ghost text-xs flex-1"
          >
            {{ t('context.cancel') }}
          </button>
          <button
            @click="saveTemplate"
            :disabled="!newTemplate.name || !newTemplate.content"
            class="btn-unified btn-unified-primary text-xs flex-1"
          >
            {{ t('ignoreModal.save') }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Default Templates -->
    <div class="space-y-2">
      <div
        v-for="template in defaultTemplates"
        :key="template.id"
        class="list-item group !p-2 cursor-pointer"
        @click="useTemplate(template)"
      >
        <div class="flex items-center gap-2">
          <div class="section-icon section-icon-orange !w-7 !h-7 flex-shrink-0">
            <svg class="!w-3.5 !h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <span class="text-sm text-white">{{ template.name }}</span>
            <p class="text-xs text-gray-500 truncate">{{ template.preview }}</p>
          </div>
          <button
            class="icon-btn-sm opacity-0 group-hover:opacity-100 transition-opacity"
            :title="t('prompts.copy')"
            @click.stop="copyTemplate(template)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Custom Templates -->
    <div v-if="customTemplates.length > 0" class="space-y-2 pt-2 border-t border-gray-700/30">
      <p class="text-[10px] text-gray-500 uppercase tracking-wider">Custom</p>
      <div
        v-for="template in customTemplates"
        :key="template.id"
        class="list-item group !p-2 cursor-pointer"
        @click="useTemplate(template)"
      >
        <div class="flex items-center gap-2">
          <div class="section-icon section-icon-purple !w-7 !h-7 flex-shrink-0">
            <svg class="!w-3.5 !h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <span class="text-sm text-white">{{ template.name }}</span>
            <p class="text-xs text-gray-500 truncate">{{ template.preview }}</p>
          </div>
          <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              class="icon-btn-sm"
              :title="t('prompts.copy')"
              @click.stop="copyTemplate(template)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            </button>
            <button
              class="icon-btn-sm icon-btn-danger"
              :title="t('prompts.delete')"
              @click.stop="deleteTemplate(template.id)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { useUIStore } from '@/stores/ui.store'
import { reactive, ref } from 'vue'

const { t } = useI18n()
const contextStore = useContextStore()
const uiStore = useUIStore()

interface PromptTemplate {
  id: string
  name: string
  content: string
  preview: string
}

const showAddForm = ref(false)
const newTemplate = reactive({ name: '', content: '' })

const defaultTemplates: PromptTemplate[] = [
  {
    id: 'review',
    name: t('prompts.default.review'),
    content: 'Please review the following code and provide feedback on:\n1. Code quality and best practices\n2. Potential bugs or issues\n3. Performance improvements\n4. Security concerns\n\n{context}',
    preview: 'Review code quality, bugs, performance...'
  },
  {
    id: 'explain',
    name: t('prompts.default.explain'),
    content: 'Please explain the following code in detail:\n- What does it do?\n- How does it work?\n- What are the key components?\n\n{context}',
    preview: 'Explain what the code does...'
  },
  {
    id: 'refactor',
    name: t('prompts.default.refactor'),
    content: 'Please refactor the following code to improve:\n- Readability\n- Maintainability\n- Performance\n- Following best practices\n\n{context}',
    preview: 'Refactor for better quality...'
  },
  {
    id: 'tests',
    name: t('prompts.default.tests'),
    content: 'Please write comprehensive unit tests for the following code:\n- Cover edge cases\n- Test error handling\n- Include both positive and negative tests\n\n{context}',
    preview: 'Write unit tests...'
  }
]

// Load custom templates from localStorage
function loadCustomTemplates(): PromptTemplate[] {
  try {
    const saved = localStorage.getItem('prompt-templates')
    if (saved) return JSON.parse(saved)
  } catch (e) {
    console.warn('Failed to load prompt templates:', e)
  }
  return []
}

const customTemplates = ref<PromptTemplate[]>(loadCustomTemplates())

function saveCustomTemplates() {
  try {
    localStorage.setItem('prompt-templates', JSON.stringify(customTemplates.value))
  } catch (e) {
    console.warn('Failed to save prompt templates:', e)
  }
}

function saveTemplate() {
  if (!newTemplate.name || !newTemplate.content) return

  const template: PromptTemplate = {
    id: `custom-${Date.now()}`,
    name: newTemplate.name,
    content: newTemplate.content,
    preview: newTemplate.content.substring(0, 50) + '...'
  }

  customTemplates.value.push(template)
  saveCustomTemplates()

  newTemplate.name = ''
  newTemplate.content = ''
  showAddForm.value = false

  uiStore.addToast(t('prompts.saved'), 'success')
}

function deleteTemplate(id: string) {
  customTemplates.value = customTemplates.value.filter(t => t.id !== id)
  saveCustomTemplates()
  uiStore.addToast(t('prompts.deleted'), 'success')
}

async function useTemplate(template: PromptTemplate) {
  let content = template.content

  // Replace variables
  if (contextStore.hasContext) {
    try {
      const contextContent = await contextStore.getFullContextContent()
      content = content.replace('{context}', contextContent)
    } catch (e) {
      content = content.replace('{context}', '[No context available]')
    }
  } else {
    content = content.replace('{context}', '[No context built]')
  }

  content = content.replace('{files}', String(contextStore.fileCount))
  content = content.replace('{tokens}', String(contextStore.tokenCount))

  await navigator.clipboard.writeText(content)
  uiStore.addToast(t('prompts.copied'), 'success')
}

async function copyTemplate(template: PromptTemplate) {
  await navigator.clipboard.writeText(template.content)
  uiStore.addToast(t('prompts.copied'), 'success')
}
</script>
