import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '#wailsjs': resolve(__dirname, 'wailsjs')
    }
  },
  server: {
    port: 3000,
    host: true,
    hmr: {
      overlay: false
    },
    // CRITICAL: File watcher causes memory leak with Wails
    // Only watch src/ directory, ignore wailsjs/ which Wails constantly regenerates
    watch: {
      usePolling: false,
      ignored: [
        '**/node_modules/**',
        '**/.git/**',
        '**/dist/**',
        '**/wailsjs/**',  // Main culprit - Wails regenerates bindings constantly
        '**/backend/**',
        '**/*.go'
      ]
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: process.env.NODE_ENV === 'production',
        drop_debugger: true
      }
    },
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Vue core
          if (id.includes('node_modules/vue') || id.includes('node_modules/pinia')) {
            return 'vue-vendor'
          }
          // UI libraries
          if (id.includes('@headlessui/vue') || id.includes('@floating-ui') || id.includes('fuse.js')) {
            return 'ui-vendor'
          }
          // Utilities
          if (id.includes('@vueuse/core') || id.includes('clsx') || id.includes('tailwind-merge') || id.includes('nanoid')) {
            return 'utils'
          }
          // Virtual scroller
          if (id.includes('vue-virtual-scroller')) {
            return 'virtual-scroller'
          }
          // Wails bindings - separate chunk
          if (id.includes('wailsjs')) {
            return 'wails-bindings'
          }
          // Heavy features - lazy loaded
          if (id.includes('features/ai-chat')) {
            return 'feature-ai-chat'
          }
          if (id.includes('features/git') || id.includes('GitSourceSelector')) {
            return 'feature-git'
          }
          if (id.includes('MemoryDashboard') || id.includes('ProjectStructurePanel')) {
            return 'feature-tools'
          }
          // Split files feature into smaller chunks
          if (id.includes('features/files')) {
            // Modals - lazy loaded on demand
            if (id.includes('Modal.vue') || id.includes('IgnoreRules') || id.includes('SelectionPresets') || id.includes('FilterSettings')) {
              return 'feature-files-modals'
            }
            // Quick filters system
            if (id.includes('QuickFilters') || id.includes('FilterDropdown') || id.includes('FilterChip') || id.includes('useQuickFilters') || id.includes('useSmartFilters') || id.includes('useFilterDropdown') || id.includes('useFilterPersistence')) {
              return 'feature-files-filters'
            }
            // Virtual tree (heavy component)
            if (id.includes('Virtual') || id.includes('FileTreeNode')) {
              return 'feature-files-tree'
            }
            // File explorer main component
            if (id.includes('FileExplorer') || id.includes('useFileExplorer') || id.includes('BreadcrumbsNav') || id.includes('FileContextMenu')) {
              return 'feature-files-explorer'
            }
            // Core: store, types, utils
            return 'feature-files-core'
          }
        },
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    },
    chunkSizeWarningLimit: 500
  },
  optimizeDeps: {
    include: ['vue', 'pinia', '@vueuse/core', '@headlessui/vue', 'fuse.js'],
    exclude: ['#wailsjs']
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts']
  }
})
