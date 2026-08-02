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
          "border-danger-border bg-danger-bg text-danger [a&]:hover:bg-danger-bg/70",
        outline: "border-border text-foreground bg-transparent [a&]:hover:bg-foreground/5",
        success: "border-success-border bg-success-bg text-success",
        warning: "border-warn-border bg-warn-bg text-warn",
        info: "border-info-border bg-info-bg text-info",
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
