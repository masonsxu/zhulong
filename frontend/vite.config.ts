import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: process.env.VITE_DEV_SERVER_HOST || '0.0.0.0',
    port: parseInt(process.env.VITE_DEV_SERVER_PORT) || 5173,
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})