import * as React from 'react';
import {cva, type VariantProps} from 'class-variance-authority';
import {cn} from '../../lib/cn';

export const badgeVariants = cva(
  'inline-flex items-center rounded-[var(--radius-full)] border px-2.5 py-0.5 text-xs font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-brand text-white',
        secondary: 'border-border bg-surface text-fg',
        success: 'border-transparent bg-success text-white',
        error: 'border-transparent bg-error text-white',
        warning: 'border-transparent bg-warning text-black',
        outline: 'border-border text-fg',
      },
    },
    defaultVariants: {variant: 'default'},
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({className, variant, ...props}: BadgeProps) {
  return <span className={cn(badgeVariants({variant}), className)} {...props} />;
}
