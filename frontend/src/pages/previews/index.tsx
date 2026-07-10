import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { GitPullRequest, Plus, Settings2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { listPreviewConfigs, type StackPreviewConfig } from "@/api/preview-configs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { EnableRepoWizard } from "./components/enable-repo-wizard/enable-repo-wizard";
import { ConfigEnvsSection } from "./components/config-envs-section";

export default function PreviewsPage() {
  const { defaultTeamName } = useResourceTeams();
  const [configs, setConfigs] = useState<StackPreviewConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName) return;
    try {
      const list = await listPreviewConfigs(orgId, defaultTeamName);
      setConfigs(list.items ?? []);
      setError(null);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [defaultTeamName]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Preview Environments</h1>
          <p className="text-sm text-muted-foreground">
            Temporary environments for pull requests on your repositories.
          </p>
        </div>
        {configs.length > 0 && (
          <Button onClick={() => setWizardOpen(true)}>
            <Plus className="h-4 w-4" />
            Enable repository
          </Button>
        )}
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}
      {loading && <p className="text-sm text-muted-foreground">Loading…</p>}

      {!loading && configs.length === 0 && (
        <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed px-6 py-16 text-center">
          <GitPullRequest className="h-8 w-8 text-muted-foreground" />
          <h2 className="text-base font-semibold">Preview every pull request</h2>
          <p className="max-w-md text-sm text-muted-foreground">
            Connect GitHub and enable a repository — each pull request can get
            its own temporary environment with a shareable URL.
          </p>
          <Button onClick={() => setWizardOpen(true)}>
            <Plus className="h-4 w-4" />
            Enable repository
          </Button>
        </div>
      )}

      {configs.map((cfg) => (
        <section key={cfg.id} className="space-y-2">
          <div className="flex items-center gap-2">
            <h2 className="font-mono text-sm font-semibold">{cfg.name}</h2>
            <span className="text-xs text-muted-foreground">
              {cfg.git_repository?.base_branch} · {cfg.stackfile_path}
            </span>
            <span className="flex-1" />
            <Button variant="ghost" size="sm" asChild>
              <Link to={`/previews/${cfg.id}`}>
                <Settings2 className="h-4 w-4" />
                Configure
              </Link>
            </Button>
          </div>
          <ConfigEnvsSection config={cfg} />
        </section>
      ))}

      <EnableRepoWizard
        open={wizardOpen}
        onOpenChange={setWizardOpen}
        onCreated={() => void refresh()}
      />
    </div>
  );
}
