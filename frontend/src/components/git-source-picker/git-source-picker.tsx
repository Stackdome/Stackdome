import { useCallback, useEffect, useState } from "react";
import { Globe, Lock, Plus, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
  listGitIntegrations,
  searchRepositories,
  getRepository,
  type GitIntegration,
  type GitRepository,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { AddIntegrationWizard } from "@/pages/git-integrations/add-integration-wizard";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
} from "@/pages/git-integrations/lib/derive-row";
import { CredentialsDropdown, usableIntegrations } from "./credentials-dropdown";
import type { PickedRepo } from "./types";

/** "https://github.com/acme/api(.git)" → "acme/api" */
export function repoTail(url: string): string {
  const trimmed = url.replace(/\.git$/, "").replace(/\/+$/, "");
  return trimmed.split("/").slice(-2).join("/");
}

function hostOf(url: string): string | null {
  try {
    return new URL(url).host.toLowerCase();
  } catch {
    return null;
  }
}

type Tab = "provider" | "url";

interface GitSourcePickerProps {
  value: PickedRepo | null;
  onChange: (repo: PickedRepo | null) => void;
  /** Shown under the Public URL tab (e.g. preview wizard's PR-automation note). */
  publicUrlHint?: string;
  /** When set, empty search results link here ("Configure in GitHub"). */
  configureUrl?: string;
}

