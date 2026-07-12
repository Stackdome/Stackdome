import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ChevronRight, GitBranch, GitPullRequest, Loader2, PlusCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, EmptyState, Panel } from "@/components/branded";
import { ProviderLogo } from "@/pages/git-integrations/components/provider-logo";
import { providerIdForHost } from "@/pages/git-integrations/lib/derive-row";
import { listAllPreviewConfigs, type StackPreviewConfig } from "@/api/preview-configs";
import { usePreviewEnvs } from "@/pages/previews/hooks/use-preview-envs";
import { EnableRepoWizard } from "@/pages/previews/components/enable-repo-wizard/enable-repo-wizard";
import { repoTail } from "@/components/git-source-picker/git-source-picker";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useCurrentUser } from "@/hooks/use-current-user";
import { useResourceTeams } from "@/hooks/use-resource-teams";

/** "https://github.com/acme/webapp.git" → "github.com" (empty on unparsable input) */
function hostOf(url?: string): string {
  if (!url) return "";
  try {
    return new URL(url).host;
  } catch {
    return "";
  }
}

export default function PreviewsPage() {
  const navigate = useNavigate();
  const { canWriteAnyTeam } = useCurrentUser();
  const { defaultTeamName } = useResourceTeams();
  const [configs, setConfigs] = useState<StackPreviewConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);
  const { envs, refresh: refreshEnvs } = usePreviewEnvs();

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName) return;
    try {
      setConfigs(await listAllPreviewConfigs(orgId, defaultTeamName));
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

  const envCount = (configId?: string) =>
    envs.filter((e) => e.config_id && e.config_id === configId).length;

  return (
    <div className="flex flex-1 flex-col p-8 space-y-6 h-full">
      <PageHeader
        eyebrow="Platform"
        title="Previews"
        subtitle="Repositories configured for preview environments — every pull request can get its own temporary stack."
        actions={
          canWriteAnyTeam && configs.length > 0 ? (
            <Button onClick={() => setWizardOpen(true)}>
              <PlusCircle className="h-4 w-4" />
              Enable repository
            </Button>
          ) : undefined
        }
      />

      {loading ? (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : error && configs.length === 0 ? (
        <EmptyState
          icon={<GitPullRequest className="h-8 w-8" />}
          title="Couldn't load preview configurations"
          description={error}
          action={<Button onClick={() => void refresh()}>Retry</Button>}
        />
      ) : configs.length === 0 ? (
        <EmptyState
          icon={<GitPullRequest className="h-8 w-8" />}
          title="Preview every pull request"
          description="Enable a repository — each pull request can get its own temporary environment with a shareable URL."
          action={
            canWriteAnyTeam ? (
              <Button onClick={() => setWizardOpen(true)}>
                <PlusCircle className="h-4 w-4" />
                Enable repository
              </Button>
            ) : undefined
          }
        />
      ) : (
        <Panel title="Configured repositories" count={configs.length}>
          <div className="divide-y divide-border">
            {configs.map((c) => {
              const count = envCount(c.id);
              const host = hostOf(c.git_repository?.repo_url);
              return (
                <button
                  key={c.id}
                  type="button"
                  onClick={() => c.id && navigate(`/previews/${c.id}`)}
                  className="flex w-full items-center gap-4 px-4 py-3 text-left hover:bg-muted/50"
                >
                  <div className="flex min-w-0 flex-1 items-center gap-3">
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card">
                      <ProviderLogo providerId={providerIdForHost(host)} className="h-5 w-5 shrink-0" />
                    </div>
                    <div className="min-w-0">
                      <p className="truncate text-[15px] font-medium text-foreground">{c.name}</p>
                      <p className="truncate font-mono text-[11.5px] text-fg-muted">
                        {c.git_repository?.repo_url ? repoTail(c.git_repository.repo_url) : ""}
                      </p>
                    </div>
                  </div>
                  <span className="inline-flex items-center gap-1.5 rounded-full border border-border px-2 py-0.5 text-xs text-muted-foreground">
                    <GitBranch className="h-3 w-3" />
                    {c.git_repository?.base_branch}
                  </span>
                  <span className="w-[130px] text-right font-mono text-[11px] text-muted-foreground tabular-nums">
                    {count} environment{count === 1 ? "" : "s"}
                  </span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
              );
            })}
          </div>
        </Panel>
      )}

      <EnableRepoWizard
        open={wizardOpen}
        onOpenChange={setWizardOpen}
        onCreated={() => {
          void refresh();
          void refreshEnvs();
        }}
      />
    </div>
  );
}
