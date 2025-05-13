import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import type { StackFormState } from "./types";

interface Props {
  value: StackFormState;
  onChange: (values: Partial<StackFormState>) => void;
  loading?: boolean;
  error?: string | null;
}

// For demo: just a placeholder for environment variables
export default function EnvironmentSection({ value, onChange, loading, error }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <Label htmlFor="env-demo">Environment Variable (Demo)</Label>
        <Input
          id="env-demo"
          name="env-demo"
          value={value.environment?.demo || ""}
          onChange={e => onChange({ environment: { ...value.environment, demo: e.target.value } })}
          placeholder="MY_ENV_VAR=value"
          className="mt-1"
          disabled={loading}
        />
      </div>
      {error && <div className="text-destructive text-sm">{error}</div>}
    </div>
  );
}
