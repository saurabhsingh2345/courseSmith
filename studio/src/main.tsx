import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { App } from "./App";
import { RunProvider } from "./state/RunContext";
import { ShortcutProvider } from "./state/ShortcutContext";
import { applyTheme, preferredMode } from "./theme/applyTheme";
import "./index.css";

// Emit design tokens as CSS variables + set the light/dark mode before paint.
applyTheme(preferredMode());

const rootEl = document.getElementById("root");
if (!rootEl) throw new Error("#root element not found");

createRoot(rootEl).render(
  <StrictMode>
    <BrowserRouter>
      <ShortcutProvider>
        <RunProvider>
          <App />
        </RunProvider>
      </ShortcutProvider>
    </BrowserRouter>
  </StrictMode>,
);
