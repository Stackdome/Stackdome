import { Ellipsis, Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { StatusText } from "@/components/branded/status-text";
import { relativeAge, absoluteAge } from "@/components/branded/entity-card";
import { cn } from "@/lib/utils";
import type { Stack } from "@/api/stack-types";
import { stackRollupState, statusReason, stateChangedAt } from "./status";

/**
 * One column template, used by the header and every row.
 *
 * **`Status` is the flexible column, not `Name`.** That is the whole layout
 * decision. `Name` used to take `1fr` and swallow every spare pixel, which left
 * a long empty run between the name and the status on a wide sheet. Now the
 * name is capped and the slack goes to the status cell — where the reason line
 * lives, and where extra width buys a sentence that finishes instead of one
 * that truncates.
 *
 * `Services` is gone. It was the stack's shape written as text, it could not be
 * recognised at a glance, and it did not survive the question every column has
 * to answer: does this help you choose a row, compare across rows, or finish
 * without leaving? The card carries the components by name instead.
 */
export const STACK_COLUMNS =
  "grid grid-cols-[minmax(240px,420px)_minmax(0,1fr)_130px_32px] items-center gap-5";

/** The branch and the commit it is pinned to, as one machine string (§6). */
function sourceRef(stack: Stack): string | null {
  const git = stack.spec?.stack_resources?.find((r) => r.source?.git)?.source?.git;
  if (git?.branch) return git.commit ? `${git.branch}@${git.commit.slice(0, 7)}` : git.branch;
  const image = stack.spec?.stack_resources?.find((r) => r.source?.image)?.source?.image?.ref;
  return image ?? null;
}

/**
 * A stack as a hairline row (§11). Separation is a 1px rule — no box, no shadow,
 * no card per item.
 *
 * The row's job is *compare and find*: the same fact in the same place on every
 * line, so the odd one out jumps. Everything that answers "what is going on
 * inside this one" belongs to the card.
 *
 * Status is said ONCE, as a word — there is no row dot (§11 names this as the
 * exact place the rule slips). The line under it is not a second reading of the
 * word; it is **why**, and it appears only when there is a why. A healthy row
 * stays one line and a broken one is visibly taller.
 *
 * The row **box** carries the 12px sheet edge and the text sits 8px inside it,
 * so the hover wash extends past the name rather than starting at it.
 */
export function DeployStackRow({
  stack,
  projectName,
  onDelete,
}: {
  stack: Stack;
  /** `useResourceProjects().projectNameById` returns null for an unresolved id. */
  projectName?: string | null;
  onDelete?: (stack: Stack) => void;
}) {
  const navigate = useNavigate();
  const ref = sourceRef(stack);
  const changed = stateChangedAt(stack);
  const reason = statusReason(stack);
  const menuDisabled = stack.lifecycle === "deleting";

  // region and author are NOT on the stack list payload — the API carries a
  // `user_id` UUID and no region at all, and a rendered UUID is worse than an
  // absent field. The line is project + branch@sha until the payload grows.
  const provenance = [projectName, ref].filter(Boolean).join(" · ");

  return (
    <div
      role="link"
      tabIndex={0}
      aria-label={`${stack.name} stack`}
      onClick={() => navigate(`/stacks/${stack.id}`)}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/stacks/${stack.id}`);
      }}
      className={cn(
        STACK_COLUMNS,
        // No rule between rows. A separator earns its place by GROUPING, and
        // measured here it groups nothing: the gap inside a row (name → branch)
        // is 0px, the gap between rows is 28px. Space was already doing all of
        // it, so the line was decorating a boundary that was never ambiguous.
        // Row extent on approach is the hover wash's job, and always was.
        //
        // The rule under the column header STAYS (see StackRowHeader) — that
        // one is the chrome/content boundary, not a row separator.
        //
        // Bring the rule back if a COMPACT row mode lands: at ~8px between
        // rows instead of 28 the grouping argument reverses.
        "group/row h-16 cursor-pointer px-2 transition-colors",
        // The row is full-bleed, so its ring turns inward — an outside ring
        // would be clipped by the sheet and lose a side.
        "hover:bg-[var(--wash-hover)] focus-ring-inset",
      )}
    >
      <div className="flex min-w-0 flex-col">
        <span className="truncate text-name font-medium text-foreground" title={stack.name}>
          {stack.name}
        </span>
        {provenance && (
          <span className="truncate font-mono text-meta text-fg-muted" title={provenance}>
            {provenance}
          </span>
        )}
      </div>

      {/* The word, and — only when there is one — why. Same glyph as the card:
          a status has to look the same in both views or switching between them
          costs a re-read. */}
      <div className="flex min-w-0 flex-col">
        <StatusText domain="stack_rollup" state={stackRollupState(stack)} icon />
        {reason && (
          <span
            className={cn(
              "truncate text-meta",
              reason.tone === "danger" ? "text-danger" : "text-fg-2",
            )}
            title={reason.text}
          >
            {reason.text}
          </span>
        )}
      </div>

      <div
        className="truncate text-meta tabular-nums text-fg-muted"
        title={absoluteAge(changed) ?? undefined}
      >
        {relativeAge(changed)}
      </div>

      {/* §11 — the row's actions appear on hover. A kebab on every row at rest is
          eight pieces of chrome competing with eight names. Hidden by opacity so
          the control keeps its tab stop and the row does not reflow. */}
      <div className="flex justify-end">
        {onDelete && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon-sm"
                shape="flat"
                aria-label={`Actions for ${stack.name}`}
                className="opacity-0 transition-opacity group-hover/row:opacity-100 focus-visible:opacity-100 data-[state=open]:opacity-100"
                onClick={(e) => e.stopPropagation()}
              >
                <Ellipsis />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[160px]" onClick={(e) => e.stopPropagation()}>
              {/* No deferral needed: the confirm service defers its own open
                  a tick past the menu close (radix-ui/primitives#1836). */}
              <DropdownMenuItem variant="destructive" disabled={menuDisabled} onSelect={() => onDelete(stack)}>
                <Trash2 />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </div>
  );
}

/** The list's column headers — sentence case, `text-label`, `fg-muted`, one 1px
 *  rule underneath and nothing else (§11). An unlabelled column makes the reader
 *  infer what a bare timestamp is measuring.
 *
 *  **The label sits in a uniform 8px inset — 8 above, 8 each side, 8 below.**
 *  It was 16 above (the sheet's content inset) and 6 below, so the one piece of
 *  chrome on the page was the only thing not square with itself. `-mt-2` gives
 *  back half the content inset: a column header is chrome and sits tighter to
 *  the band than content does. Settled on the board (node `121:885`). */
export function StackRowHeader() {
  return (
    <div className={cn(STACK_COLUMNS, "-mt-2 border-b border-border px-2 pb-2 text-label text-fg-muted")}>
      <div>Name</div>
      <div>Status</div>
      {/* Not "Updated" — this is the age of the STATE, not of the record. */}
      <div>Last change</div>
      <div />
    </div>
  );
}

/** Six rows at the real 64px pitch, so nothing moves when the data lands (§15
 *  loading). A `wash-hover` block and no shimmer — §14 rules out ambient
 *  movement on a working surface. */
export function StackRowSkeleton() {
  return (
    <div aria-hidden>
      {/* No rule, matching the loaded row — the skeleton's whole job is that
          nothing moves or appears when the data lands. */}
      {Array.from({ length: 6 }, (_, i) => (
        <div key={i} className={cn(STACK_COLUMNS, "h-16 px-2")}>
          <div className="flex min-w-0 flex-col gap-1.5">
            <div className="h-4 w-40 rounded-sm bg-[var(--wash-hover)]" />
            <div className="h-3 w-56 rounded-sm bg-[var(--wash-hover)]" />
          </div>
          <div className="h-3 w-24 rounded-sm bg-[var(--wash-hover)]" />
          <div className="h-3 w-16 rounded-sm bg-[var(--wash-hover)]" />
          <div />
        </div>
      ))}
    </div>
  );
}
