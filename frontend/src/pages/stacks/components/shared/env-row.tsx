import { X } from "lucide-react";
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
import type { PostgresAddon } from "@/api/addons";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  type CredField,
} from "@/pages/stacks/lib/addon-presets";

export type EnvFrom = FormEnvVarData["from"];

export type AddonBindingPatch = {
  addonId?: string;
  database?: string | null; // null = explicitly cleared (All databases)
  superuser?: boolean;
  credField?: CredField;
};

export type EnvRowErrors = {
  name?: string;
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
  addons: PostgresAddon[];
  addonNameById?: Map<string, string>;
  rowErrors?: EnvRowErrors;
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onChangeFrom: (from: EnvFrom) => void;
  onChangeSecret: (secretId: string, secretKey: string) => void;
  onChangeAddon: (patch: AddonBindingPatch) => void;
  onBlur?: () => void;
  onRemove: () => void;
}

export function EnvRow({
  row,
  index,
  resourceIndex,
  secrets,
  secretsLoading,
  addons,
  addonNameById,
  rowErrors,
  onChangeName,
  onChangeValue,
  onChangeFrom,
  onChangeSecret,
  onChangeAddon,
  onBlur,
  onRemove,
}: EnvRowProps) {
  const isOrphanAddon =
    row.from === "addon" &&
    addonNameById !== undefined &&
    !addonNameById.has((row as any).addonId);

  return (
    <div
      className={`border-b last:border-b-0 ${
        isOrphanAddon ? "border-l-4 border-l-yellow-500/60 pl-2 bg-yellow-500/5" : ""
      }`}
      data-testid={`env-row-${resourceIndex}-${index}`}
      onBlur={onBlur}
    >
      <div className="grid grid-cols-12 gap-2 p-3 items-start">
      {/* Key */}
      <div className="col-span-3">
        <Input
          id={`env-name-${resourceIndex}-${index}`}
          value={row.name || ""}
          onChange={(e) => onChangeName(e.target.value)}
          className={`w-full text-sm font-mono ${isOrphanAddon ? "opacity-60" : ""}`}
          placeholder="KEY"
          readOnly={isOrphanAddon}
        />
      </div>

      {/* Value */}
      <div className="col-span-6">
        {row.from === "stack" && (
          <Input
            value={row.value || ""}
            onChange={(e) => onChangeValue(e.target.value)}
            className="w-full text-sm font-mono"
            placeholder="VALUE"
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
        {row.from === "addon" && (
          <AddonInlinePickers
            row={row}
            addons={addons}
            onChangeAddon={onChangeAddon}
            rowErrors={rowErrors}
          />
        )}
      </div>

      {/* From select (Stack | Secret | Addon) */}
      <div className="col-span-2 flex justify-center items-start pt-1">
        <Select
          value={row.from}
          onValueChange={(v) => onChangeFrom(v as EnvFrom)}
        >
          <SelectTrigger className="w-[110px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="stack">Stack</SelectItem>
            <SelectItem value="secret">Secret</SelectItem>
            <SelectItem value="addon">Addon</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Remove */}
      <div className="col-span-1 flex justify-center items-start pt-1">
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 hover:bg-destructive/10 hover:text-destructive"
          onClick={onRemove}
          aria-label="Remove env var"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
      </div>
      {isOrphanAddon && (
        <p className="col-span-full text-xs text-yellow-700 dark:text-yellow-400 mt-0.5 mb-1 px-3">
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
    <div className="space-y-2">
      <Select
        value={secretId || ""}
        onValueChange={(value) => onChange(value, "")}
        disabled={loading || genericSecrets.length === 0}
      >
        <SelectTrigger className="w-full">
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
          <SelectTrigger className="w-full">
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

function AddonInlinePickers({
  row,
  addons,
  onChangeAddon,
  rowErrors: _rowErrors,
}: {
  row: Extract<FormEnvVarData, { from: "addon" }>;
  addons: PostgresAddon[];
  onChangeAddon: (patch: AddonBindingPatch) => void;
  rowErrors?: EnvRowErrors;
}) {
  void _rowErrors;
  void CRED_FIELDS;
  void CLUSTER_WIDE_FIELDS;
  return (
    <div className="flex gap-2">
      <Select
        value={row.addonId || undefined}
        onValueChange={(v) => onChangeAddon({ addonId: v })}
      >
        <SelectTrigger className="w-[160px]" data-testid="addon-picker-trigger">
          <SelectValue placeholder="Addon" />
        </SelectTrigger>
        <SelectContent>
          {addons.map((a) => (
            <SelectItem key={a.id} value={a.id!}>
              {a.name} (Postgres · {a.status?.state ?? "Unknown"})
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select value={row.database || undefined}>
        <SelectTrigger className="w-[140px]" data-testid="database-picker-trigger">
          <SelectValue placeholder="Database" />
        </SelectTrigger>
        <SelectContent></SelectContent>
      </Select>
      <Select value={row.credField || undefined}>
        <SelectTrigger className="w-[140px]" data-testid="field-picker-trigger">
          <SelectValue placeholder="Field" />
        </SelectTrigger>
        <SelectContent></SelectContent>
      </Select>
    </div>
  );
}
