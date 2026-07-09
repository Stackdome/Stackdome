import { RotateCcw, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { Secret } from "@/api/secrets";
import {
  ADDON_OUTPUT_FIELDS,
  CLUSTER_WIDE_FIELDS,
  type AddonOutputField,
} from "@/pages/stacks/lib/addon-presets";

export type EnvFrom = FormEnvVarData["from"];

export type AddonBindingPatch = {
  addonId?: string;
  database?: string | null; // null = explicitly cleared (All databases)
  superuser?: boolean;
  credField?: AddonOutputField;
};

export type EnvRowErrors = {
  name?: string;
  value?: string;
  addonId?: string;
  database?: string;
  credField?: string;
  duplicate?: string;
};

interface EnvRowProps {
  row: FormEnvVarData;
  index: number;
  resourceIndex: number;
  secrets: Secret[];
  secretsLoading: boolean;
  addonNameById?: Map<string, string>;
  /** Sibling resources (excluding this one) for the Resource source picker. */
  resourceOptions?: { name: string; outputs: string[] }[];
  /** This resource's own declared outputs for the Self source picker. */
  selfOutputs?: string[];
  rowErrors?: EnvRowErrors;
  /** Diff status vs baseline. "modified" tints + shows reset; "added" stays neutral; "unchanged" stays neutral. */
  status?: "unchanged" | "modified" | "added";
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onChangeFrom: (from: EnvFrom) => void;
  onChangeSecret: (secretId: string, secretKey: string) => void;
  onChangeAddon: (patch: AddonBindingPatch) => void;
  onChangeResource?: (resourceName: string, output: string) => void;
  onChangeSelf?: (selfOutput: string) => void;
  onBlur?: () => void;
  onRemove: () => void;
  /** When provided and row is "modified", clicking the reset arrow restores the row to baseline. */
  onReset?: () => void;
}

export function EnvRow({
  row,
  index,
  resourceIndex,
  secrets,
  secretsLoading,
  addonNameById,
  resourceOptions,
  selfOutputs,
  rowErrors,
  status = "unchanged",
  onChangeName,
  onChangeValue,
  onChangeFrom,
  onChangeSecret,
  onChangeAddon,
  onChangeResource,
  onChangeSelf,
  onBlur,
  onRemove,
  onReset,
}: EnvRowProps) {
  const isOrphanAddon =
    row.from === "addon" &&
    addonNameById !== undefined &&
    !!row.addonId &&
    !addonNameById.has(row.addonId);

  const isModified = status === "modified";
  const isAdded = status === "added";
  const isDirty = isModified || isAdded;
  return (
    <div
      className={`relative rounded-md border py-2 pl-[13px] pr-2.5 transition-colors ${
        isOrphanAddon
          ? "border-warn-border bg-warn-bg"
          : isDirty
            ? "border-brand-border bg-brand-bg"
            : "border-border bg-background"
      }`}
      data-testid={`env-row-${resourceIndex}-${index}`}
      onBlur={onBlur}
    >
      {/* 2px accent bar signalling row state (dirty = brand, orphan = warn). */}
      <span
        aria-hidden
        className={`absolute bottom-1.5 left-0 top-1.5 w-0.5 rounded-full ${
          isOrphanAddon ? "bg-warn" : isDirty ? "bg-brand" : "bg-transparent"
        }`}
      />
      {/* items-start keeps every control top-aligned so a per-cell error (which
          grows that cell downward) never knocks the other columns out of line. */}
      <div className="flex items-start gap-2">
        {/* Key */}
        <div className="w-[180px] flex-none">
          <Input
            id={`env-name-${resourceIndex}-${index}`}
            value={row.name || ""}
            onChange={(e) => onChangeName(e.target.value)}
            className={`h-8 w-full font-mono text-xs md:text-xs ${isOrphanAddon ? "opacity-60" : ""} ${
              rowErrors?.duplicate || rowErrors?.name ? "border-danger" : ""
            }`}
            placeholder="KEY"
            readOnly={isOrphanAddon}
          />
          {(rowErrors?.duplicate || rowErrors?.name) && (
            <p className="text-xs text-danger mt-1">
              {rowErrors.duplicate || rowErrors.name}
            </p>
          )}
        </div>

        {/* Value */}
        <div className="min-w-0 flex-1">
          {row.from === "stack" && (
            <Input
              value={row.value || ""}
              onChange={(e) => onChangeValue(e.target.value)}
              className={`h-8 w-full font-mono text-xs md:text-xs ${
                rowErrors?.value ? "border-danger" : ""
              }`}
              placeholder="value"
            />
          )}
          {row.from === "secret" && (
            <SecretValueCell
              secrets={secrets}
              loading={secretsLoading}
              secretId={row.secretId}
              secretKey={row.secretKey}
              onChange={onChangeSecret}
            />
          )}
          {row.from === "addon" &&
          (isOrphanAddon ? (
            <AddonOrphanReadOnly
              database={row.database}
              credField={row.credField}
              superuser={row.superuser}
            />
          ) : (
            <AddonCredFieldPicker
              credField={row.credField}
              onChange={(v) => onChangeAddon({ credField: v })}
              error={rowErrors?.credField}
              disabled={!row.addonId}
            />
          ))}
          {row.from === "resource" && (
            <ResourceOutputCell
              resourceOptions={resourceOptions ?? []}
              resourceName={row.resourceName}
              output={row.output}
              onChange={(rn, out) => onChangeResource?.(rn, out)}
            />
          )}
          {row.from === "self" && (
            <SelfOutputCell
              outputs={selfOutputs ?? []}
              selfOutput={row.selfOutput}
              onChange={(out) => onChangeSelf?.(out)}
            />
          )}
          {rowErrors?.value && (
            <p className="text-xs text-danger mt-1">{rowErrors.value}</p>
          )}
        </div>

        {/* From select (Stack | Secret | Addon) */}
        <div className="w-[106px] flex-none">
          <Select
            value={row.from}
            onValueChange={(v) => onChangeFrom(v as EnvFrom)}
          >
            <SelectTrigger size="sm" className="w-full text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="stack">Plain text</SelectItem>
              <SelectItem value="secret">Secret</SelectItem>
              <SelectItem value="addon">Addon</SelectItem>
              <SelectItem value="resource">Resource</SelectItem>
              <SelectItem value="self">Self</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Reset (modified existing row — restore baseline) or Remove (added/clean rows) */}
        <div className="flex h-8 w-6 flex-none items-center justify-center">
          {isModified && onReset ? (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-brand hover:bg-brand-bg hover:text-brand-press"
              onClick={onReset}
              aria-label="Reset env var to original value"
              title="Reset to original value"
            >
              <RotateCcw className="size-3.5" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 hover:bg-danger-bg hover:text-danger"
              onClick={onRemove}
              aria-label="Remove env var"
            >
              <X className="size-3.5" />
            </Button>
          )}
        </div>
      </div>
      {isOrphanAddon && (
        <p className="text-xs text-warn mt-1.5">
          Addon was deleted. This variable won't resolve. Remove to clean up.
        </p>
      )}
    </div>
  );
}

