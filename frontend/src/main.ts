import { createApp } from "vue";
import { createPinia } from "pinia";
import piniaPluginPersistedState from "pinia-plugin-persistedstate";
import App from "./App.vue";
import router from "./router";

import "./assets/main.css";
import "./assets/scrollbars.css";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
import "highlight.js/styles/atom-one-dark.css";

const app = createApp(App);
const pinia = createPinia();
pinia.use(piniaPluginPersistedState);

app.use(pinia);
app.use(router);

app.mount("#app");
