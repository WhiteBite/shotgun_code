import { createRouter, createWebHistory } from "vue-router";
import ProjectSelectionView from "@/presentation/views/ProjectSelectionView.vue";
import { useUiStore } from "@/stores/ui.store";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "project-selection",
      component: ProjectSelectionView,
    },
    {
      path: "/workspace",
      name: "workspace",
      component: () => import("@/presentation/views/WorkspaceView.vue"),
    },
  ],
});

// Закрываем контекстное меню и QuickLook при смене маршрута
router.beforeEach((_to, _from, next) => {
  try {
    const ui = useUiStore();
    ui.closeContextMenu();
    ui.hideQuickLook();
  } catch (error) {
    // Handle navigation errors gracefully
    console.error('Navigation error:', error)
  }
  next();
});

export default router;