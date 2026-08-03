import React, { useMemo, useState } from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Plus, Plug, X, Upload, FileText, Copy } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { envRowsDiff } from "@/pages/stacks/lib/stack-diff";

import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import { EnvRow, type EnvFrom, type EnvRowErrors, type AddonBindingPatch } from "./env-row";
import { AddonTypeIcon } from "@/components/branded/addon-type-icon";
import type { PostgresAddon } from "@/api/addons";

interface StackResourceEnvironmentTabProps {
  index: number;
  envVars: FormEnvVarData[];
  baselineEnvVars: FormEnvVarData[] | undefined;
  errors: { [field: string]: string | undefined };
  /** Sibling resources (excluding this one), each with their declared outputs. */
  resourceOptions: { name: string; outputs: string[] }[];
  /** This resource's own declared outputs, for the Self source picker. */
  selfOutputs: string[];
  secrets: import("@/api/secrets").Secret[];
  secretsLoading: boolean;
  addons: PostgresAddon[];
  addonNameById: Map<string, string>;
  /** Replace the entire environment_variables array. Identity must be stable. */
  onChangeEnvVars: (next: FormEnvVarData[]) => void;
  /** Reset a single env row to its baseline value. */
  onDiscardEnvRow?: (envIdx: number) => void;
}

/** Full-width dashed "add row" affordance shared by the variable list and each addon group. */
function AddRowButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center justify-center gap-2 rounded-md border border-dashed border-border-strong px-3 py-2 text-[12.5px] font-medium text-foreground/80 transition-colors hover:bg-muted/30"
    >
      <Plus className="size-3.5" />
      {label}
    </button>
  );
}

