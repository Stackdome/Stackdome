import { Package, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";

export function RegistriesErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <EmptyState
      icon={<Package className="h-8 w-8" />}
      title="Couldn't load registries"
      description={message}
      action={
        <Button variant="outline" onClick={onRetry}>
          Retry
        </Button>
      }
    />
  );
}

export function RegistriesEmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <EmptyState
      icon={<Package className="h-8 w-8" />}
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
