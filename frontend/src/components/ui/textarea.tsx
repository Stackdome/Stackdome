import * as React from "react"

import { cn } from "@/lib/utils"

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "border-border placeholder:text-fg-ghost aria-invalid:border-danger bg-input flex field-sizing-content min-h-16 w-full rounded-md border px-3 py-2 text-body font-normal transition-[color,box-shadow] disabled:cursor-not-allowed disabled:opacity-50",
        "focus-ring-edge",
        className
      )}
      {...props}
    />
  )
}

export { Textarea }
