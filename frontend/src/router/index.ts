import { createRouter, createWebHistory } from "vue-router";
import { useProjectStore } from "@/stores/project.store";
import ProjectSelectionView from "@/views/ProjectSelectionView.vue";

const router = createRouter({
  history: createWebHistory(), // убран import.meta.env.BASE_URL для совместимости типов
  routes: [
    {
      path: "/",
      name: "project-selection",
      component: ProjectSelectionView,
    },
    {
      path: "/workspace",
      name: "workspace",
      component: () => import("@/views/WorkspaceView.vue"),
    },
  ],
});

router.beforeEach((to, _from, next) => {
  const projectStore = useProjectStore();

  if (to.name !== "project-selection" && !projectStore.isProjectLoaded) {
    next({ name: "project-selection" });
  } else if (to.name === "project-selection" && projectStore.isProjectLoaded) {
    next({ name: "workspace" });
  } else {
    next();
  }
});

export default router;
