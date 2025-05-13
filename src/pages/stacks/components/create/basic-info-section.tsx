import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { StackFormState } from "./types";

interface Props {
  value: StackFormState;
  onChange: (values: Partial<StackFormState>) => void;
  loading?: boolean;
  error?: string | null;
}

export default function BasicInfoSection({ value, onChange, loading, error }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <Label htmlFor="name">Stack Name</Label>
        <Input
          id="name"
          name="name"
          value={value.name}
          onChange={e => onChange({ name: e.target.value })}
          placeholder="my-awesome-app"
          className="mt-1"
          disabled={loading}
        />
      </div>
      <div>
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          name="description"
          value={value.description}
          onChange={e => onChange({ description: e.target.value })}
          rows={3}
          placeholder="Describe your stack (optional)"
          className="mt-1"
          disabled={loading}
        />
      </div>
      {error && <div className="text-destructive text-sm">{error}</div>}
    </div>
  );
}
