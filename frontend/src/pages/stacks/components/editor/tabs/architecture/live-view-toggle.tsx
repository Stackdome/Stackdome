import { Eye } from "lucide-react";

export type CanvasViewMode = "draft" | "live";

interface LiveViewToggleProps {
  mode: CanvasViewMode;
  onModeChange: (mode: CanvasViewMode) => void;
  /** Undeployed edits exist — marks the Draft segment so switching away from
   *  it doesn't read as "nothing pending". */
  draftDirty: boolean;
}

const SEGMENT_BASE =
  "flex items-center gap-1.5 rounded-[5px] px-2.5 py-1 text-[12px] font-medium transition-colors";

/**
 * Draft/Live segmented control overlaid on the canvas. "Live" switches the
 * canvas to a read-only view of the converged release; the canvas footer
 * caption spells the read-only contract out.
 */
export function LiveViewToggle({ mode, onModeChange, draftDirty }: LiveViewToggleProps) {
  const segment = (value: CanvasViewMode, label: React.ReactNode) => (
    <button
      type="button"
      aria-pressed={mode === value}
      onClick={() => onModeChange(value)}
      className={`${SEGMENT_BASE} ${
        mode === value ? "bg-brand-bg text-brand" : "text-fg-muted hover:text-foreground"
      }`}
    >
      {label}
    </button>
  );

  return (
    <div
      role="group"
      aria-label="Canvas view"
      className="flex items-center gap-0.5 rounded-md border border-border bg-background p-0.5 shadow-sm"
    >
      {segment(
        "draft",
        <>
          Draft
          {draftDirty && (
            <span
              aria-hidden
              data-testid="draft-dirty-dot"
              className="inline-block size-1.5 rounded-full bg-brand"
            />
          )}
        </>,
      )}
      {segment(
        "live",
        <>
          <Eye className="size-3.5" aria-hidden />
          Live
        </>,
      )}
    </div>
  );
}
