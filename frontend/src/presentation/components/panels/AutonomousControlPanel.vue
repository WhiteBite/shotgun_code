<template>
  <div
    class="h-full flex flex-col bg-gradient-to-br from-gray-800 to-gray-900 rounded-lg shadow-xl border border-gray-700"
  >
    <!-- Header -->
    <div
      class="p-6 border-b border-gray-700 bg-gradient-to-r from-gray-800 to-gray-700"
    >
      <div class="flex items-center justify-between">
        <div>
          <h2
            class="text-2xl font-bold text-white mb-1 flex items-center gap-3"
          >
            <div
              class="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center"
            >
              <svg
                class="w-5 h-5 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            ARK Code Autonomous Mode
          </h2>
          <p class="text-gray-400 text-sm">
            Multi-step autonomous agent for intelligent code generation and
            analysis
          </p>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
          <span class="text-xs text-gray-400">Ready</span>
        </div>
      </div>
    </div>

    <!-- Task Input Section -->
    <div class="p-6 border-b border-gray-700 bg-gray-800/50">
      <div class="space-y-6">
        <!-- Task Description -->
        <div>
          <label
            class="block text-sm font-semibold text-gray-200 mb-3 flex items-center gap-2"
          >
            <svg
              class="w-4 h-4 text-blue-400"
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
            Task Description
          </label>
          <textarea
            v-model="autonomousStore.currentTask"
            placeholder="Describe what you want to accomplish... (e.g., 'Add error handling to the authentication service', 'Refactor the database connection logic', 'Implement a new API endpoint')"
            class="w-full h-28 px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none transition-all duration-200"
            :disabled="autonomousStore.isTaskRunning"
          ></textarea>
        </div>

        <!-- SLA Policy Selection -->
        <div>
          <label
            class="block text-sm font-semibold text-gray-200 mb-3 flex items-center gap-2"
          >
            <svg
              class="w-4 h-4 text-purple-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
              />
            </svg>
            SLA Policy
          </label>
          <div class="grid grid-cols-3 gap-3">
            <button
              v-for="policy in slaPolicies"
              :key="policy.value"
              :class="[
                'relative p-4 rounded-lg border-2 transition-all duration-200 transform hover:scale-105',
                autonomousStore.slaPolicy === policy.value
                  ? 'border-blue-500 bg-blue-500/10 shadow-lg shadow-blue-500/25'
                  : 'border-gray-600 bg-gray-700 hover:border-gray-500 hover:bg-gray-600',
              ]"
              :disabled="autonomousStore.isTaskRunning"
              @click="autonomousStore.slaPolicy = policy.value"
            >
              <div class="text-center">
                <div class="text-lg font-bold text-white mb-1">
                  {{ policy.label }}
                </div>
                <div class="text-xs text-gray-400">
                  {{ policy.description }}
                </div>
              </div>
              <div
                v-if="autonomousStore.slaPolicy === policy.value"
                class="absolute -top-2 -right-2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center"
              >
                <svg
                  class="w-4 h-4 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
            </button>
          </div>
        </div>

        <!-- Control Buttons -->
        <div class="flex gap-3">
          <div class="text-xs text-gray-500 mb-2">
            Debug: canStartTask={{ autonomousStore.canStartTask }}, 
            isTaskRunning={{ autonomousStore.isTaskRunning }}, 
            currentTask.length={{ autonomousStore.currentTask.length }}
          </div>
          <button
            :disabled="
              !autonomousStore.canStartTask || autonomousStore.isLoading
            "
            class="flex-1 px-6 py-3 bg-gradient-to-r from-green-600 to-green-500 hover:from-green-500 hover:to-green-400 disabled:from-gray-600 disabled:to-gray-500 disabled:cursor-not-allowed rounded-lg text-white font-semibold transition-all duration-200 transform hover:scale-105 disabled:transform-none flex items-center justify-center gap-3 shadow-lg"
            @click="startTask"
          >
            <svg
              v-if="autonomousStore.isLoading"
              class="animate-spin h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            <svg
              v-else
              class="h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            {{
              autonomousStore.isLoading
                ? "Starting..."
                : "Start Autonomous Task"
            }}
          </button>

          <button
            :disabled="!autonomousStore.canCancelTask"
            class="px-6 py-3 bg-gradient-to-r from-red-600 to-red-500 hover:from-red-500 hover:to-red-400 disabled:from-gray-600 disabled:to-gray-500 disabled:cursor-not-allowed rounded-lg text-white font-semibold transition-all duration-200 transform hover:scale-105 disabled:transform-none flex items-center gap-2 shadow-lg"
            @click="cancelTask"
          >
            <svg
              class="h-5 w-5"
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
            Cancel
          </button>
        </div>
      </div>
    </div>

    <!-- Task Status -->
    <div
      v-if="autonomousStore.taskStatus"
      class="p-6 border-b border-gray-700 bg-gray-800/30"
    >
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-white flex items-center gap-2">
            <svg
              class="w-5 h-5 text-blue-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
              />
            </svg>
            Task Status
          </h3>
          <span :class="statusBadgeClass">
            {{ autonomousStore.taskStatus.status }}
          </span>
        </div>

        <!-- Progress Bar -->
        <div v-if="autonomousStore.isTaskRunning" class="space-y-3">
          <div class="flex justify-between text-sm text-gray-300">
            <span>Progress</span>
            <span class="font-semibold"
              >{{ Math.round(autonomousStore.taskStatus.progress) }}%</span
            >
          </div>
          <div class="w-full bg-gray-700 rounded-full h-3 overflow-hidden">
            <div
              class="bg-gradient-to-r from-blue-500 to-purple-600 h-3 rounded-full transition-all duration-500 ease-out shadow-lg"
              :style="{ width: `${autonomousStore.taskStatus.progress}%` }"
            ></div>
          </div>
          <div
            v-if="autonomousStore.taskStatus.currentStep"
            class="text-sm text-gray-400 flex items-center gap-2"
          >
            <svg
              class="w-4 h-4 text-blue-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
            Current step: {{ autonomousStore.taskStatus.currentStep }}
          </div>
          <div
            v-if="autonomousStore.taskStatus.estimatedTimeRemaining"
            class="text-sm text-gray-400 flex items-center gap-2"
          >
            <svg
              class="w-4 h-4 text-green-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            Estimated time remaining:
            {{ formatTime(autonomousStore.taskStatus.estimatedTimeRemaining) }}
          </div>
        </div>

        <!-- Error Display -->
        <div
          v-if="autonomousStore.taskStatus.error"
          class="p-4 bg-red-900/30 border border-red-700 rounded-lg"
        >
          <div class="flex items-center gap-2">
            <svg
              class="w-5 h-5 text-red-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p class="text-red-300 text-sm">
              {{ autonomousStore.taskStatus.error }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- TPL Plan Visualization -->
    <div v-if="autonomousStore.tplPlan" class="flex-1 p-6 overflow-y-auto">
      <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <svg
          class="w-5 h-5 text-purple-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
          />
        </svg>
        Task Plan (TPL)
      </h3>
      <div class="space-y-4">
        <div
          v-for="step in autonomousStore.tplPlan.steps"
          :key="step.id"
          class="p-4 bg-gray-700/50 rounded-lg border-l-4 transition-all duration-200 hover:bg-gray-700/70"
          :class="stepStatusBorderClass(step.status)"
        >
          <div class="flex items-center justify-between mb-3">
            <h4 class="font-semibold text-white flex items-center gap-2">
              <svg
                class="w-4 h-4"
                :class="stepStatusIconClass(step.status)"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  v-if="step.status === 'completed'"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
                <path
                  v-else-if="step.status === 'running'"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
                <path
                  v-else-if="step.status === 'failed'"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
                <path
                  v-else
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {{ step.operation }}
            </h4>
            <span :class="stepStatusBadgeClass(step.status)">
              {{ step.status }}
            </span>
          </div>
          <p class="text-sm text-gray-300 mb-2">{{ step.description }}</p>
          <div
            v-if="step.duration"
            class="text-xs text-gray-500 flex items-center gap-1"
          >
            <svg
              class="w-3 h-3"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            Duration: {{ formatDuration(step.duration) }}
          </div>
          <div
            v-if="step.error"
            class="mt-3 p-3 bg-red-900/30 border border-red-700 rounded text-xs text-red-300"
          >
            {{ step.error }}
          </div>
        </div>
      </div>
    </div>

    <!-- Reports Section -->
    <div
      v-if="autonomousStore.reports.length > 0"
      class="p-6 border-t border-gray-700 bg-gray-800/30"
    >
      <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <svg
          class="w-5 h-5 text-green-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
        Reports
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="report in autonomousStore.reports"
          :key="report.id"
          class="p-4 bg-gray-700/50 rounded-lg hover:bg-gray-700/70 transition-all duration-200 transform hover:scale-105 cursor-pointer border border-gray-600 hover:border-gray-500"
          @click="viewReport(report)"
        >
          <div class="flex items-center justify-between mb-3">
            <h4 class="font-semibold text-white">{{ report.title }}</h4>
            <span class="text-xs px-2 py-1 bg-blue-600 text-white rounded-full">
              {{ report.type }}
            </span>
          </div>
          <p class="text-sm text-gray-400 mb-3">{{ report.summary }}</p>
          <div class="text-xs text-gray-500 flex items-center gap-1">
            <svg
              class="w-3 h-3"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            {{ formatDate(report.createdAt) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Error Display -->
    <div
      v-if="autonomousStore.error"
      class="p-6 border-t border-gray-700 bg-red-900/20"
    >
      <div class="p-4 bg-red-900/30 border border-red-700 rounded-lg">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <svg
              class="w-5 h-5 text-red-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p class="text-red-300 text-sm">{{ autonomousStore.error }}</p>
          </div>
          <button
            class="text-xs text-red-400 hover:text-red-300 transition-colors"
            @click="autonomousStore.clearError()"
          >
            Dismiss
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useAutonomousStore } from "@/stores/autonomous.store";
import { useProjectStore } from "@/stores/project.store";
import { useNotificationsStore } from "@/stores/notifications.store";

const autonomousStore = useAutonomousStore();
const projectStore = useProjectStore();
const notifications = useNotificationsStore();

const slaPolicies = [
  {
    value: "lite" as const,
    label: "Lite",
    description: "Fast execution with minimal quality gates",
  },
  {
    value: "standard" as const,
    label: "Standard",
    description: "Balanced speed and quality with comprehensive checks",
  },
  {
    value: "strict" as const,
    label: "Strict",
    description: "Maximum quality with all safety checks enabled",
  },
];

// const selectedSlaPolicy = computed(() =>
//   slaPolicies.find(p => p.value === autonomousStore.slaPolicy) || slaPolicies[1]
// )

const statusBadgeClass = computed(() => {
  const status = autonomousStore.taskStatus?.status;
  switch (status) {
    case "completed":
      return "px-3 py-1 bg-green-600 text-white text-xs rounded-full font-semibold shadow-lg";
    case "failed":
      return "px-3 py-1 bg-red-600 text-white text-xs rounded-full font-semibold shadow-lg";
    case "cancelled":
      return "px-3 py-1 bg-gray-600 text-white text-xs rounded-full font-semibold shadow-lg";
    case "running":
      return "px-3 py-1 bg-blue-600 text-white text-xs rounded-full font-semibold shadow-lg animate-pulse";
    default:
      return "px-3 py-1 bg-yellow-600 text-white text-xs rounded-full font-semibold shadow-lg";
  }
});

function stepStatusBorderClass(status: string) {
  switch (status) {
    case "completed":
      return "border-green-500";
    case "failed":
      return "border-red-500";
    case "running":
      return "border-blue-500";
    case "skipped":
      return "border-gray-500";
    default:
      return "border-gray-600";
  }
}

function stepStatusIconClass(status: string) {
  switch (status) {
    case "completed":
      return "text-green-400";
    case "failed":
      return "text-red-400";
    case "running":
      return "text-blue-400 animate-spin";
    case "skipped":
      return "text-gray-400";
    default:
      return "text-yellow-400";
  }
}

function stepStatusBadgeClass(status: string) {
  switch (status) {
    case "completed":
      return "px-2 py-1 bg-green-600 text-white text-xs rounded-full font-semibold";
    case "failed":
      return "px-2 py-1 bg-red-600 text-white text-xs rounded-full font-semibold";
    case "running":
      return "px-2 py-1 bg-blue-600 text-white text-xs rounded-full font-semibold animate-pulse";
    case "skipped":
      return "px-2 py-1 bg-gray-600 text-white text-xs rounded-full font-semibold";
    default:
      return "px-2 py-1 bg-yellow-600 text-white text-xs rounded-full font-semibold";
  }
}

function formatTime(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}s`;
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m ${seconds % 60}s`;
  } else {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  }
}

function formatDuration(milliseconds: number): string {
  return formatTime(Math.floor(milliseconds / 1000));
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString();
}

async function startTask() {
  if (!projectStore.currentProject) {
    notifications.addLog("No project loaded", "error");
    return;
  }

  const success = await autonomousStore.startTask(
    projectStore.currentProject.path,
  );
  if (success) {
    notifications.addLog("Autonomous task started successfully", "success");
  }
}

async function cancelTask() {
  const success = await autonomousStore.cancelCurrentTask();
  if (success) {
    notifications.addLog(`Task cancellation requested`, "info");
  }
}

function viewReport(report: unknown) {
  // TODO: Implement report viewer modal
  console.log("Viewing report:", report);
  notifications.addLog(`Opening report: ${(report as { title: string }).title}`, "info");
}
</script>