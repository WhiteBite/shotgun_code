
<template>
  <main class="flex-1 p-0 overflow-y-auto bg-white relative">
    <Step1PrepareContext v-if="steps.currentStep === 1" />
    <Step2ComposePrompt v-if="steps.currentStep === 2" />
    <Step3ExecutePrompt v-if="steps.currentStep === 3" />
    <Step4ApplyPatch v-if="steps.currentStep === 4" />
  </main>
</template>

<script setup>
import { useStepsStore } from '../stores/steps';
import Step1PrepareContext from './steps/Step1PrepareContext.vue';
import Step2ComposePrompt from './steps/Step2ComposePrompt.vue';
import Step3ExecutePrompt from './steps/Step3ExecutePrompt.vue';
import Step4ApplyPatch from './steps/Step4ApplyPatch.vue';

const steps = useStepsStore();
</script>