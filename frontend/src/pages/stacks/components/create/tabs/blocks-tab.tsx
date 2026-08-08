import { useState } from "react"
import { Database } from "lucide-react"

import {
  EmptyState,
  PickerList,
  PickerRow,
  PickerRowAdd,
  PickerRowCount,
  PickerRowTick,
  SearchGlyph,
} from "@/components/branded"
import { BlockGlyph } from "@/pages/stacks/components/blocks/block-glyph"
import { blockCatalog, BLOCK_CATEGORY_META } from "@/pages/stacks/data/blocks/registry"
import { blockMatchesQuery } from "@/pages/stacks/components/blocks/block-picker"
import { usePostgresAddons } from "@/hooks/use-postgres-addons"

import { SearchField } from "../search-field"
import type { BlockInstance } from "../selection"

/**
 * The block catalogue — the parts you assemble a stack out of.
 *
 * **A block adds; it does not select.** The real `addBlockToStack` de-duplicates
 * with `uniqueName()`, so the same block goes in as many times as you click it:
 * `postgres`, `postgres-2`, `postgres-3`. The row therefore carries a count and
 * a plus rather than a tick, and each instance is removed on its own from the
 * "In this stack" panel.
 *
 * Managed add-ons are the exception: an add-on is a thing that already exists,
 * so linking one **is** a selection and it gets the tick.
 */
export function BlocksTab({
  instances,
  onAddBlock,
  addonIds,
  onToggleAddon,
}: {
  instances: BlockInstance[]
  onAddBlock: (blockId: string) => void
  addonIds: string[]
  onToggleAddon: (addonId: string) => void
}) {
  const [query, setQuery] = useState("")
  const { addons } = usePostgresAddons()

  const matches = blockCatalog.filter((block) => blockMatchesQuery(block, query))
  const q = query.trim().toLowerCase()
  const addonMatches = addons.filter((a) => a.id && (!q || a.name.toLowerCase().includes(q)))
  const countOf = (blockId: string) => instances.filter((i) => i.blockId === blockId).length

  return (
    <div className="flex flex-col gap-2.5">
      <SearchField
        value={query}
        onChange={setQuery}
        placeholder="Search services, data stores and add-ons…"
        label="Search building blocks"
      />

      {matches.length === 0 && addonMatches.length === 0 ? (
        <EmptyState
          icon={<SearchGlyph />}
          title="Nothing matches that"
          description="Try a shorter word — or start from a compose file instead."
        />
      ) : (
        <>
          {BLOCK_CATEGORY_META.map((category) => {
            const inCategory = matches.filter((block) => block.category === category.id)
            if (inCategory.length === 0) return null
            return (
              <div key={category.id}>
                {/* Sentence case, always (§11) — the registry's own labels are
                    uppercase and that is what this heading is correcting. */}
                <div className="text-label text-fg-muted px-2 pb-1.5 pt-4">{category.label}</div>
                <div className="flex flex-col gap-0.5">
                  {inCategory.map((block) => {
                    const n = countOf(block.id)
                    return (
                      <PickerRow
                        key={block.id}
                        icon={<BlockGlyph icon={block.icon} size={16} />}
                        name={block.name}
                        meta={[{ text: block.summary, mono: true }]}
                        onClick={() => onAddBlock(block.id)}
                        trailing={
                          <>
                            {n > 0 && <PickerRowCount n={n} />}
                            <PickerRowAdd />
                          </>
                        }
                      />
                    )
                  })}
                </div>
              </div>
            )
          })}

          <div>
            <div className="text-label text-fg-muted px-2 pb-1.5 pt-4">Managed add-ons</div>
            {addonMatches.length === 0 ? (
              <EmptyState
                title="No managed add-ons yet"
                description="Set one up on the Addons page and it will show here, ready to link."
              />
            ) : (
              <PickerList multiple aria-label="Managed add-ons">
                {addonMatches.map((addon) => {
                  const linked = addonIds.includes(addon.id!)
                  return (
                    <PickerRow
                      key={addon.id}
                      icon={<Database />}
                      name={addon.name}
                      meta={[{ text: "managed postgres", mono: true }]}
                      selected={linked}
                      trailing={linked ? <PickerRowTick /> : null}
                      onClick={() => onToggleAddon(addon.id!)}
                    />
                  )
                })}
              </PickerList>
            )}
          </div>
        </>
      )}
    </div>
  )
}
