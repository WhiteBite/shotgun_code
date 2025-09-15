import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { container } from "@/infrastructure/container";
import { useNotificationsStore } from "@/stores/notifications.store";
// Import repository
import type { ReportsRepository } from "@/domain/repositories/ReportsRepository";

export interface Report {
  id: string;
  name: string;
  description: string;
  type: "ux" | "guardrails" | "tasks";
  timestamp: string;
  path: string;
  data?: unknown;
}

export const useReportsStore = defineStore("reports", () => {
  const reports = ref<Report[]>([]);
  const isLoading = ref(false);
  const selectedReport = ref<Report | null>(null);
  const isReportViewerOpen = ref(false);
  const notifications = useNotificationsStore();

  // Inject ReportsRepository
  const reportsRepository: ReportsRepository = container.reportsRepository;

  // Computed
  const uxReports = computed(() =>
    reports.value.filter((r) => r.type === "ux"),
  );
  const guardrailReports = computed(() =>
    reports.value.filter((r) => r.type === "guardrails"),
  );
  const taskReports = computed(() =>
    reports.value.filter((r) => r.type === "tasks"),
  );
  const hasReports = computed(() => reports.value.length > 0);

  // Actions
  async function loadReports() {
    isLoading.value = true;
    try {
      // Use ReportsRepository directly instead of use cases
      const reportsList = await reportsRepository.listReports();

      // Конвертируем GenericReport в Report
      reports.value = reportsList.map((report) => ({
        id: report.id,
        name: report.title,
        description: report.summary,
        type: mapReportType(report.type),
        timestamp: report.createdAt,
        path: report.id, // Используем ID как путь
        data: report.data,
      }));

      notifications.addLog(`Loaded ${reports.value.length} reports`, "info");
    } catch (error) {
      console.error("Failed to load reports:", error);
      notifications.addLog("Failed to load reports", "error");
      reports.value = [];
    } finally {
      isLoading.value = false;
    }
  }

  async function loadReportData(reportId: string) {
    const report = reports.value.find((r) => r.id === reportId);
    if (!report) return null;

    try {
      // Use ReportsRepository directly instead of use cases
      const reportData = await reportsRepository.getReport(report.path);
      return reportData;
    } catch (error) {
      console.error("Failed to load report data:", error);
      notifications.addLog("Failed to load report data", "error");
      return null;
    }
  }

  function mapReportType(type: string): "ux" | "guardrails" | "tasks" {
    switch (type) {
      case "why_view":
      case "time_to_green":
      case "derived_diff":
        return "ux";
      case "guardrails":
        return "guardrails";
      default:
        return "tasks";
    }
  }

  function openReport(report: Report) {
    selectedReport.value = report;
    isReportViewerOpen.value = true;
  }

  function closeReportViewer() {
    isReportViewerOpen.value = false;
    selectedReport.value = null;
  }

  async function exportAllReports() {
    try {
      // В реальном приложении здесь будет экспорт всех отчетов
      console.log("Exporting all reports...");
      notifications.addLog("Export functionality not implemented yet", "info");
      // TODO: Implement export functionality
    } catch (error) {
      console.error("Failed to export reports:", error);
      notifications.addLog("Failed to export reports", "error");
    }
  }

  async function openReportsDirectory() {
    try {
      // В реальном приложении здесь будет открытие папки с отчетами
      console.log("Opening reports directory...");
      notifications.addLog("Directory opening not implemented yet", "info");
      // TODO: Implement directory opening
    } catch (error) {
      console.error("Failed to open reports directory:", error);
      notifications.addLog("Failed to open reports directory", "error");
    }
  }

  return {
    // State
    reports,
    isLoading,
    selectedReport,
    isReportViewerOpen,

    // Computed
    uxReports,
    guardrailReports,
    taskReports,
    hasReports,

    // Actions
    loadReports,
    loadReportData,
    openReport,
    closeReportViewer,
    exportAllReports,
    openReportsDirectory,
  };
});