<template>
  <transition name="slide-fade">
    <aside
      v-if="uiStore.activeDrawer === 'settings'"
      class="absolute top-0 right-0 h-full w-[28rem] bg-gray-800 border-l border-gray-700 shadow-2xl z-30 p-4 flex flex-col"
    >
      <div class="flex items-center justify-between mb-4 flex-shrink-0">
        <h2 class="text-xl font-semibold text-white">Настройки</h2>
        <button
          @click="uiStore.closeDrawer()"
          class="p-2 rounded-md hover:bg-gray-700"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="text-gray-400"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div
        v-if="settingsStore.isLoading"
        class="flex-grow flex items-center justify-center"
      >
        <p class="text-gray-400">Загрузка настроек...</p>
      </div>

      <div
        v-else
        class="flex-grow overflow-y-auto space-y-4 text-sm pr-2 -mr-2"
      >
        <!-- Provider Selection -->
        <div>
          <label
            for="provider-select"
            class="block font-medium text-gray-300 mb-1"
            >AI Провайдер</label
          >
          <select
            id="provider-select"
            v-model="settingsStore.settings.selectedProvider"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="openai">OpenAI</option>
            <option value="gemini">Gemini</option>
            <option value="openrouter">OpenRouter</option>
            <option value="localai">Local AI</option>
          </select>
        </div>

        <!-- Provider-specific settings -->
        <div
          v-if="settingsStore.settings.selectedProvider === 'openai'"
          class="p-3 bg-gray-900/50 rounded-md space-y-3 border border-gray-700"
        >
          <h3 class="font-semibold text-white">OpenAI</h3>
          <input
            type="password"
            v-model="settingsStore.settings.openAIAPIKey"
            placeholder="OpenAI API Key (sk-...)"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
        </div>

        <div
          v-if="settingsStore.settings.selectedProvider === 'gemini'"
          class="p-3 bg-gray-900/50 rounded-md space-y-3 border border-gray-700"
        >
          <h3 class="font-semibold text-white">Google Gemini</h3>
          <input
            type="password"
            v-model="settingsStore.settings.geminiAPIKey"
            placeholder="Gemini API Key"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
          <button
            @click="settingsStore.refreshModels('gemini')"
            :disabled="settingsStore.isRefreshingModels"
            class="w-full text-xs py-1 bg-gray-600 hover:bg-gray-500 rounded-md"
          >
            Обновить модели
          </button>
        </div>

        <div
          v-if="settingsStore.settings.selectedProvider === 'openrouter'"
          class="p-3 bg-gray-900/50 rounded-md space-y-3 border border-gray-700"
        >
          <h3 class="font-semibold text-white">OpenRouter</h3>
          <input
            type="password"
            v-model="settingsStore.settings.openRouterAPIKey"
            placeholder="OpenRouter API Key (sk-or-..)"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
          <p class="text-xs text-gray-400">
            Модели жестко заданы. Введите имя модели в поле ниже, если ее нет в
            списке.
          </p>
        </div>

        <div
          v-if="settingsStore.settings.selectedProvider === 'localai'"
          class="p-3 bg-gray-900/50 rounded-md space-y-3 border border-gray-700"
        >
          <h3 class="font-semibold text-white">Local AI (OpenAI compatible)</h3>
          <input
            type="text"
            v-model="settingsStore.settings.localAIHost"
            placeholder="Host URL (e.g., http://localhost:1234/v1)"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
          <input
            type="text"
            v-model="settingsStore.settings.localAIModelName"
            placeholder="Model Name (e.g., llama3)"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
          <input
            type="password"
            v-model="settingsStore.settings.localAIAPIKey"
            placeholder="API Key (optional)"
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2"
          />
        </div>

        <!-- Model Selection -->
        <div v-if="currentModels.length > 0">
          <label for="model-select" class="block font-medium text-gray-300 mb-1"
            >Модель</label
          >
          <select
            id="model-select"
            v-model="
              settingsStore.settings.selectedModels[
                settingsStore.settings.selectedProvider
              ]
            "
            class="w-full bg-gray-700 border-gray-600 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option v-for="model in currentModels" :key="model" :value="model">
              {{ model }}
            </option>
          </select>
        </div>
      </div>

      <div class="mt-4 flex-shrink-0">
        <button
          @click="handleSave"
          class="w-full py-2 bg-blue-600 hover:bg-blue-500 rounded-md font-semibold disabled:opacity-50"
          :disabled="settingsStore.isLoading"
        >
          {{
            settingsStore.isLoading ? "Сохранение..." : "Сохранить и закрыть"
          }}
        </button>
      </div>
    </aside>
  </transition>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useUiStore } from "@/stores/ui.store";
import { useSettingsStore } from "@/stores/settings.store";

const uiStore = useUiStore();
const settingsStore = useSettingsStore();

const currentModels = computed(() => {
  const provider = settingsStore.settings.selectedProvider;
  return settingsStore.settings.availableModels[provider] || [];
});

async function handleSave() {
  await settingsStore.saveSettings();
  uiStore.closeDrawer();
}
</script>

<style scoped>
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: transform 0.22s cubic-bezier(0.165, 0.84, 0.44, 1);
}
.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translateX(100%);
}
</style>
