import { Plus, Check } from "lucide-react";
import type { BlockCategoryMeta, BlockPreset } from "@/pages/stacks/data/blocks/types";
import { cn } from "@/lib/utils";
import { BlockGlyph } from "./block-glyph";

interface BlockPickerProps {
  catalog: BlockPreset[];
  categories: BlockCategoryMeta[];
  addedIds: string[];
  onAdd: (id: string) => void;
  query: string;
}

export function BlockPicker({ catalog, categories, addedIds, onAdd, query }: BlockPickerProps) {
  const q = query.trim().toLowerCase();
  const match = (b: BlockPreset) => !q || b.name.toLowerCase().includes(q) || b.summary.toLowerCase().includes(q);
  const visible = catalog.filter(match);

  if (visible.length === 0) {
    return (
      <p className="px-1 py-6 text-sm text-muted-foreground">
        No matches for "{query}"
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-5">
      {categories.map((cat) => {
        const blocks = visible.filter((b) => b.category === cat.id);
        if (blocks.length === 0) return null;
        return (
          <div key={cat.id}>
            <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
              {cat.label}
              <span className="font-normal normal-case tracking-normal text-muted-foreground/70"> · {cat.note}</span>
            </div>
            <div className="grid grid-cols-2 gap-2.5">
              {blocks.map((b) => {
                const added = addedIds.includes(b.id);
                return (
                  <button
                    type="button"
                    key={b.id}
                    onClick={() => onAdd(b.id)}
                    className={cn(
                      "flex items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary",
                      added && "border-primary/60",
                    )}
                  >
                    <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                      <BlockGlyph icon={b.icon} size={18} />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-medium text-foreground">{b.name}</span>
                      <span className="block truncate font-mono text-[11px] text-muted-foreground">{b.summary}</span>
                    </span>
                    {added ? <Check className="h-[17px] w-[17px] text-success" /> : <Plus className="h-[17px] w-[17px] text-primary" />}
                  </button>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
