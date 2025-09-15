import { computed } from 'vue'
import type { SLAPolicy, TaskStatus, TPLPlan, Report } from '@/types/dto'
import { UIFormattingService } from '@/domain/services/UIFormattingService'
import { APP_CONFIG } from '@/config/app-config'

export class AutonomousControlService {
  private formattingService: UIFormattingService

  constructor(
    private getCurrentTask: () => string,
    private setCurrentTask: (task: string) => void,
    private getSlaPolicy: () => SLAPolicy,
    private setSlaPolicy: (policy: SLAPolicy) => void,
    private getIsTaskRunning: () => boolean,
    private getIsLoading: () => boolean,
    private getTaskStatus: () => TaskStatus | null,
    private getTplPlan: () => TPLPlan | null,
    private getReports: () => Report[],
    private getError: () => string | null,
    private getCanStartTask: () => boolean,
    private getCanCancelTask: () => boolean,
    private startTask: (projectPath: string) => Promise<boolean>,
    private cancelCurrentTask: () => Promise<boolean>,
    private clearError: () => void,
    formattingService?: UIFormattingService
  ) {
    this.formattingService = formattingService || new UIFormattingService()
  }

  // SLA Policies
  get slaPolicies(): Array<{ value: SLAPolicy; label: string; description: string }> {
    return [
      {
        value: "lite",
        label: "Lite",
        description: "Fast execution with minimal quality gates",
      },
      {
        value: "standard",
        label: "Standard",
        description: "Balanced speed and quality with comprehensive checks",
      },
      {
        value: "strict",
        label: "Strict",
        description: "Maximum quality with all safety checks enabled",
      },
    ]
  }

  // Task management
  get currentTask() {
    return this.getCurrentTask()
  }

  set currentTask(task: string) {
    this.setCurrentTask(task)
  }

  get slaPolicy() {
    return this.getSlaPolicy()
  }

  set slaPolicy(policy: SLAPolicy) {
    this.setSlaPolicy(policy)
  }

  get isTaskRunning() {
    return this.getIsTaskRunning()
  }

  get isLoading() {
    return this.getIsLoading()
  }

  get taskStatus() {
    return this.getTaskStatus()
  }

  get tplPlan() {
    return this.getTplPlan()
  }

  get reports() {
    return this.getReports()
  }

  get error() {
    return this.getError()
  }

  get canStartTask() {
    return this.getCanStartTask()
  }

  get canCancelTask() {
    return this.getCanCancelTask()
  }

  // Action handlers
  async handleStartTask(projectPath: string): Promise<boolean> {
    return await this.startTask(projectPath)
  }

  async handleCancelTask(): Promise<boolean> {
    return await this.cancelCurrentTask()
  }

  handleClearError() {
    this.clearError()
  }

  handleViewReport(report: unknown) {
    // TODO: Implement report viewer modal
    console.log("Viewing report:", report)
  }

  // Computed properties
  get statusBadgeClass() {
    const status = this.getTaskStatus()?.status
    switch (status) {
      case "completed":
        return "px-3 py-1 bg-green-600 text-white text-xs rounded-full font-semibold shadow-lg"
      case "failed":
        return "px-3 py-1 bg-red-600 text-white text-xs rounded-full font-semibold shadow-lg"
      case "cancelled":
        return "px-3 py-1 bg-gray-600 text-white text-xs rounded-full font-semibold shadow-lg"
      case "running":
        return "px-3 py-1 bg-blue-600 text-white text-xs rounded-full font-semibold shadow-lg animate-pulse"
      default:
        return "px-3 py-1 bg-yellow-600 text-white text-xs rounded-full font-semibold shadow-lg"
    }
  }

  stepStatusBorderClass(status: string) {
    switch (status) {
      case "completed":
        return "border-green-500"
      case "failed":
        return "border-red-500"
      case "running":
        return "border-blue-500"
      case "skipped":
        return "border-gray-500"
      default:
        return "border-gray-600"
    }
  }

  stepStatusIconClass(status: string) {
    switch (status) {
      case "completed":
        return "text-green-400"
      case "failed":
        return "text-red-400"
      case "running":
        return "text-blue-400 animate-spin"
      case "skipped":
        return "text-gray-400"
      default:
        return "text-yellow-400"
    }
  }

  stepStatusBadgeClass(status: string) {
    switch (status) {
      case "completed":
        return "px-2 py-1 bg-green-600 text-white text-xs rounded-full font-semibold"
      case "failed":
        return "px-2 py-1 bg-red-600 text-white text-xs rounded-full font-semibold"
      case "running":
        return "px-2 py-1 bg-blue-600 text-white text-xs rounded-full font-semibold animate-pulse"
      case "skipped":
        return "px-2 py-1 bg-gray-600 text-white text-xs rounded-full font-semibold"
      default:
        return "px-2 py-1 bg-yellow-600 text-white text-xs rounded-full font-semibold"
    }
  }

  // Formatting utilities
  formatTime(seconds: number): string {
    if (seconds < 60) {
      return `${seconds}s`
    } else if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60)
      return `${minutes}m ${seconds % 60}s`
    } else {
      const hours = Math.floor(seconds / 3600)
      const minutes = Math.floor((seconds % 3600) / 60)
      return `${hours}h ${minutes}m`
    }
  }

  formatDuration(milliseconds: number): string {
    return this.formatTime(Math.floor(milliseconds / 1000))
  }

  formatDate(dateString: string): string {
    return this.formattingService.formatDate(dateString)
  }
}