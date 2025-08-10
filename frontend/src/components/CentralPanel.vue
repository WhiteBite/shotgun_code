<template>
  <main class="flex-1 p-0 overflow-y-auto bg-white relative">
    <Step1PrepareContext v-if="steps.currentStep === 1" />
    <Step2ReviewAndApply v-if="steps.currentStep === 2" />
  </main>
</template>

<script setup>
import { useStepsStore } from '../stores/stepsStore.js';
import Step1PrepareContext from './steps/Step1PrepareContext.vue';
import Step2ReviewAndApply from './steps/Step2ReviewAndApply.vue';

const steps = useStepsStore();
</script>