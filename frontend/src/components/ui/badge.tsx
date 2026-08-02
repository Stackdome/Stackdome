import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-full border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2 aria-invalid:border-danger transition-colors overflow-hidden",
  {
    variants: {
      variant: {
        default: "border-transparent bg-foreground/5 text-fg-2 [a&]:hover:bg-foreground/10",
        secondary: "border-border bg-transparent text-fg-2 [a&]:hover:bg-foreground/5",
        destructive:
          "border-danger-border bg-danger-bg text-danger font-mono text-[10.5px] uppercase tracking-[0.14em] [a&]:hover:bg-danger-bg/70",
        outline: "border-border text-foreground bg-transparent [a&]:hover:bg-foreground/5",
        success: "border-success-border bg-success-bg text-success font-mono text-[10.5px] uppercase tracking-[0.14em]",
        warning: "border-warn-border bg-warn-bg text-warn font-mono text-[10.5px] uppercase tracking-[0.14em]",
        info: "border-info-border bg-info-bg text-info font-mono text-[10.5px] uppercase tracking-[0.14em]",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span"

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
