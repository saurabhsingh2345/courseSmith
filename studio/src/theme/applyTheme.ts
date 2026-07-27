// Emits the design tokens as CSS custom properties on :root, resolving the
// per-theme (light/dark) values for the active mode. Called once at boot and
// again whenever the mode toggles — the single place TS tokens become CSS.

import {tokens, type ColorScale} from './tokens';

export type ThemeMode = 'light' | 'dark';

const pick = (c: ColorScale, mode: ThemeMode): string => c[mode];

/** Build the flat --var → value map for a given mode. Pure; unit-testable. */
export function cssVariables(mode: ThemeMode): Record<string, string> {
  const c = tokens.colors;
  const vars: Record<string, string> = {
    '--color-brand': mode === 'dark' ? c.brand.dark : c.brand.light,
    '--color-brand-saturated': c.brand.saturated,
    '--color-success': c.semantic.success,
    '--color-error': c.semantic.error,
    '--color-warning': c.semantic.warning,
    '--color-info': c.semantic.info,
    '--color-bg': pick(c.neutral.bg, mode),
    '--color-surface': pick(c.neutral.surface, mode),
    '--color-border': pick(c.neutral.border, mode),
    '--color-text': pick(c.neutral.text, mode),
    '--color-muted': pick(c.neutral.muted, mode),
    '--font-mono': tokens.typography.mono.family,
    // Motion, exposed for CSS transitions / Framer Motion consumers.
    '--motion-fast': `${tokens.motion.timing.fast}s`,
    '--motion-normal': `${tokens.motion.timing.normal}s`,
    '--motion-slow': `${tokens.motion.timing.slow}s`,
    '--ease-entrance': tokens.motion.easing.entrance,
    '--ease-exit': tokens.motion.easing.exit,
    '--ease-subtle': tokens.motion.easing.subtle,
  };
  // The ink ramp, twice.
  //
  // `--ink-N` is the themed ramp: it is what `bg-ink-900` resolves to, and it
  // inverts with the mode (see the note on tokens.ink). `--ink-dark-N` is the
  // dark ramp unconditionally, and exists so the `.surface-dark` opt-out in
  // index.css can be written as a dozen `--ink-N: var(--ink-dark-N)` mappings
  // with no colour in it. A stylesheet that repeated the hexes would be a
  // second copy of the ramp, and the failure it produces — one step forgotten,
  // so a single line of a log pane renders in the light theme's colour — is
  // exactly the kind nobody notices until it ships.
  for (const [step, scale] of Object.entries(c.ink)) {
    vars[`--ink-${step}`] = pick(scale, mode);
    vars[`--ink-dark-${step}`] = scale.dark;
  }
  for (const [k, v] of Object.entries(tokens.spacing)) {
    vars[`--space-${k}`] = v;
  }
  for (const [k, v] of Object.entries(tokens.radius)) {
    vars[`--radius-${k}`] = v;
  }
  return vars;
}

/** Apply the token variables + color-scheme for `mode` to the document root. */
export function applyTheme(mode: ThemeMode, root: HTMLElement = document.documentElement): void {
  const vars = cssVariables(mode);
  for (const [k, v] of Object.entries(vars)) {
    root.style.setProperty(k, v);
  }
  root.style.colorScheme = mode;
  root.classList.toggle('dark', mode === 'dark');
  root.dataset.theme = mode;
}

/** Respect a stored choice, else the OS preference. */
export function preferredMode(): ThemeMode {
  const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('cs-theme') : null;
  if (stored === 'light' || stored === 'dark') return stored;
  if (typeof matchMedia !== 'undefined' && matchMedia('(prefers-color-scheme: light)').matches) {
    return 'light';
  }
  return 'dark';
}
