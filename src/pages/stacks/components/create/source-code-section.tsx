import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { StackFormState } from "./types";

interface Props {
  value: StackFormState;
  onChange: (values: Partial<StackFormState>) => void;
  loading?: boolean;
  error?: string | null;
}

export default function SourceCodeSection({ value, onChange, loading, error }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <Label htmlFor="repositoryUrl">Source Repository</Label>
        <Input
          id="repositoryUrl"
          name="repositoryUrl"
          value={value.repositoryUrl}
          onChange={e => onChange({ repositoryUrl: e.target.value })}
          placeholder="https://github.com/your/repo"
          className="mt-1"
          disabled={loading}
        />
      </div>
      {error && <div className="text-destructive text-sm">{error}</div>}
    </div>
  );
}
