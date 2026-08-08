import { useCallback, useEffect, useMemo, useState } from "react"
import { GitBranch } from "lucide-react"
import { formatDistanceToNowStrict } from "date-fns"

import { EmptyState, PickerList, PickerRow, PickerRowTick, SearchGlyph } from "@/components/branded"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SegmentedControl } from "@/components/ui/segmented-control"
import { getErrorMessage } from "@/api/client"
import { listGitIntegrations, listRepositories, type GitRepository } from "@/api/git-integrations"
import { usableIntegrations } from "@/components/git-source-picker/credentials-dropdown"
import { repoTail } from "@/components/git-source-picker/git-source-picker"
import type { PickedRepo } from "@/components/git-source-picker/types"
import { getCurrentOrganizationId } from "@/lib/common"

import { SearchField } from "../search-field"

type Mode = "provider" | "url"

const MODES = [
  { value: "provider" as const, label: "Connected provider", showLabel: true },
  { value: "url" as const, label: "Public URL", showLabel: true },
]

/**
 * Your own code — either from a provider this organisation has connected, or
 * from any public URL.
 *
 * **Nothing here asks for a branch, a port or a Dockerfile path.** Those belong
 * to a resource, and a resource is edited in the node inspector on the canvas.
 * Asking for them before the stack exists made a five-field form out of a
 * decision that is really just "which repository".
 */
export function RepositoryTab({
  mode,
  onModeChange,
  repo,
  onRepoChange,
  url,
  onUrlChange,
}: {
  mode: Mode
  onModeChange: (mode: Mode) => void
  repo: PickedRepo | null
  onRepoChange: (repo: PickedRepo | null) => void
  url: string
  onUrlChange: (url: string) => void
}) {
  const [integrationId, setIntegrationId] = useState<string | null>(null)
  const [hasProvider, setHasProvider] = useState<boolean | null>(null)
  const [repos, setRepos] = useState<GitRepository[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState("")

  const loadIntegrations = useCallback(async () => {
    const orgId = getCurrentOrganizationId()
    if (!orgId) return
    try {
      const list = usableIntegrations((await listGitIntegrations(orgId)).items ?? [])
      setHasProvider(list.length > 0)
      setIntegrationId(list[0]?.id ?? null)
    } catch (e) {
      setError(getErrorMessage(e))
      setHasProvider(false)
    }
  }, [])

  useEffect(() => {
    void loadIntegrations()
  }, [loadIntegrations])

  useEffect(() => {
    const orgId = getCurrentOrganizationId()
    if (!orgId || !integrationId) return
    let live = true
    setLoading(true)
    listRepositories(orgId, integrationId)
      .then((page) => live && setRepos(page.items ?? []))
      .catch((e) => live && setError(getErrorMessage(e)))
      .finally(() => live && setLoading(false))
    return () => {
      live = false
    }
  }, [integrationId])

  const matches = useMemo(() => {
    const q = query.trim().toLowerCase()
    return repos.filter((r) => !q || (r.full_name ?? "").toLowerCase().includes(q))
  }, [repos, query])

  return (
    <div className="flex flex-col gap-3.5">
      {/* `self-start`, or the flex column stretches the track to 620px and the
          control stops reading as a control (§11 — it hugs its segments). */}
      <SegmentedControl
        options={MODES}
        value={mode}
        onValueChange={onModeChange}
        aria-label="Where the code lives"
        className="self-start"
      />

      {mode === "url" ? (
        <div className="flex flex-col gap-1.5">
          <Input
            value={url}
            onChange={(event) => onUrlChange(event.target.value)}
            placeholder="https://github.com/acme/web-api.git"
            aria-label="Repository URL"
            className="font-mono"
          />
          <p className="text-meta text-fg-muted">
            Any public repository. A private one needs a connected provider.
          </p>
        </div>
      ) : hasProvider === false ? (
        /* A brand-new organisation has connected nothing, so the first thing
           this tab can ever show is this — not an empty list. */
        <EmptyState
          title="No git provider connected yet"
          description="Connect one and your repositories show up here. You can also paste a public URL instead."
          action={
            <Button shape="pill" asChild>
              <a href="/git-integrations">Connect a provider</a>
            </Button>
          }
        />
      ) : (
        <>
          <SearchField
            value={query}
            onChange={setQuery}
            placeholder="Search your repositories…"
            label="Search your repositories"
          />
          {error ? (
            <EmptyState title="Those repositories could not be loaded" description={error} />
          ) : loading && repos.length === 0 ? (
            <p className="text-meta text-fg-muted px-2 py-6">Loading your repositories…</p>
          ) : matches.length === 0 ? (
            <EmptyState
              icon={<SearchGlyph />}
              title="No repository matches that"
              description="Try a shorter word, or paste a public URL instead."
            />
          ) : (
            <PickerList aria-label="Repositories">
              {matches.map((r) => {
                const fullName = r.full_name ?? repoTail(r.clone_url ?? "")
                const selected = repo?.fullName === fullName
                return (
                  <PickerRow
                    key={fullName}
                    icon={<GitBranch />}
                    name={fullName}
                    meta={[
                      { text: r.default_branch || "main", mono: true },
                      ...(r.pushed_at
                        ? [{ text: `updated ${formatDistanceToNowStrict(new Date(r.pushed_at))} ago` }]
                        : []),
                    ]}
                    selected={selected}
                    trailing={selected ? <PickerRowTick /> : null}
                    onClick={() =>
                      onRepoChange({
                        fullName,
                        cloneUrl: r.clone_url ?? "",
                        defaultBranch: r.default_branch ?? "",
                        integrationId,
                      })
                    }
                  />
                )
              })}
            </PickerList>
          )}
        </>
      )}
    </div>
  )
}
