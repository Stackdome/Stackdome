import { useEffect, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Check, Loader2 } from "lucide-react";
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
import { Button } from "@/components/ui/button";

export type StickyActionBarSegment = { num: number; label: string };

export interface StickyActionBarPrimary {
  label: string;
  icon?: ReactNode;
  loadingLabel?: string;
  isLoading?: boolean;
  onClick: () => void;
}

export interface StickyActionBarSecondary {
  label: string;
  onClick: () => void;
  confirm?: {
    threshold: number;
    title: string;
    description: ReactNode;
    confirmLabel?: string;
    cancelLabel?: string;
  };
  dirtyCount?: number;
}

/**
 * pending  — unsaved/staged changes: amber left edge + pulsing amber dot.
 * deploying — a deploy is in flight: amber edge + spinner dot.
 * clean    — everything deployed: no edge, green check dot, no primary action.
 */
export type StickyActionBarTone = "pending" | "deploying" | "clean";

interface StickyActionBarProps {
  leadLabel: string;
  segments: StickyActionBarSegment[];
  /** Optional — clean/deploying states may omit the primary action. */
  primary?: StickyActionBarPrimary;
  secondary?: StickyActionBarSecondary;
  tone?: StickyActionBarTone;
}

/**
 * Sticky editorial action bar — mono caps per the design system,
 * portals into the app-layout slot so it spans the full content width.
 */
export default function StickyActionBar({
  leadLabel,
  segments,
  primary,
  secondary,
  tone = "pending",
}: StickyActionBarProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [slot, setSlot] = useState<HTMLElement | null>(null);

  useEffect(() => {
    setSlot(document.getElementById("page-sticky-bar"));
  }, []);

  const handleSecondaryClick = () => {
    if (!secondary) return;
    const dirty = secondary.dirtyCount ?? 0;
    if (secondary.confirm && dirty >= secondary.confirm.threshold) {
      setConfirmOpen(true);
    } else {
      secondary.onClick();
    }
  };

  const dot =
    tone === "clean" ? (
      <Check className="h-3.5 w-3.5 text-success" aria-hidden />
    ) : tone === "deploying" ? (
      <Loader2 className="h-3 w-3 animate-spin text-brand" aria-hidden />
    ) : (
      <span className="h-2 w-2 rounded-full bg-brand animate-pulse" aria-hidden />
    );

  const bar = (
    <div
      className="flex h-11 items-center gap-3 bg-card dark:bg-secondary px-6 text-foreground border-b border-border"
      style={tone === "clean" ? undefined : { boxShadow: "inset 3px 0 0 var(--brand)" }}
    >
      {dot}
      <span
        className={`font-mono text-[11px] font-bold uppercase tracking-[1.5px] ${tone === "clean" ? "text-success" : "text-brand"}`}
      >
        {leadLabel}
      </span>
      {segments.map((s, i) => (
        <span
          key={i}
          className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground"
        >
          <span className="mx-1.5 text-muted-foreground/60">·</span>
          <span className="font-bold text-foreground">{s.num}</span>{" "}
          <span>{s.label}</span>
        </span>
      ))}
      <div className="grow" />
      {secondary && (
        <Button
          type="button"
          variant="railGhost"
          size="rail"
          onClick={handleSecondaryClick}
          disabled={primary?.isLoading}
        >
          {secondary.label}
        </Button>
      )}
      {primary && (
        <Button
          type="button"
          variant="railPrimary"
          size="rail"
          onClick={primary.onClick}
          disabled={primary.isLoading}
        >
          {primary.isLoading ? (
            <>
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              {primary.loadingLabel ?? primary.label}
            </>
          ) : (
            <>
              {primary.icon}
              {primary.label}
            </>
          )}
        </Button>
      )}
    </div>
  );

  return (
    <>
      {slot ? createPortal(bar, slot) : null}
      {secondary?.confirm && (
        <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{secondary.confirm.title}</AlertDialogTitle>
              <AlertDialogDescription>
                {secondary.confirm.description}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>
                {secondary.confirm.cancelLabel ?? "Cancel"}
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={() => {
                  setConfirmOpen(false);
                  secondary.onClick();
                }}
              >
                {secondary.confirm.confirmLabel ?? secondary.label}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </>
  );
}
