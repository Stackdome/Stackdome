import { Container, Plus, RefreshCw, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";

export function RegistriesErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border border-danger-border bg-danger/[0.07] px-6 py-10 text-center">
      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-danger-bg text-danger">
        <TriangleAlert className="h-5 w-5" />
      </span>
      <h3 className="text-sm font-semibold text-foreground">Couldn&apos;t load registries</h3>
      <p className="font-mono text-[11.5px] text-fg-muted">{message}</p>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCw className="h-3.5 w-3.5" />
        Retry
      </Button>
    </div>
  );
}

export function RegistriesEmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <EmptyState
      icon={<Container className="h-8 w-8" />}
      title="No image registries yet"
      description="Add registry credentials so builds can pull private base images and push build artifacts."
      action={
        <Button onClick={onAdd}>
          <Plus className="h-4 w-4" />
          Add registry
        </Button>
      }
    />
  );
}
