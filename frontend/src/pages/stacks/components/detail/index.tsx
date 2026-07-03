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
import { DraftTabPlaceholder } from "@/pages/stacks/components/canvas/DraftTabPlaceholder";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import { createStack, getStackById, deleteStack, updateStack } from "@/api/stacks";
import { emptyDraftSeed, buildDraftFormData, type DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import { USER_DEFINED_LABEL_KEY } from "@/pages/stacks/lib/constants";
import { createRelease, cancelRelease, rollbackRelease } from "@/api/releases";
import { useReleases } from "@/pages/stacks/components/detail/deployments/use-releases";
import { useReleaseDetail } from "@/pages/stacks/components/detail/deployments/use-release-detail";
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
import { buildDesiredState } from "@/pages/stacks/lib/draft-sync/desired-state";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";
import type { SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";
import { stackToUpdateRequest } from "@/pages/stacks/lib/draft-sync/snapshot-to-update";
import { normalizeLabel } from "@/pages/stacks/lib/labels";
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

// Helper to map API build_spec to form schema shape

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
  const [draftLabels, setDraftLabels] = useState<FormStackData["labels"]>(seed.labels);
  const [labelSync, setLabelSync] = useState<SyncStatus>(SYNC_STATUS.idle);

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

  const baselineResources = useMemo<FormStackResourceData[]>(() => {
    const connections = stackToShow?.spec?.connections ?? [];
    return (stackToShow?.spec?.stack_resources || []).map((r) => {
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
  }, [stackToShow]);
  const baselineVolumes = useMemo<VolumeFormData[]>(
    () => (stackToShow?.spec?.volumes || []).map(mapVolumeToFormData),
    [stackToShow],
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
  useEffect(() => {
    if (isDraft || !stackToShow || !canWriteStack || session.isActive) return;
    session.start(
      { resources: baselineResources, volumes: baselineVolumes },
      { linkedAddonIds: connectionAddonIds },
    );
    // session.start is a stable useCallback; session.isActive is the only reactive
    // field we need from the session object itself.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isDraft, stackToShow, canWriteStack, session.isActive, baselineResources, baselineVolumes, connectionAddonIds]);

  // ── Deploy lifecycle (page-level: drives the status bar across all tabs) ──
  const deployIds = useMemo(() => ({
    orgId: stackToShow?.organisation_id || getCurrentOrganizationId() || "",
    teamName: (stackToShow ? teamNameById(stackToShow.team_id) : "") || defaultTeamName || "",
    stackId: stackToShow?.id || "",
  }), [stackToShow, teamNameById, defaultTeamName]);

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
      setStacks(stacks.map((s) => (s.id === fresh.id ? fresh : s)));
    },
  });

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

  const releasesResult = useReleases({ ...deployIds, enabled: !!deployIds.stackId });
  const releaseDetail = useReleaseDetail(deployIds.orgId, deployIds.teamName, deployIds.stackId);
  const lifecycle = useDeployLifecycle({
    stack: stackToShow ?? undefined,
    dirty: session.dirty,
    isActive: session.isActive,
    releases: releasesResult.releases,
    activeRelease: releasesResult.activeRelease,
    detail: releaseDetail,
  });

  // Live snapshot: already lazily fetched by useDeployLifecycle via detail.ensure;
  // peek here to gate canDiscardDraft and pass to the revert hook.
  const liveReleaseId = stackToShow?.status?.last_converged?.release_id;
  const liveSnapshot = releaseDetail.peek(liveReleaseId).data?.snapshot;

  const [revertConfirmOpen, setRevertConfirmOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const stackRevert = useStackRevert({
    ids: deployIds.stackId ? deployIds : null,
    stack: stackToShow ?? undefined,
    liveSnapshot,
    onReverted: (fresh) => {
      setFetchedStack(fresh);
      setStacks(stacks.map((s) => (s.id === fresh.id ? fresh : s)));
      draftSync.notifyExternalUpdate(fresh);
      session.discard(); // auto-start effect restarts the session on the reverted baseline
      toast({ title: "Draft discarded", description: "Stack restored to the last deployment." });
    },
  });

  const [deployBusy, setDeployBusy] = useState(false);
  const refetchReleases = releasesResult.refetch;
  const runDeploy = useCallback(async (fn: () => Promise<unknown>, ok: string) => {
    setDeployBusy(true);
    try {
      await fn();
      toast({ title: ok });
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
          const [scope0] = issue.path;
          if (scope0 === "name") {
            newNameError = issue.message;
          } else {
            topLevelMessages.push(issue.message);
          }
        }

        setNameError(newNameError);
        toast({
          title: "Validation error",
          description: topLevelMessages.length > 0
            ? topLevelMessages.join("; ")
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
      setStacks(stacks.filter((s) => s.id !== deployIds.stackId));
      toast({ title: "Stack deleted", description: `"${stackToShow?.name}" was deleted.` });
      navigate("/stacks");
    } catch {
      toast({ title: "Delete failed", description: "The stack could not be deleted.", variant: "destructive" });
    } finally {
      setDeleting(false);
      setDeleteConfirmOpen(false);
    }
  }, [deployIds, setStacks, stackToShow?.name, toast, navigate, stacks]);

  const handleNameChange = useCallback((name: string) => {
    setDraftName(name);
    setNameError(undefined);
  }, []);

  const addDraftLabel = useCallback((value: string) => {
    const normalized = normalizeLabel(value);
    if (!normalized) return;
    setDraftLabels((prev) => {
      const cur = prev ?? [];
      if (cur.some((l) => l.value === normalized)) return cur;
      return [...cur, { key: USER_DEFINED_LABEL_KEY, value: normalized }];
    });
  }, []);
  const removeDraftLabel = useCallback((idx: number) => {
    setDraftLabels((prev) => (prev ?? []).filter((_, i) => i !== idx));
  }, []);

  // Deployed stacks: persist the new label set immediately via a full PUT
  // (replace-all body built from the live server stack, labels swapped).
  // The body snapshots stackToShow at call time: an autosave op in flight can
  // land after this PUT and vice versa — bounded clobber window accepted by
  // the spec (labels never touch resources; both writers are last-write-wins).
  const persistLabels = useCallback(
    async (next: NonNullable<Stack["labels"]>) => {
      if (!stackToShow?.id || !deployIds.stackId) return;
      setLabelSync(SYNC_STATUS.saving);
      try {
        const fresh = await updateStack(
          deployIds.orgId,
          deployIds.teamName,
          deployIds.stackId,
          stackToUpdateRequest(stackToShow, next),
        );
        setFetchedStack(fresh);
        setStacks(stacks.map((s) => (s.id === fresh.id ? fresh : s)));
        setLabelSync(SYNC_STATUS.saved);
        setTimeout(() => setLabelSync(SYNC_STATUS.idle), 2000);
      } catch {
        setLabelSync(SYNC_STATUS.error);
      }
    },
    [stackToShow, deployIds, setStacks, stacks],
  );

  const addStackLabel = useCallback(
    (value: string) => {
      const normalized = normalizeLabel(value);
      if (!normalized) return;
      const cur = stackToShow?.labels ?? [];
      if (cur.some((l) => l.value === normalized)) return;
      void persistLabels([...cur, { key: USER_DEFINED_LABEL_KEY, value: normalized }]);
    },
    [stackToShow, persistLabels],
  );
  const removeStackLabel = useCallback(
    (idx: number) => {
      const cur = stackToShow?.labels ?? [];
      void persistLabels(cur.filter((_, i) => i !== idx));
    },
    [stackToShow, persistLabels],
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
  const subtitleText = `${resourceCount} ${resourceCount === 1 ? "service" : "services"} · ${volumeCount} ${
    volumeCount === 1 ? "volume" : "volumes"
  }`;

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

  return (
    <>
      <CanvasEditorShell
        stackName={isDraft ? draftName : (effectiveStack?.name ?? "")}
        stackId={effectiveStack?.id}
        isDraft={isDraft}
        nameEditable={isDraft}
        onNameChange={handleNameChange}
        nameError={nameError}
        labels={(isDraft ? draftLabels : effectiveStack?.labels) ?? []}
        labelsEditable={isDraft || canWriteStack}
        onAddLabel={isDraft ? addDraftLabel : addStackLabel}
        onRemoveLabel={isDraft ? removeDraftLabel : removeStackLabel}
        statusState={effectiveStack?.status?.state}
        subtitle={subtitleText}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        isActive={session.isActive}
        dirtyResourceCount={session.dirty.dirtyResourceIdx.size}
        dirtyTotal={dirtyTotal}
        isStaged={lifecycle.phase === "staged"}
        syncStatus={isDraft ? SYNC_STATUS.idle : labelSync !== SYNC_STATUS.idle ? labelSync : draftSync.status}
        deployBusy={deployBusy}
        canWrite={canWriteStack}
        onCreate={() => void performCreate()}
        isCreating={isCreating}
        onDeploy={onDeploy}
        onDiscardAll={() => session.discard()}
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
            connectionAddonIds={connectionAddonIds}
            addonNameById={addonNameById}
            errors={validationErrors.resources}
            onViewLogs={() => setActiveTab("logs")}
          />
        }
        deployments={deploymentsBody}
        logs={logsBody}
        metrics={metricsBody}
      />
      <AlertDialog open={revertConfirmOpen} onOpenChange={setRevertConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard draft changes?</AlertDialogTitle>
            <AlertDialogDescription>
            This restores the stack to its last deployment. Volumes added since then are deleted — their data is destroyed. This cannot be undone.
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
