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

export type EnvFrom = FormEnvVarData["from"];

interface EnvRowProps {
  row: FormEnvVarData;
  index: number;
  resourceIndex: number;
  secrets: Secret[];
  secretsLoading: boolean;
  addonNameById?: Map<string, string>;
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onChangeFrom: (from: EnvFrom) => void;
  onChangeSecret: (secretId: string, secretKey: string) => void;
  onRemove: () => void;
}

export function EnvRow({
  row,
  index,
  resourceIndex,
  secrets,
  secretsLoading,
  addonNameById,
  onChangeName,
  onChangeValue,
  onChangeFrom,
  onChangeSecret,
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
          disabled={row.from === "addon"}
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
          <AddonValueCell
            addonId={row.addonId}
            database={row.database}
            credField={row.credField}
            superuser={row.superuser}
            addonName={addonNameById?.get(row.addonId)}
            isOrphan={addonNameById !== undefined && !addonNameById.has(row.addonId)}
          />
        )}
      </div>

      {/* From select (Stack | Secret | Addon) */}
      <div className="col-span-2 flex justify-center items-start pt-1">
        <Select
          value={row.from}
          onValueChange={(v) => onChangeFrom(v as EnvFrom)}
          disabled={row.from === "addon"}
        >
          <SelectTrigger className="w-[110px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="stack">Stack</SelectItem>
            <SelectItem value="secret">Secret</SelectItem>
            <SelectItem value="addon" disabled>
              Addon
            </SelectItem>
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

function AddonValueCell({
  addonId,
  database,
  credField,
  superuser,
  addonName,
  isOrphan,
}: {
  addonId: string;
  database?: string;
  credField: string;
  superuser: boolean;
  addonName?: string;
  isOrphan?: boolean;
}) {
  const dbLabel = superuser ? "(superuser)" : database ?? "—";
  const label = isOrphan
    ? "<missing addon>"
    : addonName ?? `${addonId.slice(0, 8)}…`;
  return (
    <div
      className={`text-xs italic px-3 py-2 ${
        isOrphan ? "text-yellow-600" : "text-muted-foreground"
      }`}
    >
      ⚙ {label} · {dbLabel} · {credField}
    </div>
  );
}
