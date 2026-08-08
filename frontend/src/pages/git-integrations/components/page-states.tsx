import { GitBranch, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";

export function IntegrationsErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <EmptyState
      icon={<GitBranch className="h-8 w-8" />}
      title="Couldn't load integrations"
      description={message}
      action={
        <Button variant="outline" onClick={onRetry}>
          Retry
        </Button>
      }
    />
  );
}

export function IntegrationsEmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <EmptyState
      icon={<GitBranch className="h-8 w-8" />}
      title="No git integrations yet"
      description="Connect a provider so Stackdome can clone your repositories and trigger preview environments on every push."
      action={
        <Button onClick={onAdd}>
          <Plus className="h-4 w-4" />
          Connect provider
        </Button>
      }
    />
  );
}
