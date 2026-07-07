import { useParams, useLocation, useNavigate, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import { useMemo, useState, useEffect, useCallback, useRef } from "react";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import { useStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { StackLogsTab } from "@/pages/stacks/components/detail/logs/stack-logs-tab";
import { StackMetricsTab } from "@/pages/stacks/components/detail/metrics/stack-metrics-tab";
import { DeploymentsTab } from "@/pages/stacks/components/detail/deployments/deployments-tab";
import { StackCanvasTab } from "@/pages/stacks/components/canvas/StackCanvasTab";
import { CanvasEditorShell } from "@/pages/stacks/components/canvas/CanvasEditorShell";
import { ViewChangesModal } from "@/pages/stacks/components/canvas/ViewChangesModal";
import { DraftTabPlaceholder } from "@/pages/stacks/components/canvas/DraftTabPlaceholder";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import type { StackConnection } from "@/api/connections";
import { alignBaselineToDraft } from "@/pages/stacks/lib/stack-diff";
import { createStack, getStackById, deleteStack } from "@/api/stacks";
import { emptyDraftSeed, buildDraftFormData, type DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import { createRelease, cancelRelease, rollbackRelease } from "@/api/releases";
import { useReleases } from "@/pages/stacks/components/detail/deployments/use-releases";
import { useReleaseDetail } from "@/pages/stacks/components/detail/deployments/use-release-detail";
import { ReleaseState } from "@/pages/stacks/components/detail/deployments/release-states";
import { useDeployLifecycle } from "@/pages/stacks/components/detail/deployments/use-deploy-lifecycle";
import {
  connectionsToEnvRows,
  connectionsToMounts,
} from "@/pages/stacks/lib/connection-mapping";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { useCurrentUser } from "@/hooks/use-current-user";
import { useOrgDomains } from "@/hooks/use-org-domains";
import { pickBestIngress } from "@/pages/stacks/lib/public-endpoints";
import type { z } from "zod";
import { convertApiResourceToFormResource, convertApiVolumeToFormVolume, convertFormStackToApiStack, FormStackSchema } from "@/pages/stacks/schemas/form-schema";
import { useToast } from "@/components/ui/use-toast";
import type { ApiStackResourceSchema, ApiVolumeSchema } from "@/pages/stacks/schemas/api-schema";
import { useDraftSync } from "@/pages/stacks/hooks/use-draft-sync";
import { useStackRevert } from "@/pages/stacks/hooks/use-stack-revert";
import { useVolumeDelete } from "@/pages/stacks/hooks/use-volume-delete";
import { buildDesiredState } from "@/pages/stacks/lib/draft-sync/desired-state";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

// Helper to map an API StackResource to the form schema shape

function mapStackResourceToFormData(resource: StackResource): FormStackResourceData {
  // Remove read-only fields before converting to form data
  const { id: _id, stack_id: _stackId, revision: _revision, ...writableResource } = resource;

  const cleanedVolumeMounts = writableResource.volume_mounts?.map((volumeMount) => {
    const { stack_resource_id: _stackResourceId, source_volume_type: _sourceVolumeType, ...cleanVolumeMount } = volumeMount;
    return cleanVolumeMount;
  });

  const resourceWithCleanedMounts = {
    ...writableResource,
    volume_mounts: cleanedVolumeMounts
  };

  return convertApiResourceToFormResource(resourceWithCleanedMounts as z.infer<typeof ApiStackResourceSchema> & { status?: unknown });
}

function mapVolumeToFormData(volume: Volume): VolumeFormData {
  // Remove read-only fields before converting to form data
  const { id: _id, ...writableVolume } = volume;
  return convertApiVolumeToFormVolume(writableVolume as z.infer<typeof ApiVolumeSchema> & { status?: unknown });
}

/** Map a resource+connection set (live stack spec OR release snapshot — both use
 *  the same server shapes) into form data, folding connection-backed env rows
 *  and volume mounts into each resource. */
function formResourcesFromSpec(
  resources: StackResource[] | undefined,
  connectionsIn: StackConnection[] | undefined,
): FormStackResourceData[] {
  const connections = connectionsIn ?? [];
  return (resources || []).map((r) => {
    const form = mapStackResourceToFormData(r);
    const connRows = connectionsToEnvRows(form.name ?? "", connections) as FormEnvVarData[];
    // Populate volume_mounts from volume_mount connections — the server always
    // returns resource.volume_mounts as [] since mounts are stored in connections.
    const mountRows = connectionsToMounts(form.name ?? "", connections);
    // connectionsToMounts only emits rows with all required fields present (it
    // skips malformed connections), so the cast to the strict form type is safe.
    const withMounts: FormStackResourceData = { ...form, volume_mounts: mountRows as FormStackResourceData["volume_mounts"] };
    if (connRows.length === 0) return withMounts;
    return {
      ...withMounts,
      execution_config: {
        ...(withMounts.execution_config ?? {}),
        environment_variables: [
          ...((withMounts.execution_config?.environment_variables ?? []) as FormEnvVarData[]),
          ...connRows,
        ],
      },
    };
  });
}

export default function StackDetailPage() {
  const { id } = useParams();
  const isDraft = !id;
  const location = useLocation();
  const navigate = useNavigate();
  const seed = useMemo<DraftSeed>(
    () => ((location.state as { seed?: DraftSeed } | null)?.seed) ?? emptyDraftSeed(),
    // read once from the entry navigation; later navigations replace state
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );
  const [draftName, setDraftName] = useState(seed.name);
  // Draft labels are seeded into the create payload; there is no in-canvas label
  // editor, so the setter is intentionally dropped.
  const [draftLabels] = useState<FormStackData["labels"]>(seed.labels);

  const { stacks, setStacks } = useStacks();
  const [fetchedStack, setFetchedStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const session = useStackEditSession();
  const [activeTab, setActiveTab] = useState("configuration");
  const [isCreating, setIsCreating] = useState(false);
  const [nameError, setNameError] = useState<string | undefined>();

  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();
  const { teamNameById, defaultTeamName } = useResourceTeams();
  const { canWrite } = useCurrentUser();

  // Find the current stack in context
  const currentStack = stacks.find((stack) => stack.id === id);

  // Viewer read-only gating: only OrgAdmin / team Developer may mutate this stack.
  const stackTeamId = fetchedStack?.team_id ?? currentStack?.team_id;
  const canWriteStack = canWrite(stackTeamId ?? "");

  // Update breadcrumb with stack name
  useEffect(() => {
    if (isDraft) return;
    const path = `/stacks/${id}`;

    if (currentStack) {
      // If stack is already available in context, use its name for breadcrumb
      setCustomLabel(path, currentStack.name|| 'Stack Details');
    } else if (id) {
      // Set loading state while fetching
      setPathLoading(path, true);

      const orgId = getCurrentOrganizationId();
      if (!orgId) {
        setError("Organization ID not found.");
        setPathLoading(path, false);
        return;
      }

      setLoading(true);
      setError(null);
      // Single-stack read is team-scoped; wait for the default team to resolve
      // (this effect re-runs once it does).
      if (!defaultTeamName) {
        return;
      }
      getStackById(orgId, defaultTeamName, id)
        .then((data) => {
          setFetchedStack(data);
          setLoading(false);
          // Update breadcrumb with fetched stack name
          setCustomLabel(path, data.name || 'Stack Details');
          setPathLoading(path, false);
        })
        .catch(() => {
          setError("Failed to load stack. Please try again later.");
          setLoading(false);
          setPathLoading(path, false);
        });
    }
  }, [currentStack, id, defaultTeamName, setCustomLabel, setPathLoading, isDraft]);

  const stackToShow = currentStack || fetchedStack;

  const draftStackView = useMemo(
    () =>
      isDraft
        ? ({
          name: draftName,
          labels: draftLabels,
          spec: { stack_resources: session.draft.resources, volumes: session.draft.volumes, connections: [] },
        } as unknown as Stack)
        : null,
    [isDraft, draftName, draftLabels, session.draft.resources, session.draft.volumes],
  );
  const effectiveStack = draftStackView ?? stackToShow;

  // Publicly exposed services → best live ingress URL, for the header's
  // PUBLIC row. Drafts have no live ingress, so the row stays empty.
  const orgDomains = useOrgDomains(effectiveStack?.organisation_id ?? getCurrentOrganizationId() ?? undefined);
  const publicEndpoints = useMemo(() => {
    if (isDraft) return [];
    return (effectiveStack?.spec.stack_resources ?? []).flatMap((r) => {
      const best = pickBestIngress(r.status?.public_ingress ?? [], orgDomains);
      return best && r.name ? [{ service: r.name, url: best.url, port: best.target_port }] : [];
    });
  }, [isDraft, effectiveStack, orgDomains]);

  // ── Release plumbing (needed this early: the diff baseline is pinned to the
  // latest release's snapshot, not the autosaved server state) ──
  const deployIds = useMemo(() => ({
    orgId: stackToShow?.organisation_id || getCurrentOrganizationId() || "",
    teamName: (stackToShow ? teamNameById(stackToShow.team_id) : "") || defaultTeamName || "",
    stackId: stackToShow?.id || "",
  }), [stackToShow, teamNameById, defaultTeamName]);
  const releasesResult = useReleases({ ...deployIds, enabled: !!deployIds.stackId });
  const releaseDetail = useReleaseDetail(deployIds.orgId, deployIds.teamName, deployIds.stackId);

  // Diff anchor: the latest release (the config last shipped via Deploy), falling
  // back to the converged release until the releases list loads.
  const baselineReleaseId =
    releasesResult.activeRelease?.id ?? stackToShow?.status?.last_converged?.release_id;
  useEffect(() => {
    if (baselineReleaseId) releaseDetail.ensure(baselineReleaseId);
  }, [baselineReleaseId, releaseDetail]);
  const deployedSnapshot = releaseDetail.peek(baselineReleaseId).data?.snapshot;

  // Current server state as form data — what the canvas displays and the edit
  // session's working draft seeds from.
  const draftResources = useMemo<FormStackResourceData[]>(
    () => formResourcesFromSpec(stackToShow?.spec?.stack_resources, stackToShow?.spec?.connections),
    [stackToShow],
  );
  const draftVolumes = useMemo<VolumeFormData[]>(
    () => (stackToShow?.spec?.volumes || []).map(mapVolumeToFormData),
    [stackToShow],
  );

  // Server-computed outputs (host, port.<n>, url.<n>, public.<n>.*) keyed by
  // resource name. The working draft never computes outputs, so a resource added
  // on the canvas carries none until it is saved. The env-var OUTPUT pickers read
  // from this server-truth map (matched by name) instead of the draft copy, and
  // it refreshes with stackToShow after every autosave — no page reload needed.
  const serverOutputsByName = useMemo<Map<string, string[]>>(
    () =>
      new Map(
        (stackToShow?.spec?.stack_resources ?? [])
          .filter((r): r is StackResource & { name: string } => !!r.name)
          .map((r) => [r.name, (r.outputs ?? []).map((o) => o.name)] as [string, string[]]),
      ),
    [stackToShow?.spec?.stack_resources],
  );

  // Diff baseline: the deployed snapshot when one exists, so autosaved edits stay
  // visibly dirty/revertable until deployed. Never-deployed stacks fall back to
  // the server state (everything reads as staged for the first deploy).
  const snapshotResources = useMemo<FormStackResourceData[] | null>(
    () =>
      deployedSnapshot
        ? formResourcesFromSpec(
          deployedSnapshot.resources as StackResource[] | undefined,
          deployedSnapshot.connections as StackConnection[] | undefined,
        )
        : null,
    [deployedSnapshot],
  );
  const snapshotVolumes = useMemo<VolumeFormData[] | null>(
    () => (deployedSnapshot ? ((deployedSnapshot.volumes ?? []) as Volume[]).map(mapVolumeToFormData) : null),
    [deployedSnapshot],
  );

  // All diffing downstream is positional, but the server returns resources in
  // unstable order and the snapshot's order need not match — re-key the baseline
  // onto the order of whatever the diffs actually run against: the live session
  // draft when one is active, the server state otherwise.
  const alignResources = session.isActive ? session.draft.resources : draftResources;
  const alignVolumes = session.isActive ? session.draft.volumes : draftVolumes;
  const baselineResources = useMemo<FormStackResourceData[]>(
    () =>
      (snapshotResources
        ? alignBaselineToDraft(snapshotResources, alignResources)
        : alignBaselineToDraft(draftResources, alignResources)) as FormStackResourceData[],
    [snapshotResources, draftResources, alignResources],
  );
  const baselineVolumes = useMemo<VolumeFormData[]>(
    () =>
      (snapshotVolumes
        ? alignBaselineToDraft(snapshotVolumes, alignVolumes)
        : alignBaselineToDraft(draftVolumes, alignVolumes)) as VolumeFormData[],
    [snapshotVolumes, draftVolumes, alignVolumes],
  );

  const draftSeeded = useRef(false);
  useEffect(() => {
    if (!isDraft || draftSeeded.current) return;
    draftSeeded.current = true;
    // Baseline empty so seeded resources/volumes read as "added" and Save is enabled.
    session.start({ resources: [], volumes: [] }, { linkedAddonIds: new Set(seed.linkedAddonIds) });
    if (seed.resources.length) session.updateResources(() => seed.resources);
    if (seed.volumes.length) session.updateVolumes(() => seed.volumes);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isDraft]);

  // Addons bound to the saved stack come from its connections (from.type
  // "addon/postgres"), not from the env vars — so this is the source of truth
  // for the "Stack Addons" panel in display mode.
  const { addons: allAddons } = usePostgresAddons();
  // addonId → display name, for read-mode addon group headers.
  const addonNameById = useMemo(
    () =>
      new Map(
        allAddons
          .filter((a: PostgresAddon) => a.id && a.name)
          .map((a: PostgresAddon) => [a.id!, a.name!] as [string, string]),
      ),
    [allAddons],
  );

  const connectionAddonIds = useMemo<Set<string>>(
    () =>
      new Set(
        (stackToShow?.spec?.connections ?? [])
          .filter((c) => c.from?.type === "addon/postgres" && c.from?.id)
          .map((c) => c.from!.id as string),
      ),
    [stackToShow],
  );

  // Autosave model: the canvas is always editable for writers. The session
  // starts as soon as the stack is loaded and restarts after discard/revert.
  // Baseline = deployed snapshot (when loaded), draft = current server state:
  // they differ when the server already holds autosaved-but-undeployed edits.
  useEffect(() => {
    if (isDraft || !stackToShow || !canWriteStack || session.isActive) return;
    session.start(
      { resources: baselineResources, volumes: baselineVolumes },
      {
        linkedAddonIds: connectionAddonIds,
        draft: { resources: draftResources, volumes: draftVolumes },
      },
    );
    // session.start is a stable useCallback; session.isActive is the only reactive
    // field we need from the session object itself.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isDraft, stackToShow, canWriteStack, session.isActive, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds]);

  // When a release snapshot for a NEW anchor arrives — the lazy fetch landing, or
  // a fresh deploy creating a new release — advance the session baseline to it so
  // "dirty" always means "differs from the latest release". Guarded per release
  // id: autosave stack refreshes must never move the baseline.
  const rebasedReleaseRef = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (!session.isActive || !deployedSnapshot || !baselineReleaseId) return;
    if (rebasedReleaseRef.current === baselineReleaseId) return;
    rebasedReleaseRef.current = baselineReleaseId;
    session.rebase({ resources: baselineResources, volumes: baselineVolumes });
    // session.rebase is a stable useCallback.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session.isActive, deployedSnapshot, baselineReleaseId, baselineResources, baselineVolumes]);

  // Bumped on every autosave refresh to trigger a topology refetch.
  const [topologyRefreshKey, setTopologyRefreshKey] = useState(0);

  // Volumes that exist server-side; their spec (PVC size) is immutable, so the
  // drawer renders those fields read-only.
  const persistedVolumeNames = useMemo(
    () =>
      new Set(
        (stackToShow?.spec?.volumes ?? [])
          .map((v) => v.name)
          .filter((n): n is string => !!n),
      ),
    [stackToShow?.spec?.volumes],
  );

  // Backend validation errors from a failed autosave, mapped onto the offending
  // env-var field so the row shows them inline. Keyed by resource NAME → env-var
  // name → message (NOT by index): the resource may be reordered/removed before
  // the user fixes it, so the current index + env index are resolved at render
  // time when merging into the index-keyed errors prop.
  const [serverFieldErrors, setServerFieldErrors] = useState<{
    [resourceName: string]: { [envName: string]: string };
  }>({});

  // Dedupe for autosave-failure toasts: the backoff retry loop and re-edits after
  // a terminal 400 re-run the same ops and would re-toast the same reason every
  // cycle. Remember the last-shown message and only toast when it changes; cleared
  // on a successful sync so a later distinct failure still toasts.
  const lastSyncErrorMsgRef = useRef<string | null>(null);

  // Autosave engine: debounces draft changes and syncs thin per-resource ops to
  // the server. Disabled for drafts (nothing exists server-side to sync yet).
  const draftSync = useDraftSync({
    enabled: !isDraft && canWriteStack,
    stack: stackToShow ?? undefined,
    session,
    ids: deployIds.stackId ? deployIds : null,
    onStackRefreshed: (fresh) => {
      setFetchedStack(fresh);
      // Context write-through: stale currentStack must not win after a remote refresh.
      setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
      setTopologyRefreshKey((k) => k + 1);
    },
    onSyncError: ({ message, op }) => {
      // No op → the save actually landed and only the post-save refetch failed;
      // don't alarm the user with a "Save failed" toast or field error.
      if (!op) return;
      // Dedupe: skip if this exact reason is already on screen (retry/re-edit).
      if (lastSyncErrorMsgRef.current === message) return;
      lastSyncErrorMsgRef.current = message;
      toast({ title: "Save failed", description: message, variant: "destructive" });
      // Only resource create/update ops carry env vars; others just toast.
      if (op.kind !== "createResource" && op.kind !== "updateResource") return;
      const resourceName = op.resource.name;
      if (!resourceName) return;
      // The backend reason names the offending var: `env var '<NAME>'`.
      const envName = message.match(/env var '([^']+)'/)?.[1];
      if (!envName) return;
      // Key by resource + env-var NAME; the current row indices are resolved at
      // render time so a reorder/removal can't misattach the inline error.
      setServerFieldErrors((prev) => ({
        ...prev,
        [resourceName]: { ...(prev[resourceName] ?? {}), [envName]: message },
      }));
    },
  });

  // A successful sync clears any stale server-side field errors (the draft that
  // failed has since been fixed and persisted) and the toast-dedupe memory so a
  // later, distinct failure toasts again.
  useEffect(() => {
    if (draftSync.status === SYNC_STATUS.saved) {
      lastSyncErrorMsgRef.current = null;
      setServerFieldErrors((prev) => (Object.keys(prev).length === 0 ? prev : {}));
    }
  }, [draftSync.status]);

  // Live drawer validation: compute desired state from draft and expose zod issues
  // per resource index. Issue paths are relative to the resource root (no
  // ["spec","stack_resources",idx] prefix — drop that prefix at this boundary).
  const desiredState = useMemo(() => buildDesiredState(session.draft), [session.draft]);

  const validationErrors = useMemo(() => {
    const resources: { [index: number]: { [field: string]: string | undefined } } = {};
    desiredState.resourceIssues.forEach((issues, idx) => {
      resources[idx] = {};
      for (const issue of issues) {
        const fieldKey = issue.path.join(".");
        if (!resources[idx][fieldKey]) resources[idx][fieldKey] = issue.message;
      }
    });
    return { resources, volumes: {} };
  }, [desiredState.resourceIssues]);

  // Merge live zod validation (index-keyed) with backend field errors from a
  // failed autosave (name-keyed). Resolve each server error's resource NAME and
  // env-var NAME to their CURRENT indices here so the merged map stays index-keyed
  // for consumers but survives a reorder/removal since the error was captured.
  const mergedResourceErrors = useMemo(() => {
    const merged: { [index: number]: { [field: string]: string | undefined } } = {};
    for (const [k, v] of Object.entries(validationErrors.resources)) {
      merged[Number(k)] = { ...v };
    }
    for (const [resourceName, envErrors] of Object.entries(serverFieldErrors)) {
      const idx = session.draft.resources.findIndex((r) => r.name === resourceName);
      if (idx < 0) continue;
      const draftEnv = (session.draft.resources[idx]?.execution_config
        ?.environment_variables ?? []) as FormEnvVarData[];
      for (const [envName, message] of Object.entries(envErrors)) {
        const envIdx = draftEnv.findIndex((e) => e.name === envName);
        if (envIdx < 0) continue;
        const fieldKey = `execution_config.environment_variables.${envIdx}.value`;
        merged[idx] = { ...(merged[idx] ?? {}), [fieldKey]: message };
      }
    }
    return merged;
  }, [validationErrors.resources, serverFieldErrors, session.draft.resources]);

  const lifecycle = useDeployLifecycle({
    stack: stackToShow ?? undefined,
    // "Editing" = autosave hasn't landed yet (in flight or retrying). Saved-but-
    // undeployed content is detected by the staged-phase content diff instead.
    unsaved: draftSync.status === SYNC_STATUS.saving || draftSync.failureCount > 0,
    releases: releasesResult.releases,
    activeRelease: releasesResult.activeRelease,
    detail: releaseDetail,
  });

  // When a release converges, the server moves status.last_converged but the
  // client's stack copy still points at the previous release, so the staged
  // panel keeps diffing against the old snapshot until a manual page refresh.
  // Refetch the stack once per newly-released release to pick up the pointer.
  const convergedFetchRef = useRef<string | undefined>(undefined);
  const activeRelease = releasesResult.activeRelease;
  useEffect(() => {
    if (!deployIds.stackId || !activeRelease) return;
    if (activeRelease.state !== ReleaseState.Released) return;
    if (stackToShow?.status?.last_converged?.release_id === activeRelease.id) return;
    if (convergedFetchRef.current === activeRelease.id) return;
    convergedFetchRef.current = activeRelease.id;
    void getStackById(deployIds.orgId, deployIds.teamName, deployIds.stackId).then((fresh) => {
      setFetchedStack(fresh);
      setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
      // Server topology may have changed with the new release; re-derive edges.
      setTopologyRefreshKey((k) => k + 1);
    }).catch(() => {
      // Poll-driven refresh; the next release poll retries naturally.
      convergedFetchRef.current = undefined;
    });
  }, [deployIds, activeRelease, stackToShow?.status?.last_converged?.release_id, setStacks]);

  // Live snapshot: already lazily fetched by useDeployLifecycle via detail.ensure;
  // peek here to gate canDiscardDraft and pass to the revert hook.
  const liveReleaseId = stackToShow?.status?.last_converged?.release_id;
  const liveSnapshot = releaseDetail.peek(liveReleaseId).data?.snapshot;

  const [revertConfirmOpen, setRevertConfirmOpen] = useState(false);
  const [viewChangesOpen, setViewChangesOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const stackRevert = useStackRevert({
    ids: deployIds.stackId ? deployIds : null,
    stack: stackToShow ?? undefined,
    liveSnapshot,
    onReverted: (fresh) => {
      setFetchedStack(fresh);
      setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
      draftSync.notifyExternalUpdate(fresh);
      session.discard(); // auto-start effect restarts the session on the reverted baseline
      toast({ title: "Draft discarded", description: "Stack restored to the last deployment.", variant: "success" });
    },
  });

  // Immediate, confirm-gated volume deletion (canvas). Only wired for saved
  // stacks — the wizard (`/stacks/new`) has nothing server-side to delete yet.
  const volumeDelete = useVolumeDelete({
    ids: deployIds.stackId ? deployIds : null,
    draftSync,
    onServerRefresh: (fresh) => {
      setFetchedStack(fresh);
      setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
      setTopologyRefreshKey((k) => k + 1);
    },
    onRestoreVolume: (vol) => {
      session.updateVolumes((vs) => [...vs, mapVolumeToFormData(vol)]);
    },
    toast,
  });

  const [deployBusy, setDeployBusy] = useState(false);
  const refetchReleases = releasesResult.refetch;
  const runDeploy = useCallback(async (fn: () => Promise<unknown>, ok: string) => {
    setDeployBusy(true);
    try {
      await fn();
      toast({ title: ok, variant: "success" });
      refetchReleases();
    } catch (e) {
      toast({ title: "Action failed", description: e instanceof Error ? e.message : "", variant: "destructive" });
    } finally {
      setDeployBusy(false);
    }
  }, [toast, refetchReleases]);

  const onDeploy = useCallback(async () => {
    if (!draftSync || !deployIds.stackId) return;
    const flushed = await draftSync.flush();
    if (!flushed) {
      toast({
        title: "Deploy blocked",
        description: "Draft changes failed to save. Fix the save error and try again.",
        variant: "destructive",
      });
      return;
    }
    runDeploy(() => createRelease(deployIds.orgId, deployIds.teamName, deployIds.stackId), "Deploy started");
  }, [draftSync, deployIds, runDeploy, toast]);
  const onCancelDeploy = useCallback(
    (releaseId: string) => runDeploy(() => cancelRelease(deployIds.orgId, deployIds.teamName, deployIds.stackId, releaseId), "Release cancelled"),
    [runDeploy, deployIds],
  );
  const onRollback = useCallback(
    (releaseId: string) => runDeploy(() => rollbackRelease(deployIds.orgId, deployIds.teamName, deployIds.stackId, releaseId), "Rollback started"),
    [runDeploy, deployIds],
  );
  const onCopyId = useCallback((releaseId: string) => {
    void navigator.clipboard?.writeText(releaseId);
    toast({ title: "Release ID copied" });
  }, [toast]);

  // Draft create: validates name, creates the stack, and navigates to the new page.
  const performCreate = async () => {
    if (!isDraft) return;
    setIsCreating(true);
    setNameError(undefined);

    // A draft needs a name before it can be created. The stack name field is not
    // min-length-constrained in the schema (empty passes zod and only fails at
    // the API), so guard it here to surface the error inline on the title input.
    if (!draftName.trim()) {
      setNameError("Required");
      setIsCreating(false);
      toast({
        title: "Name your stack",
        description: "Give the stack a name before saving.",
        variant: "destructive",
      });
      return;
    }

    try {
      const orgId = getCurrentOrganizationId();
      if (!orgId) throw new Error("Organization ID not found");

      const resources = session.draft.resources as FormStackResourceData[];
      const formStackData: FormStackData = buildDraftFormData(
        draftName.trim(),
        draftLabels,
        resources,
        session.draft.volumes as VolumeFormData[],
      );

      const validation = FormStackSchema.safeParse(formStackData);
      if (!validation.success) {
        const topLevelMessages: string[] = [];
        let newNameError: string | undefined;

        for (const issue of validation.error.issues) {
          const [scope0, scope1, idx, ...fieldPath] = issue.path;
          const field = fieldPath.join(".");
          if (scope0 === "name") {
            newNameError = issue.message;
          } else if (scope0 === "spec" && scope1 === "stack_resources" && typeof idx === "number") {
            const label = resources[idx]?.name?.trim() || `Resource ${idx + 1}`;
            topLevelMessages.push(field ? `${label}: ${issue.message} (${field})` : `${label}: ${issue.message}`);
          } else if (scope0 === "spec" && scope1 === "volumes" && typeof idx === "number") {
            const vols = formStackData.spec.volumes as { name?: string }[] | undefined;
            const label = vols?.[idx]?.name?.trim() || `Volume ${idx + 1}`;
            topLevelMessages.push(field ? `${label}: ${issue.message} (${field})` : `${label}: ${issue.message}`);
          } else {
            topLevelMessages.push(issue.path.length ? `${issue.path.join(".")}: ${issue.message}` : issue.message);
          }
        }

        const uniqueMessages = [...new Set(topLevelMessages)];
        setNameError(newNameError);
        toast({
          title: "Validation error",
          description: uniqueMessages.length > 0
            ? uniqueMessages.join("; ")
            : "Please fix the highlighted errors before saving.",
          variant: "destructive",
        });
        setIsCreating(false);
        return;
      }

      const teamName = defaultTeamName;
      if (!teamName) {
        toast({
          title: "No team available",
          description: "Could not resolve a team to save into.",
          variant: "destructive",
        });
        setIsCreating(false);
        return;
      }

      const apiData = convertFormStackToApiStack(formStackData);
      const created = await createStack(orgId, teamName, apiData);
      session.discard();
      navigate(`/stacks/${created.id}`, { replace: true, state: null });
    } catch (err) {
      console.error('Failed to create stack:', err);
      toast({
        title: "Failed to create stack",
        description: err instanceof Error ? err.message : "An unexpected error occurred. Please try again.",
        variant: "destructive"
      });
    } finally {
      setIsCreating(false);
    }
  };

  const performDelete = useCallback(async () => {
    if (!deployIds) return;
    setDeleting(true);
    try {
      await deleteStack(deployIds.orgId, deployIds.teamName, deployIds.stackId);
      setStacks((prev) => prev.filter((s) => s.id !== deployIds.stackId));
      toast({ title: "Stack deleted", description: `"${stackToShow?.name}" was deleted.`, variant: "success" });
      navigate("/stacks");
    } catch {
      toast({ title: "Delete failed", description: "The stack could not be deleted.", variant: "destructive" });
    } finally {
      setDeleting(false);
      setDeleteConfirmOpen(false);
    }
  }, [deployIds, setStacks, stackToShow?.name, toast, navigate]);

  const handleNameChange = useCallback((name: string) => {
    setDraftName(name);
    setNameError(undefined);
  }, []);

  // Revert one resource/volume from the View-changes modal by name → session index.
  const discardResourceByName = useCallback(
    (name: string) => {
      const idx = session.draft.resources.findIndex((r) => (r as { name?: string }).name === name);
      if (idx >= 0) session.discardResource(idx);
    },
    [session],
  );
  const discardVolumeByName = useCallback(
    (name: string) => {
      const idx = session.draft.volumes.findIndex((v) => (v as { name?: string }).name === name);
      if (idx >= 0) session.discardVolume(idx);
    },
    [session],
  );

  if (!isDraft && loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading stack...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Error</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button asChild>
          <Link to="/stacks">Return to Stacks</Link>
        </Button>
      </div>
    );
  }

  if (!isDraft && !stackToShow) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Stack not found</h2>
        <p className="text-muted-foreground mb-4">The stack you're looking for doesn't exist or has been deleted.</p>
        <Button asChild>
          <Link to="/stacks">Return to Stacks</Link>
        </Button>
      </div>
    );
  }

  const resourceCount = effectiveStack?.spec?.stack_resources?.length || 0;
  const volumeCount = effectiveStack?.spec?.volumes?.length || 0;
  const addonCount = connectionAddonIds.size;
  const subtitleText = [
    `${resourceCount} ${resourceCount === 1 ? "resource" : "resources"}`,
    `${volumeCount} ${volumeCount === 1 ? "volume" : "volumes"}`,
    ...(addonCount > 0 ? [`${addonCount} ${addonCount === 1 ? "addon" : "addons"}`] : []),
  ].join(" · ");

  // Ops-view bodies — rendered inside the canvas shell; gated on isDraft so
  // the user sees a placeholder until the stack is saved for the first time.
  const deploymentsBody = isDraft ? <DraftTabPlaceholder label="Deployments" /> : effectiveStack?.id ? (
    <DeploymentsTab
      orgId={deployIds.orgId}
      teamName={deployIds.teamName}
      stackId={effectiveStack.id}
      stack={effectiveStack}
      onOpenLogs={() => setActiveTab("logs")}
      releases={releasesResult.releases}
      activeRelease={releasesResult.activeRelease}
      loading={releasesResult.loading}
      error={releasesResult.error}
      lifecycle={lifecycle}
      onRollback={onRollback}
      onCancel={onCancelDeploy}
      onCopyId={onCopyId}
    />
  ) : (
    <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
  );

  const logsBody = isDraft ? <DraftTabPlaceholder label="Logs" /> : effectiveStack?.id ? (
    <StackLogsTab
      stackId={effectiveStack.id}
      organizationId={effectiveStack.organisation_id || getCurrentOrganizationId() || ''}
      resources={effectiveStack.spec.stack_resources?.map(r => ({ name: r.name || '', id: r.id || '' })) || []}
    />
  ) : (
    <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
  );

  const metricsBody = isDraft ? <DraftTabPlaceholder label="Metrics" /> : effectiveStack?.id ? (
    <StackMetricsTab
      stackId={effectiveStack.id}
      organizationId={effectiveStack.organisation_id || getCurrentOrganizationId() || ''}
      resources={effectiveStack.spec.stack_resources || []}
    />
  ) : (
    <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
  );

  const dirtyTotal =
    session.dirty.dirtyResourceIdx.size + session.dirty.dirtyVolumeIdx.size + session.dirty.addonLinkCount;

  // "View changes" badge + modal count must agree with the modal's BODY. The body
  // renders lifecycle.stagedDiff (saved spec vs release), which only sees SAVED
  // content. When a sync is in flight or has errored (e.g. a fresh resource whose
  // autosave 400s), the unsaved edit is absent from the staged diff — so the diff
  // can be empty while real session dirt exists. In that unsettled state, never let
  // the (stale) staged count hide the dirt: take the larger of the two. Once the
  // sync settles, the staged diff is authoritative (session dirt can overcount,
  // e.g. a mount added then removed nets to zero staged while the session still
  // reads dirty), so trust it and fall back to dirt only until the diff resolves.
  const stagedCount = lifecycle.stagedDiff
    ? lifecycle.stagedDiff.resources.length +
      lifecycle.stagedDiff.volumes.length +
      lifecycle.stagedDiff.connections.length
    : null;
  const syncUnsettled =
    draftSync.status === SYNC_STATUS.saving || draftSync.status === SYNC_STATUS.error;
  const changeCount = syncUnsettled
    ? Math.max(stagedCount ?? 0, dirtyTotal)
    : stagedCount ?? dirtyTotal;

  return (
    <>
      <CanvasEditorShell
        stackName={isDraft ? draftName : (effectiveStack?.name ?? "")}
        stackId={effectiveStack?.id}
        isDraft={isDraft}
        nameEditable={isDraft}
        onNameChange={handleNameChange}
        nameError={nameError}
        statusState={effectiveStack?.status?.state}
        subtitle={subtitleText}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        isActive={session.isActive}
        dirtyResourceCount={session.dirty.dirtyResourceIdx.size}
        dirtyTotal={changeCount}
        isStaged={lifecycle.phase === "staged"}
        onViewChanges={() => setViewChangesOpen(true)}
        syncStatus={isDraft ? SYNC_STATUS.idle : draftSync.status}
        deployBusy={deployBusy}
        canWrite={canWriteStack}
        onCreate={() => void performCreate()}
        isCreating={isCreating}
        onDeploy={onDeploy}
        canDiscardDraft={lifecycle.phase === "staged" && !!liveSnapshot && canWriteStack}
        onDiscardDraft={() => setRevertConfirmOpen(true)}
        canDeleteStack={canWriteStack}
        onDelete={() => setDeleteConfirmOpen(true)}
        publicEndpoints={publicEndpoints}
        configuration={
          <StackCanvasTab
            session={session}
            baselineResources={baselineResources}
            baselineVolumes={baselineVolumes}
            draftResources={draftResources}
            draftVolumes={draftVolumes}
            serverOutputsByName={serverOutputsByName}
            connectionAddonIds={connectionAddonIds}
            addonNameById={addonNameById}
            errors={mergedResourceErrors}
            onViewLogs={() => setActiveTab("logs")}
            topologyIds={!isDraft && deployIds.stackId ? deployIds : null}
            topologyRefreshKey={topologyRefreshKey}
            onDeleteVolume={deployIds.stackId ? volumeDelete.deleteVolume : undefined}
            deletingVolume={volumeDelete.deleting}
            persistedVolumeNames={persistedVolumeNames}
          />
        }
        deployments={deploymentsBody}
        logs={logsBody}
        metrics={metricsBody}
      />
      <ViewChangesModal
        open={viewChangesOpen}
        onOpenChange={setViewChangesOpen}
        diff={lifecycle.stagedDiff}
        count={changeCount}
        errored={draftSync.status === SYNC_STATUS.error}
        dirty={dirtyTotal > 0}
        stackName={effectiveStack?.name ?? ""}
        onDiscardResource={discardResourceByName}
        onDiscardVolume={discardVolumeByName}
        onDiscardAll={() => {
          setViewChangesOpen(false);
          // Everything the modal lists is already autosaved server-side, so a
          // local session reset discards nothing. Route through the revert
          // flow (PUT of the release snapshot), which carries its own
          // data-loss confirm. Without a deployed snapshot to restore there is
          // nothing saved to roll back — clearing the local session is all
          // "discard" can mean.
          if (lifecycle.phase === "staged" && liveSnapshot) {
            setRevertConfirmOpen(true);
          } else {
            session.discard();
          }
        }}
        onDeploy={onDeploy}
        deployBusy={deployBusy}
        canWrite={canWriteStack}
      />
      <AlertDialog open={revertConfirmOpen} onOpenChange={setRevertConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard draft changes?</AlertDialogTitle>
            <AlertDialogDescription>
            This restores the stack to its last deployment. Volumes added since the last deployment will be deleted, and previously deleted volumes will be recreated empty. Lost volume data cannot be recovered. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep editing</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                void (async () => {
                  // Flush any pending autosave before reverting (don't block on failure)
                  await draftSync.flush();
                  const ok = await stackRevert.revert();
                  if (!ok) {
                    toast({
                      title: "Discard failed",
                      description: "The stack may be partially reverted. Reload the page to see its current state.",
                      variant: "destructive",
                    });
                  }
                })();
              }}
              disabled={stackRevert.reverting}
            >
            Discard draft
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete stack?</AlertDialogTitle>
            <AlertDialogDescription>
            This permanently deletes "{stackToShow?.name}", its resources, volumes and deployments. This cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => void performDelete()}
              disabled={deleting}
            >
            Delete stack
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
