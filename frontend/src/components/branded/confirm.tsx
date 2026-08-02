import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { releaseStaleBodyLock } from "@/lib/radix-body-lock";

export interface ConfirmOptions {
  title: string;
  description?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: "default" | "destructive";
}

export type ConfirmFn = (opts: ConfirmOptions) => Promise<boolean>;

const ConfirmContext = createContext<ConfirmFn | null>(null);

/**
 * Promise-based confirmation: `const ok = await confirm({...})`.
 *
 * The single app-wide confirm dialog lives in ConfirmProvider, which owns the
 * modal-transition sequencing that Radix's body pointer-events save/restore
 * can't survive when layers open or close in the same tick
 * (radix-ui/primitives#1836; commits 6b560665, 0bbfd378):
 *
 * - opening is deferred one tick, so a dropdown/menu closing in the same
 *   event settles first,
 * - the promise resolves one tick after the dialog's close flushes, so caller
 *   follow-up (closing a parent modal, navigating) never shares a tick with
 *   this dialog's teardown, and
 * - the lock is re-checked once the layers settle, because a parent modal
 *   still animating out when this dialog opens outlives that one-tick defer
 *   and gets its `pointer-events: none` handed back on close.
 */
export function useConfirm(): ConfirmFn {
  return useContext(ConfirmContext) as ConfirmFn;
}

interface PendingConfirm {
  opts: ConfirmOptions;
  resolve: (ok: boolean) => void;
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirm | null>(null);
  const [open, setOpen] = useState(false);
  const pendingRef = useRef<PendingConfirm | null>(null);

  const confirm = useCallback<ConfirmFn>((opts) => {
    return new Promise<boolean>((resolve) => {
      setTimeout(() => {
        // A confirm arriving while one is open supersedes it; the superseded
        // caller reads a dismissal.
        pendingRef.current?.resolve(false);
        const next = { opts, resolve };
        pendingRef.current = next;
        setPending(next);
        setOpen(true);
      }, 0);
    });
  }, []);

  // Resolves exactly once per pending confirm: the Action's onClick settles
  // true, then Radix's own close fires onOpenChange(false) whose settle(false)
  // no-ops on the cleared ref.
  const settle = useCallback((ok: boolean) => {
    const p = pendingRef.current;
    if (!p) return;
    pendingRef.current = null;
    setOpen(false);
    setTimeout(() => p.resolve(ok), 0);
    releaseStaleBodyLock();
  }, []);

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      <AlertDialog open={open} onOpenChange={(o) => !o && settle(false)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{pending?.opts.title}</AlertDialogTitle>
            {/* Always rendered so Radix's aria-describedby wiring stays valid;
                visually hidden when the caller gave no description. */}
            <AlertDialogDescription className={pending?.opts.description == null ? "sr-only" : undefined}>
              {pending?.opts.description ?? pending?.opts.title}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{pending?.opts.cancelLabel ?? "Cancel"}</AlertDialogCancel>
            <AlertDialogAction variant={pending?.opts.variant} onClick={() => settle(true)}>
              {pending?.opts.confirmLabel ?? "Confirm"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmContext.Provider>
  );
}
