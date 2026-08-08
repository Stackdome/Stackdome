import { cn } from "@/lib/utils";

/**
 * A fortnight of deploy counts, as bars.
 *
 * **The bars are neutral.** They report *volume*, not health — colouring
 * fourteen days of history with today's status was the worst error in the first
 * draft of this card: a stack that failed this morning would repaint its whole
 * good fortnight red.
 *
 * **It is labelled.** An unlabelled chart makes the reader infer what is being
 * counted, and a deploy count and a CPU trace look identical at this size.
 *
 * **A stack with no history says so in words.** `history` is absent — not
 * empty — for a stack that has never deployed, and that is a different fact
 * from a fortnight of zeroes: zeroes mean it deployed once and went quiet, and
 * drawing them as a flat line for a stack that never shipped invents a past it
 * does not have.
 */
export function DeploySparkline({
  history,
  className,
}: {
  /** Oldest first, one entry per day. Undefined when the stack never deployed. */
  history?: number[];
  className?: string;
}) {
  if (!history?.length) {
    return (
      <div className={cn("text-meta text-fg-muted", className)}>No deploys yet</div>
    );
  }

  const total = history.reduce((a, b) => a + b, 0);
  // Scale to the busiest day in the window, floored at 1 so an all-zero
  // fortnight cannot divide by zero.
  const peak = Math.max(1, ...history);

  return (
    <div className={className}>
      <div
        className="flex h-6 items-end gap-[2px]"
        role="img"
        aria-label={`${total} ${total === 1 ? "deploy" : "deploys"} in the last ${history.length} days`}
      >
        {history.map((count, day) => (
          <span
            key={day}
            // A day with no deploys keeps a 1px stub rather than a gap: the
            // fortnight has to read as a continuous span, or a quiet week looks
            // like missing data.
            className={cn(
              "min-w-[3px] flex-1 rounded-[1px]",
              count > 0 ? "bg-fg-2" : "bg-border",
            )}
            style={{ height: count > 0 ? `${Math.max(12, (count / peak) * 100)}%` : "1px" }}
          />
        ))}
      </div>
      <div className="mt-1.5 text-meta text-fg-muted">
        Deploys · {history.length}d
      </div>
    </div>
  );
}
