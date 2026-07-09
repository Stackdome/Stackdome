import { useEffect, useState } from "react";
import { Lock, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { searchRepositories, getRepository, type GitRepository } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import type { PickedRepo } from "./enable-repo-wizard";

interface RepoPickerPhaseProps {
  integrationId: string | null;
  onPicked: (repo: PickedRepo) => void;
  onBack: () => void;
}

/** "https://github.com/acme/api(.git)" → "acme/api" */
function repoTail(url: string): string {
  const trimmed = url.replace(/\.git$/, "").replace(/\/+$/, "");
  const parts = trimmed.split("/");
  return parts.slice(-2).join("/");
}

export function RepoPickerPhase({ integrationId, onPicked, onBack }: RepoPickerPhaseProps) {
  const [manual, setManual] = useState(integrationId == null);
  const [query, setQuery] = useState("");
  const [repos, setRepos] = useState<GitRepository[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [manualUrl, setManualUrl] = useState("");

  // Debounced repo search whenever discovery is available.
  useEffect(() => {
    if (manual || !integrationId) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    setLoading(true);
    const t = setTimeout(() => {
      searchRepositories(orgId, integrationId, { query: query || undefined })
        .then((page) => {
          setRepos(page.items ?? []);
          setError(null);
        })
        .catch((e) => setError(getErrorMessage(e)))
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(t);
  }, [manual, integrationId, query]);

  const pick = async (repo: GitRepository) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integrationId || !repo.full_name) return;
    const [owner, name] = repo.full_name.split("/");
    try {
      const detail = await getRepository(orgId, integrationId, owner, name);
      onPicked({
        fullName: detail.full_name ?? repo.full_name,
        cloneUrl: detail.clone_url ?? repo.clone_url ?? "",
        defaultBranch: detail.default_branch ?? "",
        integrationId,
      });
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  if (manual) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex-1 space-y-4 overflow-y-auto p-6">
          <div>
            <h3 className="text-sm font-semibold">Repository URL</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Public repositories work without a connection; private ones need a
              matching git integration.
            </p>
          </div>
          <Input
            placeholder="https://github.com/acme/webapp"
            value={manualUrl}
            onChange={(e) => setManualUrl(e.target.value)}
          />
          {integrationId != null && (
            <Button variant="link" size="sm" className="px-0" onClick={() => setManual(false)}>
              Pick from your GitHub repositories instead
            </Button>
          )}
        </div>
        <WizardFooter
          onBack={onBack}
          onContinue={() =>
            onPicked({
              fullName: repoTail(manualUrl),
              cloneUrl: manualUrl.trim(),
              defaultBranch: "",
              integrationId: null,
            })
          }
          continueDisabled={manualUrl.trim() === ""}
        />
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 space-y-3 overflow-y-auto p-6">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            className="pl-8"
            placeholder="Search repositories…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
        {loading && <p className="text-sm text-muted-foreground">Searching…</p>}
        <ul className="divide-y rounded-md border">
          {repos.map((r) => (
            <li key={r.full_name}>
              <button
                type="button"
                className="flex w-full items-center gap-2 px-3 py-2.5 text-left text-sm hover:bg-accent"
                onClick={() => void pick(r)}
              >
                <span className="flex-1 truncate font-mono text-xs">{r.full_name}</span>
                {r.private && (
                  <Badge variant="outline" className="gap-1 text-[10px]">
                    <Lock className="h-3 w-3" />
                    private
                  </Badge>
                )}
              </button>
            </li>
          ))}
          {!loading && repos.length === 0 && (
            <li className="px-3 py-6 text-center text-sm text-muted-foreground">
              No repositories found. The app may not have access — manage
              repository access on GitHub, then search again.
            </li>
          )}
        </ul>
        <Button variant="link" size="sm" className="px-0" onClick={() => setManual(true)}>
          Enter repository URL instead
        </Button>
      </div>
      <WizardFooter onBack={onBack} onContinue={() => {}} continueDisabled hint="Pick a repository to continue" />
    </div>
  );
}
