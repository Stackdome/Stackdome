import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full text-[13px] font-medium transition-[background-color,box-shadow] duration-150 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2 aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-[var(--edge)] hover:bg-primary-hover active:shadow-[var(--press)]",
        destructive:
          "bg-danger text-white shadow-[var(--edge)] hover:opacity-90 active:shadow-[var(--press)]",
        outline:
          "bg-control text-foreground border border-border hover:bg-control-hover shadow-[var(--edge)] active:shadow-[var(--press)]",
        // Same material as `outline`, no border — the two looked doubled up
        // side by side in stories, so secondary drops the hairline.
        secondary:
          "bg-control text-foreground hover:bg-control-hover shadow-[var(--edge)] active:shadow-[var(--press)]",
        ghost:
          "bg-transparent text-foreground hover:bg-foreground/5 active:shadow-[var(--press)]",
        link: "text-foreground underline-offset-4 hover:underline",
        // Inverse foreground CTA — used on auth pages to contrast with the cream/navy band.
        // Foreground fill at rest; hover is a one-step opacity shift, same as
        // `destructive` (rubric: black is the only action colour, orange is
        // never a button fill).
        inverse:
          "bg-foreground text-background shadow-[var(--edge)] hover:opacity-90 active:shadow-[var(--press)]",
        // Quiet neutral text button (D15: buttons are never mono — this used
        // to be a monospace terminal CTA). Kept for API stability; call sites
        // should migrate to `ghost` during their page's graphite pass.
        mono:
          "bg-transparent text-foreground hover:bg-foreground/5 active:shadow-[var(--press)]",
        // Console-rail primary — control-fill material, ONLY for persistent
        // action rails (session edit bar, command palette, debug overlays)
        // where the surrounding chrome is already terminal-flavored. Do NOT
        // use for in-page CTAs — use `default` instead.
        railPrimary:
          "bg-control text-foreground hover:bg-control-hover shadow-[var(--edge)] active:shadow-[var(--press)]",
        // Console-rail secondary — paired with railPrimary (Cancel, Discard).
        railGhost:
          "bg-transparent text-foreground hover:bg-foreground/5 active:shadow-[var(--press)]",
        // Console-rail destructive — paired with railPrimary in destructive flows.
        railDanger:
          "bg-transparent text-danger hover:bg-danger-bg",
      },
      size: {
        // --ctl-h-l ladder (40px height / 15px inline padding)
        default: "h-10 px-[15px] has-[>svg]:px-3",
        sm: "h-8 gap-1.5 px-3 has-[>svg]:px-2.5",
        lg: "h-11 px-8 has-[>svg]:px-6",
        icon: "size-10",
        rail: "h-8 gap-1.5 px-3",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
