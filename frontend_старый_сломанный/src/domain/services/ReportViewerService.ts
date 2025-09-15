import { ref, type Ref } from 'vue'
import type { Report } from '@/types/dto'
import { UIFormattingService } from '@/domain/services/UIFormattingService'
import { APP_CONFIG } from '@/config/app-config'

export class ReportViewerService {
  private reportData: Ref<unknown> = ref(null)
  private isLoading: Ref<boolean> = ref(false)
  private formattingService: UIFormattingService

  constructor(
    private getIsReportViewerOpen: () => boolean,
    private getSelectedReport: () => Report | null,
    private closeReportViewer: () => void,
    private loadReportData: (reportId: string) => Promise<unknown>,
    formattingService?: UIFormattingService
  ) {
    this.formattingService = formattingService || new UIFormattingService()
  }

  // State accessors
  get isVisible() {
    return this.getIsReportViewerOpen()
  }

  get report() {
    return this.getSelectedReport()
  }

  get data() {
    return this.reportData.value
  }

  get loading() {
    return this.isLoading.value
  }

  // Action handlers
  closeModal() {
    this.closeReportViewer()
  }

  async loadReport() {
    const report = this.getSelectedReport()
    if (!report) return

    this.isLoading.value = true
    try {
      const data = await this.loadReportData(report.id)
      this.reportData.value = data
    } catch (error) {
      console.error("Failed to load report data:", error)
      this.reportData.value = null
    } finally {
      this.isLoading.value = false
    }
  }

  exportReport() {
    if (!this.reportData.value) return

    const dataStr = JSON.stringify(this.reportData.value, null, 2)
    const dataBlob = new Blob([dataStr], { type: "application/json" })
    const url = URL.createObjectURL(dataBlob)

    const link = document.createElement("a")
    link.href = url
    link.download = `${this.report?.name || "report"}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  // Utility methods
  getReportTypeLabel(type?: string): string {
    switch (type) {
      case "ux":
        return "UX Метрики"
      case "guardrails":
        return "Guardrails"
      case "tasks":
        return "Задачи"
      default:
        return "Неизвестно"
    }
  }

  formatDate(dateString?: string): string {
    if (!dateString) return "Неизвестно"
    return this.formattingService.formatDate(dateString)
  }

  getValuePreview(value: unknown): string {
    if (typeof value === "string") {
      return value.length > 50 ? value.substring(0, 50) + "..." : value
    }
    if (Array.isArray(value)) {
      return `[${value.length} элементов]`
    }
    if (typeof value === "object" && value !== null) {
      return `{${Object.keys(value).length} полей}`
    }
    return String(value)
  }

  scrollToSection(key: string) {
    // В реальном приложении здесь будет прокрутка к секции
    console.log("Scroll to section:", key)
  }
}