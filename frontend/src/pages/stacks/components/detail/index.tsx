import { useParams, useLocation, useNavigate, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import { useMemo, useState, useEffect, useCallback, useRef } from "react";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import { useStackEditSession, type EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
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
import { StackLogsTab } from "@/pages/stacks/components/detail/logs/stack-logs-tab";
import { StackMetricsTab } from "@/pages/stacks/components/detail/metrics/stack-metrics-tab";
import { DeploymentsTab } from "@/pages/stacks/components/detail/deployments/deployments-tab";
import { StackCanvasTab } from "@/pages/stacks/components/canvas/StackCanvasTab";
import { CanvasEditorShell } from "@/pages/stacks/components/canvas/CanvasEditorShell";
import { DraftTabPlaceholder } from "@/pages/stacks/components/canvas/DraftTabPlaceholder";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import { createStack, getStackById, updateStack } from "@/api/stacks";
import { emptyDraftSeed, buildDraftFormData, type DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
import { createRelease, cancelRelease, rollbackRelease } from "@/api/releases";
import { useReleases } from "@/pages/stacks/components/detail/deployments/use-releases";
import { useReleaseDetail } from "@/pages/stacks/components/detail/deployments/use-release-detail";
import { useDeployLifecycle } from "@/pages/stacks/components/detail/deployments/use-deploy-lifecycle";
import {
  connectionsToEnvRows,
} from "@/pages/stacks/lib/connection-mapping";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { useCurrentUser } from "@/hooks/use-current-user";
import type { z } from "zod";
import { convertApiResourceToFormResource, convertApiVolumeToFormVolume, convertFormStackToApiStack, FormStackSchema } from "@/pages/stacks/schemas/form-schema";
import { useToast } from "@/components/ui/use-toast";
import type { ApiStackResourceSchema, ApiVolumeSchema } from "@/pages/stacks/schemas/api-schema";

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

  const { stacks } = useStacks();
  const [fetchedStack, setFetchedStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const session = useStackEditSession();
  const [activeTab, setActiveTab] = useState("configuration");
  const [isSaving, setIsSaving] = useState(false);
  const [detachConfirmOpen, setDetachConfirmOpen] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{
    resources: { [index: number]: { [field: string]: string | undefined } };
    volumes: { [index: number]: { [field: string]: string | undefined } };
  }>({ resources: {}, volumes: {} });

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

  const baselineResources = useMemo<FormStackResourceData[]>(() => {
    const connections = stackToShow?.spec?.connections ?? [];
    return (stackToShow?.spec?.stack_resources || []).map((r) => {
      const form = mapStackResourceToFormData(r);
      const connRows = connectionsToEnvRows(form.name ?? "", connections) as FormEnvVarData[];
      if (connRows.length === 0) return form;
      return {
        ...form,
        execution_config: {
          ...(form.execution_config ?? {}),
          environment_variables: [
            ...((form.execution_config?.environment_variables ?? []) as FormEnvVarData[]),
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

  const activateEdit = (opts?: { resourceIdx?: number; volumeIdx?: number; openTab?: EditSessionTab }) => {
    session.start(
      { resources: baselineResources, volumes: baselineVolumes },
      {
        openResourceIdx: opts?.resourceIdx ?? null,
        openVolumeIdx: opts?.volumeIdx ?? null,
        openTab: opts?.openTab ?? null,
        linkedAddonIds: connectionAddonIds,
      },
    );
  };

  // Compute "linked but unbound" — addons in linkedAddonIds with zero env
  // bindings across the draft. Env vars no longer carry addon-backed sources,
  // so every linked addon is unbound.
  const computeUnboundLinked = (): string[] => Array.from(session.linkedAddonIds);

  // Env vars are no longer addon-backed, so there is nothing to convert on
  // detach; resources pass through unchanged with empty provenance.
  const applyPendingDetach = (): {
    resources: Partial<FormStackResourceData>[];
    provenance: Map<string, { addonName: string; credField?: string }>;
  } => ({
    resources: session.draft.resources,
    provenance: new Map<string, { addonName: string; credField?: string }>(),
  });

  // ── Deploy lifecycle (page-level: drives the status bar across all tabs) ──
  const deployIds = useMemo(() => ({
    orgId: stackToShow?.organisation_id || getCurrentOrganizationId() || "",
    teamName: (stackToShow ? teamNameById(stackToShow.team_id) : "") || defaultTeamName || "",
    stackId: stackToShow?.id || "",
  }), [stackToShow, teamNameById, defaultTeamName]);

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

  const onDeploy = useCallback(
    () => runDeploy(() => createRelease(deployIds.orgId, deployIds.teamName, deployIds.stackId), "Deploy started"),
    [runDeploy, deployIds],
  );
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

  const performSave = async () => {
    if (!session.isActive) return;
    if (!isDraft && !stackToShow) return;
    setIsSaving(true);
    setValidationErrors({ resources: {}, volumes: {} });

    try {
      const orgId = getCurrentOrganizationId();
      if (!orgId) throw new Error("Organization ID not found");

      const detachResult = session.pendingDetach.size > 0 ? applyPendingDetach() : null;
      const resources = (detachResult?.resources ?? session.draft.resources) as FormStackResourceData[];

      const formStackData: FormStackData = isDraft
        ? buildDraftFormData(draftName.trim(), draftLabels, resources, session.draft.volumes as VolumeFormData[])
        : {
          name: stackToShow!.name || "",
          labels: stackToShow!.labels || [],
          spec: { stack_resources: resources, volumes: session.draft.volumes as VolumeFormData[] },
        };

      const validation = FormStackSchema.safeParse(formStackData);
      if (!validation.success) {
        const nextErrors: typeof validationErrors = { resources: {}, volumes: {} };
        for (const issue of validation.error.issues) {
          const [scope0, scope1, idxRaw, ...rest] = issue.path;
          if (scope0 !== "spec" || (scope1 !== "stack_resources" && scope1 !== "volumes")) continue;
          const idx = typeof idxRaw === "number" ? idxRaw : Number(idxRaw);
          if (Number.isNaN(idx)) continue;
          const bucket = scope1 === "stack_resources" ? nextErrors.resources : nextErrors.volumes;
          if (!bucket[idx]) bucket[idx] = {};
          const fieldKey = rest.join(".");
          if (!bucket[idx][fieldKey]) bucket[idx][fieldKey] = issue.message;
        }
        setValidationErrors(nextErrors);
        toast({
          title: "Validation error",
          description: "Please fix the highlighted errors before deploying.",
          variant: "destructive",
        });
        setIsSaving(false);
        return;
      }

      const teamName = isDraft ? defaultTeamName : teamNameById(fetchedStack?.team_id ?? currentStack?.team_id);
      if (!teamName) {
        toast({
          title: isDraft ? "No team available" : "Failed to update stack",
          description: "Could not resolve a team to save into.",
          variant: "destructive",
        });
        setIsSaving(false);
        return;
      }

      const apiData = convertFormStackToApiStack(formStackData);
      if (isDraft) {
        const created = await createStack(orgId, teamName, apiData);
        session.discard();
        navigate(`/stacks/${created.id}`, { replace: true, state: null });
        return;
      }

      // The stack PUT carries the full desired connection set in spec.connections;
      // the backend replaces the connection set atomically (upsert-by-id) and
      // returns the stack with its reconciled connections. No separate diff.
      const updatedStack = await updateStack(orgId, teamName, id!, apiData);
      setFetchedStack(updatedStack);

      session.discard();

      toast({
        title: "Stack updated successfully",
        description: "Your stack configuration has been saved.",
        variant: "default"
      });

    } catch (err) {
      console.error('Failed to save stack:', err);
      toast({
        title: "Failed to save stack",
        description: err instanceof Error ? err.message : "An unexpected error occurred. Please try again.",
        variant: "destructive"
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleSave = () => {
    if (!session.isActive) return;
    if (session.pendingDetach.size > 0) {
      setDetachConfirmOpen(true);
      return;
    }
    const unbound = computeUnboundLinked();
    if (unbound.length > 0) {
      // Silently drop phantom links — addons that were added to the stack
      // but never referenced in any env var. They have no API representation
      // anyway (links are derived from env vars), so this is just cleanup.
      session.setLinkedAddonIds((prev) => {
        const next = new Set(prev);
        for (const id of unbound) next.delete(id);
        return next;
      });
    }
    void performSave();
  };

  const addDraftLabel = useCallback((value: string) => {
    setDraftLabels((prev) => [...(prev ?? []), { key: "stackdome.io/user-defined-value", value }]);
  }, []);
  const removeDraftLabel = useCallback((idx: number) => {
    setDraftLabels((prev) => (prev ?? []).filter((_, i) => i !== idx));
  }, []);

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

  const detachDialog = (
    <AlertDialog open={detachConfirmOpen} onOpenChange={setDetachConfirmOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Detach {session.pendingDetach.size} {session.pendingDetach.size === 1 ? "addon" : "addons"}?
          </AlertDialogTitle>
          <AlertDialogDescription>
            Bound env keys will be converted to plain stack vars with their last-known values. Confirm to continue.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => {
              setDetachConfirmOpen(false);
              const unbound = computeUnboundLinked();
              if (unbound.length > 0) {
                session.setLinkedAddonIds((prev) => {
                  const next = new Set(prev);
                  for (const id of unbound) next.delete(id);
                  return next;
                });
              }
              void performSave();
            }}
          >
            Confirm and detach
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );

  const dirtyTotal =
    session.dirty.dirtyResourceIdx.size + session.dirty.dirtyVolumeIdx.size + session.dirty.addonLinkCount;
  return (
    <>
      <CanvasEditorShell
        stackName={isDraft ? draftName : (effectiveStack?.name ?? "")}
        isDraft={isDraft}
        nameEditable={isDraft}
        onNameChange={setDraftName}
        labels={(isDraft ? draftLabels : effectiveStack?.labels) ?? []}
        labelsEditable={isDraft}
        onAddLabel={addDraftLabel}
        onRemoveLabel={removeDraftLabel}
        statusState={effectiveStack?.status?.state}
        subtitle={subtitleText}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        isActive={session.isActive}
        dirtyResourceCount={session.dirty.dirtyResourceIdx.size}
        dirtyTotal={dirtyTotal}
        isStaged={lifecycle.phase === "staged"}
        isSaving={isSaving}
        deployBusy={deployBusy}
        canWrite={canWriteStack}
        onSave={handleSave}
        onDeploy={onDeploy}
        onDiscardAll={() => session.discard()}
        onEdit={() => activateEdit({})}
        onDelete={() =>
          toast({ title: "Not implemented", description: "Delete stack will land in a follow-up." })
        }
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
      {detachDialog}
    </>
  );
}
