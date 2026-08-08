import { useEffect, useState } from "react";
import { ChevronDown, Link2, Lock, Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
  listGitIntegrations,
  listRepositories,
  getRepository,
  type GitIntegration,
  type GitRepository,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import {
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
} from "@/lib/git-integrations";
import { usableIntegrations } from "./credentials-dropdown";
import { repoTail } from "./git-source-picker";

export interface RepoPick {
  repo_url: string;
  integration_id: string | undefined;
}

interface RepoComboboxProps {
  id: string;
  /** Current source.git.repo_url */
  value: string;
  /** Set when the current URL was picked through an integration. */
  integrationId?: string;
  onChange: (pick: RepoPick) => void;
  hasError?: boolean;
}

/** Compact repository combobox for the stack drawer: searches repos across
    usable git integrations, accepts free text as a repository URL. */
export function RepoCombobox({ id, value, integrationId, onChange, hasError }: RepoComboboxProps) {
  const [open, setOpen] = useState(false);
  const [integrations, setIntegrations] = useState<GitIntegration[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [repos, setRepos] = useState<GitRepository[]>([]);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selected = integrations.find((i) => i.id === selectedId) ?? null;

  // Integrations load lazily on first open; failure degrades to URL-only entry.
  useEffect(() => {
    if (!open || loaded) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setLoaded(true);
      return;
    }
    listGitIntegrations(orgId)
      .then((res) => {
        const list = usableIntegrations(res.items ?? []);
        setIntegrations(list);
        setSelectedId((current) => {
          if (current && list.some((i) => i.id === current)) return current;
          const app = list.find((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);
          return app?.id ?? list[0]?.id ?? null;
        });
      })
      .catch((e) => setError(getErrorMessage(e)))
      .finally(() => setLoaded(true));
  }, [open, loaded]);

  // Repos load once per integration; the query filters them client-side.
  useEffect(() => {
    if (!open || selected?.type !== GIT_INTEGRATION_TYPE_GITHUB_APP) {
      setRepos([]);
      return;
    }
    const orgId = getCurrentOrganizationId();
    if (!orgId || !selected.id) return;
    let cancelled = false;
    setSearching(true);
    listRepositories(orgId, selected.id!)
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
    return () => {
      cancelled = true;
    };
  }, [open, selected?.id, selected?.type]);

  const pickRepo = async (repo: GitRepository) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !selected?.id || !repo.full_name) return;
    const [owner, name] = repo.full_name.split("/");
    try {
      const detail = await getRepository(orgId, selected.id, owner, name);
      onChange({
        repo_url: detail.clone_url ?? repo.clone_url ?? "",
        integration_id: selected.id,
      });
      setOpen(false);
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  // Trimmed once and reused everywhere the query is validated or emitted —
  // stray leading/trailing whitespace from a paste should never reach the
  // form value.
  const trimmedQuery = query.trim();

  const useAsUrl = () => {
    onChange({ repo_url: trimmedQuery, integration_id: undefined });
    setOpen(false);
  };

  // Credentials hosts can't list repos; typed paths compose against the host.
  // Strip any leading/trailing slashes and an existing trailing ".git" from
  // the typed path first, so pasting "acme/api.git" doesn't double up into
  // "acme/api.git.git".
  const useOnHost = () => {
    if (!selected?.host || !selected.id) return;
    const path = trimmedQuery
      .replace(/^\/+/, "")
      .replace(/\/+$/, "")
      .replace(/\.git\/?$/, "");
    onChange({
      repo_url: `https://${selected.host}/${path}.git`,
      integration_id: selected.id,
    });
    setOpen(false);
  };

  const display = value ? (integrationId ? repoTail(value) : value) : null;
  // scp-style refs (user@host:path, e.g. deploy@git.corp.io:acme/api.git)
  // aren't limited to a "git" user, so match the general `user@host:` shape.
  const looksLikeUrl =
    /^(https?|ssh|git):\/\//i.test(trimmedQuery) || /^[\w.-]+@[\w.-]+:/.test(trimmedQuery);
  const needle = trimmedQuery.toLowerCase();
  const filteredRepos = needle
    ? repos.filter((r) => r.full_name?.toLowerCase().includes(needle))
    : repos;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          id={id}
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-invalid={hasError || undefined}
          className={cn(
            "h-9 w-full justify-between font-mono text-meta font-normal",
            !display && "text-muted-foreground",
            hasError && "border-danger",
          )}
        >
          <span className="truncate">{display ?? "Select repository or enter URL"}</span>
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command shouldFilter={false}>
          {integrations.length > 1 && (
            <div className="border-b p-2">
              <Select value={selectedId ?? undefined} onValueChange={setSelectedId}>
                <SelectTrigger className="text-meta" aria-label="Integration">
                  <SelectValue placeholder="Integration" />
                </SelectTrigger>
                <SelectContent>
                  {integrations.map((i) => (
                    <SelectItem key={i.id} value={i.id!}>
                      {i.host}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
          <CommandInput
            placeholder="Search repositories or paste a URL…"
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {error && <div className="px-3 py-2 text-meta text-danger">{error}</div>}
            {selected?.type === GIT_INTEGRATION_TYPE_GITHUB_APP && (
              <CommandGroup>
                {filteredRepos.map((repo) => (
                  <CommandItem
                    key={repo.full_name}
                    value={repo.full_name!}
                    onSelect={() => void pickRepo(repo)}
                  >
                    {repo.private ? (
                      <Lock className="h-3.5 w-3.5 text-muted-foreground" />
                    ) : (
                      <Globe className="h-3.5 w-3.5 text-muted-foreground" />
                    )}
                    <span className="flex-1 truncate font-mono text-meta">{repo.full_name}</span>
                    {repo.default_branch && (
                      <span className="text-label text-muted-foreground">{repo.default_branch}</span>
                    )}
                  </CommandItem>
                ))}
                {!searching && filteredRepos.length === 0 && (
                  <CommandEmpty>No repositories found.</CommandEmpty>
                )}
              </CommandGroup>
            )}
            {query && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  {looksLikeUrl && (
                    <CommandItem value={`url-${query}`} onSelect={useAsUrl}>
                      <Link2 className="h-3.5 w-3.5 text-muted-foreground" />
                      <span className="truncate text-meta">Use &quot;{query}&quot; as repository URL</span>
                    </CommandItem>
                  )}
                  {!looksLikeUrl && selected?.type === GIT_INTEGRATION_TYPE_CREDENTIALS && (
                    <CommandItem value={`host-${query}`} onSelect={useOnHost}>
                      <Link2 className="h-3.5 w-3.5 text-muted-foreground" />
                      <span className="truncate text-meta">
                        Use {selected.host}/{query}
                      </span>
                    </CommandItem>
                  )}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
