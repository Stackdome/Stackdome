import { Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatusText } from "@/components/branded/status-text";
import { relativeAge, absoluteAge } from "@/components/branded/entity-card";
import { cn } from "@/lib/utils";
import type { Stack } from "@/api/stack-types";
import { stackRollupState, statusReason, stateChangedAt } from "./status";
import { stackComponents, fitChips, ComponentChip } from "./stack-components";

/**
 * One grid template, used by the cards view and its skeleton — the row has
 * `STACK_COLUMNS` for the same reason: a loading state on a different rhythm
 * moves the whole page when the data lands.
 *
 * **12 across, 16 down.** The gutter is tighter than the row gap on purpose: at
 * 1440 that puts three 379.33px cards on the 1186px sheet, and a card's
 * neighbours read as one band rather than as a grid of separate boxes.
 */
export const STACK_CARD_GRID = "grid grid-cols-1 gap-x-3 gap-y-4 sm:grid-cols-2 xl:grid-cols-3";

/** The branch and the commit it is pinned to, as one machine string (§6). */
function sourceRef(stack: Stack): string | null {
  const git = stack.spec?.stack_resources?.find((r) => r.source?.git)?.source?.git;
  if (git?.branch) return git.commit ? `${git.branch}@${git.commit.slice(0, 7)}` : git.branch;
  const image = stack.spec?.stack_resources?.find((r) => r.source?.image)?.source?.image?.ref;
  return image ?? null;
}

/**
 * The same stack as a card — the row's facts **unabbreviated**, never a second
 * dataset (§11).
 *
 * The card answers the half of the job a row structurally cannot: *where does
 * this estate stand, to someone who does not already know it.* A row is built
 * for the person who opens `orders-api` every morning; eight names tell a new
 * joiner nothing about which one is a static site and which is a six-service
 * platform.
 *
 * **It carries no topology diagram and no sparkline.** Both were tried and both
 * drew the *shape* of substance without supplying any: four unlabelled boxes
 * need notation the reader does not have, and fourteen bars with no axis cannot
 * distinguish eleven deploys from four. The components are **named** instead —
 * the diagram's labels, kept, and the diagram dropped — and each carries the
 * canvas's own glyph, so the kind of thing it is reads before the name does.
 *
 * **Every card is the same height**, so a failing one does not shove the grid
 * around. That has a cost worth knowing: the reason line gets ONE line at 344px,
 * where the row's status column has 504px. The row therefore shows *more* of a
 * failure than the card does — the card's advantage is the component list, not
 * the reason.
 *
 * Material is §11's: white fill, 1px hairline, `rounded-lg`, **no shadow** —
 * content is flat. Hover is a 4% ink wash plus the stronger hairline; no lift,
 * no scale, no shadow.
 */
