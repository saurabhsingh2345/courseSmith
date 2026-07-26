import * as React from 'react';
import {cn} from '../../lib/cn';

const sizes = {sm: 'size-4 border-2', md: 'size-6 border-2', lg: 'size-10 border-[3px]'} as const;

export interface LoaderProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: keyof typeof sizes;
  label?: string;
}

/** Accessible spinner. `label` is announced to screen readers via role=status. */
export function Loader({className, size = 'md', label = 'Loading', ...props}: LoaderProps) {
  return (
    <div role="status" aria-live="polite" className={cn('inline-flex', className)} {...props}>
      <span
        className={cn(
          'animate-spin rounded-full border-brand border-t-transparent',
          sizes[size],
        )}
      />
      <span className="sr-only">{label}</span>
    </div>
  );
}
