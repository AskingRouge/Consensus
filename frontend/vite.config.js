import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Fixed port so it always matches the backend's FRONTEND_URL (used
    // for CORS and the WebSocket Origin check).
    port: 5173,
    strictPort: true,
  },
})
