import type { StackPreviewConfig } from "@/api/preview-configs";
import { usePreviewEnvs } from "../hooks/use-preview-envs";
import { PreviewEnvRow } from "./preview-env-row";

interface ConfigEnvsSectionProps {
  config: StackPreviewConfig;
}

export function ConfigEnvsSection({ config }: ConfigEnvsSectionProps) {
  const { envs, loading, error } = usePreviewEnvs(config.id);

  if (loading) return <p className="text-sm text-muted-foreground">Loading environments…</p>;
  if (error) return <p className="text-sm text-destructive">{error}</p>;
  if (envs.length === 0) {
    return (
      <p className="rounded-md border border-dashed px-3 py-4 text-sm text-muted-foreground">
        No preview environments yet.
      </p>
    );
  }
  return (
    <div className="space-y-2">
      {envs.map((env) => (
        <PreviewEnvRow key={env.id} env={env} onSync={() => {}} onDelete={() => {}} />
      ))}
    </div>
  );
}
