import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSecrets } from "@/pages/secrets/hooks/use-secrets";

export interface EnvVarFormRow {
  name: string;
  value: string;
}

interface EnvVarsEditorProps {
  value: EnvVarFormRow[];
  onChange: (rows: EnvVarFormRow[]) => void;
}

/** A value that is exactly one secret reference, e.g. "{{ secret.api-token }}".
 *  The name may be empty — that's a Secret-mode row whose secret isn't picked yet. */
const EXACT_SECRET_REF = /^\{\{\s*secret\.([\w-]*)\s*\}\}$/;

const secretRef = (name: string) => `{{ secret.${name} }}`;

/** Add/remove env-var list for preview configs. Each row's value is either a
 *  literal string or a reference to a saved org secret — the Secret source is
 *  sugar that writes the {{ secret.NAME }} form the backend already resolves
 *  into a secret connection at deploy; hand-typed refs still round-trip. */
export function EnvVarsEditor({ value, onChange }: EnvVarsEditorProps) {
  const { secrets } = useSecrets();
  const genericSecrets = secrets.filter((s) => s.type === "Generic");

  const addRow = () => onChange([...value, { name: "", value: "" }]);

  const updateRow = (index: number, patch: Partial<EnvVarFormRow>) => {
    onChange(value.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  };

  const removeRow = (index: number) => onChange(value.filter((_, i) => i !== index));

  return (
    <div className="space-y-2">
      {value.map((row, index) => {
        const refMatch = EXACT_SECRET_REF.exec(row.value);
        const fromSecret = refMatch != null;
        return (
          <div key={index} className="flex items-center gap-2">
            <Input
              value={row.name}
              onChange={(e) => updateRow(index, { name: e.target.value })}
              placeholder="NAME"
              aria-label="Variable name"
              className="w-[36%] font-mono text-xs"
            />
            {fromSecret ? (
              <Select
                value={refMatch[1] || ""}
                onValueChange={(name) => updateRow(index, { value: secretRef(name) })}
              >
                <SelectTrigger aria-label="Secret" className="flex-1 text-xs">
                  <SelectValue placeholder="Select secret…" />
                </SelectTrigger>
                <SelectContent>
                  {genericSecrets.length === 0 && refMatch[1] === "" && (
                    <div className="px-2 py-1.5 text-xs text-muted-foreground">
                      No secrets yet. Create one on the Secrets page.
                    </div>
                  )}
                  {genericSecrets.map((s) => (
                    <SelectItem key={s.id} value={s.name}>
                      {s.name}
                    </SelectItem>
                  ))}
                  {/* A hand-typed ref to a secret that no longer exists stays selectable
                      so the row renders instead of blanking. */}
                  {refMatch[1] !== "" && !genericSecrets.some((s) => s.name === refMatch[1]) && (
                    <SelectItem value={refMatch[1]}>{refMatch[1]}</SelectItem>
                  )}
                </SelectContent>
              </Select>
            ) : (
              <Input
                value={row.value}
                onChange={(e) => updateRow(index, { value: e.target.value })}
                placeholder="value"
                aria-label="Variable value"
                className="flex-1 font-mono text-xs"
              />
            )}
            <Select
              value={fromSecret ? "secret" : "plain"}
              onValueChange={(source) => updateRow(index, { value: source === "secret" ? secretRef("") : "" })}
            >
              <SelectTrigger aria-label="Value source" className="w-[110px] flex-none text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="plain">Plain text</SelectItem>
                <SelectItem value="secret">Secret</SelectItem>
              </SelectContent>
            </Select>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-8 w-8 flex-none hover:bg-danger-bg hover:text-danger"
              onClick={() => removeRow(index)}
              aria-label="Remove variable"
            >
              <X className="size-3.5" />
            </Button>
          </div>
        );
      })}
      <button
        type="button"
        onClick={addRow}
        className="flex w-full items-center justify-center gap-2 rounded-md border border-dashed border-border-strong px-3 py-2 text-[12.5px] font-medium text-foreground/80 transition-colors hover:bg-muted/30"
      >
        <Plus className="size-3.5" />
        Add variable
      </button>
    </div>
  );
}
