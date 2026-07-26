import * as React from 'react';
import {cn} from '../../lib/cn';

/** Loading placeholder. Pulse pace comes from the shared motion vars. */
export function Skeleton({className, ...props}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('animate-pulse rounded-[var(--radius-md)] bg-ink-700/60', className)}
      aria-hidden="true"
      {...props}
    />
  );
}
