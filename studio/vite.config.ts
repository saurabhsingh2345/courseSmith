import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  base: "/",
  plugins: [react()],
  // The showcase page imports the real renderer components (../renderer/src)
  // and drives them with @remotion/player. Dedupe the shared singletons so
  // there's one React/Remotion instance, and allow the dev server to read the
  // sibling renderer directory.
  resolve: {
    dedupe: ["react", "react-dom", "remotion", "@remotion/player"],
  },
  server: {
    fs: {
      allow: [".."],
    },
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8787",
        changeOrigin: true,
      },
      "/artifacts": {
        target: "http://127.0.0.1:8787",
        changeOrigin: true,
      },
    },
  },
  test: {
    // Default to node; component/a11y specs opt into jsdom via a per-file
    // `// @vitest-environment jsdom` docblock.
    environment: "node",
    include: ["src/**/*.test.{ts,tsx}"],
  },
});
