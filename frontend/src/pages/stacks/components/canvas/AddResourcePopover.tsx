import { useState } from "react";
import { Check, ExternalLink, HardDrive, Plus, Search } from "lucide-react";
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import { blockCatalog, BLOCK_CATEGORY_META } from "@/pages/stacks/data/blocks/registry";
import { BlockPicker, blockMatchesQuery } from "@/pages/stacks/components/wizard/block-picker";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";

const SERVICE_CATEGORIES = BLOCK_CATEGORY_META.filter((c) => c.id === "services");
const DATA_CATEGORIES = BLOCK_CATEGORY_META.filter((c) => c.id !== "services");

export interface AddResourcePanelProps {
  /** Block ids already present in the stack (shows the added badge). */
  addedIds: string[];
  onAdd: (blockId: string) => void;
  /** Provisioned managed addons available to link. */
  addons: { id: string; name: string }[];
  /** Addon ids already linked to this stack. */
  linkedAddonIds: ReadonlySet<string>;
  onLinkAddon: (addonId: string) => void;
  /** False while the stack has no services — volumes are born attached. */
  canAddVolume: boolean;
  onAddVolume: () => void;
  /** Close the hosting surface (popover / context panel) before a modal opens. */
  onRequestClose: () => void;
}

/**
 * The add-resource picker body: search plus Services, Storage, managed
 * add-ons, and the data-store groups. Shared between the "+ Add resource"
 * popover and the canvas right-click panel so both open the same catalog.
 */
export function AddResourcePanel({
  addedIds,
  onAdd,
  addons,
  linkedAddonIds,
  onLinkAddon,
  canAddVolume,
  onAddVolume,
  onRequestClose,
}: AddResourcePanelProps) {
  const [query, setQuery] = useState("");

  const visibleAddons = addons.filter(
    (a) => !query.trim() || a.name.toLowerCase().includes(query.trim().toLowerCase()),
  );
  const trimmedQuery = query.trim().toLowerCase();
  // "Storage" is the section header shown for the Volume tile — match either name.
  const showStorageSection = !trimmedQuery || "volume".includes(trimmedQuery) || "storage".includes(trimmedQuery);
  const anyBlockMatches = blockCatalog.some((b) => blockMatchesQuery(b, query));
  const nothingMatches = !anyBlockMatches && !showStorageSection && visibleAddons.length === 0;

  const volumeTile = (
    <button
      type="button"
      disabled={!canAddVolume}
      onClick={() => {
        onRequestClose();
        onAddVolume();
      }}
      className="flex min-h-[60px] items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary disabled:cursor-not-allowed disabled:opacity-50"
    >
      <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
        <HardDrive className="size-[18px]" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-sm font-medium text-foreground">Volume</span>
        <span className="block truncate font-mono text-[11px] text-muted-foreground">persistent storage</span>
      </span>
      <Plus className="h-[17px] w-[17px] text-primary" />
    </button>
  );

  return (
    <>
      <div className="relative border-b border-border p-2">
        <Search className="absolute left-4 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search services, data stores…"
          className="pl-9"
        />
      </div>
      <div className="max-h-[440px] overflow-y-auto p-3">
        {nothingMatches && (
          <p className="px-1 py-6 text-sm text-muted-foreground">No matches for "{query}"</p>
        )}
        {/* Section order: services first, storage + managed add-ons next, and the
            long data-store list last so the common picks stay above the fold. */}
        <BlockPicker
          catalog={blockCatalog}
          categories={SERVICE_CATEGORIES}
          addedIds={addedIds}
          query={query}
          onAdd={onAdd}
          hideEmptyMessage
        />
        {showStorageSection ? (
          <div className="mt-5">
            <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
              Storage
            </div>
            {/* Same half-width card size as the block tiles. */}
            <div className="grid grid-cols-2 gap-2.5">
              {!canAddVolume ? (
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <span>{volumeTile}</span>
                  </TooltipTrigger>
                  <TooltipContent side="top">
                    Add a service first — volumes attach to a service.
                  </TooltipContent>
                </Tooltip>
              ) : (
                volumeTile
              )}
            </div>
          </div>
        ) : null}
        <div className="mt-5">
          <BlockPicker
            catalog={blockCatalog}
            categories={DATA_CATEGORIES}
            addedIds={addedIds}
            query={query}
            onAdd={onAdd}
            hideEmptyMessage
          />
        </div>
        {/* Last section, matching the wizard's block composer order. */}
        {(visibleAddons.length > 0 || (addons.length === 0 && !trimmedQuery)) && (
          <div className="mt-5">
            <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
              Managed add-ons
            </div>
            {addons.length === 0 ? (
              // No managed addons in the workspace yet — offer a way to create
              // one (same empty state as the wizard's block composer).
              <a
                href="/addons"
                target="_blank"
                rel="noopener noreferrer"
                className="flex min-h-[60px] items-center gap-3 rounded-md border border-dashed bg-card/40 px-3 py-3 text-left transition-colors hover:border-primary hover:bg-card"
              >
                <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                  <Plus className="h-[18px] w-[18px]" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block text-sm font-medium text-foreground">Create an add-on</span>
                  <span className="block text-[12px] text-muted-foreground">
                    No managed add-ons yet. Set one up in the Addons page.
                  </span>
                </span>
                <ExternalLink className="h-4 w-4 flex-none text-muted-foreground" />
              </a>
            ) : (
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
            )}
          </div>
        )}
      </div>
    </>
  );
}

type AddResourcePopoverProps = Omit<AddResourcePanelProps, "onRequestClose">;

/**
 * "+ Add resource" control for the canvas. Reuses the wizard's BlockPicker grid
 * and block catalog so the add experience matches the create flow. Stays open
 * after an add so several blocks can be dropped in a row.
 */
export function AddResourcePopover(props: AddResourcePopoverProps) {
  const [open, setOpen] = useState(false);

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
      {/* align="end": the trigger sits at the canvas's top-right, so the wide
          panel must open leftward to stay on screen. */}
      <PopoverContent align="end" className="w-[560px] p-0">
        <AddResourcePanel {...props} onRequestClose={() => setOpen(false)} />
      </PopoverContent>
    </Popover>
  );
}
