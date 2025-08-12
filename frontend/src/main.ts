import { createApp } from 'vue';
import { createPinia } from 'pinia';
import piniaPluginPersistedState from 'pinia-plugin-persistedstate';
import App from './App.vue';
import router from './router';
import { eventService } from './services/event.service';
import { useKeyboardShortcuts } from './composables/useKeyboardShortcuts';

import './assets/main.css';
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css';

const app = createApp(App);
const pinia = createPinia();
pinia.use(piniaPluginPersistedState);

app.use(pinia);
app.use(router);

// Initialize event listener service after Pinia is ready
eventService.initialize();

// Mount the app, then initialize things that need the DOM or component context
app.mount('#app');

// Initialize global shortcuts. This composable will handle its own lifecycle.
useKeyboardShortcuts();