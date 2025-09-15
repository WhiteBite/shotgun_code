<template>
  <div
    v-if="isVisible"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
  >
    <div
      class="bg-gray-900 border border-gray-700 rounded-lg shadow-xl w-full max-w-4xl h-full max-h-[90vh] flex flex-col"
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between p-4 border-b border-gray-700"
      >
        <div>
          <h2 class="text-lg font-semibold text-white">{{ report?.name }}</h2>
          <p class="text-sm text-gray-400">{{ report?.description }}</p>
        </div>
        <div class="flex items-center space-x-2">
          <button
            class="p-2 text-gray-400 hover:text-white transition-colors"
            title="Экспорт отчета"
            @click="exportReport"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </button>
          <button
            class="p-2 text-gray-400 hover:text-white transition-colors"
            title="Закрыть"
            @click="closeModal"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-grow overflow-hidden">
        <div class="h-full flex">
          <!-- Report Info -->
          <div
            class="w-64 bg-gray-800/50 border-r border-gray-700 p-4 overflow-y-auto"
          >
            <div class="space-y-4">
              <div>
                <h3 class="text-sm font-medium text-gray-300 mb-2">
                  Информация
                </h3>
                <div class="space-y-2 text-xs text-gray-400">
                  <div>
                    <span class="font-medium">Тип:</span>
                    <span class="ml-1">{{
                      getReportTypeLabel(report?.type)
                    }}</span>
                  </div>
                  <div>
                    <span class="font-medium">Дата:</span>
                    <span class="ml-1">{{
                      formatDate(report?.timestamp)
                    }}</span>
                  </div>
                  <div>
                    <span class="font-medium">Путь:</span>
                    <span class="ml-1 break-all">{{ report?.path }}</span>
                  </div>
                </div>
              </div>

              <div v-if="reportData">
                <h3 class="text-sm font-medium text-gray-300 mb-2">
                  Структура
                </h3>
                <div class="space-y-1">
                  <div
                    v-for="(value, key) in reportData"
                    :key="key"
                    class="text-xs text-gray-400 cursor-pointer hover:text-gray-300"
                    @click="scrollToSection(key)"
                  >
                    <span class="font-medium">{{ key }}:</span>
                    <span class="ml-1">{{ getValuePreview(value) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- JSON Viewer -->
          <div class="flex-grow overflow-hidden">
            <div class="h-full overflow-auto">
              <pre
                v-if="reportData"
                class="p-4 text-xs text-gray-300 font-mono whitespace-pre-wrap"
                >{{ JSON.stringify(reportData, null, 2) }}</pre
              >
              <div v-else class="p-4 text-center text-gray-500">
                <svg
                  class="w-8 h-8 mx-auto mb-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <div>Загрузка отчета...</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useReportsStore } from "@/stores/reports.store";

const reportsStore = useReportsStore();
const reportData = ref<unknown>(null);
const isLoading = ref(false);

const isVisible = computed({
  get: () => reportsStore.isReportViewerOpen,
  set: (value) => {
    if (!value) {
      reportsStore.closeReportViewer();
    }
  },
});

const report = computed(() => reportsStore.selectedReport);

function closeModal() {
  reportsStore.closeReportViewer();
}

function getReportTypeLabel(type?: string): string {
  switch (type) {
    case "ux":
      return "UX Метрики";
    case "guardrails":
      return "Guardrails";
    case "tasks":
      return "Задачи";
    default:
      return "Неизвестно";
  }
}

function formatDate(dateString?: string): string {
  if (!dateString) return "Неизвестно";
  const date = new Date(dateString);
  return date.toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function getValuePreview(value: unknown): string {
  if (typeof value === "string") {
    return value.length > 50 ? value.substring(0, 50) + "..." : value;
  }
  if (Array.isArray(value)) {
    return `[${value.length} элементов]`;
  }
  if (typeof value === "object" && value !== null) {
    return `{${Object.keys(value).length} полей}`;
  }
  return String(value);
}

function scrollToSection(key: string) {
  // В реальном приложении здесь будет прокрутка к секции
  console.log("Scroll to section:", key);
}

async function loadReportData() {
  if (!report.value) return;

  isLoading.value = true;
  try {
    const data = await reportsStore.loadReportData(report.value.id);
    reportData.value = data;
  } catch (error) {
    console.error("Failed to load report data:", error);
    reportData.value = null;
  } finally {
    isLoading.value = false;
  }
}

function exportReport() {
  if (!reportData.value) return;

  const dataStr = JSON.stringify(reportData.value, null, 2);
  const dataBlob = new Blob([dataStr], { type: "application/json" });
  const url = URL.createObjectURL(dataBlob);

  const link = document.createElement("a");
  link.href = url;
  link.download = `${report.value?.name || "report"}.json`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

watch(
  () => report.value,
  () => {
    if (report.value) {
      loadReportData();
    }
  },
  { immediate: true },
);
</script>