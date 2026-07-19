import { useParams, useLocation, useNavigate, Link } from "react-router-dom";
import { getErrorMessage } from "@/api/client";
import { parseApiError, type ParsedFieldError } from "@/api/errors";
import { mapFieldErrors } from "@/pages/stacks/lib/map-field-errors";
import { formatDraftValidationIssues } from "@/pages/stacks/lib/format-draft-validation";
import { buildBannerItems } from "@/pages/stacks/components/editor/lib/banner-items";
import { ValidationBanner, type ValidationBannerItem } from "@/pages/stacks/components/editor/validation-banner";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import { useMemo, useState, useEffect, useCallback, useRef } from "react";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import { useStackEditSession, type EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
import { statusVariant } from "@/components/branded/status-variant";
import { LogsTab } from "@/pages/stacks/components/editor/tabs/logs/logs-tab";
import { MetricsTab } from "@/pages/stacks/components/editor/tabs/metrics/metrics-tab";
import { DeploymentsTab } from "@/pages/stacks/components/editor/tabs/deployments/deployments-tab";
import { jumpTargetIndex } from "@/pages/stacks/components/editor/tabs/deployments/release-errors";
import { ArchitectureTab } from "@/pages/stacks/components/editor/tabs/architecture/architecture-tab";
import { CanvasEditorShell } from "@/pages/stacks/components/editor/canvas-editor-shell";
import { EDITOR_TABS, type EditorTabId } from "@/pages/stacks/components/editor/editor-tabs";
import { ViewChangesModal } from "@/pages/stacks/components/editor/view-changes-modal";
import { DraftTabPlaceholder } from "@/pages/stacks/components/editor/draft-tab-placeholder";
import { resolveChangeCount } from "@/pages/stacks/components/editor/lib/resolve-change-count";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import type { StackConnection } from "@/api/connections";
import { alignBaselineToDraft, renameFingerprint } from "@/pages/stacks/lib/stack-diff";
import { applyStackByName, getStackById, deleteStack } from "@/api/stacks";
import { emptyDraftSeed, buildDraftFormData, type DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import { createRelease, cancelRelease, rollbackRelease } from "@/api/releases";
import { useReleases } from "@/pages/stacks/components/editor/tabs/deployments/use-releases";
import { useReleaseDetail, ReleaseDetailProvider } from "@/pages/stacks/components/editor/tabs/deployments/use-release-detail";
import { deriveHeaderHealth, latestDeployFailed, shouldRefetchStackSummaries, stripUnpinnedGitRevisions } from "@/pages/stacks/components/editor/tabs/deployments/derive";
import { useDeployLifecycle } from "@/pages/stacks/components/editor/tabs/deployments/use-deploy-lifecycle";
import { useReleaseAnchors } from "@/pages/stacks/components/editor/hooks/use-release-anchors";
import { mapVolumeToFormData, formResourcesFromSpec } from "@/pages/stacks/lib/spec-to-form";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useCurrentUser } from "@/hooks/use-current-user";
import { useOrgDomains } from "@/hooks/use-org-domains";
import { pickBestIngress } from "@/pages/stacks/lib/public-endpoints";
import { convertFormStackToApiStack, FormStackSchema } from "@/pages/stacks/schemas/form-schema";
import { useToast } from "@/components/ui/use-toast";
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

export default function CanvasEditorPage() {
  const { id } = useParams();
  const isNewStack = !id;
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
  const [activeTab, setActiveTab] = useState<EditorTabId>(EDITOR_TABS.architecture);
  // Resource pre-selected in the Logs tab filter when arriving via a drawer's
  // "View logs". Cleared on direct tab navigation so the filter doesn't stick.
  const [logsInitialSource, setLogsInitialSource] = useState<string | undefined>();
  const [draftDeploying, setDraftDeploying] = useState(false);
  const [nameError, setNameError] = useState<string | undefined>();

  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();
  const { projectNameById, defaultProjectName } = useResourceProjects();
  const { canWrite } = useCurrentUser();

  const currentStack = stacks.find((stack) => stack.id === id);

  // Viewer read-only gating: only OrgAdmin / project Developer may mutate this stack.
  const stackProjectId = fetchedStack?.project_id ?? currentStack?.project_id;
  const canWriteStack = canWrite(stackProjectId ?? "");

  useEffect(() => {
    if (isNewStack) return;
    const path = `/stacks/${id}`;

    if (currentStack) {
      setCustomLabel(path, currentStack.name|| 'Stack Details');
    } else if (id) {
      setPathLoading(path, true);

      const orgId = getCurrentOrganizationId();
      if (!orgId) {
        setError("Organization ID not found.");
        setPathLoading(path, false);
        return;
      }

      setLoading(true);
      setError(null);
      // Single-stack read is project-scoped; wait for the default project to resolve
      // (this effect re-runs once it does).
      if (!defaultProjectName) {
        return;
      }
      getStackById(orgId, defaultProjectName, id)
        .then((data) => {
          setFetchedStack(data);
          setLoading(false);
          setCustomLabel(path, data.name || 'Stack Details');
          setPathLoading(path, false);
        })
        .catch(() => {
          setError("Failed to load stack. Please try again later.");
          setLoading(false);
          setPathLoading(path, false);
        });
    }
  }, [currentStack, id, defaultProjectName, setCustomLabel, setPathLoading, isNewStack]);

  const savedStack = currentStack || fetchedStack;

  const draftStackView = useMemo(
    () =>
      isNewStack
        ? ({
          name: draftName,
          labels: draftLabels,
          spec: { stack_resources: session.draft.resources, volumes: session.draft.volumes, connections: [] },
        } as unknown as Stack)
        : null,
    [isNewStack, draftName, draftLabels, session.draft.resources, session.draft.volumes],
  );
  const effectiveStack = draftStackView ?? savedStack;

  const orgDomains = useOrgDomains(effectiveStack?.organisation_id ?? getCurrentOrganizationId() ?? undefined);

  // ── Release plumbing (needed this early: the diff baseline is pinned to the
  // latest release's snapshot, not the autosaved server state) ──
  const deployIds = useMemo(() => ({
    orgId: savedStack?.organisation_id || getCurrentOrganizationId() || "",
    projectName: (savedStack ? projectNameById(savedStack.project_id) : "") || defaultProjectName || "",
    stackId: savedStack?.id || "",
  }), [savedStack, projectNameById, defaultProjectName]);
  const idsReady = !!deployIds.stackId && !!deployIds.projectName;
  const releasesResult = useReleases({ ...deployIds, enabled: idsReady });
  const releaseDetail = useReleaseDetail(deployIds.orgId, deployIds.projectName, deployIds.stackId);

  const { baselineReleaseId, deployedSnapshot, convergedReleaseDetail, statusLiveStatus } =
    useReleaseAnchors({
      stack: savedStack ?? null,
      releases: releasesResult.releases,
      activeRelease: releasesResult.activeRelease,
      releaseDetail,
    });

  // Publicly exposed services → best live ingress URL, for the header's
  // PUBLIC row. Drafts have no live ingress, so the row stays empty. Each chip's
  // dot carries its OWN resource's rollout state (same status release as the
  // canvas node dots), not the stack-level rollup — one crashed service must not
  // paint every endpoint red.
  const publicEndpoints = useMemo(() => {
    if (isNewStack) return [];
    const liveResources = convergedReleaseDetail?.live_status?.resources ?? {};
    const statusResources = statusLiveStatus?.resources ?? {};
    return (effectiveStack?.spec.stack_resources ?? []).flatMap((r) => {
      const ingress = r.name ? liveResources[r.name]?.public_ingress ?? [] : [];
      const best = pickBestIngress(ingress, orgDomains);
      if (!best || !r.name) return [];
      const variant = statusVariant("rollout", statusResources[r.name]?.state);
      return [{ service: r.name, url: best.url, port: best.target_port, variant }];
    });
  }, [isNewStack, effectiveStack, orgDomains, convergedReleaseDetail, statusLiveStatus]);

  // Current server state as form data — what the canvas displays and the edit
  // session's working draft seeds from.
  const serverResources = useMemo<FormStackResourceData[]>(
    () => formResourcesFromSpec(savedStack?.spec?.stack_resources, savedStack?.spec?.connections),
    [savedStack],
  );
  const serverVolumes = useMemo<VolumeFormData[]>(
    () => (savedStack?.spec?.volumes || []).map(mapVolumeToFormData),
    [savedStack],
  );

  // Server-computed outputs (host, port/url or port.<name>/url.<name>,
  // public_host/public_url or public_host.<name>/public_url.<name>) keyed by
  // resource name. The working draft never computes outputs, so a resource added
  // on the canvas carries none until it is saved. The env-var OUTPUT pickers read
  // from this server-truth map (matched by name) instead of the draft copy, and
  // it refreshes with savedStack after every autosave — no page reload needed.
  const serverOutputsByName = useMemo<Map<string, string[]>>(
    () =>
      new Map(
        (savedStack?.spec?.stack_resources ?? [])
          .filter((r): r is StackResource & { name: string } => !!r.name)
          .map((r) => [r.name, (r.outputs ?? []).map((o) => o.name)] as [string, string[]]),
      ),
    [savedStack?.spec?.stack_resources],
  );

  // Diff baseline: the deployed snapshot when one exists, so autosaved edits stay
  // visibly dirty/revertable until deployed. Never-deployed stacks fall back to
  // the server state (everything reads as staged for the first deploy).
  // The snapshot stores the RESOLVED git revision (branch/commit written by the
  // pin resolver at deploy time). When the saved spec doesn't pin one, those are
  // deploy-time facts, not config drift — strip them so the baseline compares
  // intent with intent instead of reading every unpinned git resource as dirty.
  const snapshotResources = useMemo<FormStackResourceData[] | null>(() => {
    if (!deployedSnapshot) return null;
    return formResourcesFromSpec(
      stripUnpinnedGitRevisions(
        (deployedSnapshot.resources ?? []) as StackResource[],
        savedStack?.spec?.stack_resources ?? [],
      ),
      deployedSnapshot.connections as StackConnection[] | undefined,
    );
  }, [deployedSnapshot, savedStack?.spec?.stack_resources]);
  const snapshotVolumes = useMemo<VolumeFormData[] | null>(
    () => (deployedSnapshot ? ((deployedSnapshot.volumes ?? []) as Volume[]).map(mapVolumeToFormData) : null),
    [deployedSnapshot],
  );

  // All diffing downstream is positional, but the server returns resources in
  // unstable order and the snapshot's order need not match — re-key the baseline
  // onto the order of whatever the diffs actually run against: the live session
  // draft when one is active, the server state otherwise.
  const alignResources = session.isActive ? session.draft.resources : serverResources;
  const alignVolumes = session.isActive ? session.draft.volumes : serverVolumes;
  const baselineResources = useMemo<FormStackResourceData[]>(
    () =>
      (snapshotResources
        ? alignBaselineToDraft(snapshotResources, alignResources, renameFingerprint)
        : alignBaselineToDraft(serverResources, alignResources, renameFingerprint)) as FormStackResourceData[],
    [snapshotResources, serverResources, alignResources],
  );
  const baselineVolumes = useMemo<VolumeFormData[]>(
    () =>
      (snapshotVolumes
        ? alignBaselineToDraft(snapshotVolumes, alignVolumes, renameFingerprint)
        : alignBaselineToDraft(serverVolumes, alignVolumes, renameFingerprint)) as VolumeFormData[],
    [snapshotVolumes, serverVolumes, alignVolumes],
  );

  const draftSeeded = useRef(false);
  useEffect(() => {
    if (!isNewStack || draftSeeded.current) return;
    draftSeeded.current = true;
    // Baseline empty so seeded resources/volumes read as "added" and Save is enabled.
    session.start({ resources: [], volumes: [] }, { linkedAddonIds: new Set(seed.linkedAddonIds) });
    if (seed.resources.length) session.updateResources(() => seed.resources);
    if (seed.volumes.length) session.updateVolumes(() => seed.volumes);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isNewStack]);

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
  // addonId → live state, for canvas addon node dots.
  const addonStateById = useMemo(
    () =>
      new Map(
        allAddons
          .filter((a: PostgresAddon) => a.id)
          .map((a: PostgresAddon) => [a.id!, a.status?.state ?? ""] as [string, string]),
      ),
    [allAddons],
  );

  const connectionAddonIds = useMemo<Set<string>>(
    () =>
      new Set(
        (savedStack?.spec?.connections ?? [])
          .filter((c) => c.from?.type === "addon/postgres" && c.from?.id)
          .map((c) => c.from!.id as string),
      ),
    [savedStack],
  );

  // Autosave model: the canvas is always editable for writers. The session
  // starts as soon as the stack is loaded and restarts after discard/revert.
  // Baseline = deployed snapshot (when loaded), draft = current server state:
  // they differ when the server already holds autosaved-but-undeployed edits.
  useEffect(() => {
    if (isNewStack || !savedStack || !canWriteStack || session.isActive) return;
    session.start(
      { resources: baselineResources, volumes: baselineVolumes },
      {
        linkedAddonIds: connectionAddonIds,
        draft: { resources: serverResources, volumes: serverVolumes },
      },
    );
    // session.start is a stable useCallback; session.isActive is the only reactive
    // field we need from the session object itself.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isNewStack, savedStack, canWriteStack, session.isActive, baselineResources, baselineVolumes, serverResources, serverVolumes, connectionAddonIds]);

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

  // Server topology may have changed; re-derive edges alongside the fresh stack.
  const applyFreshStack = useCallback((fresh: Stack) => {
    setFetchedStack(fresh);
    // Context write-through: stale currentStack must not win after a remote refresh.
    setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
    setTopologyRefreshKey((k) => k + 1);
  }, [setStacks]);

  // Volumes that exist server-side; their spec (PVC size) is immutable, so the
  // drawer renders those fields read-only.
  const persistedVolumeNames = useMemo(
    () =>
      new Set(
        (savedStack?.spec?.volumes ?? [])
          .map((v) => v.name)
          .filter((n): n is string => !!n),
      ),
    [savedStack?.spec?.volumes],
  );

  // Backend structured field errors (from a failed autosave or draft deploy),
  // mapped onto the editor's inner field keys so the offending inputs show them
  // inline. Keyed by resource NAME → field key → message (NOT by index): the
  // resource may be reordered/removed before the user fixes it, so the current
  // index is resolved at render time when merging into the index-keyed errors prop.
  const [serverFieldErrors, setServerFieldErrors] = useState<{
    [resourceName: string]: { [fieldKey: string]: string };
  }>({});
  // Raw structured errors from the last draft-deploy attempt, used to render the
  // summary banner. Cleared when a deploy is retried or the banner is dismissed.
  const [deployFieldErrors, setDeployFieldErrors] = useState<ParsedFieldError[]>([]);
  const [bannerDismissed, setBannerDismissed] = useState(false);
  // Bumped to ask the canvas to open a resource drawer (banner "jump to error").
  const [openResourceSignal, setOpenResourceSignal] = useState<
    { index: number; tab: EditSessionTab; nonce: number } | null
  >(null);

  // Dedupe for autosave-failure toasts: the backoff retry loop and re-edits after
  // a terminal 400 re-run the same ops and would re-toast the same reason every
  // cycle. Remember the last-shown message and only toast when it changes; cleared
  // on a successful sync so a later distinct failure still toasts.
  const lastSyncErrorMsgRef = useRef<string | null>(null);

  // Autosave engine: debounces draft changes and syncs thin per-resource ops to
  // the server. Disabled for drafts (nothing exists server-side to sync yet).
  const draftSync = useDraftSync({
    enabled: !isNewStack && canWriteStack,
    stack: savedStack ?? undefined,
    session,
    ids: idsReady ? deployIds : null,
    onStackRefreshed: (fresh) => applyFreshStack(fresh),
    onSyncError: ({ message, op, fieldErrors }) => {
      // No op → the save actually landed and only the post-save refetch failed;
      // don't alarm the user with a "Save failed" toast or field error.
      if (!op) return;
      // Dedupe the toast: skip if this exact reason is already on screen.
      if (lastSyncErrorMsgRef.current !== message) {
        lastSyncErrorMsgRef.current = message;
        toast({ title: "Save failed", description: message, variant: "destructive" });
      }
      // Only resource create/update ops carry mappable field errors.
      if (op.kind !== "createResource" && op.kind !== "updateResource") return;
      const resourceName = op.resource.name;
      if (!resourceName || fieldErrors.length === 0) return;
      // Thin op → all errors belong to this one resource; map to editor field keys
      // and store by resource NAME so a reorder/removal can't misattach them.
      const fields = mapFieldErrors(fieldErrors, { dialect: "thin", resourceIndex: 0 }).resources[0];
      if (!fields) return;
      setServerFieldErrors((prev) => ({
        ...prev,
        [resourceName]: { ...(prev[resourceName] ?? {}), ...fields },
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

  // Merge live zod validation (index-keyed) with backend field errors (name-keyed).
  // Resolve each server error's resource NAME to its CURRENT index here so the
  // merged map stays index-keyed for consumers but survives a reorder/removal
  // since the error was captured.
  const mergedResourceErrors = useMemo(() => {
    const merged: { [index: number]: { [field: string]: string | undefined } } = {};
    for (const [k, v] of Object.entries(validationErrors.resources)) {
      merged[Number(k)] = { ...v };
    }
    for (const [resourceName, fields] of Object.entries(serverFieldErrors)) {
      const idx = session.draft.resources.findIndex((r) => r.name === resourceName);
      if (idx < 0) continue;
      merged[idx] = { ...(merged[idx] ?? {}), ...fields };
    }
    return merged;
  }, [validationErrors.resources, serverFieldErrors, session.draft.resources]);

  const bannerItems = useMemo<ValidationBannerItem[]>(
    () => buildBannerItems(deployFieldErrors, session.draft.resources as ReadonlyArray<{ name?: string }>),
    [deployFieldErrors, session.draft.resources],
  );

  const lifecycle = useDeployLifecycle({
    stack: savedStack ?? undefined,
    // "Editing" = autosave hasn't landed yet (in flight or retrying). Saved-but-
    // undeployed content is detected by the staged-phase content diff instead.
    unsaved: draftSync.status === SYNC_STATUS.saving || draftSync.failureCount > 0,
    releases: releasesResult.releases,
    activeRelease: releasesResult.activeRelease,
    detail: releaseDetail,
  });

  // When a release settles into ANY terminal state, the stack's converged_release /
  // latest_release summaries are stale until refetched — the staged panel would
  // keep diffing against the old snapshot otherwise. Refetch once per transition,
  // keyed on the polled releases list (not on the stack's own pointer).
  const refetchStack = useCallback(() => {
    if (!deployIds.stackId) return;
    void getStackById(deployIds.orgId, deployIds.projectName, deployIds.stackId).then((fresh) => {
      applyFreshStack(fresh);
    }).catch(() => {
      // Transient fetch failure; the next terminal transition retries.
    });
  }, [deployIds, applyFreshStack]);
  const activeRelease = releasesResult.activeRelease;
  const prevActiveReleaseRef = useRef<{ id?: string; state?: string } | undefined>(
    activeRelease && { id: activeRelease.id, state: activeRelease.state },
  );
  useEffect(() => {
    const prev = prevActiveReleaseRef.current;
    prevActiveReleaseRef.current = activeRelease && { id: activeRelease.id, state: activeRelease.state };
    if (shouldRefetchStackSummaries(prev, activeRelease, savedStack?.latest_release)) {
      refetchStack();
    }
  }, [activeRelease, savedStack?.latest_release, refetchStack]);

  // Live snapshot: already lazily fetched above (convergedReleaseId ensure); peek
  // here to gate canDiscardDraft and pass to the revert hook.
  const liveSnapshot = convergedReleaseDetail?.snapshot;

  const [revertConfirmOpen, setRevertConfirmOpen] = useState(false);
  const [viewChangesOpen, setViewChangesOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const stackRevert = useStackRevert({
    ids: idsReady ? deployIds : null,
    stack: savedStack ?? undefined,
    liveSnapshot,
    onReverted: (fresh) => {
      setFetchedStack(fresh);
      setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
      draftSync.notifyExternalUpdate(fresh);
      session.discard(); // auto-start effect restarts the session on the reverted baseline
      toast({ title: "Draft discarded", description: "Stack restored to the last deployment.", variant: "success" });
    },
    onError: (message) => {
      // Revert isn't atomic (apply + volume deletes), so warn about partial state
      // alongside the backend reason.
      toast({
        title: "Discard failed",
        description: `${message} The stack may be partially reverted; reload to see its current state.`,
        variant: "destructive",
      });
    },
  });

  // Immediate, confirm-gated volume deletion (canvas). Only wired for saved
  // stacks — the wizard (`/stacks/new`) has nothing server-side to delete yet.
  const volumeDelete = useVolumeDelete({
    ids: idsReady ? deployIds : null,
    draftSync,
    onServerRefresh: applyFreshStack,
    onRestoreVolume: (vol) => {
      session.updateVolumes((vs) => [...vs, mapVolumeToFormData(vol)]);
    },
    toast,
  });

  // Shared validation-failure handler for release actions (draft deploy, deploy,
  // cancel, rollback). Rich (fieldErrors-bearing) 400s paint inline field errors
  // and the summary banner instead of a generic toast. Returns true when the
  // error was consumed this way, so the caller can skip its own fallback toast.
  const applyValidationFailure = useCallback((err: unknown): boolean => {
    const parsed = parseApiError(err);
    if (parsed.fieldErrors.length === 0) return false;
    const mapped = mapFieldErrors(parsed.fieldErrors, { dialect: "fat" });
    if (mapped.stackName) setNameError(mapped.stackName);
    setServerFieldErrors((prev) => {
      const next = { ...prev };
      for (const [idxStr, fields] of Object.entries(mapped.resources)) {
        const nm = session.draft.resources[Number(idxStr)]?.name;
        if (nm) next[nm] = { ...(next[nm] ?? {}), ...fields };
      }
      return next;
    });
    setDeployFieldErrors(parsed.fieldErrors);
    setBannerDismissed(false);
    toast({
      title: "Deploy failed",
      description: `${parsed.fieldErrors.length} validation ${parsed.fieldErrors.length === 1 ? "error" : "errors"}`,
      variant: "destructive",
    });
    return true;
  }, [session.draft.resources, toast]);

  const [deployBusy, setDeployBusy] = useState(false);
  const refetchReleases = releasesResult.refetch;
  const runDeploy = useCallback(async (fn: () => Promise<unknown>, ok: string) => {
    setDeployBusy(true);
    try {
      await fn();
      toast({ title: ok, variant: "success" });
      refetchReleases();
    } catch (e) {
      if (!applyValidationFailure(e)) {
        toast({ title: "Action failed", description: e instanceof Error ? e.message : "", variant: "destructive" });
      }
    } finally {
      setDeployBusy(false);
    }
  }, [toast, refetchReleases, applyValidationFailure]);

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
    runDeploy(() => createRelease(deployIds.orgId, deployIds.projectName, deployIds.stackId), "Deploy started");
  }, [draftSync, deployIds, runDeploy, toast]);
  const onCancelDeploy = useCallback(
    (releaseId: string) => runDeploy(() => cancelRelease(deployIds.orgId, deployIds.projectName, deployIds.stackId, releaseId), "Release cancelled"),
    [runDeploy, deployIds],
  );
  const onRollback = useCallback(
    (releaseId: string) => runDeploy(() => rollbackRelease(deployIds.orgId, deployIds.projectName, deployIds.stackId, releaseId), "Rollback started"),
    [runDeploy, deployIds],
  );
  const onCopyId = useCallback((releaseId: string) => {
    void navigator.clipboard?.writeText(releaseId);
    toast({ title: "Release ID copied" });
  }, [toast]);

  // Draft deploy: validates name, creates the stack, starts the first release,
  // and navigates to the new page. There is no separate "create" step — the
  // draft stays local until the user deploys.
  const performDraftDeploy = async () => {
    if (!isNewStack) return;
    setDraftDeploying(true);
    setNameError(undefined);
    // Clear stale validation state from a previous failed attempt.
    setDeployFieldErrors([]);
    setServerFieldErrors({});

    // A draft needs a name before it can be created. The stack name field is not
    // min-length-constrained in the schema (empty passes zod and only fails at
    // the API), so guard it here to surface the error inline on the title input.
    if (!draftName.trim()) {
      setNameError("Required");
      setDraftDeploying(false);
      toast({
        title: "Name your stack",
        description: "Give the stack a name before deploying.",
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
        const { nameError: newNameError, messages } = formatDraftValidationIssues(
          validation.error.issues,
          resources,
          formStackData.spec.volumes as { name?: string }[] | undefined,
        );
        setNameError(newNameError);
        toast({
          title: "Validation error",
          description: messages.length > 0 ? messages.join("; ") : "Please fix the highlighted errors before saving.",
          variant: "destructive",
        });
        setDraftDeploying(false);
        return;
      }

      const projectName = defaultProjectName;
      if (!projectName) {
        toast({
          title: "No project available",
          description: "Could not resolve a project to save into.",
          variant: "destructive",
        });
        setDraftDeploying(false);
        return;
      }

      const apiData = convertFormStackToApiStack(formStackData);
      // Single atomic upsert: the name-addressed apply validates the full
      // document server-side and creates the stack + children in one
      // transaction, so a validation failure persists nothing and retrying
      // after a fix cannot 409 on an orphaned shell.
      const created = await applyStackByName(orgId, projectName, apiData);
      if (!created.id) {
        throw new Error("Created stack is missing an id");
      }
      session.discard();
      // The stack exists from here on — the show page is the source of truth,
      // so navigate regardless of whether the first release starts cleanly.
      try {
        await createRelease(orgId, projectName, created.id);
        toast({ title: "Deploy started", variant: "success" });
      } catch (releaseErr) {
        toast({
          title: "Stack created, but deploy failed",
          description: getErrorMessage(releaseErr),
          variant: "destructive",
        });
      }
      navigate(`/stacks/${created.id}`, { replace: true, state: null });
    } catch (err) {
      console.error('Failed to create stack:', err);
      if (!applyValidationFailure(err)) {
        toast({
          title: "Failed to create stack",
          description: parseApiError(err).topLevel,
          variant: "destructive",
        });
      }
    } finally {
      setDraftDeploying(false);
    }
  };

  const performDelete = useCallback(async () => {
    if (!deployIds) return;
    setDeleting(true);
    try {
      await deleteStack(deployIds.orgId, deployIds.projectName, deployIds.stackId);
      setStacks((prev) => prev.filter((s) => s.id !== deployIds.stackId));
      toast({ title: "Stack deleted", description: `"${savedStack?.name}" was deleted.`, variant: "success" });
      navigate("/stacks");
    } catch {
      toast({ title: "Delete failed", description: "The stack could not be deleted.", variant: "destructive" });
    } finally {
      setDeleting(false);
      setDeleteConfirmOpen(false);
    }
  }, [deployIds, setStacks, savedStack?.name, toast, navigate]);

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

  if (!isNewStack && loading) {
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

  if (!isNewStack && !savedStack) {
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

  const headerHealth = effectiveStack ? deriveHeaderHealth(effectiveStack) : undefined;
  const showDeployFailedHint = !!effectiveStack && latestDeployFailed(effectiveStack);

  const resourceCount = effectiveStack?.spec?.stack_resources?.length || 0;
  const volumeCount = effectiveStack?.spec?.volumes?.length || 0;
  const addonCount = connectionAddonIds.size;
  const subtitleText = [
    `${resourceCount} ${resourceCount === 1 ? "resource" : "resources"}`,
    `${volumeCount} ${volumeCount === 1 ? "volume" : "volumes"}`,
    ...(addonCount > 0 ? [`${addonCount} ${addonCount === 1 ? "addon" : "addons"}`] : []),
  ].join(" · ");

  // Ops-view bodies — rendered inside the canvas shell; gated on isNewStack so
  // the user sees a placeholder until the stack is saved for the first time.
  const deploymentsBody = isNewStack ? <DraftTabPlaceholder label="Deployments" /> : effectiveStack?.id ? (
    <DeploymentsTab
      orgId={deployIds.orgId}
      projectName={deployIds.projectName}
      stackId={effectiveStack.id}
      stack={effectiveStack}
      onJumpToResource={(resourceName, tab) => {
        // Resolve against the list the canvas drawer actually indexes into
        // (the live draft while a session is active), at click time — the
        // banner's render-time list may have drifted.
        const index = jumpTargetIndex(
          resourceName,
          session.isActive ? session.draft.resources : serverResources,
        );
        if (index === undefined) return;
        setActiveTab(EDITOR_TABS.architecture);
        setOpenResourceSignal({ index, tab, nonce: Date.now() });
      }}
      refetchReleases={refetchReleases}
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

  const logsBody = isNewStack ? <DraftTabPlaceholder label="Logs" /> : effectiveStack?.id ? (
    <LogsTab
      stackId={effectiveStack.id}
      organizationId={effectiveStack.organisation_id || getCurrentOrganizationId() || ''}
      resources={effectiveStack.spec.stack_resources?.map(r => ({ name: r.name || '', id: r.id || '' })) || []}
      initialSources={logsInitialSource ? [logsInitialSource] : undefined}
    />
  ) : (
    <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
  );

  const metricsBody = isNewStack ? <DraftTabPlaceholder label="Metrics" /> : effectiveStack?.id ? (
    <MetricsTab
      stackId={effectiveStack.id}
      organizationId={effectiveStack.organisation_id || getCurrentOrganizationId() || ''}
      resources={effectiveStack.spec.stack_resources || []}
    />
  ) : (
    <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
  );

  const dirtyTotal =
    session.dirty.dirtyResourceIdx.size + session.dirty.dirtyVolumeIdx.size + session.dirty.addonLinkCount;

  const changeCount = resolveChangeCount(lifecycle.stagedDiff, dirtyTotal, isNewStack ? SYNC_STATUS.idle : draftSync.status);

  return (
    <ReleaseDetailProvider value={releaseDetail}>
      <CanvasEditorShell
        stackName={isNewStack ? draftName : (effectiveStack?.name ?? "")}
        stackId={effectiveStack?.id}
        isNewStack={isNewStack}
        nameEditable={isNewStack}
        onNameChange={handleNameChange}
        nameError={nameError}
        headerHealth={headerHealth}
        latestDeployFailed={showDeployFailedHint}
        lifecycle={effectiveStack?.lifecycle}
        subtitle={subtitleText}
        activeTab={activeTab}
        onTabChange={(tab) => {
          setLogsInitialSource(undefined);
          setActiveTab(tab);
        }}
        isActive={session.isActive}
        dirtyResourceCount={session.dirty.dirtyResourceIdx.size}
        dirtyTotal={changeCount}
        isStaged={lifecycle.phase === "staged"}
        onViewChanges={() => setViewChangesOpen(true)}
        syncStatus={isNewStack ? SYNC_STATUS.idle : draftSync.status}
        deployBusy={deployBusy}
        canWrite={canWriteStack}
        hasResources={(session.isActive ? session.draft.resources : serverResources).length > 0}
        onDraftDeploy={() => void performDraftDeploy()}
        draftDeploying={draftDeploying}
        onDeploy={onDeploy}
        canDiscardDraft={lifecycle.phase === "staged" && !!liveSnapshot && canWriteStack}
        onDiscardDraft={() => setRevertConfirmOpen(true)}
        canDeleteStack={canWriteStack}
        onDelete={() => setDeleteConfirmOpen(true)}
        publicEndpoints={publicEndpoints}
        architecture={
          <>
            {!bannerDismissed && bannerItems.length > 0 && (
              <div className="px-4 pt-3">
                <ValidationBanner
                  items={bannerItems}
                  onJump={(index, tab) => setOpenResourceSignal({ index, tab, nonce: Date.now() })}
                  onDismiss={() => setBannerDismissed(true)}
                />
              </div>
            )}
            <ArchitectureTab
              session={session}
              openResourceSignal={openResourceSignal}
              baselineResources={baselineResources}
              baselineVolumes={baselineVolumes}
              draftResources={serverResources}
              draftVolumes={serverVolumes}
              serverOutputsByName={serverOutputsByName}
              connectionAddonIds={connectionAddonIds}
              addonNameById={addonNameById}
              addonStateById={addonStateById}
              errors={mergedResourceErrors}
              onViewLogs={(resourceName) => {
                setLogsInitialSource(resourceName);
                setActiveTab(EDITOR_TABS.logs);
              }}
              topologyIds={!isNewStack && idsReady ? deployIds : null}
              topologyRefreshKey={topologyRefreshKey}
              onDeleteVolume={idsReady ? volumeDelete.deleteVolume : undefined}
              deletingVolume={volumeDelete.deleting}
              persistedVolumeNames={persistedVolumeNames}
              releaseInFlight={deployBusy || lifecycle.phase === "deploying"}
              liveStatusResources={statusLiveStatus?.resources}
            />
          </>
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
                  // Failure surfaces via the hook's onError toast.
                  await stackRevert.revert();
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
            This permanently deletes "{savedStack?.name}", its resources, volumes and deployments. This cannot be undone.
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
    </ReleaseDetailProvider>
  );
}
