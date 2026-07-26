// vitest-axe ships its matcher augmentation against the legacy `Vi` namespace,
// which Vitest 2 no longer reads. Re-augment the `vitest` module directly so
// `expect(results).toHaveNoViolations()` typechecks.
import 'vitest';
import type {AxeMatchers} from 'vitest-axe/matchers';

declare module 'vitest' {
  interface Assertion<T = any> extends AxeMatchers {}
  interface AsymmetricMatchersContaining extends AxeMatchers {}
}
