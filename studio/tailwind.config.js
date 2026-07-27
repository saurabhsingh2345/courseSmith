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
        // The ink ramp is ordered by depth, not lightness (see tokens.ts), and
        // applyTheme() rewrites these vars per mode — so `bg-ink-900` is the
        // surface behind content in both themes and no page needs a `dark:`
        // variant. A region that must stay dark whatever the mode opts out with
        // `.surface-dark`.
        //
        // The fallback in each var() is the dark value, and it is what paints
        // the first frame: the stylesheet is in <head> and applyTheme() runs
        // from a module, so without it every ink colour resolves to nothing for
        // the frame before React boots and the page flashes unstyled. The
        // ink.test.ts drift test asserts these equal tokens.colors.ink[*].dark.
        ink: {
          950: "var(--ink-950, #09090b)",
          900: "var(--ink-900, #0e0e11)",
          850: "var(--ink-850, #131316)",
          800: "var(--ink-800, #18181c)",
          750: "var(--ink-750, #1e1e23)",
          700: "var(--ink-700, #26262c)",
          600: "var(--ink-600, #33333b)",
          500: "var(--ink-500, #7a7a86)",
          400: "var(--ink-400, #8e8e9a)",
          300: "var(--ink-300, #9d9da8)",
          200: "var(--ink-200, #c6c6cf)",
          100: "var(--ink-100, #e6e6ec)"
        }
      },
      animation: {
        "pulse-fast": "pulse 1.1s cubic-bezier(0.4, 0, 0.6, 1) infinite"
      }
    },
  },
  plugins: [tailwindcssAnimate],
};
