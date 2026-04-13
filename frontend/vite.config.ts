import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src")
    }
  },
  server: {
    port: 3000,
    // host: true binds to 0.0.0.0 so the dev server is reachable inside Docker.
    host: true,
    proxy: {
      "/api": {
        // In Docker the backend is reached by service name; locally by localhost.
        // Set BACKEND_URL=http://backend:8080 in docker-compose.override.yml.
        target: process.env.BACKEND_URL ?? "http://localhost:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "")
      }
    }
  }
});
