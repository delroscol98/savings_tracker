import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    react({
      include: [".js", ".jsx", ".ts", ".tsx"],
      exclude: ["node_modules/"],
      jsxImportSource: "react",
      jsxRuntime: "automatic",
      reactRefreshHost: "localhost",
    }),
    tailwindcss({
      optimize: false,
    }),
  ],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  server: {
    proxy: {
      "/api": { target: "http://localhost:8080", changeOrigin: true },
    },
  },
});
