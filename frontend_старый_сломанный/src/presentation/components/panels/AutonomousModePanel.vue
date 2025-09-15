<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAutonomousStore } from '@/stores/autonomous.store'
import { 
  CpuChipIcon, 
  CheckCircleIcon, 
  ExclamationCircleIcon 
} from '@/presentation/components/icons'
import LoadingSpinner from '@/presentation/components/shared/LoadingSpinner.vue'

// Use the autonomous store
const autonomousStore = useAutonomousStore()

// Get reactive refs from the store
const { 
  currentTask,
  taskStatus,
  tplPlan,
  reports,
  isLoading,
  error,
  isTaskRunning,
  canStartTask
} = storeToRefs(autonomousStore)

// Local refs for success criteria and constraints (not in store)
const successCriteria = ref('')
const constraints = ref('')

// Computed properties that were previously in the service
const executionStatus = computed(() => {
  if (!isTaskRunning.value) return 'Ready to start'
  return 'Executing pipeline...'
})

const canExecute = computed(() => currentTask.value.trim().length > 20)

const currentStep = computed(() => taskStatus.value?.currentStep || 0)
const totalSteps = computed(() => taskStatus.value?.totalSteps || 5)

// Mock pipeline steps since they're not in the store
const pipelineSteps = computed(() => [
  {
    id: 1,
    name: 'Analyze Goal',
    description: 'Breaking down the high-level goal into actionable tasks',
    status: (currentStep.value >= 1 && isTaskRunning.value) ? 'active' : (currentStep.value > 1 ? 'completed' : 'pending'),
    progress: currentStep.value >= 1 ? 25 : 0,
    duration: null
  },
  {
    id: 2,
    name: 'Plan Architecture',
    description: 'Designing the overall system architecture and approach',
    status: (currentStep.value >= 2 && isTaskRunning.value) ? 'active' : (currentStep.value > 2 ? 'completed' : 'pending'),
    progress: currentStep.value >= 2 ? 50 : 0,
    duration: null
  },
  {
    id: 3,
    name: 'Generate Code',
    description: 'Creating the necessary code components and implementations',
    status: (currentStep.value >= 3 && isTaskRunning.value) ? 'active' : (currentStep.value > 3 ? 'completed' : 'pending'),
    progress: currentStep.value >= 3 ? 75 : 0,
    duration: null
  },
  {
    id: 4,
    name: 'Run Tests',
    description: 'Executing tests and validating the implementation',
    status: (currentStep.value >= 4 && isTaskRunning.value) ? 'active' : (currentStep.value > 4 ? 'completed' : 'pending'),
    progress: currentStep.value >= 4 ? 90 : 0,
    duration: null
  },
  {
    id: 5,
    name: 'Final Review',
    description: 'Reviewing the complete implementation against success criteria',
    status: (currentStep.value >= 5 && isTaskRunning.value) ? 'active' : (currentStep.value > 5 ? 'completed' : 'pending'),
    progress: currentStep.value >= 5 ? 100 : 0,
    duration: null
  }
])

// Methods
function startExecution() {
  // This would trigger the actual autonomous execution
  console.log('Starting autonomous execution with task:', currentTask.value)
  // In a real implementation, this would call a method on the store
}

function pauseExecution() {
  console.log('Pausing execution')
  // In a real implementation, this would call a method on the store
}
</script>