export function GitSourcePicker({ value, onChange, publicUrlHint, configureUrl }: GitSourcePickerProps) {
  const [tab, setTab] = useState<Tab>("provider");
  const [integrations, setIntegrations] = useState<GitIntegration[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [connectOpen, setConnectOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [repos, setRepos] = useState<GitRepository[]>([]);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hostUrl, setHostUrl] = useState("");
  const [publicUrl, setPublicUrl] = useState("");

  const selected = integrations.find((i) => i.id === selectedId) ?? null;
  const hasGithubApp = integrations.some((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);

  const loadIntegrations = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    try {
      const list = usableIntegrations((await listGitIntegrations(orgId)).items ?? []);
      setIntegrations(list);
      setSelectedId((current) => {
        if (current && list.some((i) => i.id === current)) return current;
        const app = list.find((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);
        return app?.id ?? list[0]?.id ?? null;
      });
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    void loadIntegrations();
  }, [loadIntegrations]);

  // Debounced repo search — only meaningful for GitHub App integrations.
  useEffect(() => {
    if (tab !== "provider" || selected?.type !== GIT_INTEGRATION_TYPE_GITHUB_APP) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId || !selected.id) return;
    let cancelled = false;
    setSearching(true);
    const t = setTimeout(() => {
      searchRepositories(orgId, selected.id!, { query: query || undefined })
        .then((page) => {
          if (cancelled) return;
          setRepos(page.items ?? []);
          setError(null);
        })
        .catch((e) => {
          if (!cancelled) setError(getErrorMessage(e));
        })
        .finally(() => {
          if (!cancelled) setSearching(false);
        });
    }, 300);
    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [tab, selected?.id, selected?.type, query]);

  const pickRepo = async (repo: GitRepository) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !selected?.id || !repo.full_name) return;
    const [owner, name] = repo.full_name.split("/");
    try {
      const detail = await getRepository(orgId, selected.id, owner, name);
      onChange({
        fullName: detail.full_name ?? repo.full_name,
        cloneUrl: detail.clone_url ?? repo.clone_url ?? "",
        defaultBranch: detail.default_branch ?? "",
        integrationId: selected.id,
      });
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  const selectIntegration = (integration: GitIntegration) => {
    setSelectedId(integration.id ?? null);
    setHostUrl(integration.type === GIT_INTEGRATION_TYPE_CREDENTIALS ? `https://${integration.host}/` : "");
    setError(null);
    onChange(null);
  };

  const switchTab = (next: Tab) => {
    setTab(next);
    setError(null);
    setPublicUrl("");
    setHostUrl(selected?.type === GIT_INTEGRATION_TYPE_CREDENTIALS ? `https://${selected.host}/` : "");
    onChange(null);
  };

  const emitHostUrl = (url: string, integration: GitIntegration) => {
    setHostUrl(url);
    const trimmed = url.trim();
    let host: string | null = null;
    let path = "";
    try {
      const parsed = new URL(trimmed);
      host = parsed.host.toLowerCase();
      path = parsed.pathname.replace(/^\/+|\/+$/g, "");
    } catch {
      host = null;
    }
    if (host === (integration.host ?? "").toLowerCase() && path.includes("/")) {
      onChange({
        fullName: repoTail(trimmed),
        cloneUrl: trimmed,
        defaultBranch: "",
        integrationId: integration.id ?? null,
      });
    } else {
      onChange(null);
    }
  };

  const emitPublicUrl = (url: string) => {
    setPublicUrl(url);
    const trimmed = url.trim();
    if (trimmed) {
      onChange({ fullName: repoTail(trimmed), cloneUrl: trimmed, defaultBranch: "", integrationId: null });
    } else {
      onChange(null);
    }
  };

  const hostMismatch =
    selected?.type === GIT_INTEGRATION_TYPE_CREDENTIALS &&
    hostUrl.trim() !== "" &&
    hostOf(hostUrl.trim()) !== (selected.host ?? "").toLowerCase();

  return (
    <div className="space-y-4">
      <div role="tablist" className="flex gap-6 border-b border-border">
        {(
          [
            { key: "provider", label: "Connected provider" },
            { key: "url", label: "Public URL" },
          ] as { key: Tab; label: string }[]
        ).map((t) => (
          <button
            key={t.key}
            type="button"
            role="tab"
            aria-selected={tab === t.key}
            onClick={() => switchTab(t.key)}
            className={cn(
              "-mb-px border-b-2 px-0.5 py-2 font-mono text-[11px] uppercase tracking-[1.5px] transition-colors",
              tab === t.key
                ? "border-brand text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "provider" && loaded && integrations.length === 0 && (
        <div className="flex flex-col items-center gap-3 rounded-md border border-dashed py-10 text-center">
          <p className="text-sm text-muted-foreground">
            Connect a git provider to pick from your repositories.
          </p>
          <Button onClick={() => setConnectOpen(true)}>
            <Plus className="h-4 w-4" />
            Connect provider
          </Button>
        </div>
      )}

      {tab === "provider" && integrations.length > 0 && (
        <>
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                className="pl-8"
                placeholder="Search repositories…"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                disabled={selected?.type !== GIT_INTEGRATION_TYPE_GITHUB_APP}
              />
            </div>
            <CredentialsDropdown
              integrations={integrations}
              selectedId={selectedId}
              onSelect={selectIntegration}
              onConnectNew={() => setConnectOpen(true)}
            />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          {selected?.type === GIT_INTEGRATION_TYPE_GITHUB_APP && (
            <>
              {searching && <p className="text-sm text-muted-foreground">Searching…</p>}
              <ul className="divide-y rounded-md border">
                {repos.map((r) => (
                  <li key={r.full_name}>
                    <button
                      type="button"
                      className={cn(
                        "flex w-full items-center gap-2 px-3 py-2.5 text-left text-sm hover:bg-accent",
                        value?.cloneUrl && value.cloneUrl === (r.clone_url ?? "") && "bg-brand-bg",
                      )}
                      onClick={() => void pickRepo(r)}
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
                {!searching && repos.length === 0 && (
                  <li className="space-y-2 px-3 py-6 text-center text-sm text-muted-foreground">
                    <p>No repositories found.</p>
                    {configureUrl && (
                      <a
                        href={configureUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="text-brand hover:underline"
                      >
                        Don&apos;t see your repository? Configure in GitHub →
                      </a>
                    )}
                  </li>
                )}
              </ul>
            </>
          )}

          {selected?.type === GIT_INTEGRATION_TYPE_CREDENTIALS && (
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Repository listing isn&apos;t available for token connections — paste the
                repository URL on {selected.host}. Clones use the stored credentials.
              </p>
              <Input
                placeholder={`https://${selected.host}/group/project`}
                value={hostUrl}
                onChange={(e) => emitHostUrl(e.target.value, selected)}
              />
              {hostMismatch && (
                <p className="text-sm text-destructive">
                  URL must be on {selected.host} to use this connection.
                </p>
              )}
            </div>
          )}
        </>
      )}

      {tab === "url" && (
        <div className="space-y-2">
          <div className="relative">
            <Globe className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              className="pl-8"
              placeholder="https://github.com/acme/webapp"
              value={publicUrl}
              onChange={(e) => emitPublicUrl(e.target.value)}
            />
          </div>
          <p className="text-xs text-muted-foreground">
            Public repositories work without a connection; private ones need a matching
            git integration.
          </p>
          {publicUrlHint && <p className="text-xs text-warn">{publicUrlHint}</p>}
        </div>
      )}

      <AddIntegrationWizard
        open={connectOpen}
        onOpenChange={setConnectOpen}
        hasGithubApp={hasGithubApp}
        onCreated={() => void loadIntegrations()}
      />
    </div>
  );
}
