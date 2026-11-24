import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// API_TARGET はコンテナ内では compose のサービス名（http://backend:8080）を指す
const apiTarget = process.env.API_TARGET ?? 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': apiTarget,
    },
  },
})
