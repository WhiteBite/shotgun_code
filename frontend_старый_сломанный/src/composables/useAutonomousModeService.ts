import { computed } from 'vue'
import { AutonomousModeService } from '@/domain/services/AutonomousModeService'

export function useAutonomousModeService() {
  // Create service instance
  const autonomousModeService = new AutonomousModeService()
  
  // Computed properties that were previously in the component
  const executionStatus = computed(() => autonomousModeService.executionStatus)
  const canExecute = computed(() => autonomousModeService.canExecute)
  
  return {
    autonomousModeService,
    executionStatus,
    canExecute
  }
}