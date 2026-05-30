import React, { useMemo, useState } from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PlusCircle, X, Upload, FileText, Copy } from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { envRowsDiff } from "@/pages/stacks/lib/stack-diff";

import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import { EnvRow, type EnvRowErrors } from "./env-row";

interface StackResourceEnvironmentTabProps {
  index: number;
  envVars: FormEnvVarData[];
  baselineEnvVars: FormEnvVarData[] | undefined;
  errors: { [field: string]: string | undefined };
  /** Replace the entire environment_variables array. Identity must be stable. */
  onChangeEnvVars: (next: FormEnvVarData[]) => void;
  /** Reset a single env row to its baseline value. */
  onDiscardEnvRow?: (envIdx: number) => void;
}

function StackResourceEnvironmentTabImpl({
  index,
  envVars,
  baselineEnvVars,
  errors,
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
        description: "All variables already exist or are invalid",
        variant: "destructive",
      });
      return;
    }
    onChangeEnvVars([...(envVars || []), ...newVars]);
    toast({
      title: "Environment variables added",
      description: `Added ${newVars.length} new environment variables`,
      variant: "default",
    });
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

  return (
    <TabsContent value="environment" className="pt-4">
      <div className="flex items-center mb-3">
        <h3 className="text-sm font-semibold text-foreground">Environment Variables</h3>
        <div className="ml-auto flex gap-2">
          <Button
            variant="ghost"
            className="text-danger hover:text-danger hover:bg-danger-bg"
            size="sm"
            onClick={() => {
              if (envVars?.length) {
                onChangeEnvVars([]);
                toast({
                  title: "Environment variables cleared",
                  description: "All environment variables have been removed",
                });
              }
            }}
            disabled={!envVars?.length}
          >
            <X className="h-4 w-4 mr-1" />
            <span>Clear All</span>
          </Button>
          {/* Paste Variables button */}
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="sm" className="gap-2">
                <Copy className="h-4 w-4" />
                <span>Paste Variables</span>
              </Button>
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
              <Button variant="ghost" size="sm">
                <Upload className="h-4 w-4 mr-2" /> Import File
              </Button>
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
      <div className="border border-muted rounded-md">
        {/* Header Row */}
        <div className="grid grid-cols-12 gap-2 p-3 border-b bg-muted/30 text-sm font-medium">
          <div className="col-span-4">Key</div>
          <div className="col-span-7">Value</div>
          <div className="col-span-1"></div>
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
            if ((dirty || errPath("name")) && !r.name) out.name = "Required";
            return Object.keys(out).length === 0 ? undefined : out;
          };

          if (envVars.length === 0) {
            return (
              <div className="p-8 text-center text-muted-foreground">
                No environment variables defined
              </div>
            );
          }

          return envVars.map((env, envIdx) => (
            <EnvRow
              key={envIdx}
              row={env}
              index={envIdx}
              resourceIndex={index}
              rowErrors={rowErrorsForIndex(envIdx)}
              status={envRowStatuses[envIdx] ?? "unchanged"}
              onReset={onDiscardEnvRow ? () => onDiscardEnvRow(envIdx) : undefined}
              onChangeName={(name) => {
                replaceEnvVar(envIdx, { ...env, name });
              }}
              onChangeValue={(value) => {
                replaceEnvVar(envIdx, { ...env, value });
              }}
              onBlur={() => markEnvRowDirty(envIdx)}
              onRemove={() => removeEnvVar(envIdx)}
            />
          ));
        })()}
      </div>
      {/* Add Variable button */}
      <div className="mt-2">
        <Button variant="ghost" size="sm" onClick={() => addEnvVar()}>
          <PlusCircle className="h-4 w-4 mr-2" />Add variable
        </Button>
      </div>
    </TabsContent>
  );
}

export const StackResourceEnvironmentTab = React.memo(StackResourceEnvironmentTabImpl);
