import { useMemo, useState } from "react";
import { ExternalLink, PlusCircle, Puzzle, X } from "lucide-react";
import { Panel } from "@/components/branded";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

interface AddonsInStackPanelProps {
  // Kept for API compatibility with callers; env vars no longer derive addon
  // links so this is unused.
  resources?: Partial<FormStackResourceData>[];
  linkedAddonIds: Set<string>;
  onLinkAddon: (addonId: string) => void;
  onRemoveLinkedAddon?: (addonId: string) => void;
}

interface DerivedRow {
  addonId: string;
  resourceNames: string[];
}

export default function AddonsInStackPanel({
  linkedAddonIds,
  onLinkAddon,
  onRemoveLinkedAddon,
}: AddonsInStackPanelProps) {
  const { addons } = usePostgresAddons();
  // Each entry is a unique slot id for a not-yet-picked addon row.
  const [pendingSlots, setPendingSlots] = useState<string[]>([]);

  // Env vars no longer carry addon-backed sources, so addon links come solely
  // from the explicit `linkedAddonIds` set.
  const allRows: DerivedRow[] = useMemo(
    () => Array.from(linkedAddonIds).map((id) => ({ addonId: id, resourceNames: [] })),
    [linkedAddonIds],
  );

  const availablePostgres = addons.filter(
    (a) => a.id && !linkedAddonIds.has(a.id),
  );

  const findAddonName = (id: string) => addons.find((a) => a.id === id)?.name ?? id;

  const addPendingRow = () => {
    setPendingSlots((prev) => [...prev, `slot-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`]);
  };

  const removePendingRow = (slot: string) => {
    setPendingSlots((prev) => prev.filter((s) => s !== slot));
  };

  const handlePick = (slot: string, addonId: string) => {
    onLinkAddon(addonId);
    removePendingRow(slot);
  };

  const noAddonsInSystem = addons.length === 0;
  const totalRows = allRows.length + pendingSlots.length;

  return (
    <Panel
      title="Stack Addons"
      count={totalRows}
      bodyClassName="p-0"
    >
      {totalRows === 0 ? (
        <div className="flex flex-col items-center justify-center text-center py-14 px-6">
          <Puzzle className="h-6 w-6 mb-4 text-muted-foreground/70" />
          <h3 className="text-sm font-semibold text-foreground">
            No addons attached
            <span className="ml-1.5 font-normal text-muted-foreground">(optional)</span>
          </h3>
          <p className="text-[12.5px] text-muted-foreground mt-2 max-w-sm leading-relaxed">
            {noAddonsInSystem
              ? "You haven't created any addons yet. Create one in the Addons page, then come back to attach it here."
              : "Attach a managed Postgres, Redis, or other addon and bind its credentials into your services."}
          </p>
          {noAddonsInSystem ? (
            <Button
              type="button"
              variant="outline"
              className="mt-5"
              onClick={() => window.open("/addons", "_blank")}
            >
              <PlusCircle className="h-4 w-4" />
              Create Addon
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              className="mt-5"
              onClick={addPendingRow}
            >
              <PlusCircle className="h-4 w-4" />
              Add Addon
            </Button>
          )}
        </div>
      ) : null}
      <div className="divide-y divide-border">
        {allRows.map((row) => {
          const isLinkedOnly = row.resourceNames.length === 0;
          const name = findAddonName(row.addonId);
          return (
            <div
              key={row.addonId}
              className="flex items-center gap-3 px-5 py-3"
            >
              <div className="flex flex-col min-w-0">
                <span className="text-sm font-medium text-foreground truncate flex items-center gap-2">
                  <AddonTypeIcon type="postgres" size={16} />
                  {name}
                  <a
                    href={`/addons/postgres/${row.addonId}/edit`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label={`Open ${name} in new tab`}
                  >
                    <ExternalLink className="h-3.5 w-3.5" />
                  </a>
                </span>
                {isLinkedOnly ? (
                  <span className="text-[11.5px] italic text-muted-foreground">
                    Not yet linked to any resource
                  </span>
                ) : (
                  <span className="text-[11.5px] text-muted-foreground truncate">
                    Linked with: {row.resourceNames.join(", ")}
                  </span>
                )}
              </div>
              <div className="grow" />
              {isLinkedOnly && onRemoveLinkedAddon && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 hover:bg-danger-bg hover:text-danger"
                  onClick={() => onRemoveLinkedAddon(row.addonId)}
                  aria-label={`Unlink ${name}`}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          );
        })}
        {pendingSlots.map((slot) => (
          <div key={slot} className="flex items-center gap-3 px-5 py-3">
            <Select onValueChange={(v) => handlePick(slot, v)}>
              <SelectTrigger className="h-8 w-[280px] text-[13px]">
                <SelectValue
                  placeholder={
                    availablePostgres.length === 0
                      ? "No unlinked Postgres addons available"
                      : "Select an addon"
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {availablePostgres.map((a) => (
                  <SelectItem key={a.id} value={a.id!}>
                    <span className="flex items-center gap-2">
                      <AddonTypeIcon type="postgres" size={14} />
                      {a.name}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div className="grow" />
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 hover:bg-danger-bg hover:text-danger"
              onClick={() => removePendingRow(slot)}
              aria-label="Remove pending addon row"
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          </div>
        ))}
      </div>
      {totalRows > 0 && (
        <div className="flex justify-center mt-4">
          {(() => {
            const allLinked = !noAddonsInSystem && availablePostgres.length === 0;
            const button = (
              <Button
                type="button"
                variant="ghost"
                disabled={allLinked}
                onClick={noAddonsInSystem ? () => window.open("/addons", "_blank") : addPendingRow}
              >
                <PlusCircle className="h-4 w-4" />
                {noAddonsInSystem ? "Create Addon" : "Add Addon"}
              </Button>
            );
            if (!allLinked) return button;
            return (
              <TooltipProvider>
                <Tooltip delayDuration={200}>
                  <TooltipTrigger asChild>
                    <span className="inline-block">{button}</span>
                  </TooltipTrigger>
                  <TooltipContent>All addons linked</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            );
          })()}
        </div>
      )}
    </Panel>
  );
}
