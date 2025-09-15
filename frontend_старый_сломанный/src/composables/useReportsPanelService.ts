import { storeToRefs } from 'pinia'
import { useReportsStore } from '@/stores/reports.store'
import { ReportsPanelService } from '@/domain/services/ReportsPanelService'

export function useReportsPanelService() {
  // Get stores
  const reportsStore = useReportsStore()
  
  // Get reactive refs from stores
  const { 
    reports,
    isLoading,
    selectedReport
  } = storeToRefs(reportsStore)
  
  // Create service instance
  const reportsPanelService = new ReportsPanelService()
  
  return {
    reportsPanelService,
    reportsStore,
    // Store refs
    reports,
    isLoading,
    selectedReport
  }
}