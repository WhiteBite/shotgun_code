<template>
  <div class="h-full flex flex-col p-6">
    <!-- Autonomous Mode Header -->
    <div class="mb-6">
      <div class="flex items-center space-x-3 mb-2">
        <div class="w-8 h-8 bg-purple-600 rounded-lg flex items-center justify-center">
          <CpuChipIcon class="w-5 h-5 text-white" />
        </div>
        <h2 class="text-xl font-semibold text-gray-100">Autonomous Mode</h2>
      </div>
      <p class="text-gray-400">Define high-level goals and let AI handle the implementation</p>
    </div>
    
    <!-- Goal Definition Section -->
    <div class="mb-6">
      <h3 class="text-lg font-medium text-gray-200 mb-3">Goal Definition</h3>
      
      <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-300 mb-2">
            High-Level Goal
          </label>
          <textarea
            v-model="goal"
            class="w-full h-24 p-3 bg-gray-700 text-gray-200 rounded-lg border border-gray-600 placeholder-gray-500 resize-none focus:outline-none focus:ring-2 focus:ring-purple-500"
            placeholder="Describe the overall objective you want to achieve...

Example:
- Build a complete user authentication system
- Refactor the codebase for better maintainability
- Implement a real-time chat feature"
          />
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Success Criteria
            </label>
            <textarea
              v-model="successCriteria"
              class="w-full h-20 p-3 bg-gray-700 text-gray-200 rounded-lg border border-gray-600 placeholder-gray-500 resize-none focus:outline-none focus:ring-2 focus:ring-purple-500"
              placeholder="Define what success looks like..."
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">
              Constraints
            </label>
            <textarea
              v-model="constraints"
              class="w-full h-20 p-3 bg-gray-700 text-gray-200 rounded-lg border border-gray-600 placeholder-gray-500 resize-none focus:outline-none focus:ring-2 focus:ring-purple-500"
              placeholder="Any limitations or requirements..."
            />
          </div>
        </div>
      </div>
    </div>
    
    <!-- Execution Pipeline Section -->
    <div class="flex-1 flex flex-col min-h-0">
      <h3 class="text-lg font-medium text-gray-200 mb-3">Execution Pipeline</h3>
      
      <div class="bg-gray-800 rounded-lg border border-gray-700 flex-1 flex flex-col">
        <div class="p-4 border-b border-gray-700">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-4">
              <div class="flex items-center space-x-2">
                <div class="w-3 h-3 bg-purple-500 rounded-full animate-pulse" />
                <span class="text-sm text-gray-300">{{ executionStatus }}</span>
              </div>
              
              <div class="text-sm text-gray-400">
                Step {{ currentStep }} of {{ totalSteps }}
              </div>
            </div>
            
            <div class="flex items-center space-x-2">
              <button
                v-if="!isExecuting"
                class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium"
                :disabled="!canExecute"
                @click="startExecution"
              >
                Start Execution
              </button>
              
              <button
                v-else
                class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium"
                @click="pauseExecution"
              >
                Pause
              </button>
            </div>
          </div>
        </div>
        
        <!-- Pipeline Steps -->
        <div class="flex-1 p-4 overflow-auto">
          <div class="space-y-3">
            <div
              v-for="step in pipelineSteps"
              :key="step.id"
              class="flex items-center space-x-3 p-3 rounded-lg transition-colors"
              :class="[
                step.status === 'completed' ? 'bg-green-900/20 border border-green-700' :
                step.status === 'active' ? 'bg-purple-900/20 border border-purple-700' :
                step.status === 'error' ? 'bg-red-900/20 border border-red-700' :
                'bg-gray-700'
              ]"
            >
              <!-- Step Icon -->
              <div class="flex-shrink-0">
                <CheckCircleIcon
                  v-if="step.status === 'completed'"
                  class="w-5 h-5 text-green-400"
                />
                <LoadingSpinner
                  v-else-if="step.status === 'active'"
                  size="sm"
                  color="primary"
                />
                <ExclamationCircleIcon
                  v-else-if="step.status === 'error'"
                  class="w-5 h-5 text-red-400"
                />
                <div
                  v-else
                  class="w-5 h-5 bg-gray-600 rounded-full"
                />
              </div>
              
              <!-- Step Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-medium text-gray-200">{{ step.name }}</h4>
                  <span class="text-xs text-gray-400">{{ step.duration || 'Pending' }}</span>
                </div>
                <p class="text-xs text-gray-400 mt-1">{{ step.description }}</p>
                
                <!-- Progress bar for active step -->
                <div v-if="step.status === 'active'" class="mt-2">
                  <div class="w-full h-1 bg-gray-600 rounded-full overflow-hidden">
                    <div 
                      class="h-full bg-purple-500 transition-all duration-500"
                      :style="{ width: `${step.progress || 0}%` }"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  CpuChipIcon, 
  CheckCircleIcon, 
  ExclamationCircleIcon 
} from '@heroicons/vue/24/outline'
import LoadingSpinner from '@/presentation/components/shared/LoadingSpinner.vue'

// Form state
const goal = ref('')
const successCriteria = ref('')
const constraints = ref('')

// Execution state
const isExecuting = ref(false)
const currentStep = ref(0)
const totalSteps = ref(5)

const executionStatus = computed(() => {
  if (!isExecuting.value) return 'Ready to start'
  return 'Executing pipeline...'
})

const canExecute = computed(() => {
  return goal.value.trim().length > 20
})

// Pipeline steps
const pipelineSteps = ref([
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

const startExecution = () => {
  console.log('Starting autonomous execution with goal:', goal.value)
  isExecuting.value = true
  currentStep.value = 1
  
  // Simulate execution progress
  pipelineSteps.value[0].status = 'active'
  pipelineSteps.value[0].progress = 25
  
  // This would trigger the actual autonomous execution
}

const pauseExecution = () => {
  console.log('Pausing execution')
  isExecuting.value = false
  
  // Reset active steps
  pipelineSteps.value.forEach(step => {
    if (step.status === 'active') {
      step.status = 'pending'
      step.progress = 0
    }
  })
}
</script>