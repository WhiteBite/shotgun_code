import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { defineConfig } from 'vite'

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
      overlay: false // Disable error overlay to reduce memory usage
    },
    watch: {
      // Reduce file watching overhead
      ignored: ['**/node_modules/**', '**/.git/**', '**/dist/**']
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: process.env.NODE_ENV !== 'production', // Sourcemaps only in dev
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
    exclude: [
      // Exclude any packages that should not be pre-bundled
    ]
  },

  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts']
  }
})