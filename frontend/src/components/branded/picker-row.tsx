import * as React from "react"
import { Check, Plus, X } from "lucide-react"

import { cn } from "@/lib/utils"

/**
 * One thing you can pick out of a list — a repository, a ready-made app, a
 * building block, a managed add-on, or an item already in the stack.
 *
 * **It is a row, not a card.** §11: no box, no shadow, no rule between rows.
 * What separates one row from the next is space and the hover wash, and what
 * marks the chosen one is the selection wash plus a single blue tick — one
 * colour, said once (§7).
 *
 * **Two sizes, one object.** 56 is the catalogue rung; 40 is the same row made
 * dense for the "In this stack" panel, where the name is a resource instance
 * and therefore machine-set (§6). They are not two components.
 *
 * The trailing slot takes whatever the list needs — a tick, a count, a plus, a
 * remove. The five pieces the board draws are exported below so a call site
 * composes them rather than hand-rolling a copy (§2).
 */
export function PickerRow({
  size = 56,
  icon,
  name,
  meta,
  endText,
  selected,
  trailing,
  onClick,
  className,
  ...props
}: {
  /** 56 is the catalogue rung; 40 is the dense "In this stack" rung. */
  size?: 56 | 40
  icon: React.ReactNode
  name: string
  /** Dot-separated parts under the name. `mono` for machine values (§6). */
  meta?: { text: string; mono?: boolean }[]
  /** Right-aligned reading matter — "12 services". Sits before `trailing`. */
  endText?: string
  /**
   * Omit entirely for a row that *acts* — a block you add. Pass a boolean and
   * the row becomes an option, which means its list must be a `listbox`; use
   * `PickerList`.
   */
  selected?: boolean
  trailing?: React.ReactNode
  className?: string
} & Omit<React.ComponentPropsWithoutRef<"button">, "children">) {
  const dense = size === 40
  const option = selected !== undefined

  // A row you can click is a button. A row whose only target is the control at
  // its end — "In this stack", with its remove — is not: a button cannot hold
  // another button, and nesting them is what breaks both for the keyboard.
  const Element = (onClick ? "button" : "div") as React.ElementType

  return (
    <Element
      type={onClick ? "button" : undefined}
      onClick={onClick}
      role={option ? "option" : undefined}
      aria-selected={option ? selected : undefined}
      className={cn(
        "group/row flex w-full items-center rounded-md text-left transition-colors",
        onClick && "focus-ring-edge",
        // 4px, not the 5 the prototype used: its two lines were ~15px each and
        // ours are the scale's 16, so 4 is what lands the rung on exactly 40.
        dense ? "min-h-10 gap-[9px] px-2 py-1" : "min-h-14 gap-3 px-3 py-2",
        // Branched, never stacked as `hover:` variants — a selected row being
        // hovered matches both and which wins is not predictable (§4).
        selected
          ? "bg-[var(--wash-selected)] hover:bg-[var(--wash-selected-hover)]"
          : "hover:bg-[var(--wash-hover)]",
        className,
      )}
      {...props}
    >
      <span
        aria-hidden
        className={cn(
          "border-border bg-control text-fg-2 flex flex-none items-center justify-center border",
          "[&_svg]:pointer-events-none [&_svg]:shrink-0",
          dense
            ? "size-6 rounded-sm [&_svg]:size-3.5"
            : "size-8 rounded-md [&_svg]:size-4",
        )}
      >
        {icon}
      </span>

      <span className="flex min-w-0 flex-1 flex-col">
        <span
          className={cn(
            "text-foreground truncate",
            // The dense rung names a resource instance — `postgres-2` is a
            // value a machine produced and will read back (§6).
            dense ? "text-meta font-mono" : "text-name font-medium",
          )}
        >
          {name}
        </span>
        {meta && meta.length > 0 && (
          <span
            className={cn(
              "flex gap-2 overflow-hidden whitespace-nowrap",
              dense ? "text-label text-fg-muted" : "text-meta text-fg-2",
            )}
          >
            {meta.map((part, index) => (
              <React.Fragment key={`${part.text}-${index}`}>
                {index > 0 && <span className="text-fg-ghost">·</span>}
                <span className={cn("truncate", part.mono && "font-mono")}>{part.text}</span>
              </React.Fragment>
            ))}
          </span>
        )}
      </span>

      {(endText || trailing) && (
        <span className="text-meta text-fg-muted flex flex-none items-center gap-2.5">
          {endText}
          {trailing}
        </span>
      )}
    </Element>
  )
}

/**
 * The list the selectable rows need. A `role="option"` outside a `listbox` is
 * not a valid option — this is the container that makes it one, and it is the
 * row's own list rather than a second component.
 *
 * A row that *acts* rather than selects — a block you add — does not belong in
 * here. Put those in a plain `<div className="flex flex-col gap-0.5">`.
 */
export function PickerList({
  multiple,
  className,
  children,
  ...props
}: {
  /** Add-ons can be linked several at a time; a repository cannot. */
  multiple?: boolean
  className?: string
  children: React.ReactNode
} & Omit<React.ComponentPropsWithoutRef<"div">, "children">) {
  return (
    <div
      role="listbox"
      aria-multiselectable={multiple || undefined}
      className={cn("flex flex-col gap-0.5", className)}
      {...props}
    >
      {children}
    </div>
  )
}

/** Selection, said once and in the one colour that carries it (§5). */
export function PickerRowTick() {
  return <Check aria-hidden className="text-ring size-4" />
}

/** How many of this block are in the stack already — `postgres`, `postgres-2`. */
export function PickerRowCount({ n }: { n: number }) {
  return (
    <span className="border-border bg-control text-foreground rounded-sm border px-1.5 py-px font-mono text-[11px] leading-4">
      ×{n}
    </span>
  )
}

/** The affordance on a row that adds rather than selects. Inks on approach. */
export function PickerRowAdd() {
  return <Plus aria-hidden className="group-hover/row:text-foreground size-4 transition-colors" />
}

/**
 * Takes one instance back out. It is the only target on its row, which is why
 * the row around it is not itself clickable.
 *
 * It is `opacity-0` at rest and appears on approach (§11 — actions on hover),
 * but it never *goes* on focus, or the keyboard could not reach it.
 */
export function PickerRowRemove({
  label,
  onRemove,
}: {
  /** "Remove postgres-2" — the row's name, so a screen reader knows which. */
  label: string
  onRemove: () => void
}) {
  return (
    <button
      type="button"
      aria-label={label}
      onClick={onRemove}
      className={cn(
        "focus-ring-edge flex size-[22px] items-center justify-center rounded-sm opacity-0 transition-opacity",
        "group-hover/row:opacity-100 focus-visible:opacity-100",
        "hover:bg-[var(--wash-hover)] hover:text-foreground",
        "[&_svg]:size-3.5",
      )}
    >
      <X aria-hidden />
    </button>
  )
}
