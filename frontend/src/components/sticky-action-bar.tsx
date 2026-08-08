import { useEffect, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Check, Loader2 } from "lucide-react";
import { useConfirm } from "@/components/branded/confirm";
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
  const [slot, setSlot] = useState<HTMLElement | null>(null);
  const confirm = useConfirm();

  useEffect(() => {
    setSlot(document.getElementById("page-sticky-bar"));
  }, []);

  const handleSecondaryClick = async () => {
    if (!secondary) return;
    const dirty = secondary.dirtyCount ?? 0;
    if (secondary.confirm && dirty >= secondary.confirm.threshold) {
      const ok = await confirm({
        title: secondary.confirm.title,
        description: secondary.confirm.description,
        confirmLabel: secondary.confirm.confirmLabel ?? secondary.label,
        cancelLabel: secondary.confirm.cancelLabel,
      });
      if (!ok) return;
    }
    secondary.onClick();
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
        className={`font-mono text-label font-semibold  ${tone === "clean" ? "text-success" : "text-brand"}`}
      >
        {leadLabel}
      </span>
      {segments.map((s, i) => (
        <span
          key={i}
          className="font-mono text-label text-muted-foreground"
        >
          <span className="mx-1.5 text-muted-foreground/60">·</span>
          <span className="font-semibold text-foreground">{s.num}</span>{" "}
          <span>{s.label}</span>
        </span>
      ))}
      <div className="grow" />
      {secondary && (
        <Button
          type="button"
          variant="ghost"
          onClick={() => void handleSecondaryClick()}
          disabled={primary?.isLoading}
        >
          {secondary.label}
        </Button>
      )}
      {primary && (
        <Button
          type="button"
          variant="secondary"
          onClick={primary.onClick}
          loading={primary.isLoading}
          loadingText={primary.loadingLabel ?? primary.label}
        >
          {primary.icon}
          {primary.label}
        </Button>
      )}
    </div>
  );

  return slot ? createPortal(bar, slot) : null;
}
