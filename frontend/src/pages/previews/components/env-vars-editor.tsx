import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface EnvVarFormRow {
  name: string;
  value: string;
}

interface EnvVarsEditorProps {
  value: EnvVarFormRow[];
  onChange: (rows: EnvVarFormRow[]) => void;
}

/** Standalone add/remove env-var list for preview configs. Unlike the stack
 *  resource's EnvRow, there is no addon/secret/resource source picker here —
 *  every value is a literal string (secrets are referenced via
 *  {{ secret.NAME }} text, not resolved in the UI). */
export function EnvVarsEditor({ value, onChange }: EnvVarsEditorProps) {
  const addRow = () => onChange([...value, { name: "", value: "" }]);

  const updateRow = (index: number, patch: Partial<EnvVarFormRow>) => {
    onChange(value.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  };

  const removeRow = (index: number) => onChange(value.filter((_, i) => i !== index));

  return (
    <div className="space-y-2">
      {value.map((row, index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            value={row.name}
            onChange={(e) => updateRow(index, { name: e.target.value })}
            placeholder="NAME"
            aria-label="Variable name"
            className="w-[45%] font-mono text-xs"
          />
          <Input
            value={row.value}
            onChange={(e) => updateRow(index, { value: e.target.value })}
            placeholder="value"
            aria-label="Variable value"
            className="flex-1 font-mono text-xs"
          />
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
      ))}
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
