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
    watch: {
      ignored: [
        '**/node_modules/**',
        '**/.git/**',
        '**/dist/**',
        '**/backend/**',
        '**/*.go',
        '**/*.sum',
        '**/*.mod',
        '**/build/**'
      ]
    },
    // Reduce memory by not pre-warming
    warmup: {
      clientFiles: []
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false, // Disabled - causes memory leaks in dev
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: process.env.NODE_ENV === 'production', // Remove console.log in production
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.debug'] // Remove specific console methods
      }
    },
    rollupOptions: {
      output: {
        // Manual chunking for better code splitting
        manualChunks: {
          // Vue core
          'vue-vendor': ['vue', 'pinia'],

          // UI libraries
          'ui-vendor': [
            '@headlessui/vue',
            '@floating-ui/dom',
            'fuse.js'
          ],

          // Utilities
          'utils': [
            '@vueuse/core',
            'clsx',
            'tailwind-merge',
            'nanoid'
          ],

          // Virtual scrolling (if used)
          'virtual-scroller': ['vue-virtual-scroller']
        },

        // Asset file names
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    },

    // Chunk size warnings
    chunkSizeWarningLimit: 500 // Warning if chunk > 500kb
  },

  // Optimized dependency pre-bundling
  optimizeDeps: {
    include: [
      'vue',
      'pinia',
      '@vueuse/core',
      '@headlessui/vue',
      'fuse.js'
    ],
    // Reduce memory usage
    holdUntilCrawlEnd: false,
    // Don't force re-optimization
    force: false
  },

  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts']
  }
})