
<template>
  <MainLayout />
</template>

<script setup>
import MainLayout from './components/MainLayout.vue';
import { useAppCoordinator } from './stores/appCoordinator';
import { onMounted } from 'vue';

onMounted(() => {
  // Запускаем координатор, который настроит взаимодействие между сторами
  useAppCoordinator();
});
</script>