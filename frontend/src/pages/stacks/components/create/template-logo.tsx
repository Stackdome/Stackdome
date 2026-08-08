import { useState } from "react"

import { cn } from "@/lib/utils"
import type { Template } from "@/pages/stacks/data/templates/types"

/**
 * A ready-made app's own mark.
 *
 * **It is never tinted** (§7). A Lucide glyph takes `ink`/`fg-2` because it is
 * our drawing; a Grafana or Immich logo keeps its own colour because it is
 * theirs. No `currentColor`, no filter, no opacity.
 *
 * Falls back to the record's two-letter `initials` when the art is missing or
 * fails to load — a template with a broken asset still has to be pickable.
 */
export function TemplateLogo({
  template,
  size = 18,
  className,
}: {
  template: Template
  /** Logos read optically lighter than a stroked glyph, so they run a rung
   *  larger than the 16px Lucide icons they sit beside. */
  size?: number
  className?: string
}) {
  const [broken, setBroken] = useState(false)

  if (!template.icon || broken) {
    return (
      <span
        aria-hidden
        className={cn("text-foreground font-mono font-semibold", className)}
        style={{ fontSize: Math.round(size * 0.62) }}
      >
        {template.initials}
      </span>
    )
  }

  return (
    <img
      src={template.icon}
      alt=""
      aria-hidden
      width={size}
      height={size}
      className={cn("object-contain", className)}
      onError={() => setBroken(true)}
    />
  )
}
