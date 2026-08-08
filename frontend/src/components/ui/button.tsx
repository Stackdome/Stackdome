import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { Loader2 } from "lucide-react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  // Radius is NOT set here — it belongs to `shape`, because a button's radius
  // reports what KIND of action it is, not how big it is. See the shape variant.
  //
  // DISABLED: dimmed, not-allowed cursor, inert. The dim is the whole signal,
  // so the button must still RECEIVE the pointer for the cursor to show —
  // `pointer-events-none` would hand the cursor back to whatever is underneath.
  // Nothing blocks the click in its place: the native `disabled` attribute
  // already does that. What did need replacing is hover and press, which still
  // match a disabled button in CSS; every interaction state is now guarded by
  // `not-disabled:`. Not `enabled:` — that matches form controls only, so an
  // `asChild` <a> would lose its hover entirely.
  //
  // `loading` keeps full contrast (§6 — a request in flight is not a disabled
  // control) and says so with a progress cursor rather than a refusal.
  "inline-flex items-center justify-center gap-1 whitespace-nowrap text-body font-medium transition-[background-color,box-shadow] duration-150 not-disabled:active:[&>*]:translate-y-px [&>*]:transition-transform [&>*]:duration-75 disabled:cursor-not-allowed data-[loading]:cursor-progress disabled:[&:not([data-loading])]:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-3.5 shrink-0 [&_svg]:shrink-0 focus-ring-edge aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        // The three DARK-FACED variants take the ring WITH its gap. Everywhere
        // else the ring lands flush, but a solid ink ring on a near-black face
        // is invisible — the gap is the entire signal there, not a refinement.
        //
        // Press and focus are BOTH box-shadow, so `active:` would simply
        // overwrite `focus-visible:` and a pressed button would lose its ring
        // mid-click. Each variant therefore spells the combined state out:
        // `focus-visible:active:` is two pseudo-classes against one, so it wins
        // on specificity rather than on declaration order.
        default:
          "bg-primary text-primary-foreground not-disabled:hover:bg-primary-hover not-disabled:active:bg-primary-press not-disabled:active:shadow-[var(--btn-press-strong)] [--press-shadow:var(--btn-press-strong)] focus-ring",
        destructive:
          "bg-danger text-white not-disabled:hover:opacity-90 not-disabled:active:opacity-100 not-disabled:active:shadow-[var(--btn-press-mid)] [--press-shadow:var(--btn-press-mid)] focus-ring",
        outline:
          "bg-control text-foreground border border-border not-disabled:hover:bg-control-hover not-disabled:active:bg-control not-disabled:active:shadow-[var(--btn-press-soft)] [--press-shadow:var(--btn-press-soft)]",
        // Same material as `outline`, no border — the two looked doubled up
        // side by side in stories, so secondary drops the hairline.
        secondary:
          "bg-control text-foreground not-disabled:hover:bg-control-hover not-disabled:active:bg-control not-disabled:active:shadow-[var(--btn-press-soft)] [--press-shadow:var(--btn-press-soft)]",
        ghost:
          "bg-transparent text-foreground not-disabled:hover:bg-[var(--wash-hover)] not-disabled:active:bg-[var(--wash-pressed)] not-disabled:active:shadow-[var(--btn-press-soft)] [--press-shadow:var(--btn-press-soft)]",
        link: "text-foreground underline-offset-4 not-disabled:hover:underline",
        // Inverse foreground CTA — used on auth pages to contrast with the cream/navy band.
        // Foreground fill at rest; hover is a one-step opacity shift, same as
        // `destructive` (rubric: black is the only action colour, orange is
        // never a button fill).
        inverse:
          "bg-foreground text-background not-disabled:hover:opacity-90 not-disabled:active:opacity-100 not-disabled:active:shadow-[var(--btn-press-strong)] [--press-shadow:var(--btn-press-strong)] focus-ring",
      },
      // Three heights, and that is the whole ladder. Height follows DENSITY,
      // never importance — an important button gets filled, not taller.
      //   sm   28px  chips, in-row actions, anything inside a dense row
      //   default 32px  page toolbars, dialogs footers — the everyday button
      //   lg   40px  form flows only, paired with 40px inputs
      //
      // Horizontal padding is OPTICALLY corrected, not mathematically equal.
      // The eye aligns on centre of mass, not on bounding boxes: an icon
      // carries invisible safe area inside its box (a 14px lucide glyph
      // measures ~2.9px of air each side, a chevron ~4.5px) while a letterform
      // meets the edge with almost none.
      //
      // One rule, derived from the size's base inset — NOT hand-set per size,
      // which is how `sm` drifted to only +0.5px of correction:
      //     icon side  = base - 3
      //     label side = base + 2
      // The -3 pays back the glyph's safe area; the +2 is the optical margin
      // dense letterforms need to read level against it.
      //
      //   sm       base 10  ->   7 / 12
      //   default  base 12  ->   9 / 14
      //   lg       base 15  ->  12 / 17
      //
      // Symmetric stays symmetric when there is nothing to correct against:
      // text-only, icon-only, and icons on both sides (the not-has guards).
      // A pill needs no curvature correction here — at the vertical midline,
      // where single-line text sits, the rounded edge is at its full extent.
      size: {
        default: "h-8 px-3 has-[svg]:px-3 has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pl-[9px] has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pr-[14px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pr-[9px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pl-[14px]",
        sm: "h-7 gap-1 px-2.5 has-[svg]:px-2.5 has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pl-[7px] has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pr-[12px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pr-[7px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pl-[12px]",
        lg: "h-10 px-[15px] has-[svg]:px-[15px] has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pl-[12px] has-[>span>svg:first-child]:not-has-[>span>svg:last-child]:pr-[17px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pr-[12px] has-[>span>svg:last-child]:not-has-[>span>svg:first-child]:pl-[17px]",
        icon: "size-8",
        "icon-sm": "size-7",
      },
      // SHAPE REPORTS THE CLASS OF THE ACTION — not its size. It is the fastest
      // signal on the screen: you read it before you read the label.
      //
      //   pill  this COMMITS. It finishes a flow or ships something.
      //         Auth `Continue`, the action in an empty state, `Deploy`,
      //         `Publish`, a wizard's final step. At most ONE per screen —
      //         if a screen has two, one of them is not a commitment.
      //
      //   flat  a WORKING control. Toolbars, filters, row actions, dialog
      //         footers, `Save`, `Cancel`. The bulk of every screen.
      //
      // Measured off OpenAI's Platform console across 417 screens: their flat
      // radius is 6px on a 36px control. Our step is 32px on a 13px base
      // (§3, §5), so the proportional radius is 5.3 — `rounded-sm` is the rung
      // that lands there, and it is the token chips already use.
      //
      // `flat` carries NO radius of its own — it defers to the height ladder in
      // §2, set by the compoundVariants below. Anything the same height takes
      // the same radius, so a flat button matches the input and the select
      // standing next to it in a toolbar. `pill` is the only value that
      // overrides height.
      //
      // `flat` is the DEFAULT, so new work is correct without thinking about
      // it. Existing screens are corrected as we work each journey — deliberate
      // pills get restored screen by screen, never retrofitted in a sweep.
      shape: {
        pill: "rounded-full",
        flat: "",
      },
    },
    // §2 — radius is a function of HEIGHT: 28/6 · 32/8 · 40/12.
    // Three unrelated corner values in one row is the kind of thing the eye
    // catches even when the mind does not.
    compoundVariants: [
      { shape: "flat", size: "sm", class: "rounded-sm" },          // 28px
      { shape: "flat", size: "icon-sm", class: "rounded-sm" },     // 28px
      { shape: "flat", size: "default", class: "rounded-md" },     // 32px
      { shape: "flat", size: "icon", class: "rounded-md" },        // 32px
      { shape: "flat", size: "lg", class: "rounded-lg" },          // 40px
    ],
    defaultVariants: {
      variant: "default",
      size: "default",
      shape: "flat",
    },
  }
)

