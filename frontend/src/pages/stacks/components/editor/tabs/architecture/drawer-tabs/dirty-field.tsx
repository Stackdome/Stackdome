import type { ReactNode } from "react";
import { RotateCcw } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { isPathDirty } from "@/pages/stacks/lib/stack-diff";

interface DirtyFieldProps {
  /** The draft object to compare. Pass the resource (or whatever subtree the path resolves against). */
  draft: unknown;
  /** Baseline counterpart. When undefined, the field renders unwrapped. */
  baseline: unknown;
  /** Dot-path within draft/baseline to compare. */
  path: string;
  /**
   * Called when the user clicks the reset arrow. Should restore the path on
   * the underlying source-of-truth (typically `session.discardResourceField`).
   */
  onReset?: () => void;
  /** Compact mode tightens the inset for inline cells (e.g. ports table rows). */
  compact?: boolean;
  /** When true, hides the reset arrow even if dirty (use in read-only previews). */
  hideReset?: boolean;
  className?: string;
  children: ReactNode;
}

/**
 * Visual wrapper that flags a field as modified vs baseline. When dirty:
 * adds a brand-tinted background + amber left edge and surfaces a small
 * RotateCcw button to reset that one field.
 *
 * Critically, the DOM tree shape is IDENTICAL whether dirty or clean — only
 * classes toggle and the reset button's `visibility` flips. If the tree
 * structure changed on dirty, React would unmount/remount the input inside
 * and the user would lose focus on the very first keystroke (which is what
 * triggers the dirty state).
 */
export function DirtyField({
  draft,
  baseline,
  path,
  onReset,
  compact = false,
  hideReset = false,
  className,
  children,
}: DirtyFieldProps) {
  // No baseline means there's no edit session to diff against (e.g., the
  // create-stack page). Skip the visual frame entirely so inputs sit flush
  // with their labels instead of leaving a permanently-empty inset.
  if (baseline === undefined) {
    return <>{children}</>;
  }

  const dirty = isPathDirty(draft, baseline, path);
  const showReset = dirty && !hideReset && !!onReset;

  // Negative margin extends the wrapper out to the accordion's left edge so the
  // 4px brand stripe aligns with the accordion border. Padding pushes children
  // back to the original FieldShell edge, so labels and inputs stay flush.
  // Compact mode (ledger rows) skips the stripe entirely — the brand tint and
  // reset arrow already mark the row, and rows sit flush against hairlines. Its
  // own pl/-ml pair bleeds the tint slightly past the control's left edge so
  // the wash doesn't look cut off flush against the input border.
  return (
    <div
      className={cn(
        "transition-colors",
        compact
          ? "rounded-md py-1 pr-1 pl-1.5 -ml-1.5"
          : cn(
            "border-l-4 -ml-5 pl-4 py-1 pr-2",
            dirty ? "border-l-brand" : "border-l-transparent",
          ),
        dirty && "bg-brand-bg",
        className,
      )}
    >
      <div className="flex items-center gap-2">
        <div className="grow min-w-0">{children}</div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className={cn(
            "shrink-0 text-brand hover:bg-brand-bg-hover hover:text-brand-press",
            !showReset && "invisible pointer-events-none",
          )}
          onClick={onReset}
          aria-label="Reset to original value"
          aria-hidden={!showReset}
          tabIndex={showReset ? 0 : -1}
          title="Reset to original value"
        >
          <RotateCcw className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}
