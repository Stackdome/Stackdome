import { useState } from "react";
import { Plus, Search } from "lucide-react";
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { blockCatalog, BLOCK_CATEGORY_META } from "@/pages/stacks/data/blocks/registry";
import { BlockPicker } from "@/pages/stacks/components/wizard/block-picker";

interface AddResourcePopoverProps {
  /** Block ids already present in the stack (shows the added badge). */
  addedIds: string[];
  onAdd: (blockId: string) => void;
}

/**
 * "+ Add resource" control for the canvas. Reuses the wizard's BlockPicker grid
 * and block catalog so the add experience matches the create flow. Stays open
 * after an add so several blocks can be dropped in a row.
 */
export function AddResourcePopover({ addedIds, onAdd }: AddResourcePopoverProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");

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
        </div>
      </PopoverContent>
    </Popover>
  );
}
