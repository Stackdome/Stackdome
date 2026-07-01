import { useState } from "react";
import { Check, Plus, Search } from "lucide-react";
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { blockCatalog, BLOCK_CATEGORY_META } from "@/pages/stacks/data/blocks/registry";
import { BlockPicker } from "@/pages/stacks/components/wizard/block-picker";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";

interface AddResourcePopoverProps {
  /** Block ids already present in the stack (shows the added badge). */
  addedIds: string[];
  onAdd: (blockId: string) => void;
  /** Provisioned managed addons available to link. */
  addons: { id: string; name: string }[];
  /** Addon ids already linked to this stack. */
  linkedAddonIds: ReadonlySet<string>;
  onLinkAddon: (addonId: string) => void;
}

/**
 * "+ Add resource" control for the canvas. Reuses the wizard's BlockPicker grid
 * and block catalog so the add experience matches the create flow. Stays open
 * after an add so several blocks can be dropped in a row.
 */
export function AddResourcePopover({ addedIds, onAdd, addons, linkedAddonIds, onLinkAddon }: AddResourcePopoverProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");

  const visibleAddons = addons.filter(
    (a) => !query.trim() || a.name.toLowerCase().includes(query.trim().toLowerCase()),
  );

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-1.5 rounded-md border border-border bg-card px-2.5 py-1.5 text-[13px] font-medium text-foreground shadow-xs transition-colors hover:bg-muted"
        >
          <Plus className="size-3.5" />
          Add resource
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-[420px] p-0">
        <div className="relative border-b border-border p-2">
          <Search className="absolute left-4 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search services, data stores…"
            className="pl-9"
          />
        </div>
        <div className="max-h-[360px] overflow-y-auto p-3">
          <BlockPicker
            catalog={blockCatalog}
            categories={BLOCK_CATEGORY_META}
            addedIds={addedIds}
            query={query}
            onAdd={onAdd}
          />
          {visibleAddons.length > 0 && (
            <div className="mt-5">
              <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
                Managed add-ons
              </div>
              <div className="grid grid-cols-2 gap-2.5">
                {visibleAddons.map((a) => {
                  const linked = linkedAddonIds.has(a.id);
                  return (
                    <button
                      type="button"
                      key={a.id}
                      onClick={() => onLinkAddon(a.id)}
                      className="flex min-h-[60px] items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary"
                    >
                      <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                        <AddonTypeIcon type="postgres" size={18} />
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block text-sm font-medium text-foreground">
                          {a.name}
                        </span>
                        <span className="block truncate font-mono text-[11px] text-muted-foreground">
                          managed postgres
                        </span>
                      </span>
                      {linked ? (
                        <Check className="h-[17px] w-[17px] text-success" />
                      ) : (
                        <Plus className="h-[17px] w-[17px] text-primary" />
                      )}
                    </button>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
