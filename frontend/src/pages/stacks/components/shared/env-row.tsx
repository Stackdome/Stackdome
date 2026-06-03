import { RotateCcw, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

export type EnvRowErrors = {
  name?: string;
  duplicate?: string;
};

interface EnvRowProps {
  row: FormEnvVarData;
  index: number;
  resourceIndex: number;
  rowErrors?: EnvRowErrors;
  /** Diff status vs baseline. "modified" tints + shows reset; "added" stays neutral; "unchanged" stays neutral. */
  status?: "unchanged" | "modified" | "added";
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onBlur?: () => void;
  onRemove: () => void;
  /** When provided and row is "modified", clicking the reset arrow restores the row to baseline. */
  onReset?: () => void;
}

export function EnvRow({
  row,
  index,
  resourceIndex,
  rowErrors,
  status = "unchanged",
  onChangeName,
  onChangeValue,
  onBlur,
  onRemove,
  onReset,
}: EnvRowProps) {
  const isModified = status === "modified";
  const isAdded = status === "added";
  const isDirty = isModified || isAdded;
  return (
    <div
      className={`border-b last:border-b-0 ${
        isDirty ? "border-l-4 border-l-brand bg-brand-bg" : ""
      }`}
      data-testid={`env-row-${resourceIndex}-${index}`}
      onBlur={onBlur}
    >
      <div className="grid grid-cols-12 gap-2 p-3 items-start">
        {/* Key */}
        <div className="col-span-4">
          <Input
            id={`env-name-${resourceIndex}-${index}`}
            value={row.name || ""}
            onChange={(e) => onChangeName(e.target.value)}
            className={`w-full text-sm font-mono ${
              rowErrors?.duplicate || rowErrors?.name ? "border-danger" : ""
            }`}
            placeholder="KEY"
          />
          {(rowErrors?.duplicate || rowErrors?.name) && (
            <p className="text-xs text-danger mt-1">
              {rowErrors.duplicate || rowErrors.name}
            </p>
          )}
        </div>

        {/* Value */}
        <div className="col-span-7">
          <Input
            value={row.value || ""}
            onChange={(e) => onChangeValue(e.target.value)}
            className="w-full text-sm font-mono"
            placeholder="VALUE"
          />
        </div>

        {/* Reset (when dirty — restore baseline, removes added rows) or Remove (when clean) */}
        <div className="col-span-1 flex justify-center items-start pt-1">
          {isDirty && onReset ? (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-brand hover:bg-brand-bg hover:text-brand-press"
              onClick={onReset}
              aria-label={isAdded ? "Remove this newly added env var" : "Reset env var to original value"}
              title={isAdded ? "Remove (newly added)" : "Reset to original value"}
            >
              <RotateCcw className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 hover:bg-danger-bg hover:text-danger"
              onClick={onRemove}
              aria-label="Remove env var"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