function SecretValueCell({
  secrets,
  loading,
  secretId,
  secretKey,
  onChange,
}: {
  secrets: Secret[];
  loading: boolean;
  secretId: string;
  secretKey: string;
  onChange: (secretId: string, secretKey: string) => void;
}) {
  const genericSecrets = secrets.filter((s) => s.type === "Generic");
  const selected = genericSecrets.find((s) => s.id === secretId);
  const availableKeys = selected?.data?.map((d) => d.key) || [];

  return (
    <div className="flex flex-col gap-1.5">
      <Select
        value={secretId || ""}
        onValueChange={(value) => onChange(value, "")}
        disabled={loading || genericSecrets.length === 0}
      >
        <SelectTrigger size="sm" className="w-full text-xs">
          <SelectValue
            placeholder={
              genericSecrets.length === 0
                ? "No generic secrets available"
                : "select secret..."
            }
          />
        </SelectTrigger>
        <SelectContent>
          {genericSecrets.map((secret) => (
            <SelectItem key={secret.id} value={secret.id!}>
              {secret.name}
              {secret.description && (
                <span className="text-muted-foreground ml-2">
                  - {secret.description}
                </span>
              )}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {secretId && (
        <Select
          value={secretKey || ""}
          onValueChange={(value) => onChange(secretId, value)}
          disabled={availableKeys.length === 0}
        >
          <SelectTrigger size="sm" className="w-full text-xs">
            <SelectValue
              placeholder={
                availableKeys.length === 0
                  ? "No keys available in secret"
                  : "select key..."
              }
            />
          </SelectTrigger>
          <SelectContent>
            {availableKeys.map((key) => (
              <SelectItem key={key} value={key}>
                {key}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}

function ResourceOutputCell({
  resourceOptions,
  resourceName,
  output,
  onChange,
}: {
  resourceOptions: { name: string; outputs: string[] }[];
  resourceName: string;
  output: string;
  onChange: (resourceName: string, output: string) => void;
}) {
  const selected = resourceOptions.find((r) => r.name === resourceName);
  const outputs = selected?.outputs ?? [];
  return (
    <div className="flex flex-col gap-1.5">
      <Select
        value={resourceName || ""}
        onValueChange={(v) => onChange(v, "")}
        disabled={resourceOptions.length === 0}
      >
        <SelectTrigger size="sm" className="w-full text-xs" data-testid="resource-picker-trigger">
          <SelectValue placeholder={resourceOptions.length === 0 ? "No other resources" : "select resource..."} />
        </SelectTrigger>
        <SelectContent>
          {resourceOptions.map((r) => (
            <SelectItem key={r.name} value={r.name}>{r.name}</SelectItem>
          ))}
        </SelectContent>
      </Select>
      {resourceName && (
        <Select value={output || ""} onValueChange={(v) => onChange(resourceName, v)} disabled={outputs.length === 0}>
          <SelectTrigger size="sm" className="w-full text-xs" data-testid="resource-output-trigger">
            <SelectValue placeholder={outputs.length === 0 ? "No outputs" : "select output..."} />
          </SelectTrigger>
          <SelectContent>
            {outputs.map((o) => (<SelectItem key={o} value={o}>{o}</SelectItem>))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}

function SelfOutputCell({
  outputs,
  selfOutput,
  onChange,
}: {
  outputs: string[];
  selfOutput: string;
  onChange: (selfOutput: string) => void;
}) {
  return (
    <Select value={selfOutput || ""} onValueChange={onChange} disabled={outputs.length === 0}>
      <SelectTrigger size="sm" className="w-full text-xs" data-testid="self-output-trigger">
        <SelectValue placeholder={outputs.length === 0 ? "No outputs declared" : "select output..."} />
      </SelectTrigger>
      <SelectContent>
        {outputs.map((o) => (<SelectItem key={o} value={o}>{o}</SelectItem>))}
      </SelectContent>
    </Select>
  );
}

function AddonOrphanReadOnly({
  database,
  credField,
  superuser,
}: {
  database?: string;
  credField?: string;
  superuser: boolean;
}) {
  const dbLabel = superuser ? "(superuser)" : database ?? "—";
  return (
    <div className="text-xs italic py-1.5 text-warn">
      ⚙ &lt;missing addon&gt; · {dbLabel} · {credField ?? "—"}
    </div>
  );
}

function AddonCredFieldPicker({
  credField,
  onChange,
  error,
  disabled,
}: {
  credField?: AddonOutputField;
  onChange: (v: AddonOutputField) => void;
  error?: string;
  disabled?: boolean;
}) {
  return (
    <div>
      <Select
        value={credField || undefined}
        onValueChange={(v) => onChange(v as AddonOutputField)}
        disabled={disabled}
      >
        <SelectTrigger
          size="sm"
          className={`w-full text-xs ${error ? "border-danger" : ""}`}
          data-testid="field-picker-trigger"
        >
          <SelectValue placeholder={disabled ? "Pick an addon first" : "Select field"} />
        </SelectTrigger>
        <SelectContent>
          {ADDON_OUTPUT_FIELDS.map((f) => (
            <SelectItem key={f} value={f}>
              <span className="flex items-center gap-2">
                <span>{f}</span>
                {CLUSTER_WIDE_FIELDS.has(f) && (
                  <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
                    cluster
                  </span>
                )}
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {error && <p className="text-xs text-danger mt-1">{error}</p>}
    </div>
  );
}

