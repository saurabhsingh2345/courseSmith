import {describe, expect, it} from 'vitest';
import {tokens} from './tokens';
import {cssVariables, preferredMode} from './applyTheme';
import {motion} from './motionTokens';

const HEX = /^#[0-9a-fA-F]{6}$/;

describe('design tokens', () => {
  it('every colour is a valid 6-digit hex', () => {
    const hexes = [
      tokens.colors.brand.light,
      tokens.colors.brand.dark,
      tokens.colors.brand.saturated,
      ...Object.values(tokens.colors.semantic),
      ...Object.values(tokens.colors.neutral).flatMap((s) => [s.light, s.dark]),
    ];
    for (const h of hexes) expect(h, h).toMatch(HEX);
  });

  it('typography sizes carry units', () => {
    for (const t of [tokens.typography.heading1, tokens.typography.heading2, tokens.typography.body]) {
      expect(t.size).toMatch(/rem|px|em$/);
      expect(t.lineHeight).toBeGreaterThan(1);
    }
  });
});

describe('cssVariables', () => {
  it('light and dark resolve neutral surfaces differently', () => {
    const light = cssVariables('light');
    const dark = cssVariables('dark');
    expect(light['--color-bg']).not.toBe(dark['--color-bg']);
    expect(light['--color-text']).not.toBe(dark['--color-text']);
  });

  it('brand var swaps per mode but saturated stays fixed', () => {
    expect(cssVariables('light')['--color-brand']).toBe(tokens.colors.brand.light);
    expect(cssVariables('dark')['--color-brand']).toBe(tokens.colors.brand.dark);
    expect(cssVariables('light')['--color-brand-saturated']).toBe(
      cssVariables('dark')['--color-brand-saturated'],
    );
  });

  it('emits a variable per spacing and radius token', () => {
    const vars = cssVariables('dark');
    for (const k of Object.keys(tokens.spacing)) expect(vars[`--space-${k}`]).toBeTruthy();
    for (const k of Object.keys(tokens.radius)) expect(vars[`--radius-${k}`]).toBeTruthy();
  });

  it('motion tokens surface as CSS time values', () => {
    expect(cssVariables('dark')['--motion-fast']).toBe(`${motion.timing.fast}s`);
    expect(cssVariables('dark')['--ease-entrance']).toBe(motion.easing.entrance);
  });
});

describe('preferredMode', () => {
  it('falls back to dark with no stored choice / no matchMedia', () => {
    expect(preferredMode()).toBe('dark');
  });
});