function StackResourceEnvironmentTabImpl({
  index,
  envVars,
  baselineEnvVars,
  errors,
  resourceOptions,
  selfOutputs,
  secrets,
  secretsLoading,
  addons,
  addonNameById,
  onChangeEnvVars,
  onDiscardEnvRow,
}: StackResourceEnvironmentTabProps) {
  // Per-row diff status for env vars, used to tint modified rows + render reset arrow.
  const envRowStatuses = useMemo(() => {
    if (!baselineEnvVars) return [] as ReturnType<typeof envRowsDiff>;
    return envRowsDiff(
      envVars as unknown as Array<Record<string, unknown>>,
      baselineEnvVars as unknown as Array<Record<string, unknown>>,
    );
  }, [envVars, baselineEnvVars]);

  const [dirtyEnvRows, setDirtyEnvRows] = useState<Set<number>>(new Set());
  const markEnvRowDirty = (envIdx: number) => {
    setDirtyEnvRows((prev) => {
      if (prev.has(envIdx)) return prev;
      const next = new Set(prev);
      next.add(envIdx);
      return next;
    });
  };

  // Helper: insert one literal env var.
  const addEnvVar = (next: FormEnvVarData = { from: "stack", name: "", value: "" }) => {
    onChangeEnvVars([...(envVars || []), next]);
  };

  // Helper: replace a single env-var row entirely.
  const replaceEnvVar = (envIdx: number, next: FormEnvVarData) => {
    onChangeEnvVars((envVars || []).map((env, i) => (i === envIdx ? next : env)));
  };

  const removeEnvVar = (envIdx: number) => {
    onChangeEnvVars((envVars || []).filter((_, i) => i !== envIdx));
  };

  const addMultipleEnvVars = (incoming: Array<{ name: string; value: string }>) => {
    const filtered = incoming.filter((env) => env.name.trim() !== "");
    const existing = new Set((envVars || []).map((e) => e.name));
    const newVars: FormEnvVarData[] = filtered
      .filter((env) => !existing.has(env.name))
      .map((env) => ({ from: "stack" as const, name: env.name, value: env.value }));
    if (newVars.length === 0) {
      toast({
        title: "No new variables added",
        description: "All variables already exist or are invalid.",
        variant: "destructive",
      });
      return;
    }
    onChangeEnvVars([...(envVars || []), ...newVars]);
    toast({
      title: "Environment variables added",
      description: `Added ${newVars.length} new environment variables.`,
      variant: "success",
    });
  };

  // Produce a fresh env row for a chosen source, preserving the row's name.
  // Covers all five FormEnvVarData arms.
  const switchRowFrom = (envIdx: number, from: EnvFrom) => {
    const current = envVars?.[envIdx];
    if (!current) return;
    const name = current.name;
    if (from === "stack") {
      replaceEnvVar(envIdx, { from: "stack", name, value: "" });
    } else if (from === "secret") {
      replaceEnvVar(envIdx, { from: "secret", name, secretId: "", secretKey: "" });
    } else if (from === "addon") {
      replaceEnvVar(envIdx, {
        from: "addon",
        name,
        addonId: "",
        database: undefined,
        superuser: false,
        credField: undefined,
      });
    } else if (from === "resource") {
      replaceEnvVar(envIdx, { from: "resource", name, resourceName: "", output: "" });
    } else if (from === "self") {
      replaceEnvVar(envIdx, { from: "self", name, selfOutput: "" });
    }
  };

  // Per-row addon patch (credField / addonId / database / superuser).
  const onChangeAddonForRow = (envIdx: number, patch: AddonBindingPatch) => {
    const current = envVars?.[envIdx];
    if (!current) return;
    const base =
      current.from === "addon"
        ? current
        : { from: "addon" as const, name: current.name, addonId: "", superuser: false };
    const nextDatabase =
      patch.database === null
        ? undefined
        : patch.database === undefined
          ? base.database
          : patch.database;
    replaceEnvVar(envIdx, {
      ...base,
      addonId: patch.addonId ?? base.addonId,
      database: nextDatabase,
      superuser: patch.superuser ?? base.superuser,
      credField: patch.credField ?? base.credField,
    });
    markEnvRowDirty(envIdx);
  };

  const parseEnvContent = (content: string): Array<{ name: string; value: string }> =>
    content
      .split("\n")
      .filter((line) => line.trim() && !line.trim().startsWith("#"))
      .map((line) => {
        const [name, ...valueParts] = line.split("=");
        return { name: name.trim(), value: valueParts.join("=").trim() };
      });

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (!event.target.files || event.target.files.length === 0) return;
    const file = event.target.files[0];
    const reader = new FileReader();
    reader.onload = (e) => {
      if (!e.target?.result) return;
      const content = e.target.result.toString();
      addMultipleEnvVars(parseEnvContent(content));
    };
    reader.readAsText(file);
    event.target.value = "";
  };

  // Design ghost chip: mono 11px bordered pill, hover swings to brand.
  const ghostChip =
    "inline-flex items-center gap-1.5 rounded-md border border-border px-2 py-1 font-mono text-[11px] text-muted-foreground transition-colors hover:border-brand hover:text-brand disabled:pointer-events-none disabled:opacity-50";

  return (
    <TabsContent value="environment" className="pt-4">
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <span className="text-[12.5px] text-muted-foreground">
          Environment{" "}
            <span className="font-mono text-muted-foreground/70">
            · {(envVars || []).length} variables
            </span>
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              className="inline-flex items-center gap-1.5 rounded-md border border-border px-2 py-1 font-mono text-[11px] text-muted-foreground transition-colors hover:border-danger hover:text-danger disabled:pointer-events-none disabled:opacity-50"
              onClick={() => {
                if (envVars?.length) {
                  onChangeEnvVars([]);
                  toast({
                    title: "Environment variables cleared",
                    description: "All environment variables have been removed.",
                    variant: "success",
                  });
                }
              }}
              disabled={!envVars?.length}
            >
              <X className="size-3.5" />
            clear all
            </button>
            {/* Paste .env button */}
            <Dialog>
              <DialogTrigger asChild>
                <button type="button" className={ghostChip}>
                  <Copy className="size-3.5" />
                paste .env
                </button>
              </DialogTrigger>
              <DialogContent className="w-[95vw] max-w-4xl p-0 overflow-auto">
                <div className="p-6">
                  <DialogHeader>
                    <DialogTitle className="text-lg font-medium">
                    Paste Environment Variables
                    </DialogTitle>
                  </DialogHeader>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor={`env-paste-${index}`} className="text-sm font-medium">
                      Paste in KEY=VALUE format (one per line)
                      </Label>
                      <div className="relative">
                        <Textarea
                          id={`env-paste-${index}`}
                          placeholder={
                            "DATABASE_URL=postgres://user:pass@localhost:5432/db\n" +
                          "API_KEY=your_api_key\n" +
                          "# NODE_ENV=development"
                          }
                          className="font-mono text-sm min-h-[180px] w-full"
                        />
                      </div>
                      <p className="text-xs text-muted-foreground">
                      Lines starting with # will be ignored as comments
                      </p>
                    </div>
                    <Button
                      onClick={() => {
                        const textarea = document.getElementById(
                          `env-paste-${index}`,
                        ) as HTMLTextAreaElement | null;
                        if (textarea) {
                          const content = textarea.value.trim();
                          if (content) {
                            addMultipleEnvVars(parseEnvContent(content));
                          }
                        }
                      }}
                    >
                    Add Variables
                    </Button>
                  </div>
                </div>
              </DialogContent>
            </Dialog>
            {/* Import from file button */}
            <Dialog>
              <DialogTrigger asChild>
                <button type="button" className={ghostChip}>
                  <Upload className="size-3.5" />
                import file
                </button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>Import Environment Variables</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor={`env-file-upload-${index}`} className="text-sm font-medium">
                    Upload .env File
                    </Label>
                    <div className="flex items-center justify-center w-full">
                      <label
                        htmlFor={`env-file-upload-${index}`}
                        className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer bg-muted/20 hover:bg-muted/30"
                      >
                        <div className="flex flex-col items-center justify-center pt-5 pb-6">
                          <FileText className="w-8 h-8 mb-2 text-muted-foreground" />
                          <p className="mb-2 text-sm text-muted-foreground">Click to upload or drag and drop</p>
                          <p className="text-xs text-muted-foreground">Supports .env files</p>
                        </div>
                        <input
                          id={`env-file-upload-${index}`}
                          type="file"
                          accept=".env,text/plain"
                          className="hidden"
                          onChange={handleFileUpload}
                        />
                      </label>
                    </div>
                  </div>
                </div>
              </DialogContent>
            </Dialog>
          </div>
        </div>
        {/* Environment Variables Rows */}
        {(() => {
          // Live duplicate detection (always on, regardless of dirty state)
          const nameCounts = new Map<string, number>();
          envVars.forEach((r) => {
            const k = r.name?.trim();
            if (!k) return;
            nameCounts.set(k, (nameCounts.get(k) ?? 0) + 1);
          });
          const rowErrorsForIndex = (envIdx: number): EnvRowErrors | undefined => {
            const r = envVars[envIdx];
            if (!r) return undefined;
            const out: EnvRowErrors = {};
            if (r.name && (nameCounts.get(r.name.trim()) ?? 0) > 1) {
              out.duplicate = `Duplicate name "${r.name}"`;
            }
            const dirty = dirtyEnvRows.has(envIdx);
            const errPath = (field: string) =>
              errors[`execution_config.environment_variables.${envIdx}.${field}`];
            if (r.from === "addon") {
              if ((dirty || errPath("addonId")) && !r.addonId) out.addonId = "Pick an addon";
              if ((dirty || errPath("database")) && !r.superuser && !r.database)
                out.database = "Pick a database";
              if ((dirty || errPath("credField")) && !r.credField) out.credField = "Pick a field";
            }
            if ((dirty || errPath("name")) && !r.name) out.name = "Required";
            // Server-reported value error (e.g. a backend validation reason pinned
            // to this env var by the autosave engine); always surfaced when present.
            const valueErr = errPath("value");
            if (valueErr) out.value = valueErr;
            return Object.keys(out).length === 0 ? undefined : out;
          };

          // "Add variable" appends a literal row; its own "From" selector then
          // switches the source. Shared by the empty state and the populated list.
          const addVariableButton = (
            <AddRowButton
              label="Add variable"
              onClick={() => addEnvVar({ from: "stack", name: "", value: "" })}
            />
          );

          if (envVars.length === 0) {
            return addVariableButton;
          }

          // Partition rows so all ungrouped (non-addon) rows render FIRST (in
          // their original order), then addon groups render LAST:
          //  - plain list => every non-addon row, keeping its true envIdx.
          //  - addon groups => addon rows sharing (addonId, database), in
          //    first-appearance order, each rendered inside a dashed-border
          //    wrapper with a group-level addon + database picker and an
          //    "Add binding" button.
          type AddonGroup = {
            addonId: string;
            database: string;
            items: { env: FormEnvVarData; envIdx: number }[];
          };
          const plainItems: { env: FormEnvVarData; envIdx: number }[] = [];
          const addonGroups: AddonGroup[] = [];
          const addonGroupByKey = new Map<string, AddonGroup>();
          envVars.forEach((env, envIdx) => {
            if (env.from === "addon") {
              const aid = env.addonId || "";
              const db = env.database || "";
              const key = `${aid}|${db}`;
              let g = addonGroupByKey.get(key);
              if (!g) {
                g = { addonId: aid, database: db, items: [] };
                addonGroupByKey.set(key, g);
                addonGroups.push(g);
              }
              g.items.push({ env, envIdx });
            } else {
              plainItems.push({ env, envIdx });
            }
          });

          const renderRow = ({ env, envIdx }: { env: FormEnvVarData; envIdx: number }) => (
            <EnvRow
              key={envIdx}
              row={env}
              index={envIdx}
              resourceIndex={index}
              secrets={secrets}
              secretsLoading={secretsLoading}
              addonNameById={addonNameById}
              resourceOptions={resourceOptions}
              selfOutputs={selfOutputs}
              rowErrors={rowErrorsForIndex(envIdx)}
              status={envRowStatuses[envIdx] ?? "unchanged"}
              onReset={onDiscardEnvRow ? () => onDiscardEnvRow(envIdx) : undefined}
              onChangeName={(name) => {
                replaceEnvVar(envIdx, { ...env, name });
              }}
              onChangeValue={(value) => {
                replaceEnvVar(envIdx, { from: "stack", name: env.name, value });
              }}
              onChangeFrom={(from) => {
                switchRowFrom(envIdx, from);
                markEnvRowDirty(envIdx);
              }}
              onChangeSecret={(secretId, secretKey) =>
                replaceEnvVar(envIdx, { from: "secret", name: env.name, secretId, secretKey })}
              onChangeAddon={(patch) => onChangeAddonForRow(envIdx, patch)}
              onChangeResource={(resourceName, output) =>
                replaceEnvVar(envIdx, { from: "resource", name: env.name, resourceName, output })}
              onChangeSelf={(selfOutput) =>
                replaceEnvVar(envIdx, { from: "self", name: env.name, selfOutput })}
              onBlur={() => markEnvRowDirty(envIdx)}
              onRemove={() => removeEnvVar(envIdx)}
            />
          );

          return (
            <>
              {plainItems.length > 0 && (
                <div className="flex flex-col gap-1.5">{plainItems.map(renderRow)}</div>
              )}
              {/* Sits directly below the plain rows so a new var appears at the click point. */}
              {addVariableButton}
              {addonGroups.map((g, gIdx) => {
                const aid = g.addonId;
                const db = g.database;
                const selectedAddon = addons.find((a) => a.id === aid);
                const databases = ((selectedAddon?.spec as unknown as { databases?: { name?: string }[] })
                  ?.databases ?? []) as { name?: string }[];
                const name = aid ? (addonNameById?.get(aid) ?? aid) : null;

                const updateAllInGroup = (patch: { addonId?: string; database?: string | undefined }) => {
                  const next = (envVars || []).map((e, i) => {
                    if (g.items.some((it) => it.envIdx === i) && e.from === "addon") {
                      return { ...e, ...patch };
                    }
                    return e;
                  });
                  onChangeEnvVars(next);
                };

                // Disallow picking an (addon, db) combo that already has its own group.
                const usedDbsByAddon = new Map<string, Set<string>>();
                const usedAddonsByDb = new Map<string, Set<string>>();
                for (const og of addonGroups) {
                  if (og === g) continue;
                  if (og.addonId) {
                    if (!usedDbsByAddon.has(og.addonId)) usedDbsByAddon.set(og.addonId, new Set());
                    if (og.database) usedDbsByAddon.get(og.addonId)!.add(og.database);
                  }
                  if (og.database) {
                    if (!usedAddonsByDb.has(og.database)) usedAddonsByDb.set(og.database, new Set());
                    if (og.addonId) usedAddonsByDb.get(og.database)!.add(og.addonId);
                  }
                }
                const dbBlocked = (dbName: string) =>
                  aid !== "" && (usedDbsByAddon.get(aid)?.has(dbName) ?? false);
                const addonBlocked = (addonId: string) =>
                  db !== "" && (usedAddonsByDb.get(db)?.has(addonId) ?? false);

                const handleAddonChange = (newAid: string) => {
                  const a = addons.find((x) => x.id === newAid);
                  const dbs = ((a?.spec as unknown as { databases?: { name?: string }[] })?.databases ??
                []) as { name?: string }[];
                  const usedDbs = usedDbsByAddon.get(newAid) ?? new Set();
                  const firstFreeDb = dbs.find((d) => d.name && !usedDbs.has(d.name))?.name;
                  const newDb = dbs.length === 1 ? dbs[0].name : firstFreeDb;
                  updateAllInGroup({ addonId: newAid, database: newDb });
                };
                const handleDbChange = (newDb: string) => {
                  updateAllInGroup({ database: newDb });
                };
                const handleAddBinding = () => {
                  addEnvVar({
                    from: "addon",
                    name: "",
                    addonId: aid,
                    database: db || undefined,
                    superuser: false,
                    credField: undefined,
                  });
                };

                return (
                  <div
                    key={`a-${gIdx}-${aid}-${db}`}
                    className="relative mt-2 rounded-lg border border-border-strong bg-muted/20 px-2.5 pb-2.5 pt-6"
                    data-testid="env-addon-group"
                  >
                    <div className="absolute -top-3.5 left-3 inline-flex items-center gap-2 rounded-md bg-background px-1.5 py-0.5">
                      <span className="inline-flex items-center gap-1.5 font-mono text-[9.5px] font-semibold uppercase tracking-[0.15em] text-foreground/80">
                        <Plug className="size-3" />
                        Addon
                      </span>
                      <Select value={aid || undefined} onValueChange={handleAddonChange}>
                        <SelectTrigger
                          size="sm"
                          className="w-[180px] gap-2 text-xs"
                          data-testid="addon-picker-trigger"
                        >
                          <span className="flex items-center gap-2 min-w-0">
                            {aid && <AddonTypeIcon type="postgres" size={14} />}
                            <SelectValue placeholder="Pick addon">
                              {aid ? name : undefined}
                            </SelectValue>
                          </span>
                        </SelectTrigger>
                        <SelectContent>
                          {addons.length === 0 ? (
                            <div className="px-3 py-2 text-xs text-muted-foreground">
                              No add-ons on this stack. Add one from the canvas with Add resource.
                            </div>
                          ) : (
                            addons.map((a) => (
                              <SelectItem
                                key={a.id}
                                value={a.id!}
                                disabled={a.id !== aid && addonBlocked(a.id!)}
                              >
                                <span className="flex items-center gap-2">
                                  <AddonTypeIcon type="postgres" size={14} />
                                  <span>{a.name}</span>
                                  {a.id !== aid && addonBlocked(a.id!) && (
                                    <span className="ml-1 text-[10px] text-muted-foreground">
                                  in use
                                    </span>
                                  )}
                                </span>
                              </SelectItem>
                            ))
                          )}
                        </SelectContent>
                      </Select>
                      {aid && (
                        <>
                          <span className="text-muted-foreground/60">·</span>
                          {databases.length > 1 ? (
                            <Select value={db || undefined} onValueChange={handleDbChange}>
                              <SelectTrigger
                                size="sm"
                                className="w-[140px] text-xs"
                                data-testid="database-picker-trigger"
                              >
                                <SelectValue placeholder="Pick database" />
                              </SelectTrigger>
                              <SelectContent>
                                {databases.map((d) =>
                                  d.name ? (
                                    <SelectItem
                                      key={d.name}
                                      value={d.name}
                                      disabled={d.name !== db && dbBlocked(d.name)}
                                    >
                                      {d.name}
                                      {d.name !== db && dbBlocked(d.name) && (
                                        <span className="ml-2 text-[10px] text-muted-foreground">
                                      in use
                                        </span>
                                      )}
                                    </SelectItem>
                                  ) : null,
                                )}
                              </SelectContent>
                            </Select>
                          ) : (
                            <span className="font-mono text-[11px] text-muted-foreground">
                          db: {db || databases[0]?.name || "—"}
                            </span>
                          )}
                        </>
                      )}
                    </div>
                    <div className="flex flex-col gap-1.5">{g.items.map(renderRow)}</div>
                    <div className="pt-1.5">
                      <AddRowButton label="Add binding" onClick={handleAddBinding} />
                    </div>
                  </div>
                );
              })}
            </>
          );
        })()}
      </div>
    </TabsContent>
  );
}

export const StackResourceEnvironmentTab = React.memo(StackResourceEnvironmentTabImpl);
