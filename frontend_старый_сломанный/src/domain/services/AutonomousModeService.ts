import { ref, computed, type Ref } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export interface PipelineStep {
  id: number
  name: string
  description: string
  status: 'pending' | 'active' | 'completed' | 'error'
  progress: number | null
  duration: string | null
}

export class AutonomousModeService {
  // Form state
  private _goal: Ref<string> = ref('')
  private _successCriteria: Ref<string> = ref('')
  private _constraints: Ref<string> = ref('')

  // Execution state
  private _isExecuting: Ref<boolean> = ref(false)
  private _currentStep: Ref<number> = ref(0)
  private _totalSteps: Ref<number> = ref(APP_CONFIG.autonomous.slaPolicies.standard.maxConcurrentTasks)

  // Pipeline steps
  private _pipelineSteps: Ref<PipelineStep[]> = ref([
    {
      id: 1,
      name: 'Analyze Goal',
      description: 'Breaking down the high-level goal into actionable tasks',
      status: 'pending',
      progress: 0,
      duration: null
    },
    {
      id: 2,
      name: 'Plan Architecture',
      description: 'Designing the overall system architecture and approach',
      status: 'pending',
      progress: 0,
      duration: null
    },
    {
      id: 3,
      name: 'Generate Code',
      description: 'Creating the necessary code components and implementations',
      status: 'pending',
      progress: 0,
      duration: null
    },
    {
      id: 4,
      name: 'Run Tests',
      description: 'Executing tests and validating the implementation',
      status: 'pending',
      progress: 0,
      duration: null
    },
    {
      id: 5,
      name: 'Final Review',
      description: 'Reviewing the complete implementation against success criteria',
      status: 'pending',
      progress: 0,
      duration: null
    }
  ])

  // Computed properties
  get executionStatus() {
    if (!this._isExecuting.value) return 'Ready to start'
    return 'Executing pipeline...'
  }

  get canExecute() {
    return this._goal.value.trim().length > 20
  }

  // Reactive state accessors
  get goal() {
    return this._goal.value
  }

  set goal(value: string) {
    this._goal.value = value
  }

  get successCriteria() {
    return this._successCriteria.value
  }

  set successCriteria(value: string) {
    this._successCriteria.value = value
  }

  get constraints() {
    return this._constraints.value
  }

  set constraints(value: string) {
    this._constraints.value = value
  }

  get isExecuting() {
    return this._isExecuting.value
  }

  get currentStep() {
    return this._currentStep.value
  }

  get totalSteps() {
    return this._totalSteps.value
  }

  get pipelineSteps() {
    return this._pipelineSteps.value
  }

  // Action handlers
  startExecution() {
    console.log('Starting autonomous execution with goal:', this._goal.value)
    this._isExecuting.value = true
    this._currentStep.value = 1

    // Reset all steps
    this._pipelineSteps.value.forEach(step => {
      step.status = 'pending'
      step.progress = 0
      step.duration = null
    })

    // Set first step as active
    if (this._pipelineSteps.value.length > 0) {
      this._pipelineSteps.value[0].status = 'active'
      this._pipelineSteps.value[0].progress = 25
    }

    // This would trigger the actual autonomous execution
  }

  pauseExecution() {
    console.log('Pausing execution')
    this._isExecuting.value = false

    // Reset active steps
    this._pipelineSteps.value.forEach(step => {
      if (step.status === 'active') {
        step.status = 'pending'
        step.progress = 0
      }
    })
  }

  // Service methods for updating step status
  updateStepStatus(stepId: number, status: PipelineStep['status'], progress?: number, duration?: string) {
    const step = this._pipelineSteps.value.find(s => s.id === stepId)
    if (step) {
      step.status = status
      if (progress !== undefined) step.progress = progress
      if (duration !== undefined) step.duration = duration
    }
  }

  completeStep(stepId: number, duration?: string) {
    this.updateStepStatus(stepId, 'completed', 100, duration)
  }

  failStep(stepId: number, duration?: string) {
    this.updateStepStatus(stepId, 'error', null, duration)
  }

  activateStep(stepId: number) {
    // Set all steps to pending first
    this._pipelineSteps.value.forEach(step => {
      if (step.id !== stepId) {
        step.status = 'pending'
      }
    })
    
    // Activate the specified step
    this.updateStepStatus(stepId, 'active', 0)
    this._currentStep.value = stepId
  }
}