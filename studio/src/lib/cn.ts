// Class-name merge helper used by every base component. clsx resolves
// conditionals/arrays; tailwind-merge dedupes conflicting Tailwind utilities so
// a caller's `className` reliably overrides a component's defaults.
import {clsx, type ClassValue} from 'clsx';
import {twMerge} from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
