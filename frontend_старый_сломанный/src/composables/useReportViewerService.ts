import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useReportsStore } from '@/stores/reports.store'
import { ReportViewerService } from '@/domain/services/ReportViewerService'
import { UIFormattingService } from '@/domain/services/UIFormattingService'

export function useReportViewerService() {
  // Get store
  const reportsStore = useReportsStore()
  
  // Get reactive refs
  const { isReportViewerOpen, selectedReport } = storeToRefs(reportsStore)
  
  // Create formatting service
  const formattingService = new UIFormattingService()
  
  // Create service instance
  const reportViewerService = new ReportViewerService(
    () => isReportViewerOpen.value,
    () => selectedReport.value,
    () => reportsStore.closeReportViewer(),
    (reportId) => reportsStore.loadReportData(reportId),
    formattingService
  )
  
  // Watch for report changes
  watch(
    () => selectedReport.value,
    () => {
      if (selectedReport.value) {
        reportViewerService.loadReport()
      }
    },
    { immediate: true }
  )
  
  return {
    reportViewerService,
    reportsStore
  }
}