import { GitBranch, Plus, RefreshCw, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";

export function IntegrationsErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border border-danger-border bg-danger/[0.07] px-6 py-10 text-center">
      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-danger-bg text-danger">
        <TriangleAlert className="h-5 w-5" />
      </span>
      <h3 className="text-sm font-semibold text-foreground">Couldn&apos;t load integrations</h3>
      <p className="max-w-md text-xs text-muted-foreground">
        We couldn&apos;t reach the server. Your integrations are safe — this is a display issue.
      </p>
      <p className="font-mono text-[11.5px] text-fg-muted">Error: {message}</p>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="h-3.5 w-3.5" />
        Retry
      </Button>
    </div>
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
