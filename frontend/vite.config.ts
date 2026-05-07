import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // outDir is consumed by //go:embed in pkg/web/web.go;
  // emptyOutDir is required because it lives outside the project root.
  build: {
    outDir: "../pkg/web/dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        // Optionally remove /api prefix if your backend does not expect it:
        // rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