export function DeployStackCard({
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
  const parts = stackComponents(stack);
  const { shown, hidden } = fitChips(parts);
  const deleteDisabled = stack.lifecycle === "deleting";

  // Same line the row carries, so switching views never changes what a stack
  // says about itself.
  const provenance = [projectName, ref ?? "never deployed"].filter(Boolean).join(" · ");

  return (
    <Card
      role="link"
      tabIndex={0}
      aria-label={`${stack.name} stack`}
      onClick={() => navigate(`/stacks/${stack.id}`)}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/stacks/${stack.id}`);
      }}
      /* 162, every state — measured, not derived. Inside the hairline that is
         160, less 32 of padding leaves 128. The FULLEST head is 88 (24 name +
         2 + 16 provenance + 2 + 16 reason, then 8 + 20 chips) and the footer
         is 16, so 24 of air is left over even in the worst case and the
         lighter states hand their spare rows to the same gap via
         `justify-between`. The tallest card breathes rather than fitting
         exactly, and no state shoves the grid around on a poll. */
      /* Focus is a box-shadow ring, never an outline. Flush to the card's own
         edge: the ring is solid ink against an 11% hairline, so there is no
         chance of reading it as a border and no need for the gap. */
      className="group/card h-[162px] cursor-pointer flex-col justify-between gap-0 p-4 transition-colors hover:border-border-strong hover:bg-[var(--wash-hover)] focus-ring-edge"
    >
      {/* Head: the text block, then the chips 8px under it. The chips belong to
          the head because they say what the stack IS — parking them by the
          footer made them read as metadata about the last change. */}
      <div className="min-w-0">
        {/* Baseline, not centre: the 16px name and the 12px status word sit on
            one written line instead of two boxes centred against each other. */}
        <div className="flex items-baseline gap-2.5">
          <div className="min-w-0 flex-1 truncate text-title font-semibold text-foreground" title={stack.name}>
            {stack.name}
          </div>
          {/* Glyph + word, in one hue. The glyph used to be off here because
              the set was per-FAMILY, so Degraded, Unavailable and Failed all
              drew one triangle — chrome that could not make the distinction
              the word already made. The set is per-STATE now, so the mark
              carries its own information and earns the space (§7). */}
          <StatusText domain="stack_rollup" state={stackRollupState(stack)} icon />
        </div>

        {/* Two typefaces, one line (§6): the project is a word and the ref is a
            machine value. Setting the whole line in mono makes the project name
            look like something you could paste into a terminal. */}
        <div data-slot="provenance" className="mt-0.5 truncate text-meta text-fg-muted" title={provenance}>
          {projectName ? `${projectName} · ` : null}
          {ref ? <span className="font-mono">{ref}</span> : "never deployed"}
        </div>

        {/* Part of the head group, 2px under the provenance — it elaborates the
            status word, so it must not float off on its own.
            `fg-2` unless it CONTRADICTS the word above it: "Last deploy failed"
            under a green `Healthy` is the one line here carrying news the status
            does not, so it keeps the danger hue. A failure's own message under
            a red `Failed` is already accounted for and stays quiet. */}
        {reason && (
          <div
            className={cn("mt-0.5 truncate text-meta", reason.tone === "danger" ? "text-danger" : "text-fg-2")}
            title={reason.text}
          >
            {reason.text}
          </div>
        )}

        {/* Chips only — no "N components" label above them. A count is the
            stack's shape written as a number and there is nothing to do with
            it; the names are the fact. One row, never wrapped: the card's
            height is fixed, so anything that does not fit is counted. */}
        <div data-slot="components" className="mt-2 flex gap-[5px] overflow-hidden">
          {shown.map((part) => (
            <ComponentChip key={part.name} part={part} />
          ))}
          {hidden > 0 && (
            <span className="inline-flex h-5 shrink-0 items-center rounded-sm border border-border px-[7px] text-meta text-fg-muted">
              +{hidden}
            </span>
          )}
        </div>
      </div>

      {/* No rule above it. The 18px of air and the card's own edge already say
          where the footer starts; a hairline through the middle of a 162px box
          divides it into two cards. */}
      <div data-slot="card-footer" className="flex h-4 items-center gap-2">
        <span className="min-w-0 flex-1 truncate text-meta text-fg-muted" title={absoluteAge(changed) ?? undefined}>
          {changed ? `Last change ${relativeAge(changed)}` : "Never deployed"}
        </span>
        {/* A trash, not an overflow menu. One destination behind a kebab is a
              menu that exists to hide a single item — the icon says what it does
              and the §10 retype gate is what makes it safe, not the extra click. */}
        {onDelete && (
          <Button
            variant="ghost"
            size="icon-sm"
            shape="flat"
            disabled={deleteDisabled}
            aria-label={`Delete ${stack.name}`}
            /* The button is 28px and this row is 16. The negative margin
                 bleeds the hit area into the padding instead of growing the
                 row, so the card keeps its height and the target stays 28. */
            className="-my-1.5 -mr-1.5 flex-none opacity-0 transition-opacity group-hover/card:opacity-100 focus-visible:opacity-100"
            onClick={(e) => {
              e.stopPropagation();
              onDelete(stack);
            }}
          >
            <Trash2 />
          </Button>
        )}
      </div>
    </Card>
  );
}

/** The cards view's loading state, at the real card rhythm so nothing jumps
 *  when the data lands. No shimmer (§14). */
export function StackCardSkeleton() {
  return (
    <div className={STACK_CARD_GRID} aria-hidden>
      {Array.from({ length: 6 }, (_, i) => (
        <div key={i} className="flex h-[162px] flex-col justify-between rounded-lg border border-border p-4">
          {/* The real card's head group, block for block: 24 name, 2, 16
              provenance, 8, 20 chips. A skeleton on a different rhythm moves
              every card on the page the moment the data lands. */}
          <div>
            <div className="h-6 w-40 rounded-sm bg-[var(--wash-hover)]" />
            <div className="mt-0.5 h-4 w-48 rounded-sm bg-[var(--wash-hover)]" />
            <div className="mt-2 flex gap-[5px]">
              <div className="h-5 w-14 rounded-sm bg-[var(--wash-hover)]" />
              <div className="h-5 w-20 rounded-sm bg-[var(--wash-hover)]" />
            </div>
          </div>
          <div className="flex h-4 items-center">
            <div className="h-4 w-32 rounded-sm bg-[var(--wash-hover)]" />
          </div>
        </div>
      ))}
    </div>
  );
}
