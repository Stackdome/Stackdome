import { Button } from "@/components/ui/button";
import type { StackFormState } from "./types";

interface Props {
  value: StackFormState;
  onSubmit: () => void;
  loading?: boolean;
  error?: string | null;
}

export default function ReviewDeploySection({ value, onSubmit, loading, error }: Props) {
  return (
    <div className="space-y-6">
      <div className="p-4 bg-gray-50 rounded-md border border-gray-200 dark:bg-gray-900 dark:border-gray-800">
        <pre className="text-xs whitespace-pre-wrap break-all font-mono">
{JSON.stringify(value, null, 2)}
        </pre>
      </div>
      {error && <div className="text-destructive text-sm">{error}</div>}
      <div className="flex justify-end">
        <Button onClick={onSubmit} disabled={loading} className="bg-black text-white">
          {loading ? "Deploying..." : "Deploy Stack"}
        </Button>
      </div>
    </div>
  );
}
