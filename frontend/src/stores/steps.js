
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useStepsStore = defineStore('steps', () => {
  const currentStep = ref(1);
  const steps = ref([
    { id: 1, title: 'Prepare Context', completed: false },
    { id: 2, title: 'Compose Prompt', completed: false },
    { id: 3, title: 'Execute Prompt', completed: false },
    { id: 4, title: 'Apply Patch', completed: false },
  ]);

  function navigateToStep(stepId) {
    const targetStep = steps.value.find(s => s.id === stepId);
    if (!targetStep) return;

    // Разрешаем навигацию на любой завершенный шаг или на следующий незавершенный
    const firstUncompleted = steps.value.find(s => !s.completed);
    if (targetStep.completed || (firstUncompleted && stepId === firstUncompleted.id)) {
      currentStep.value = stepId;
    }
  }

  function completeStep(stepId) {
    const step = steps.value.find(s => s.id === stepId);
    if (step && !step.completed) {
      step.completed = true;
    }
  }

  function resetSteps() {
    steps.value.forEach(s => s.completed = false);
    currentStep.value = 1;
  }

  return {
    currentStep,
    steps,
    navigateToStep,
    completeStep,
    resetSteps,
  };
});