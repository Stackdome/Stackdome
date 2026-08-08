import { useState } from "react"
import { EmptyState, PickerList, PickerRow, PickerRowTick, SearchGlyph } from "@/components/branded"
import { Button } from "@/components/ui/button"
import { templates } from "@/pages/stacks/data/templates/registry"
import type { Template } from "@/pages/stacks/data/templates/types"

import { SearchField } from "../search-field"
import { templateServices } from "../template-services"
import { TemplateLogo } from "../template-logo"

/**
 * `v2.27.4`, once.
 *
 * Three of the seven records already carry the `v` and four do not — prefixing
 * blindly printed `vv3.20.189-lts`. Normalising here rather than editing the
 * records keeps a template's `version` the value its upstream publishes.
 */
function versionLabel(version: string) {
  return `v${version.replace(/^v/i, "")}`
}

/**
 * The curated apps. One list, and a detail panel that appears with the first
 * pick — the same rule the "In this stack" panel follows.
 *
 * The detail shows **every field the record actually carries** — category,
 * version, the long blurb, the website and the docs. The old browser panel
 * showed all of it and an earlier pass of this redesign showed none of it,
 * which quietly made the templates less useful than the thing being replaced.
 */
export function TemplateTab({
  picked,
  onPick,
}: {
  picked: Template | null
  onPick: (template: Template) => void
}) {
  const [query, setQuery] = useState("")
  const q = query.trim().toLowerCase()
  const matches = templates.filter(
    (t) => !q || t.name.toLowerCase().includes(q) || t.shortDescription.toLowerCase().includes(q),
  )

  return (
    <div className="flex flex-col gap-2.5">
      <SearchField
        value={query}
        onChange={setQuery}
        placeholder="Search ready-made apps…"
        label="Search ready-made apps"
      />

      {matches.length === 0 ? (
        <EmptyState
          icon={<SearchGlyph />}
          title="No app matches that"
          description="Try a shorter word, or deploy your own code instead."
        />
      ) : (
        <PickerList aria-label="Ready-made apps">
          {matches.map((template) => {
            const services = templateServices(template)
            return (
              <PickerRow
                key={template.id}
                icon={<TemplateLogo template={template} />}
                name={template.name}
                meta={[
                  { text: template.shortDescription },
                  { text: versionLabel(template.version), mono: true },
                ]}
                endText={`${services.length} ${services.length === 1 ? "service" : "services"}`}
                selected={picked?.id === template.id}
                trailing={picked?.id === template.id ? <PickerRowTick /> : null}
                onClick={() => onPick(template)}
              />
            )
          })}
        </PickerList>
      )}
    </div>
  )
}

/** The right-hand detail for the picked app. Rendered by the page beside the
 *  list, in the same slot the "In this stack" panel uses. */
export function TemplateDetail({ template }: { template: Template }) {
  const services = templateServices(template)
  return (
    <aside className="border-border-subtle sticky top-0 w-[300px] flex-none border-l pl-6">
      <span className="border-border bg-control flex size-11 items-center justify-center rounded-lg border">
        <TemplateLogo template={template} size={26} />
      </span>
      <h2 className="text-title text-foreground mt-2.5 font-semibold">{template.name}</h2>
      {/* The category is a word, so it is set in the interface face. Only the
          version is a machine value, and only it takes mono (§6). */}
      <p className="text-label text-fg-muted mt-0.5">
        {template.category} · <span className="font-mono">{versionLabel(template.version)}</span>
      </p>
      <p className="text-meta text-fg-2 mt-2 leading-[18px]">{template.longDescription}</p>
      <div className="mt-3 flex gap-1.5">
        <Button asChild variant="outline" size="sm">
          <a href={template.website} target="_blank" rel="noopener noreferrer">
            Website
          </a>
        </Button>
        <Button asChild variant="outline" size="sm">
          <a href={template.docs} target="_blank" rel="noopener noreferrer">
            Docs
          </a>
        </Button>
      </div>

      {/* What you are actually about to get. "A ready-made app" is one row in
          the list but four services on the canvas, and this is the only place
          before Create stack where that is stated. */}
      <div className="border-border-subtle mt-3.5 border-t pt-3">
        <div className="text-label text-fg-muted">Includes · {services.length}</div>
        <div className="mt-1.5 flex flex-col gap-px">
          {services.map((service) => (
            <div key={service} className="text-meta text-foreground font-mono">
              {service}
            </div>
          ))}
        </div>
      </div>
    </aside>
  )
}
