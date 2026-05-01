import { useEffect, useMemo, useRef, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  DEFAULT_ENV_NAMES,
  applyPreset,
  type CredField,
  type Preset,
} from "@/pages/stacks/lib/addon-presets";
import type { PostgresAddon } from "@/api/addons";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

interface AddFromAddonDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  addons: PostgresAddon[];
  existingEnvNames: Set<string>;
  onAdd: (rows: FormEnvVarData[]) => void;
}

export function AddFromAddonDialog({
  open,
  onOpenChange,
  addons,
  existingEnvNames,
  onAdd,
}: AddFromAddonDialogProps) {
  const [addonId, setAddonId] = useState<string>(addons[0]?.id ?? "");
  const [database, setDatabase] = useState<string>("");
  const [superuser, setSuperuser] = useState(false);
  const [selected, setSelected] = useState<Set<CredField>>(new Set());
  const [envNames, setEnvNames] = useState<Partial<Record<CredField, string>>>({});

  const addon = addons.find((a) => a.id === addonId);
  // The OpenAPI generated type may not surface `databases` directly; cast safely.
  const databases = useMemo(
    () =>
      ((addon?.spec as unknown as { databases?: { name?: string }[] })?.databases ?? []) as {
        name?: string;
      }[],
    [addon],
  );
  const supportsSuperuser =
    (addon?.spec as unknown as { configuration?: { enable_superuser_access?: boolean } })
      ?.configuration?.enable_superuser_access === true;

  // Auto-select first database (covers single-db case and provides a sensible
  // default for multi-db). The user can change it via the dropdown.
  useEffect(() => {
    if (databases.length > 0 && !database) {
      setDatabase(databases[0]?.name ?? "");
    }
  }, [databases, database]);

  // Reset state when addon switches (skip first render so initial addonId
  // doesn't clobber the auto-selected database).
  const prevAddonId = useRef(addonId);
  useEffect(() => {
    if (prevAddonId.current === addonId) return;
    prevAddonId.current = addonId;
    setDatabase("");
    setSuperuser(false);
    setSelected(new Set());
    setEnvNames({});
  }, [addonId]);

  const toggleField = (field: CredField, on: boolean) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (on) next.add(field);
      else next.delete(field);
      return next;
    });
    setEnvNames((prev) => {
      if (on && !prev[field]) {
        return { ...prev, [field]: DEFAULT_ENV_NAMES[field] };
      }
      if (!on) {
        const next = { ...prev };
        delete next[field];
        return next;
      }
      return prev;
    });
  };

  const onPreset = (preset: Preset) => {
    const result = applyPreset(preset);
    setSelected(result.selected);
    setEnvNames(result.envNames);
  };

  const collisions = useMemo(() => {
    const out: string[] = [];
    selected.forEach((f) => {
      const name = envNames[f] ?? "";
      if (name && existingEnvNames.has(name)) out.push(name);
    });
    return out;
  }, [selected, envNames, existingEnvNames]);

  const validationError = useMemo(() => {
    if (!addon) return "Pick an addon.";
    if (selected.size === 0) return "Tick at least one credential field.";
    for (const f of selected) {
      const name = envNames[f] ?? "";
      if (!name.trim()) return `Env name for "${f}" is empty.`;
    }
    if (collisions.length) {
      return `Env name "${collisions[0]}" already exists in this resource.`;
    }
    const names = [...selected].map((f) => envNames[f]);
    if (new Set(names).size !== names.length) {
      return "Two ticked fields share the same env name.";
    }
    if (!superuser && !database) return "Pick a database.";
    return null;
  }, [addon, superuser, database, selected, envNames, collisions]);

  const onConfirm = () => {
    if (validationError) return;
    const rows: FormEnvVarData[] = [...selected].map((credField) => ({
      from: "addon",
      name: envNames[credField] ?? DEFAULT_ENV_NAMES[credField],
      addonType: "postgres",
      addonId,
      database: superuser ? undefined : database,
      superuser,
      credField,
    }));
    try {
      onAdd(rows);
      onOpenChange(false);
    } catch {
      // onAdd failed — keep dialog open so user can review
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>Add from Addon</DialogTitle>
          <DialogDescription>
            Inject credentials from an addon as environment variables.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label className="text-sm">Addon</Label>
            {addons.length === 0 ? (
              <div className="rounded-md border bg-muted/40 px-3 py-3 text-sm">
                <p className="text-muted-foreground mb-2">
                  No addons yet. Create a Postgres addon to inject its credentials here.
                </p>
                <a
                  href="/addons/create/postgres"
                  target="_blank"
                  rel="noreferrer"
                  className="text-primary underline"
                >
                  + Create Postgres addon
                </a>
              </div>
            ) : (
              <Select value={addonId} onValueChange={setAddonId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select an addon" />
                </SelectTrigger>
                <SelectContent>
                  {addons.map((a) => (
                    <SelectItem key={a.id} value={a.id!}>
                      {a.name} (Postgres · {a.status?.state ?? "Unknown"})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {!superuser && (
            <div>
              <Label className="text-sm">Database</Label>
              <Select value={database} onValueChange={setDatabase}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a database" />
                </SelectTrigger>
                <SelectContent>
                  {databases.filter((d) => d.name).map((d) => (
                    <SelectItem key={d.name} value={d.name!}>
                      {d.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground mt-1" data-testid="selected-database">
                {database ? `Selected: ${database}` : ""}
              </p>
            </div>
          )}

          {supportsSuperuser && (
            <div className="flex items-center gap-2">
              <Switch
                id="superuser-toggle"
                checked={superuser}
                onCheckedChange={setSuperuser}
              />
              <Label htmlFor="superuser-toggle" className="text-sm">
                Superuser
              </Label>
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <Label className="text-sm">Inject credentials</Label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="sm">
                    Apply preset
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onClick={() => onPreset("postgres-conventions")}>
                    Postgres conventions
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onPreset("connection-string")}>
                    Connection string only
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onPreset("clear")}>
                    Clear
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div className="border rounded-md divide-y">
              {CRED_FIELDS.map((field) => (
                <div key={field} className="flex items-center gap-2 px-3 py-2">
                  <input
                    type="checkbox"
                    id={`cred-${field}`}
                    checked={selected.has(field)}
                    onChange={(e) => toggleField(field, e.target.checked)}
                  />
                  <Label htmlFor={`cred-${field}`} className="w-44 text-sm font-mono">
                    {field}
                  </Label>
                  <span className="text-muted-foreground">→</span>
                  <Input
                    value={envNames[field] ?? ""}
                    placeholder={DEFAULT_ENV_NAMES[field]}
                    disabled={!selected.has(field)}
                    onChange={(e) =>
                      setEnvNames((p) => ({ ...p, [field]: e.target.value }))
                    }
                    className="flex-1"
                  />
                  {CLUSTER_WIDE_FIELDS.has(field) && (
                    <span className="text-xs text-muted-foreground">cluster</span>
                  )}
                </div>
              ))}
            </div>

            {validationError ? (
              <p className="text-xs text-destructive mt-2">{validationError}</p>
            ) : (
              <p className="text-xs text-muted-foreground mt-2">
                Adds {selected.size} environment variable{selected.size === 1 ? "" : "s"}.
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={validationError !== null}>
            Add
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
