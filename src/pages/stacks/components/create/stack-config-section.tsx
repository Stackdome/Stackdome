import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import type { StackFormState } from "./types";

interface Props {
  value: StackFormState;
  onChange: (values: Partial<StackFormState>) => void;
  loading?: boolean;
  error?: string | null;
}

export default function StackConfigSection({ value, onChange, loading, error }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <Label htmlFor="yamlConfig">Stack YAML Config</Label>
        <Textarea
          id="yamlConfig"
          name="yamlConfig"
          value={value.yamlConfig}
          onChange={e => onChange({ yamlConfig: e.target.value })}
          rows={8}
          placeholder="Paste your stack YAML config here"
          className="mt-1 font-mono"
          disabled={loading}
        />
      </div>
      {error && <div className="text-destructive text-sm">{error}</div>}
    </div>
  );
}
