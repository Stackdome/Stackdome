import * as React from "react"
import * as SwitchPrimitive from "@radix-ui/react-switch"

import { cn } from "@/lib/utils"

function Switch({
  className,
  ...props
}: React.ComponentProps<typeof SwitchPrimitive.Root>) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        "peer data-[state=checked]:bg-brand data-[state=checked]:border-brand data-[state=unchecked]:bg-muted data-[state=unchecked]:border-border-strong focus-visible:ring-brand/30 focus-visible:ring-offset-2 focus-visible:ring-offset-background inline-flex h-[1.15rem] w-8 shrink-0 items-center rounded-full border shadow-xs transition-colors outline-none focus-visible:ring-2 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        className={cn(
          "pointer-events-none block size-3.5 translate-x-0.5 rounded-full bg-background shadow-sm ring-0 transition-transform data-[state=checked]:translate-x-[calc(100%-2px)] data-[state=unchecked]:bg-muted-foreground"
        )}
      />
    </SwitchPrimitive.Root>
  )
}

export { Switch }
