import { createApp } from "vue";
import { createPinia } from "pinia";
import { autoAnimatePlugin } from "@formkit/auto-animate/vue";

import App from "./App.vue";
import "./assets/main.css";
import "./assets/design-system.css";
import { useUIStore } from "@/stores/ui.store";
import 'highlight.js/styles/github-dark.css';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(autoAnimatePlugin);

// Глобальный обработчик ошибок
app.config.errorHandler = (error, instance, info) => {
  console.error("Global error handler:", error, info);

  // Получаем стор уведомлений для отображения ошибки
  const uiStore = useUIStore();
  const errorMessage = error instanceof Error ? error.message : String(error);
  uiStore.addToast(`Application error: ${errorMessage}`, "error");

  // Дополнительная информация для разработки
  if (process.env.NODE_ENV === "development") {
    console.error("Error details:", { error, instance, info });
  }
};

// Обработчик необработанных промисов
window.addEventListener("unhandledrejection", (event) => {
  console.error("Unhandled promise rejection:", event.reason);

  const uiStore = useUIStore();
  const errorMessage =
    event.reason instanceof Error ? event.reason.message : String(event.reason);
  uiStore.addToast(`Unhandled promise rejection: ${errorMessage}`, "error");

  // Предотвращаем стандартную обработку ошибки
  event.preventDefault();
});

app.mount("#app");