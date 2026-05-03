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
    !!row.addonId &&
    !addonNameById.has(row.addonId);

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
            className={`w-full text-sm font-mono ${isOrphanAddon ? "opacity-60" : ""} ${
              rowErrors?.duplicate || rowErrors?.name ? "border-destructive" : ""
            }`}
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
          {row.from === "addon" &&
          (isOrphanAddon ? (
            <AddonOrphanReadOnly
              database={row.database}
              credField={row.credField}
              superuser={row.superuser}
            />
          ) : (
            <AddonInlinePickers
              row={row}
              addons={addons}
              onChangeAddon={onChangeAddon}
              rowErrors={rowErrors}
            />
          ))}
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
      {rowErrors?.duplicate && (
        <p className="col-span-full text-xs text-destructive mt-0.5 mb-1 px-3">
          {rowErrors.duplicate}
        </p>
      )}
      {rowErrors?.name && (
        <p className="col-span-full text-xs text-destructive mt-0.5 mb-1 px-3">
          {rowErrors.name}
        </p>
      )}
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

const ALL_DATABASES_VALUE = "__ALL_DATABASES__";

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
    <div className="text-xs italic px-3 py-2 text-yellow-600">
      ⚙ &lt;missing addon&gt; · {dbLabel} · {credField ?? "—"}
    </div>
  );
}

function AddonInlinePickers({
  row,
  addons,
  onChangeAddon,
  rowErrors,
}: {
  row: Extract<FormEnvVarData, { from: "addon" }>;
  addons: PostgresAddon[];
  onChangeAddon: (patch: AddonBindingPatch) => void;
  rowErrors?: EnvRowErrors;
}) {
  const selectedAddon = addons.find((a) => a.id === row.addonId);
  const databases = ((selectedAddon?.spec as unknown as { databases?: { name?: string }[] })
    ?.databases ?? []) as { name?: string }[];
  const supportsSuperuser =
    (selectedAddon?.spec as unknown as {
      configuration?: { enable_superuser_access?: boolean };
    })?.configuration?.enable_superuser_access === true;

  const handleAddonChange = (addonId: string) => {
    const a = addons.find((x) => x.id === addonId);
    const dbs = ((a?.spec as unknown as { databases?: { name?: string }[] })?.databases ?? []) as {
      name?: string;
    }[];
    const aSupportsSU =
      (a?.spec as unknown as {
        configuration?: { enable_superuser_access?: boolean };
      })?.configuration?.enable_superuser_access === true;
    if (dbs.length === 1 && !aSupportsSU && dbs[0]?.name) {
      onChangeAddon({ addonId, database: dbs[0].name, superuser: false });
    } else {
      onChangeAddon({ addonId, database: null, superuser: false });
    }
  };

  const handleDatabaseChange = (value: string) => {
    if (value === ALL_DATABASES_VALUE) {
      onChangeAddon({ database: null, superuser: true });
    } else {
      onChangeAddon({ database: value, superuser: false });
    }
  };

  return (
    <div>
      <div className="flex gap-2">
        <Select
          value={row.addonId || undefined}
          onValueChange={handleAddonChange}
        >
          <SelectTrigger
            className={`w-[160px] ${rowErrors?.addonId ? "border-destructive" : ""}`}
            data-testid="addon-picker-trigger"
          >
            <SelectValue placeholder="Addon" />
          </SelectTrigger>
          <SelectContent>
            {addons.length === 0 ? (
              <div className="px-3 py-3 text-sm">
                <p className="text-muted-foreground mb-2">No Postgres addons yet.</p>
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
              addons.map((a) => (
                <SelectItem key={a.id} value={a.id!}>
                  {a.name} (Postgres · {a.status?.state ?? "Unknown"})
                </SelectItem>
              ))
            )}
          </SelectContent>
        </Select>
        <Select
          value={row.superuser ? ALL_DATABASES_VALUE : row.database || undefined}
          onValueChange={handleDatabaseChange}
          disabled={!row.addonId}
        >
          <SelectTrigger
            className={`w-[140px] ${rowErrors?.database ? "border-destructive" : ""}`}
            data-testid="database-picker-trigger"
          >
            <SelectValue placeholder={row.addonId ? "Database" : "Pick an addon first"} />
          </SelectTrigger>
          <SelectContent>
            {supportsSuperuser && (
              <SelectItem value={ALL_DATABASES_VALUE}>─ All databases ─</SelectItem>
            )}
            {databases.map((d) =>
              d.name ? (
                <SelectItem key={d.name} value={d.name}>
                  {d.name}
                </SelectItem>
              ) : null,
            )}
          </SelectContent>
        </Select>
        <Select
          value={row.credField || undefined}
          onValueChange={(v) => onChangeAddon({ credField: v as CredField })}
          disabled={!row.addonId}
        >
          <SelectTrigger
            className={`w-[140px] ${rowErrors?.credField ? "border-destructive" : ""}`}
            data-testid="field-picker-trigger"
          >
            <SelectValue placeholder={row.addonId ? "Field" : "Pick an addon first"} />
          </SelectTrigger>
          <SelectContent>
            {CRED_FIELDS.map((f) => (
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
      </div>
      {(rowErrors?.addonId || rowErrors?.database || rowErrors?.credField) && (
        <div className="mt-1 space-y-0.5">
          {rowErrors?.addonId && <p className="text-xs text-destructive">{rowErrors.addonId}</p>}
          {rowErrors?.database && <p className="text-xs text-destructive">{rowErrors.database}</p>}
          {rowErrors?.credField && <p className="text-xs text-destructive">{rowErrors.credField}</p>}
        </div>
      )}
    </div>
  );
}
