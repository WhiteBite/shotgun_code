import { createRouter, createWebHistory } from 'vue-router';
import { useProjectStore } from '@/stores/project.store';
import ProjectSelectionView from '@/views/ProjectSelectionView.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'project-selection',
      component: ProjectSelectionView,
    },
    {
      path: '/workspace',
      name: 'workspace',
      component: () => import('@/views/WorkspaceView.vue'),
    },
  ],
});

router.beforeEach((to, from, next) => {
  const projectStore = useProjectStore();

  if (to.name !== 'project-selection' && !projectStore.isProjectLoaded) {
    // If trying to access workspace without a project, redirect to selection.
    next({ name: 'project-selection' });
  } else if (to.name === 'project-selection' && projectStore.isProjectLoaded) {
    // If trying to access selection with a project loaded, go to workspace.
    next({ name: 'workspace' });
  } else {
    next();
  }
});

export default router;