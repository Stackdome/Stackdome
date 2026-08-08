import type { ReactNode } from "react"

import { PickerRow, PickerRowRemove } from "@/components/branded"

/** One thing the chosen starting point puts on the canvas. */
export interface StackItem {
  key: string
  /** The resource's own name — machine-set, so the row sets it in mono (§6). */
  name: string
  /** What it is, in words: "service", "linked add-on", "from a public URL". */
  sub: string
  icon: ReactNode
  /** Present only where the user can take this one instance back out. */
  onRemove?: () => void
}

/**
 * The stack so far — the old wizard's one genuinely good panel, kept.
 *
 * **It appears with the first selection and not before.** An empty panel
 * reserving 300px of the sheet to show nothing is the bug this flow already
 * had; the body simply widens when there is nothing to put here.
 *
 * The rows are the dense 40px rung of the same picker row the catalogue uses —
 * the same object, made dense, not a second component.
 */
export function InThisStack({ items }: { items: StackItem[] }) {
  return (
    <aside className="border-border-subtle sticky top-0 w-[300px] flex-none border-l pl-6">
      <div className="text-label text-fg-muted px-2 pb-2 pt-0.5">In this stack · {items.length}</div>
      <div className="flex flex-col gap-px">
        {items.map((item) => (
          <PickerRow
            key={item.key}
            size={40}
            icon={item.icon}
            name={item.name}
            meta={[{ text: item.sub }]}
            trailing={
              item.onRemove && <PickerRowRemove label={`Remove ${item.name}`} onRemove={item.onRemove} />
            }
          />
        ))}
      </div>
      <p className="text-label text-fg-muted px-2 pt-2 leading-4">
        You can rename, connect and configure all of this on the canvas.
      </p>
    </aside>
  )
}
