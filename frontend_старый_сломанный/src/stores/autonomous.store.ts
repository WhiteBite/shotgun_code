import { defineStore } from "pinia";
import { ref, computed, readonly } from "vue";
import { container } from "@/infrastructure/container";
import { useNotificationsStore } from "@/stores/notifications.store";
import { useAdvancedErrorHandler, type ErrorContext } from "@/composables/useErrorHandler";
import type {
  AutonomousTaskRequest,
  AutonomousTaskResponse,
  AutonomousTaskStatus,
  TPLPlan,
  GenericReport,
  // SLAPolicy, // Unused
} from "@/types/dto";
// import { eventService } from "@/infrastructure/events/event.service"; // Unused

export const useAutonomousStore = defineStore("autonomous", () => {
  const notifications = useNotificationsStore();
  const { handleStructuredError, executeWithRetry } = useAdvancedErrorHandler();
  
  // Inject autonomous repository
  const autonomousRepository = container.autonomousRepository;

  // State
  const currentTaskId = ref<string | null>(null);
  const currentTask = ref<string>("");
  const slaPolicy = ref<"lite" | "standard" | "strict">("standard");
  const taskStatus = ref<AutonomousTaskStatus | null>(null);
  const tplPlan = ref<TPLPlan | null>(null);
  const reports = ref<GenericReport[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // Computed
  const isTaskRunning = computed(
    () =>
      taskStatus.value?.status === "running" ||
      taskStatus.value?.status === "pending",
  );

  const isTaskCompleted = computed(
    () => taskStatus.value?.status === "completed",
  );

  const isTaskFailed = computed(() => taskStatus.value?.status === "failed");

  const isTaskCancelled = computed(
    () => taskStatus.value?.status === "cancelled",
  );

  const canStartTask = computed(
    () => !isTaskRunning.value && currentTask.value.trim().length > 0,
  );

  const canCancelTask = computed(() => isTaskRunning.value);

  // Methods
  async function startTask(projectPath: string): Promise<boolean> {
    if (!canStartTask.value) {
      const validationError = "Cannot start task: invalid state or empty task description";
      error.value = validationError;
      handleStructuredError(new Error(validationError), {
        operation: 'startTask',
        component: 'autonomous'
      });
      return false;
    }

    isLoading.value = true;
    error.value = null;

    const context: ErrorContext = {
      operation: 'startAutonomousTask',
      component: 'autonomous',
      retry: () => startTask(projectPath)
    };

    try {
      const request: AutonomousTaskRequest = {
        task: currentTask.value.trim(),
        slaPolicy: slaPolicy.value,
        projectPath,
        options: {
          enableStaticAnalysis: true,
          enableTests: true,
          enableSBOM: true,
        },
      };

      // Use executeWithRetry for network-related errors
      const response: AutonomousTaskResponse = await executeWithRetry(
        () => autonomousRepository.startAutonomousTask(request),
        context,
        {
          maxAttempts: 3,
          retryableErrors: ['NETWORK_ERROR', 'RATE_LIMIT_EXCEEDED']
        }
      );

      if (response.status === "accepted") {
        currentTaskId.value = response.taskId;
        taskStatus.value = {
          taskId: response.taskId,
          status: "pending",
          progress: 0,
          startedAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };

        notifications.addLog(
          `Autonomous task started: ${response.taskId}`,
          "info",
        );
        return true;
      } else {
        const rejectionError = response.message || "Task was rejected by the system";
        error.value = rejectionError;
        handleStructuredError(new Error(rejectionError), context);
        return false;
      }
    } catch (err) {
      const errorInfo = handleStructuredError(err, context);
      error.value = errorInfo.classification.userMessage;
      notifications.addLog(`Error starting task: ${errorInfo.classification.userMessage}`, "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  async function cancelCurrentTask(): Promise<boolean> {
    if (!currentTaskId.value || !canCancelTask.value) {
      return false;
    }

    const context: ErrorContext = {
      operation: 'cancelAutonomousTask',
      component: 'autonomous',
      retry: () => cancelCurrentTask()
    };

    try {
      await autonomousRepository.cancelAutonomousTask(currentTaskId.value);
      notifications.addLog(`Task ${currentTaskId.value} cancelled`, "info");
      return true;
    } catch (err) {
      const errorInfo = handleStructuredError(err, context);
      error.value = errorInfo.classification.userMessage;
      notifications.addLog(`Error cancelling task: ${errorInfo.classification.userMessage}`, "error");
      return false;
    }
  }

  async function getTaskStatus(
    taskId: string,
  ): Promise<AutonomousTaskStatus | null> {
    try {
      const status = await autonomousRepository.getAutonomousTaskStatus(taskId);
      taskStatus.value = status;
      return status;
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to get task status";
      error.value = errorMessage;
      return null;
    }
  }

  async function loadReports(): Promise<void> {
    try {
      // Note: For reports, we're using the reports repository
      const reportsList = await container.reportsRepository.listReports();
      reports.value = reportsList;
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to load reports";
      error.value = errorMessage;
      notifications.addLog(`Error loading reports: ${errorMessage}`, "error");
    }
  }

  async function loadReportData(
    reportId: string,
  ): Promise<GenericReport | null> {
    try {
      // Note: For reports, we're using the reports repository
      const report = await container.reportsRepository.getReport(reportId);
      // Update the report in the list
      const index = reports.value.findIndex((r) => r.id === reportId);
      if (index !== -1) {
        reports.value[index] = report;
      }
      return report;
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to load report data";
      error.value = errorMessage;
      return null;
    }
  }

  // Event handlers for Wails Events
  function handleTaskStatusUpdate(status: AutonomousTaskStatus) {
    if (status.taskId === currentTaskId.value) {
      taskStatus.value = status;

      // Update notifications based on status
      if (status.status === "completed") {
        notifications.addLog(
          "Autonomous task completed successfully",
          "success",
        );
      } else if (status.status === "failed") {
        notifications.addLog(
          `Task failed: ${status.error || "Unknown error"}`,
          "error",
        );
      } else if (status.status === "cancelled") {
        notifications.addLog("Task was cancelled", "info");
      }
    }
  }

  function handleReportUpdate(report: GenericReport) {
    if (report.taskId === currentTaskId.value) {
      // Update or add the report
      const index = reports.value.findIndex((r) => r.id === report.id);
      if (index !== -1) {
        reports.value[index] = report;
      } else {
        reports.value.push(report);
      }

      notifications.addLog(`Report updated: ${report.title}`, "info");
    }
  }

  function handleTaskCompleted(_: unknown) {
    if (taskStatus.value) {
      taskStatus.value.status = "completed";
      taskStatus.value.progress = 100;
    }
    notifications.addLog(`Task ${currentTaskId.value} completed`, "success");
    currentTaskId.value = null;
  }

  function handleTaskFailed(error: unknown) {
    if (taskStatus.value) {
      taskStatus.value.status = "failed";
      taskStatus.value.error =
        error instanceof Error
          ? error.message
          : "An unknown error occurred during task execution.";
    }
    notifications.addLog(
      `Task ${currentTaskId.value} failed: ${taskStatus.value?.error}`,
      "error",
    );
    currentTaskId.value = null;
  }

  function handleTaskCancelled(taskId: string) {
    if (taskId === currentTaskId.value) {
      taskStatus.value = {
        ...taskStatus.value!,
        status: "cancelled",
        updatedAt: new Date().toISOString(),
      };
    }
  }

  // Reset state
  function reset() {
    currentTaskId.value = null;
    currentTask.value = "";
    slaPolicy.value = "standard";
    taskStatus.value = null;
    tplPlan.value = null;
    reports.value = [];
    isLoading.value = false;
    error.value = null;
  }

  function clearError() {
    error.value = null;
  }

  return {
    // State
    currentTaskId: readonly(currentTaskId),
    currentTask,
    slaPolicy,
    taskStatus: readonly(taskStatus),
    tplPlan: readonly(tplPlan),
    reports: readonly(reports),
    isLoading: readonly(isLoading),
    error: readonly(error),

    // Computed
    isTaskRunning,
    isTaskCompleted,
    isTaskFailed,
    isTaskCancelled,
    canStartTask,
    canCancelTask,

    // Methods
    startTask,
    cancelCurrentTask,
    getTaskStatus,
    loadReports,
    loadReportData,
    handleTaskStatusUpdate,
    handleReportUpdate,
    handleTaskCompleted,
    handleTaskFailed,
    handleTaskCancelled,
    reset,
    clearError,
  };
});
