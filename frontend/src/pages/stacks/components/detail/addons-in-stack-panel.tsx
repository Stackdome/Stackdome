import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { ExternalLink, PlusCircle, X } from "lucide-react";
import { Panel } from "@/components/branded";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import { AddonTypePickerDialog, type AddonType } from "@/pages/addons/components/addon-type-picker-dialog";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

interface AddonsInStackPanelProps {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: Set<string>;
  onLinkAddon: (addonId: string) => void;
  onRemoveLinkedAddon?: (addonId: string) => void;
}

interface DerivedRow {
  addonId: string;
  keyCount: number;
  resourceCount: number;
}

export default function AddonsInStackPanel({
  resources,
  linkedAddonIds,
  onLinkAddon,
  onRemoveLinkedAddon,
}: AddonsInStackPanelProps) {
  const { addons } = usePostgresAddons();
  const [typePickerOpen, setTypePickerOpen] = useState(false);
  const [pgPickerOpen, setPgPickerOpen] = useState(false);
  const [pendingAddonId, setPendingAddonId] = useState<string>("");

  const derived = useMemo<DerivedRow[]>(() => {
    const keyCount = new Map<string, number>();
    const resourceSet = new Map<string, Set<number>>();
    resources.forEach((r, idx) => {
      const envs = (r.execution_config?.environment_variables || []) as FormEnvVarData[];
      for (const e of envs) {
        if (e.from === "addon" && e.addonId) {
          keyCount.set(e.addonId, (keyCount.get(e.addonId) ?? 0) + 1);
          if (!resourceSet.has(e.addonId)) resourceSet.set(e.addonId, new Set());
          resourceSet.get(e.addonId)!.add(idx);
        }
      }
    });
    return Array.from(keyCount.entries()).map(([addonId, count]) => ({
      addonId,
      keyCount: count,
      resourceCount: resourceSet.get(addonId)?.size ?? 0,
    }));
  }, [resources]);

  const derivedIds = useMemo(() => new Set(derived.map((d) => d.addonId)), [derived]);

  const linkedOnly = useMemo(
    () => Array.from(linkedAddonIds).filter((id) => !derivedIds.has(id)),
    [linkedAddonIds, derivedIds],
  );

  const allRows: Array<DerivedRow | { addonId: string; keyCount: 0; resourceCount: 0 }> = [
    ...derived,
    ...linkedOnly.map((id) => ({ addonId: id, keyCount: 0 as const, resourceCount: 0 as const })),
  ];

  if (allRows.length === 0) return null;

  const availablePostgres = addons.filter(
    (a) => a.id && !derivedIds.has(a.id) && !linkedAddonIds.has(a.id),
  );

  const handleTypeSelect = (type: AddonType) => {
    setTypePickerOpen(false);
    if (type === "postgres") {
      setPendingAddonId("");
      setPgPickerOpen(true);
    }
  };

  const confirmLink = () => {
    if (pendingAddonId) {
      onLinkAddon(pendingAddonId);
    }
    setPgPickerOpen(false);
    setPendingAddonId("");
  };

  const findAddonName = (id: string) => addons.find((a) => a.id === id)?.name ?? id;

  return (
    <Panel
      title="Addons in this stack"
      count={allRows.length}
      bodyClassName="p-0"
    >
      <div className="divide-y divide-border">
        {allRows.map((row) => {
          const isLinkedOnly = row.keyCount === 0;
          const name = findAddonName(row.addonId);
          return (
            <div
              key={row.addonId}
              className="flex items-center gap-3 px-5 py-3"
            >
              <AddonTypeIcon type="postgres" size={18} />
              <div className="flex flex-col min-w-0">
                <span className="text-sm font-medium text-foreground truncate">{name}</span>
                {isLinkedOnly ? (
                  <span className="text-[11.5px] italic text-muted-foreground">
                    Not yet linked to any resource
                  </span>
                ) : (
                  <span className="text-[11.5px] text-muted-foreground">
                    {row.keyCount} {row.keyCount === 1 ? "key" : "keys"} across {row.resourceCount}{" "}
                    {row.resourceCount === 1 ? "resource" : "resources"}
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
              <Link
                to={`/addons/postgres/${row.addonId}`}
                className="text-[12.5px] text-brand hover:text-brand-press inline-flex items-center gap-1"
              >
                View detail <ExternalLink className="h-3 w-3" />
              </Link>
            </div>
          );
        })}
      </div>
      <div className="px-5 py-3 border-t border-border">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-[12.5px] text-brand hover:text-brand-press"
          onClick={() => setTypePickerOpen(true)}
        >
          <PlusCircle className="h-3.5 w-3.5" />
          Link addon
        </Button>
      </div>
      <AddonTypePickerDialog
        open={typePickerOpen}
        onOpenChange={setTypePickerOpen}
        onSelect={handleTypeSelect}
      />
      <Dialog open={pgPickerOpen} onOpenChange={setPgPickerOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Link a Postgres addon</DialogTitle>
            <DialogDescription>
              Pick an existing Postgres addon to make available to this stack's resources.
            </DialogDescription>
          </DialogHeader>
          <Select value={pendingAddonId} onValueChange={setPendingAddonId}>
            <SelectTrigger>
              <SelectValue
                placeholder={
                  availablePostgres.length === 0
                    ? "No unlinked Postgres addons available"
                    : "Select a Postgres addon"
                }
              />
            </SelectTrigger>
            <SelectContent>
              {availablePostgres.map((a) => (
                <SelectItem key={a.id} value={a.id!}>
                  {a.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setPgPickerOpen(false)}>
              Cancel
            </Button>
            <Button onClick={confirmLink} disabled={!pendingAddonId}>
              Link addon
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Panel>
  );
}
