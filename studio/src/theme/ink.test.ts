// The ink ramp is the studio's working scale — `bg-ink-900`, `text-ink-100` —
// and it now exists in three places that have to agree: tokens.ts (the source),
// the :root fallback in index.css (so the first paint is not unstyled), and the
// .surface-dark opt-out. Nothing at build time connects them, so these tests do.
//
// The contrast assertions matter more than they look. The ramp's light values
// are derived by inverting a scale that was authored for a dark stage, and a
// derived colour's first appearance in a browser is far too late to discover it
// is grey-on-grey.

import {describe, expect, it} from 'vitest';
// @ts-expect-error -- plain JS config, no types, and adding a .d.ts for it would
// be more machinery than the one import is worth.
import tailwindConfig from '../../tailwind.config.js';
import {tokens} from './tokens';
import {cssVariables} from './applyTheme';

const STEPS = Object.keys(tokens.colors.ink);

const tailwindInk: Record<string, string> = (
  tailwindConfig as {theme: {extend: {colors: {ink: Record<string, string>}}}}
).theme.extend.colors.ink;

// --- WCAG relative luminance / contrast -------------------------------------

function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

describe('ink ramp', () => {
  it('every step is a 6-digit hex in both modes', () => {
    for (const [step, scale] of Object.entries(tokens.colors.ink)) {
      expect(scale.light, `ink-${step} light`).toMatch(/^#[0-9a-fA-F]{6}$/);
      expect(scale.dark, `ink-${step} dark`).toMatch(/^#[0-9a-fA-F]{6}$/);
    }
  });

  it('applyTheme emits a variable for every step, resolved per mode', () => {
    const light = cssVariables('light');
    const dark = cssVariables('dark');
    for (const step of STEPS) {
      expect(light[`--ink-${step}`]).toBe(tokens.colors.ink[step].light);
      expect(dark[`--ink-${step}`]).toBe(tokens.colors.ink[step].dark);
    }
  });

  it('is ordered by depth: 950 is the page, 100 is the type — in both modes', () => {
    // The ramp's whole premise. In dark, luminance rises from 950 to 100; in
    // light it falls. Either way the *roles* are stable, which is what lets one
    // class name be correct in both themes.
    const asc = [...STEPS].sort((a, b) => Number(b) - Number(a)); // 950 → 100
    const darkL = asc.map((s) => luminance(tokens.colors.ink[s].dark));
    const lightL = asc.map((s) => luminance(tokens.colors.ink[s].light));
    for (let i = 1; i < asc.length; i++) {
      expect(darkL[i], `dark ink-${asc[i]} vs ink-${asc[i - 1]}`).toBeGreaterThan(darkL[i - 1]);
      expect(lightL[i], `light ink-${asc[i]} vs ink-${asc[i - 1]}`).toBeLessThan(lightL[i - 1]);
    }
  });

  it.each(['light', 'dark'] as const)('%s mode meets WCAG AA for text on the page', (mode) => {
    const page = tokens.colors.ink['950'][mode];
    const surface = tokens.colors.ink['900'][mode];
    // 100-300 are type; 400/500 are the muted/secondary steps, and `text-ink-500`
    // is the single most-used colour in the studio. All must pass AA (4.5:1) on
    // both the page and the surface sitting on it.
    //
    // 600 and below are excluded because they are structure, not type:
    // borders, dividers and fills. That split is now real — the 28 places that
    // wrote text at ink-600 (empty states, hints, placeholders, "(optional)")
    // measured 1.6:1 and were moved to 500. Raising the 600 step instead would
    // have dragged `border-ink-600` and `bg-ink-600` with it, which is the
    // wrong fix for a text problem.
    for (const step of ['100', '200', '300', '400', '500']) {
      const c = tokens.colors.ink[step][mode];
      expect(contrast(c, page), `ink-${step} on page`).toBeGreaterThanOrEqual(4.5);
      expect(contrast(c, surface), `ink-${step} on surface`).toBeGreaterThanOrEqual(4.5);
    }
  });

  it('the tailwind fallback ramp is the dark ramp', () => {
    // These paint the frame before applyTheme runs. If they drift, that frame
    // is a different dark than the one installed a moment later — a flicker on
    // every load, and one nobody would think to look for.
    for (const step of STEPS) {
      const declared = tailwindInk[step];
      expect(declared, `ink-${step} missing from tailwind.config.js`).toBeTruthy();
      expect(declared, `ink-${step}`).toBe(
        `var(--ink-${step}, ${tokens.colors.ink[step].dark})`,
      );
    }
  });

  it('tailwind declares exactly the steps the ramp has', () => {
    // A step in one and not the other is a class that silently resolves to
    // nothing (`bg-ink-650`) or a token nothing can reach.
    expect(Object.keys(tailwindInk).sort()).toEqual([...STEPS].sort());
  });

  it('applyTheme emits an always-dark ramp for the .surface-dark opt-out', () => {
    // index.css maps --ink-N to these inside a terminal. A step missing here is
    // one line of a log pane rendering in the light theme's colour.
    for (const mode of ['light', 'dark'] as const) {
      const vars = cssVariables(mode);
      for (const step of STEPS) {
        expect(vars[`--ink-dark-${step}`], `--ink-dark-${step} in ${mode}`).toBe(
          tokens.colors.ink[step].dark,
        );
      }
    }
  });
});
