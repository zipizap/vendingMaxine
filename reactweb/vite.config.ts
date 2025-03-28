import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/login': {
        target: 'http://127.0.0.1:8080', // Go server address
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, ''), // Ensure the /api prefix is not removed
      },
      '/logout': {
        target: 'http://127.0.0.1:8080', // Go server address
        changeOrigin: true,
      },
      '/callback': {
        target: 'http://127.0.0.1:8080', // Go server address
        changeOrigin: true,
      },
      '/api': {
        target: 'http://127.0.0.1:8080', // Go server address
        changeOrigin: true,
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true, // Ensure the directory is empty before building
  },
})
