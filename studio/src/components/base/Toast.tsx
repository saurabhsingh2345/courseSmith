import * as React from 'react';
import * as ToastPrimitive from '@radix-ui/react-toast';
import {cva, type VariantProps} from 'class-variance-authority';
import {X} from 'lucide-react';
import {cn} from '../../lib/cn';

export const ToastProvider = ToastPrimitive.Provider;

export const ToastViewport = React.forwardRef<
  React.ElementRef<typeof ToastPrimitive.Viewport>,
  React.ComponentPropsWithoutRef<typeof ToastPrimitive.Viewport>
>(({className, ...props}, ref) => (
  <ToastPrimitive.Viewport
    ref={ref}
    className={cn(
      'fixed bottom-0 right-0 z-[100] flex max-h-screen w-full flex-col-reverse gap-2 p-4 sm:max-w-sm',
      className,
    )}
    {...props}
  />
));
ToastViewport.displayName = ToastPrimitive.Viewport.displayName;

const toastVariants = cva(
  'group pointer-events-auto relative flex w-full items-center justify-between gap-3 overflow-hidden rounded-[var(--radius-md)] border p-4 pr-8 shadow-lg transition-all data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-80 data-[swipe=end]:animate-out',
  {
    variants: {
      variant: {
        default: 'border-border bg-surface text-fg',
        success: 'border-success/40 bg-success/10 text-fg',
        error: 'border-error/40 bg-error/10 text-fg',
      },
    },
    defaultVariants: {variant: 'default'},
  },
);

export const Toast = React.forwardRef<
  React.ElementRef<typeof ToastPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof ToastPrimitive.Root> & VariantProps<typeof toastVariants>
>(({className, variant, ...props}, ref) => (
  <ToastPrimitive.Root ref={ref} className={cn(toastVariants({variant}), className)} {...props} />
));
Toast.displayName = ToastPrimitive.Root.displayName;

export const ToastTitle = React.forwardRef<
  React.ElementRef<typeof ToastPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof ToastPrimitive.Title>
>(({className, ...props}, ref) => (
  <ToastPrimitive.Title ref={ref} className={cn('text-sm font-semibold', className)} {...props} />
));
ToastTitle.displayName = ToastPrimitive.Title.displayName;

export const ToastDescription = React.forwardRef<
  React.ElementRef<typeof ToastPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof ToastPrimitive.Description>
>(({className, ...props}, ref) => (
  <ToastPrimitive.Description ref={ref} className={cn('text-sm opacity-90', className)} {...props} />
));
ToastDescription.displayName = ToastPrimitive.Description.displayName;

export const ToastClose = React.forwardRef<
  React.ElementRef<typeof ToastPrimitive.Close>,
  React.ComponentPropsWithoutRef<typeof ToastPrimitive.Close>
>(({className, ...props}, ref) => (
  <ToastPrimitive.Close
    ref={ref}
    className={cn(
      'absolute right-1.5 top-1.5 rounded-[var(--radius-sm)] p-1 text-fg/60 transition-colors hover:text-fg',
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand',
      className,
    )}
    aria-label="Close"
    {...props}
  >
    <X className="size-4" />
  </ToastPrimitive.Close>
));
ToastClose.displayName = ToastPrimitive.Close.displayName;

// --- Lightweight imperative API -------------------------------------------
// useToast() gives a queue + push() helper; <Toaster/> renders the queue inside
// a single ToastProvider/Viewport. Mount <Toaster/> once near the app root.

export interface ToastData {
  id: number;
  title?: string;
  description?: string;
  variant?: VariantProps<typeof toastVariants>['variant'];
  duration?: number;
}

type ToastInput = Omit<ToastData, 'id'>;

const listeners = new Set<(toasts: ToastData[]) => void>();
let queue: ToastData[] = [];
let counter = 0;

function emit() {
  for (const l of listeners) l(queue);
}

/** Enqueue a toast from anywhere (even outside React). Returns its id. */
export function toast(input: ToastInput): number {
  const id = ++counter;
  queue = [...queue, {id, duration: 4000, ...input}];
  emit();
  return id;
}

export function dismissToast(id: number) {
  queue = queue.filter((t) => t.id !== id);
  emit();
}

export function useToast() {
  const [toasts, setToasts] = React.useState<ToastData[]>(queue);
  React.useEffect(() => {
    listeners.add(setToasts);
    return () => {
      listeners.delete(setToasts);
    };
  }, []);
  return {toasts, toast, dismiss: dismissToast};
}

export function Toaster() {
  const {toasts} = useToast();
  return (
    <ToastProvider>
      {toasts.map(({id, title, description, variant, duration}) => (
        <Toast
          key={id}
          variant={variant}
          duration={duration}
          onOpenChange={(open) => {
            if (!open) dismissToast(id);
          }}
        >
          <div className="grid gap-1">
            {title && <ToastTitle>{title}</ToastTitle>}
            {description && <ToastDescription>{description}</ToastDescription>}
          </div>
          <ToastClose />
        </Toast>
      ))}
      <ToastViewport />
    </ToastProvider>
  );
}
