import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAutonomousStore } from '@/stores/autonomous.store'
import { useProjectStore } from '@/stores/project.store'
import { useNotificationsStore } from '@/stores/notifications.store'
import { AutonomousControlService } from '@/domain/services/AutonomousControlService'
import { UIFormattingService } from '@/domain/services/UIFormattingService'

export function useAutonomousControlService() {
  // Get stores
  const autonomousStore = useAutonomousStore()
  const projectStore = useProjectStore()
  const notificationsStore = useNotificationsStore()
  
  // Get reactive refs
  const { 
    currentTask,
    slaPolicy,
    isTaskRunning,
    isLoading,
    taskStatus,
    tplPlan,
    reports,
    error,
    canStartTask,
    canCancelTask
  } = storeToRefs(autonomousStore)
  
  // Create formatting service
  const formattingService = new UIFormattingService()
  
  // Create service instance
  const autonomousControlService = new AutonomousControlService(
    () => currentTask.value,
    (task) => { autonomousStore.currentTask = task },
    () => slaPolicy.value,
    (policy) => { autonomousStore.slaPolicy = policy },
    () => isTaskRunning.value,
    () => isLoading.value,
    () => taskStatus.value,
    () => tplPlan.value,
    () => reports.value,
    () => error.value,
    () => canStartTask.value,
    () => canCancelTask.value,
    (projectPath) => autonomousStore.startTask(projectPath),
    () => autonomousStore.cancelCurrentTask(),
    () => autonomousStore.clearError(),
    formattingService
  )
  
  // Additional computed properties that depend on multiple stores
  const selectedSlaPolicy = computed(() =>
    autonomousControlService.slaPolicies.find(p => p.value === autonomousControlService.slaPolicy) || 
    autonomousControlService.slaPolicies[1]
  )
  
  return {
    autonomousControlService,
    autonomousStore,
    projectStore,
    notificationsStore,
    selectedSlaPolicy
  }
}