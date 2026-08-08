import * as React from "react"

import { cn } from "@/lib/utils"

export interface SegmentedControlOption<T extends string> {
  value: T
  /** Omit for an icon-only segment — `label` is still required as the a11y name. */
  label: string
  icon?: React.ReactNode
  /** Renders the label visually; an icon-only segment keeps it for screen readers. */
  showLabel?: boolean
  disabled?: boolean
}

const SIZES = {
  // §8 — the track takes the radius of its own height, and the SELECTED segment
  // repeats it on whichever outer corners it owns. The track's `overflow-hidden`
  // would round them anyway; the segment carries its own so that its **shadow**
  // traces the rounded corner instead of a square one.
  sm: { track: "h-7 rounded-sm", segment: "min-w-7 px-2 gap-1", start: "rounded-l-sm", end: "rounded-r-sm" },
  default: { track: "h-8 rounded-md", segment: "min-w-8 px-2.5 gap-1.5", start: "rounded-l-md", end: "rounded-r-md" },
} as const

/**
 * Two or three mutually exclusive views of the same thing (§7 — the list/cards
 * toggle). Not two buttons and not a dropdown: a segmented control shows every
 * option and which one is live in a single glance.
 *
 * **Selection is carried by ink, never by opacity.** A dimmed icon reads as
 * disabled, and the one state this control exists to report would then be
 * indistinguishable from the one state it must never be confused with.
 *
 * **The track and the selected segment both carry a hairline, and the track has
 * no padding.** That is the whole trick. A bordered segment inset inside a
 * bordered track puts two edges 2px apart and the eye reads a doubled line — so
 * the segment is not inset. It runs flush, its outer edges land on the same
 * pixel as the track's, and the only border you actually see between them is
 * the one dividing the two segments.
 *
 * **The selected segment is raised.** It carries `shadow-sm`, which the track
 * clips on three sides — so the lift reads along the divider and nowhere else.
 * Measured off the board: the divider goes from `#E2DFD8` (the hairline over
 * the control fill) to `#DEDAD3`, and the two pixels beyond it darken by one
 * level. That is the whole effect, and it is deliberately almost nothing.
 *
 * The track is `--control`, not a grey well. With a hairline around it the
 * control already reads as a control, and a darker fill under the white segment
 * would make the unselected side look switched off.
 *
 * Keyboard: `role="radiogroup"` with roving focus — arrows move the selection,
 * Home/End jump to the ends, and Tab enters and leaves the group once.
 */
export function SegmentedControl<T extends string>({
  options,
  value,
  onValueChange,
  size = "default",
  disabled = false,
  className,
  "aria-label": ariaLabel,
}: {
  options: SegmentedControlOption<T>[]
  value: T
  onValueChange: (value: T) => void
  size?: keyof typeof SIZES
  disabled?: boolean
  className?: string
  "aria-label"?: string
}) {
  const refs = React.useRef<(HTMLButtonElement | null)[]>([])
  const s = SIZES[size]

  const selectable = options
    .map((o, i) => ({ o, i }))
    .filter(({ o }) => !o.disabled && !disabled)

  function move(from: number, step: number) {
    if (!selectable.length) return
    const at = selectable.findIndex(({ i }) => i === from)
    const next = selectable[(at + step + selectable.length) % selectable.length]
    onValueChange(next.o.value)
    refs.current[next.i]?.focus()
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
      const edge = e.key === "Home" ? selectable[0] : selectable[selectable.length - 1]
      if (!edge) return
      onValueChange(edge.o.value)
      refs.current[edge.i]?.focus()
    }
  }

  return (
    <div
      role="radiogroup"
      aria-label={ariaLabel}
      className={cn(
        // No padding: the segments run flush, so their outer edges land on the
        // same pixel as the track's border instead of 1px inside it.
        // overflow-hidden gives their square corners the track's radius.
        "bg-control border border-border inline-flex items-center overflow-hidden",
        s.track,
        disabled && "opacity-50",
        className,
      )}
    >
      {options.map((option, index) => {
        const selected = option.value === value
        const isDisabled = disabled || option.disabled
        return (
          <button
            key={option.value}
            ref={(el) => { refs.current[index] = el }}
            type="button"
            role="radio"
            aria-checked={selected}
            aria-label={option.label}
            disabled={isDisabled}
            // Roving tabstop: the group is one Tab stop, arrows move inside it.
            tabIndex={selected ? 0 : -1}
            onClick={() => !isDisabled && onValueChange(option.value)}
            onKeyDown={(e) => onKeyDown(e, index)}
            className={cn(
              "relative inline-flex h-full items-center justify-center whitespace-nowrap text-body font-medium transition-colors",
              "focus-ring-inset",
              "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
              s.segment,
              // Ink vs fg-muted. NOT opacity — see the component note.
              //
              // `shadow-sm` is the one place a piece of CONTENT is allowed an
              // elevation (§5). The selected segment is not floating over the
              // page — it is a raised face inside a well, the same idea as a
              // key on a keyboard, and the board draws it that way. The track's
              // `overflow-hidden` clips the shadow at three sides, so what you
              // actually see is a soft lift along the divider.
              //
              // On focus the ring (unlayered CSS) overwrites this shadow, which
              // is fine: a 3.5% lift under a solid 2px blue ring is invisible.
              selected
                ? "bg-card text-foreground shadow-sm"
                : "text-fg-muted hover:text-foreground",
              // The outer corners the selected segment owns, so its shadow
              // follows the track's radius rather than cutting a square corner.
              selected && index === 0 && s.start,
              selected && index === options.length - 1 && s.end,
              // The divider is the selected segment's own edge, so a separate
              // rule can never end up doubled beside it. At either end of the
              // track that edge lands on the track's own border — drawn there
              // too, it would composite to 22% and read as `border-strong`.
              selected && index > 0 && "border-l border-border",
              selected && index < options.length - 1 && "border-r border-border",
              isDisabled && "cursor-not-allowed text-fg-ghost hover:text-fg-ghost",
            )}
          >
            {option.icon}
            {(option.showLabel ?? !option.icon) && <span>{option.label}</span>}
          </button>
        )
      })}
    </div>
  )
}
