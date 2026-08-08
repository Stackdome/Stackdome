import * as React from "react"
import { Check, Minus } from "lucide-react"

import { cn } from "@/lib/utils"

/**
 * A checkbox, on a native `<input type="checkbox">`.
 *
 * There is no Radix checkbox in this project and no dependency for one. A
 * native input needs neither: it is already focusable, already announced, and
 * already keyboard-operable, so the whole component is a styled box drawn over
 * it. The input itself stays in the DOM — hidden with `sr-only`, not
 * `display:none` — which is what keeps the label click, the space key and the
 * accessibility tree working.
 *
 * A Switch is NOT a substitute. A switch is a setting you change; a checkbox is
 * a statement you are making. §6a's level-2 gate is the second thing.
 */
function Checkbox({
  className,
  checked,
  defaultChecked,
  indeterminate = false,
  onCheckedChange,
  disabled,
  ...props
}: Omit<React.ComponentProps<"input">, "type" | "onChange"> & {
  indeterminate?: boolean
  onCheckedChange?: (checked: boolean) => void
}) {
  const ref = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    if (ref.current) ref.current.indeterminate = indeterminate
  }, [indeterminate])

  return (
    <span className={cn("relative inline-flex size-4 shrink-0", className)}>
      <input
        ref={ref}
        type="checkbox"
        data-slot="checkbox"
        checked={checked}
        defaultChecked={defaultChecked}
        disabled={disabled}
        onChange={(e) => onCheckedChange?.(e.target.checked)}
        // Transparent and stretched over the box rather than `sr-only`: a
        // clipped 1px input is not where the pointer lands, so the click falls
        // through to the decorative span and is swallowed. This way the real
        // control IS the hit target, and the styled box is just paint under it.
        className="peer absolute inset-0 z-10 m-0 cursor-pointer opacity-0 disabled:cursor-not-allowed"
        {...props}
      />
      <span
        aria-hidden
        className={cn(
          // 4px on a 16px box — the same proportion §2's control ladder runs
          // (28/6 · 32/8 · 40/12), continued down to the smallest element.
          "pointer-events-none inline-flex size-4 items-center justify-center rounded-[4px] border border-border-strong bg-input transition-colors",
          "peer-checked:border-primary peer-checked:bg-primary peer-checked:text-primary-foreground",
          "peer-indeterminate:border-primary peer-indeterminate:bg-primary peer-indeterminate:text-primary-foreground",
          "focus-ring-peer",
          "peer-disabled:cursor-not-allowed peer-disabled:opacity-50",
          // The glyph is a DESCENDANT of this span, not a sibling of the input,
          // so `peer-checked:` cannot reach it directly — the variant has to be
          // applied here and reach down.
          "[&_svg]:opacity-0 peer-checked:[&_svg]:opacity-100 peer-indeterminate:[&_svg]:opacity-100",
        )}
      >
        {indeterminate ? (
          <Minus className="size-3 transition-opacity" strokeWidth={3} />
        ) : (
          <Check className="size-3 transition-opacity" strokeWidth={3} />
        )}
      </span>
    </span>
  )
}

export { Checkbox }
