import { createApp } from "vue";
import { createPinia, setActivePinia } from "pinia";
import { autoAnimatePlugin } from "@formkit/auto-animate/vue";

import App from "./App.vue";
import router from "./router";
import tooltipDirective from "./directives/tooltip";
import smartTooltipDirective from "./directives/smartTooltip";

import "./assets/main.css";
import { useNotificationsStore } from "@/stores/notifications.store";
import { eventService } from "@/infrastructure/events/event.service";
import { useProjectStore } from "@/stores/project.store"; // Add project store import
import 'highlight.js/styles/github-dark.css';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
setActivePinia(pinia);
app.use(router);
app.use(autoAnimatePlugin);
app.use(tooltipDirective);
app.use(smartTooltipDirective);

// Initialize stores after Pinia is properly set up
const projectStore = useProjectStore();
projectStore.initialize().then(async () => {
  // In development mode, try to auto-load current directory
  if (process.env.NODE_ENV === 'development') {
    try {
      await projectStore.tryAutoOpenLastProject();
    } catch (error) {
      console.warn('Failed to auto-load project on startup:', error);
    }
  }
}).catch(error => {
  console.error('Failed to initialize project store:', error);
});

// Initialize EventService with a delay to ensure Wails runtime is fully loaded
// Only initialize if we're in a Wails environment
const isWailsEnvironment = typeof window !== 'undefined' &&
                          typeof (window as any).runtime !== 'undefined' &&
                          typeof (window as any).runtime.EventsOnMultiple !== 'undefined';

if (isWailsEnvironment) {
  // Use a more robust initialization with retries
  const initializeEventService = async () => {
    try {
      await eventService.initialize();
    } catch (error) {
      console.warn('EventService initialization failed, continuing without events:', error);
    }
  };

  setTimeout(() => {
    initializeEventService();
  }, 100); // 100ms delay to ensure Wails is ready
} else {
  console.log('⚠️  Not in Wails environment, skipping EventService initialization');
}

// Глобальный обработчик ошибок
app.config.errorHandler = (error, instance, info) => {
  console.error("Global error handler:", error, info);

  // Получаем стор уведомлений для отображения ошибки
  const notifications = useNotificationsStore();
  const errorMessage = error instanceof Error ? error.message : String(error);
  notifications.addLog(`Application error: ${errorMessage}`, "error");

  // Дополнительная информация для разработки
  if (process.env.NODE_ENV === "development") {
    console.error("Error details:", { error, instance, info });
  }
};

// Обработчик необработанных промисов
window.addEventListener("unhandledrejection", (event) => {
  console.error("Unhandled promise rejection:", event.reason);

  const notifications = useNotificationsStore();
  const errorMessage =
    event.reason instanceof Error ? event.reason.message : String(event.reason);
  notifications.addLog(`Unhandled promise rejection: ${errorMessage}`, "error");

  // Предотвращаем стандартную обработку ошибки
  event.preventDefault();
});

app.mount("#app");
