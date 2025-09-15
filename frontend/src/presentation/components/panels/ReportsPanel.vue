<template>
  <div class="h-full flex flex-col p-4 bg-gray-800 rounded-lg">
    <h2 class="text-lg font-semibold mb-4">Reports</h2>
    <div class="flex-grow overflow-y-auto">
      <div v-if="reportsStore.isLoading" class="text-center text-gray-400">
        Loading reports...
      </div>
      <div v-else-if="reportsStore.reports.length === 0" class="text-center text-gray-400">
        No reports available.
      </div>
      <div v-else class="space-y-2">
        <div
          v-for="report in reportsStore.reports"
          :key="report.id"
          class="p-3 bg-gray-700 rounded-md cursor-pointer hover:bg-gray-600"
          @click="openReport(report)"
        >
          <div class="flex justify-between items-center">
            <span class="font-medium">{{ report.name }}</span>
            <span class="text-xs text-gray-400">{{ report.type }}</span>
          </div>
          <div class="text-sm text-gray-300 mt-1">{{ report.description }}</div>
          <div class="text-xs text-gray-500 mt-2">{{ new Date(report.timestamp).toLocaleString() }}</div>
        </div>
      </div>
    </div>
    <div class="mt-4">
      <button class="w-full py-2 bg-blue-600 hover:bg-blue-500 rounded-md" @click="reportsStore.loadReports">
        Refresh Reports
      </button>
    </div>

    <ReportViewerModal />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useReportsStore } from "@/stores/reports.store";
import ReportViewerModal from "../modals/ReportViewerModal.vue";

const reportsStore = useReportsStore();

function openReport(report: any) {
  reportsStore.openReport(report);
}

onMounted(() => {
  // Загрузка отчетов при монтировании компонента
  reportsStore.loadReports();
});
</script>