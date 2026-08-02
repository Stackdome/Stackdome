import * as React from "react"

import { cn } from "@/lib/utils"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      data-slot="input"
      className={cn(
        "file:text-foreground placeholder:text-fg-ghost selection:bg-primary selection:text-primary-foreground bg-input border-border flex h-10 w-full min-w-0 rounded-full border px-[15px] text-[13px] font-normal transition-[color,box-shadow] file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
        "focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2",
        "aria-invalid:border-danger",
        className
      )}
      {...props}
    />
  )
}

export { Input }
