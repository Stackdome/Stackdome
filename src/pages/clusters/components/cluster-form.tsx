import type { Cluster } from "../types";
import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface ClusterFormProps {
  initial?: Partial<Cluster>;
  onSubmit: (values: Partial<Cluster>) => void;
  onCancel: () => void;
  loading?: boolean;
}

export function ClusterForm({ initial = {}, onSubmit, onCancel, loading }: ClusterFormProps) {
  const [name, setName] = useState(initial.name || "");
  // Add more fields as needed

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSubmit({ ...initial, name });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <div>
        <label className="block text-sm font-medium mb-1">Name</label>
        <Input value={name} onChange={e => setName(e.target.value)} required />
      </div>
      {/* Add more fields here */}
      <div className="flex gap-2 justify-end">
        <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
        <Button type="submit" disabled={loading}>
          {loading ? "Saving..." : "Save"}
        </Button>
      </div>
    </form>
  );
}
