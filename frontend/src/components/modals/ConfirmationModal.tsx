'use client';

import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { AlertTriangle, Info, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface ConfirmationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'destructive' | 'warning' | 'primary';
  isLoading?: boolean;
  onConfirm: () => void | Promise<void>;
  icon?: 'trash' | 'warning' | 'info' | React.ReactNode;
}

export function ConfirmationModal({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  cancelLabel = 'Cancel',
  variant = 'destructive',
  isLoading = false,
  onConfirm,
  icon,
}: ConfirmationModalProps) {
  const [internalLoading, setInternalLoading] = React.useState(false);
  const effectiveLoading = isLoading || internalLoading;

  const handleConfirm = async (e: React.MouseEvent) => {
    e.preventDefault();
    try {
      setInternalLoading(true);
      await onConfirm();
    } finally {
      setInternalLoading(false);
    }
  };

  const renderIcon = () => {
    if (React.isValidElement(icon)) return icon;

    if (icon === 'trash' || (!icon && variant === 'destructive')) {
      return (
        <div className="h-9 w-9 rounded-md bg-rose-950/50 border border-rose-900/60 flex items-center justify-center text-rose-400 shrink-0">
          <Trash2 className="h-4.5 w-4.5" />
        </div>
      );
    }

    if (icon === 'warning' || (!icon && variant === 'warning')) {
      return (
        <div className="h-9 w-9 rounded-md bg-amber-950/50 border border-amber-900/60 flex items-center justify-center text-amber-400 shrink-0">
          <AlertTriangle className="h-4.5 w-4.5" />
        </div>
      );
    }

    return (
      <div className="h-9 w-9 rounded-md bg-zinc-800/60 border border-zinc-700/60 flex items-center justify-center text-zinc-300 shrink-0">
        <Info className="h-4.5 w-4.5" />
      </div>
    );
  };

  const defaultConfirmText =
    confirmLabel ?? (variant === 'destructive' ? 'Delete' : 'Confirm');

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!effectiveLoading) {
          onOpenChange(isOpen);
        }
      }}
    >
      <DialogContent className="sm:max-w-md bg-zinc-950 border border-zinc-800 text-zinc-100 p-6 shadow-2xl">
        <div className="flex items-start gap-3.5">
          {renderIcon()}
          <div className="flex-1 space-y-1.5 pt-0.5">
            <DialogHeader className="text-left space-y-1">
              <DialogTitle className="text-sm font-semibold text-zinc-100 leading-snug">
                {title}
              </DialogTitle>
              <DialogDescription className="text-xs text-zinc-400 leading-relaxed">
                {description}
              </DialogDescription>
            </DialogHeader>
          </div>
        </div>

        <DialogFooter className="pt-4 mt-2 border-t border-zinc-800/80 flex sm:justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => onOpenChange(false)}
            disabled={effectiveLoading}
            className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-850 hover:text-zinc-100"
          >
            {cancelLabel}
          </Button>
          <Button
            type="button"
            variant={variant === 'destructive' ? 'destructive' : 'primary'}
            size="sm"
            onClick={handleConfirm}
            isLoading={effectiveLoading}
            disabled={effectiveLoading}
            className={cn(
              'text-xs',
              variant === 'destructive' && 'bg-rose-600 hover:bg-rose-500 text-white'
            )}
          >
            {defaultConfirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default ConfirmationModal;
