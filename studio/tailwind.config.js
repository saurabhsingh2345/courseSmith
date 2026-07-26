import tailwindcssAnimate from "tailwindcss-animate";

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "Roboto",
          "Helvetica Neue",
          "Arial",
          "sans-serif",
        ],
        mono: [
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Consolas",
          "Liberation Mono",
          "monospace",
        ],
      },
      colors: {
        // Semantic aliases resolve to the CSS custom properties emitted by
        // applyTheme(). Because applyTheme rewrites the same --color-* vars per
        // mode (and toggles the `.dark` class), `bg-brand`/`text-fg` follow the
        // active theme automatically — no dark: variants needed for colour.
        brand: {
          DEFAULT: "var(--color-brand)",
          saturated: "var(--color-brand-saturated)",
        },
        success: "var(--color-success)",
        error: "var(--color-error)",
        warning: "var(--color-warning)",
        info: "var(--color-info)",
        bg: "var(--color-bg)",
        surface: "var(--color-surface)",
        border: "var(--color-border)",
        fg: "var(--color-text)",
        muted: "var(--color-muted)",
        ink: {
          950: "#09090b",
          900: "#0e0e11",
          850: "#131316",
          800: "#18181c",
          750: "#1e1e23",
          700: "#26262c",
          600: "#33333b",
          500: "#4c4c56",
          400: "#71717c",
          300: "#9d9da8",
          200: "#c6c6cf",
          100: "#e6e6ec"
        }
      },
      animation: {
        "pulse-fast": "pulse 1.1s cubic-bezier(0.4, 0, 0.6, 1) infinite"
      }
    },
  },
  plugins: [tailwindcssAnimate],
};
