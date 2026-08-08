import * as React from "react"

import { cn } from "@/lib/utils"

export interface StartingPoint<T extends string> {
  value: T
  /**
   * **The action, phrased as a verb** — "Deploy your own code", not "A
   * repository". A tab strip is scanned in any order, so each tab has to be
   * complete on its own.
   *
   * These were nouns with a `Start from` label above them, and the label was
   * the only thing making the nouns into a sentence. It sat 966px from the last
   * tab's title and 11px tall in the muted tier — the least visible mark in the
   * band was carrying its grammar. With the verb in the title there is nothing
   * left for it to explain, so it is gone.
   */
  name: string
  /** The specifics — which provider, which apps, which parts. Never a second
   *  explanation of the title. */
  description: string
  icon: React.ReactNode
  /**
   * The blank canvas is the odd one out, so it looks different: a dashed chip
   * reports "nothing in it yet", which is a fact rather than decoration. The
   * chip still fills when selected, so "selected" reads the same on all five.
   */
  dashed?: boolean
}

/**
 * The five peer starting points for a new stack, as one full-bleed strip.
 *
 * **These were cards, and cards were the wrong shape.** Five bordered boxes at
 * the top of the sheet is a wall — the same mistake §11 names for lists, one
 * level up. The strip keeps every option visible and one always live, at a
 * fraction of the weight.
 *
 * **The strip is full-bleed; its content is not.** The hairline under it spans
 * the whole sheet so the band reads as one object, while each tab carries the
 * sheet's own 16px inset — so the first chip lands on the same column as the
 * page title above it and the body below it. Getting that wrong is what made an
 * earlier pass read as two unrelated regions stacked on each other.
 *
 * **The strip carries no label of its own.** It had one — a `Start from`
 * eyebrow — and it was the wrong fix for a real problem: the titles were nouns
 * ("A repository") that only made sense once you read the stem above them. The
 * verb now lives in the title, so there is nothing left to introduce, and the
 * band is 36px shorter for it.
 *
 * **Selection is the underline and the chip, and nothing else.** One colour
 * (`--ring`), two marks, no box and no wash. §7's "say it once" applies to
 * selection exactly as it does to status.
 *
 * **The underline is an inset shadow, never a border.** A `border-b`
 * participates in layout, so a selected tab would stand 2px shorter than its
 * siblings and the whole strip would shift as you clicked along it.
 *
 * The dividers are the only thing separating the tabs once the boxes are gone,
 * so unlike a list rule (§11) they genuinely earn their place. They are inset
 * to 64px rather than run full height, so they describe the group without
 * boxing it.
 *
 * Keyboard: `role="tablist"` with roving focus — arrows move the selection,
 * Home/End jump to the ends, and Tab enters and leaves the strip once.
 */
export function StartingPointTabs<T extends string>({
  options,
  value,
  onValueChange,
  className,
  "aria-label": ariaLabel = "Start from",
}: {
  options: StartingPoint<T>[]
  value: T
  onValueChange: (value: T) => void
  className?: string
  /** The group's accessible name. It is not drawn — see `name` above. */
  "aria-label"?: string
}) {
  const refs = React.useRef<(HTMLButtonElement | null)[]>([])

  function move(from: number, step: number) {
    const next = (from + step + options.length) % options.length
    onValueChange(options[next].value)
    refs.current[next]?.focus()
  }

  function onKeyDown(e: React.KeyboardEvent, index: number) {
    const step =
      e.key === "ArrowRight" || e.key === "ArrowDown" ? 1
        : e.key === "ArrowLeft" || e.key === "ArrowUp" ? -1
          : 0
    if (step) {
      e.preventDefault()
      move(index, step)
      return
    }
    if (e.key === "Home" || e.key === "End") {
      e.preventDefault()
      const at = e.key === "Home" ? 0 : options.length - 1
      onValueChange(options[at].value)
      refs.current[at]?.focus()
    }
  }

  return (
    <div className={className}>
      {/* `sheet-edge-b`, not `border-b`: a real border participates in layout,
          so the 117px strip measured 118 and every row below it moved down by
          a pixel. The inset shadow is the same line at no cost (§12a). */}
      <div role="tablist" aria-label={ariaLabel} className="sheet-edge-b flex">
        {options.map((option, index) => {
          const selected = option.value === value
          return (
            <React.Fragment key={option.value}>
              {index > 0 && (
                <div aria-hidden className="bg-border my-auto h-16 w-px flex-none" />
              )}
              <button
                ref={(el) => { refs.current[index] = el }}
                type="button"
                role="tab"
                aria-selected={selected}
                // Roving tabstop: the strip is one Tab stop, arrows move inside it.
                tabIndex={selected ? 0 : -1}
                onClick={() => onValueChange(option.value)}
                onKeyDown={(e) => onKeyDown(e, index)}
                className={cn(
                  "flex flex-1 flex-col items-start gap-[11px] px-4 py-5 text-left transition-colors",
                  "focus-ring-inset",
                  // Transparent at rest, so it borrows the wash (§4). Branched
                  // rather than stacked as `hover:` variants, because a selected
                  // tab being hovered would match both and which wins is not
                  // reliably predictable.
                  selected
                    ? "shadow-[inset_0_-2px_0_var(--ring)]"
                    : "hover:bg-[var(--wash-hover)]",
                )}
              >
                <span
                  className={cn(
                    "flex size-[30px] flex-none items-center justify-center rounded-md border",
                    "[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
                    selected
                      ? "border-ring bg-ring text-primary-foreground"
                      : option.dashed
                        ? "border-border border-dashed text-fg-2"
                        : "border-border bg-control text-fg-2",
                  )}
                >
                  {option.icon}
                </span>
                <span className="flex flex-col">
                  <span className="text-body text-foreground font-medium">{option.name}</span>
                  <span className="text-meta text-fg-2">{option.description}</span>
                </span>
              </button>
            </React.Fragment>
          )
        })}
      </div>
    </div>
  )
}
