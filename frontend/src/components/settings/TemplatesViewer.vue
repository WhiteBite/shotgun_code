frontend/s<template>
  <div>
    <h3 class="text-lg font-medium text-gray-800 mb-2">Prompt Templates</h3>
    <p class="text-sm text-gray-500 mb-4">
      This is a read-only view of the system prompts used for different tasks. You can select which one to use on the main screen.
    </p>
    <div class="space-y-4">
      <div v-for="(template, key) in promptTemplate.templates" :key="key" class="border rounded-md">
        <div class="bg-gray-100 p-2 border-b">
          <h4 class="font-semibold text-gray-700">{{ template.name }} ({{ key }})</h4>
        </div>
        <pre class="w-full p-3 bg-gray-50 font-mono text-xs overflow-x-auto whitespace-pre-wrap">{{ template.content }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup>
import { usePromptTemplateStore } from '../../stores/promptTemplateStore';

const promptTemplate = usePromptTemplateStore();
</script>rc/components/settings/TemplatesViewer.vue