/** Text nodes cannot be transformed or targeted by `:first-child`, so every
 *  child becomes an element: it is what makes the press travel and the optical
 *  padding work. */
function wrapChildren(children: React.ReactNode) {
  return React.Children.map(children, (child) =>
    typeof child === "string" || typeof child === "number" ? <span>{child}</span> : child
  )
}

function Button({
  className,
  variant,
  size,
  shape,
  asChild = false,
  loading = false,
  loadingText,
  children,
  disabled,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
    /** Swaps the content for a spinner and makes the button inert. */
    loading?: boolean
    /**
     * What the button says while it works — "Creating…", "Deploying…". Say what
     * is happening, not that something is: a spinner already reports *that*.
     * Replaces `children` wholesale, so a leading icon disappears on its own.
     */
    loadingText?: React.ReactNode
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      data-loading={loading ? "" : undefined}
      aria-busy={loading || undefined}
      disabled={disabled || loading}
      className={cn(
        buttonVariants({ variant, size, shape }),
        // An `asChild` button is usually an <a>, where `disabled` is not a real
        // attribute and blocks nothing. This is the one case that still needs
        // pointer-events to go — it costs the cursor, but a navigable link that
        // is meant to be inert is the worse failure.
        asChild && (disabled || loading) && "pointer-events-none",
        className
      )}
      {...props}
    >
      {asChild ? (
        children
      ) : (
        <span data-slot="button-content" className="inline-flex items-center justify-center gap-1">
          {loading ? (
            <>
              <Loader2 className="animate-spin" aria-hidden />
              {wrapChildren(loadingText ?? children)}
            </>
          ) : (
            wrapChildren(children)
          )}
        </span>
      )}
    </Comp>
  )
}

export { Button, buttonVariants }
