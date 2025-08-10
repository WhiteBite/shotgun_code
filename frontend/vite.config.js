import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { defineProject } from 'vitest/config'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  test: {
    // enable vitest's browser-like APIs
    globals: true,
    // to simulate DOM
    environment: 'jsdom',
  },
})