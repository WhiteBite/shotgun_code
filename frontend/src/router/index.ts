import { createRouter, createWebHistory } from 'vue-router';
import ProjectScreen from '../views/ProjectScreen.vue';
import { useProjectStore } from '@/stores/projectStore';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'project',
      component: ProjectScreen
    },
    {
      path: '/workspace',
      name: 'workspace',
      component: () => import('../views/WorkspaceScreen.vue')
    },
    {
      path: '/review',
      name: 'review',
      component: () => import('../views/ReviewScreen.vue')
    }
  ]
});

router.beforeEach((to, _from, next) => {
  // Pinia needs to be initialized before this guard runs.
  // This happens in main.ts, so we can safely call it here.
  const projectStore = useProjectStore();

  if (to.name === 'project' && projectStore.isProjectLoaded) {
    // If user tries to access project selection but a project is already loaded,
    // redirect them to the workspace.
    next({ name: 'workspace' });
  } else if (to.name !== 'project' && !projectStore.isProjectLoaded) {
    // If user tries to access any other page without a loaded project,
    // redirect them to the project selection screen.
    next({ name: 'project' });
  } else {
    next();
  }
});


export default router;