import { useParams, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Button } from "@/components/ui/button";
import { Loader2, Pencil, Rocket, Trash2 } from "lucide-react";
import { PageHeader, Panel, StatusPill, variantFromState } from "@/components/branded";
import { useMemo, useState, useEffect, useCallback } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import StackResourcesForm, { getDefaultResource } from "@/pages/stacks/components/shared/stack-resources-form";
import StackVolumesForm, { getDefaultVolume } from "@/pages/stacks/components/shared/stack-volumes-form";
import StackResourcesDetail from "@/pages/stacks/components/detail/stack-resources-detail";
import StackVolumesDetail from "@/pages/stacks/components/detail/stack-volumes-detail";
import StickyActionBar, { type StickyActionBarSegment } from "@/pages/stacks/components/shared/sticky-action-bar";
import AddonsInStackPanel from "@/pages/stacks/components/detail/addons-in-stack-panel";
import { useStackEditSession, type EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
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
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { AddonGroupStateMap } from "@/pages/stacks/components/shared/stack-resource-item";
import { StackLogsTab } from "@/pages/stacks/components/detail/logs/stack-logs-tab";
import { StackMetricsTab } from "@/pages/stacks/components/detail/metrics/stack-metrics-tab";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import { getStackById, updateStack } from "@/api/stacks";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
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
  const { stacks } = useStacks();
  const [fetchedStack, setFetchedStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const session = useStackEditSession();
  const { addons: postgresAddons } = usePostgresAddons();
  const addonNameById = useMemo(
    () => new Map(postgresAddons.filter((a) => a.id).map((a) => [a.id!, a.name])),
    [postgresAddons],
  );
  const [isSaving, setIsSaving] = useState(false);
  const [editingBindingIds, setEditingBindingIds] = useState<Set<string>>(new Set());
  // Per-addonId provenance for converted env rows after a successful detach
  // save. Keyed by `${resourceIdx}::${envName}`. Page-state only — vanishes
  // on reload by design (no API field for it).
  const [detachedProvenance, setDetachedProvenance] = useState<Map<string, { addonName: string; credField?: string }>>(
    new Map(),
  );
  const [detachConfirmOpen, setDetachConfirmOpen] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{
    resources: { [index: number]: { [field: string]: string | undefined } };
    volumes: { [index: number]: { [field: string]: string | undefined } };
  }>({ resources: {}, volumes: {} });

  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();

  // Find the current stack in context
  const currentStack = stacks.find((stack) => stack.id === id);

  // Update breadcrumb with stack name
  useEffect(() => {
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
      getStackById(orgId, id)
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
  }, [currentStack, id, setCustomLabel, setPathLoading]);

  const stackToShow = currentStack || fetchedStack;

  const baselineResources = useMemo<FormStackResourceData[]>(
    () => (stackToShow?.spec.stack_resources || []).map(mapStackResourceToFormData),
    [stackToShow],
  );
  const baselineVolumes = useMemo<VolumeFormData[]>(
    () => (stackToShow?.spec?.volumes || []).map(mapVolumeToFormData),
    [stackToShow],
  );

  const activateEdit = (opts?: { resourceIdx?: number; volumeIdx?: number; openTab?: EditSessionTab }) => {
    session.start(
      { resources: baselineResources, volumes: baselineVolumes },
      {
        openResourceIdx: opts?.resourceIdx ?? null,
        openVolumeIdx: opts?.volumeIdx ?? null,
        openTab: opts?.openTab ?? null,
      },
    );
    setEditingBindingIds(new Set());
  };

  // Page-derived per-addonId state map. Detaching wins over editing.
  const addonGroupState = useMemo<AddonGroupStateMap>(() => {
    const m = new Map<string, "idle" | "editing-binding" | "detaching">();
    for (const id of editingBindingIds) m.set(id, "editing-binding");
    for (const id of session.pendingDetach) m.set(id, "detaching");
    return m;
  }, [editingBindingIds, session.pendingDetach]);

  const handleEditAddonBinding = useCallback((addonId: string) => {
    setEditingBindingIds((prev) => {
      const next = new Set(prev);
      next.add(addonId);
      return next;
    });
  }, []);

  const handleDetachAddon = useCallback((addonId: string) => {
    setEditingBindingIds((prev) => {
      const next = new Set(prev);
      next.delete(addonId);
      return next;
    });
    session.setPendingDetach((prev) => {
      const next = new Set(prev);
      next.add(addonId);
      return next;
    });
  }, [session]);

  const handleCancelDetachAddon = useCallback((addonId: string) => {
    session.setPendingDetach((prev) => {
      const next = new Set(prev);
      next.delete(addonId);
      return next;
    });
    setEditingBindingIds((prev) => {
      const next = new Set(prev);
      next.add(addonId);
      return next;
    });
  }, [session]);

  // Compute "linked but unbound" — addons in linkedAddonIds with zero env
  // bindings across the draft.
  const computeUnboundLinked = (): string[] => {
    const referenced = new Set<string>();
    for (const r of session.draft.resources) {
      const envs = (r?.execution_config?.environment_variables || []) as FormEnvVarData[];
      for (const e of envs) {
        if (e.from === "addon" && e.addonId) referenced.add(e.addonId);
      }
    }
    return Array.from(session.linkedAddonIds).filter((id) => !referenced.has(id));
  };

  // Convert addon rows in pendingDetach to plain stack rows. Returns the new
  // resources array and the provenance map of resourceIdx::envName entries.
  const applyPendingDetach = (): {
    resources: Partial<FormStackResourceData>[];
    provenance: Map<string, { addonName: string; credField?: string }>;
  } => {
    const provenance = new Map<string, { addonName: string; credField?: string }>();
    const resources = session.draft.resources.map((r, rIdx) => {
      const envs = (r?.execution_config?.environment_variables || []) as FormEnvVarData[];
      const newEnvs = envs.map((e) => {
        if (e.from === "addon" && session.pendingDetach.has(e.addonId)) {
          provenance.set(`${rIdx}::${e.name}`, {
            addonName: addonNameById.get(e.addonId) ?? e.addonId,
            credField: e.credField,
          });
          return { from: "stack" as const, name: e.name, value: "" };
        }
        return e;
      });
      return {
        ...r,
        execution_config: {
          ...(r.execution_config || {}),
          environment_variables: newEnvs,
        },
      };
    });
    return { resources, provenance };
  };

  const performSave = async () => {
    if (!stackToShow || !session.isActive || !id) return;
    setIsSaving(true);
    setValidationErrors({ resources: {}, volumes: {} });

    try {
      const orgId = getCurrentOrganizationId();
      if (!orgId) throw new Error("Organization ID not found");

      const detachResult = session.pendingDetach.size > 0 ? applyPendingDetach() : null;
      const resources = (detachResult?.resources ?? session.draft.resources) as FormStackResourceData[];

      const formStackData: FormStackData = {
        name: stackToShow.name || '',
        labels: stackToShow.labels || [],
        spec: {
          stack_resources: resources,
          volumes: session.draft.volumes as VolumeFormData[],
        }
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

      const apiData = convertFormStackToApiStack(formStackData);
      const updatedStack = await updateStack(orgId, id, apiData);

      setFetchedStack(updatedStack);
      if (detachResult) setDetachedProvenance(detachResult.provenance);
      else setDetachedProvenance(new Map());
      session.discard();
      setEditingBindingIds(new Set());

      toast({
        title: "Stack updated successfully",
        description: "Your stack configuration has been saved.",
        variant: "default"
      });

    } catch (err) {
      console.error('Failed to update stack:', err);
      toast({
        title: "Failed to update stack",
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

  // NOTE: this used to be wrapped in useTransition, which improved
  // typing throughput but broke caret position on controlled inputs —
  // when React commits the deferred state, it re-applies the input's
  // value prop and the browser moves the caret to the end. Mid-string
  // edits became impossible. The win from Cycle 6 (per-tab memoization)
  // gives us most of the throughput back without the side effect.
  const handleResourcesChange = useCallback((updatedResources: Partial<FormStackResourceData>[]) => {
    session.updateResources(updatedResources as FormStackResourceData[]);
  }, [session]);

  const handleVolumesChange = useCallback((updatedVolumes: Partial<VolumeFormData>[]) => {
    session.updateVolumes(updatedVolumes as VolumeFormData[]);
  }, [session]);

  const handleDiscardEnvRow = useCallback(
    (rIdx: number, eIdx: number) => session.discardEnvRow(rIdx, eIdx),
    [session],
  );
  const handleDiscardResource = useCallback(
    (rIdx: number) => session.discardResource(rIdx),
    [session],
  );
  const handleDiscardResourceField = useCallback(
    (rIdx: number, path: string) => session.discardResourceField(rIdx, path),
    [session],
  );

  // Stable Set of every addonId currently available to env rows: explicit
  // links + addons referenced by existing env vars. Without useMemo this
  // would be a fresh Set every render of detail/index.tsx, breaking memo
  // on every StackResourceItem child via the addons → availableAddonIds
  // → addons.filter chain.
  const availableAddonIds = useMemo(() => {
    const ids = new Set(session.linkedAddonIds);
    for (const r of session.draft.resources) {
      const envs = (r?.execution_config?.environment_variables || []) as FormEnvVarData[];
      for (const e of envs) {
        if (e.from === "addon" && e.addonId) ids.add(e.addonId);
      }
    }
    return ids;
  }, [session.linkedAddonIds, session.draft.resources]);

  if (loading) {
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

  if (!stackToShow) {
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

  const resourceCount = stackToShow.spec?.stack_resources?.length || 0;
  const volumeCount = stackToShow.spec?.volumes?.length || 0;
  const subtitleParts: React.ReactNode[] = [
    `${resourceCount} ${resourceCount === 1 ? "service" : "services"}`,
    `${volumeCount} ${volumeCount === 1 ? "volume" : "volumes"}`,
  ];

  const headerActions = (
    <div className="flex items-center gap-2">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => activateEdit({})}
        disabled={session.isActive}
        aria-label={session.isActive ? "Stack is being edited" : "Edit stack"}
      >
        <Pencil className="h-3.5 w-3.5" />
        {session.isActive ? "Editing" : "Edit stack"}
      </Button>
      <Button
        variant="ghost"
        size="sm"
        className="text-muted-foreground/70 hover:text-danger hover:bg-danger-bg"
        onClick={() =>
          toast({
            title: "Not implemented",
            description: "Delete stack will land in a follow-up.",
          })
        }
        aria-label="Delete stack"
      >
        <Trash2 className="h-3.5 w-3.5" />
        Delete stack
      </Button>
    </div>
  );

  return (
    <div className="p-8 space-y-8">
      {session.isActive && (() => {
        const resourceCount = session.dirty.dirtyResourceIdx.size;
        const volumeCount = session.dirty.dirtyVolumeIdx.size;
        const dirtyEntities = resourceCount + volumeCount;
        const segments: StickyActionBarSegment[] = [];
        if (resourceCount > 0) {
          segments.push({ num: resourceCount, label: resourceCount === 1 ? "RESOURCE MODIFIED" : "RESOURCES MODIFIED" });
        }
        if (volumeCount > 0) {
          segments.push({ num: volumeCount, label: volumeCount === 1 ? "VOLUME MODIFIED" : "VOLUMES MODIFIED" });
        }
        if (session.dirty.addonLinkCount > 0) {
          segments.push({ num: session.dirty.addonLinkCount, label: session.dirty.addonLinkCount === 1 ? "ADDON LINK" : "ADDON LINKS" });
        }
        return (
          <StickyActionBar
            leadLabel="Draft"
            segments={segments}
            primary={{
              label: "Deploy",
              loadingLabel: "Deploying",
              icon: <Rocket className="h-3.5 w-3.5" />,
              isLoading: isSaving,
              onClick: handleSave,
            }}
            secondary={{
              label: "Discard all",
              onClick: () => session.discard(),
              dirtyCount: dirtyEntities,
              confirm: {
                threshold: 2,
                title: "Discard all changes?",
                description: (
                  <>
                    You have unsaved edits across {dirtyEntities}{" "}
                    {dirtyEntities === 1 ? "item" : "items"}. This will revert every
                    change in this session.
                  </>
                ),
                confirmLabel: "Discard all",
                cancelLabel: "Keep editing",
              },
            }}
          />
        );
      })()}

      <PageHeader
        title={stackToShow.name}
        status={
          <span className="flex items-center gap-2">
            {stackToShow.status?.state && (
              <StatusPill variant={variantFromState(stackToShow.status.state)}>
                {stackToShow.status.state}
              </StatusPill>
            )}
          </span>
        }
        subtitle={subtitleParts.map((p, i) => (
          <span key={i}>
            {i > 0 && <span className="mx-2 text-muted-foreground/50">·</span>}
            {p}
          </span>
        ))}
        actions={headerActions}
      />

      <Tabs defaultValue="configuration" className="w-full">
        <TabsList className="mb-6 w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-6">
          <TabsTrigger
            value="configuration"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-brand data-[state=active]:text-brand data-[state=active]:bg-transparent data-[state=active]:shadow-none px-1 pb-3 -mb-px font-medium"
          >
            Configuration
          </TabsTrigger>
          <TabsTrigger
            value="logs"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-brand data-[state=active]:text-brand data-[state=active]:bg-transparent data-[state=active]:shadow-none px-1 pb-3 -mb-px font-medium"
          >
            Logs
          </TabsTrigger>
          <TabsTrigger
            value="metrics"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-brand data-[state=active]:text-brand data-[state=active]:bg-transparent data-[state=active]:shadow-none px-1 pb-3 -mb-px font-medium"
          >
            Metrics
          </TabsTrigger>
        </TabsList>

        {/* Configuration Tab: Stack Resources and Volumes */}
        <TabsContent value="configuration" className="space-y-8">
          <Panel
            title="Stack Resources"
            count={baselineResources.length}
            bodyClassName="p-0"
          >
            {session.isActive ? (
              <StackResourcesForm
                resources={session.draft.resources}
                onResourcesChange={handleResourcesChange}
                errors={validationErrors.resources}
                volumes={session.draft.volumes}
                accordionDefaultOpen={false}
                defaultOpenResourceIdx={session.openResourceIdx}
                addonGroupState={addonGroupState}
                onEditAddonBinding={handleEditAddonBinding}
                onDetachAddon={handleDetachAddon}
                onCancelDetachAddon={handleCancelDetachAddon}
                baselineResources={session.baseline.resources}
                onDiscardEnvRow={handleDiscardEnvRow}
                onDiscardResource={handleDiscardResource}
                onDiscardResourceField={handleDiscardResourceField}
                availableAddonIds={availableAddonIds}
              />
            ) : (
              <StackResourcesDetail
                resources={baselineResources}
                accordionDefaultOpen={false}
                onEditResource={(idx) =>
                  activateEdit({ resourceIdx: idx, openTab: "environment" })
                }
                onAddResource={() => {
                  const nextIdx = baselineResources.length;
                  session.start(
                    {
                      resources: [...baselineResources, getDefaultResource() as FormStackResourceData],
                      volumes: baselineVolumes,
                    },
                    { openResourceIdx: nextIdx, openTab: "configuration" },
                  );
                  setEditingBindingIds(new Set());
                }}
                detachedProvenance={detachedProvenance}
              />
            )}
          </Panel>

          <Panel
            title="Stack Volumes"
            count={baselineVolumes.length}
            bodyClassName="p-0"
          >
            {session.isActive ? (
              <StackVolumesForm
                volumes={session.draft.volumes}
                onVolumesChange={handleVolumesChange}
                errors={validationErrors.volumes}
                stackResources={session.draft.resources}
                baselineVolumes={baselineVolumes}
                accordionDefaultOpen={false}
                defaultOpenVolumeIdx={session.openVolumeIdx}
              />
            ) : (
              <StackVolumesDetail
                volumes={baselineVolumes}
                stackResources={baselineResources}
                accordionDefaultOpen={false}
                onEditVolume={(idx) => activateEdit({ volumeIdx: idx })}
                onAddVolume={() => {
                  const nextIdx = baselineVolumes.length;
                  session.start(
                    {
                      resources: baselineResources,
                      volumes: [...baselineVolumes, getDefaultVolume() as VolumeFormData],
                    },
                    { openVolumeIdx: nextIdx },
                  );
                  setEditingBindingIds(new Set());
                }}
              />
            )}
          </Panel>

          {(() => {
            const baseline = { resources: baselineResources, volumes: baselineVolumes };
            const ensureActive = () => {
              if (!session.isActive) session.start(baseline);
            };
            return (
              <AddonsInStackPanel
                resources={(session.isActive ? session.draft.resources : baselineResources) as Partial<FormStackResourceData>[]}
                linkedAddonIds={session.linkedAddonIds}
                onLinkAddon={(addonId) => {
                  ensureActive();
                  session.setLinkedAddonIds((prev) => {
                    const next = new Set(prev);
                    next.add(addonId);
                    return next;
                  });
                }}
                onRemoveLinkedAddon={(addonId) => {
                  session.setLinkedAddonIds((prev) => {
                    const next = new Set(prev);
                    next.delete(addonId);
                    return next;
                  });
                }}
              />
            );
          })()}
        </TabsContent>

        {/* Logs Tab */}
        <TabsContent value="logs">
          {stackToShow.id ? (
            <StackLogsTab
              stackId={stackToShow.id}
              organizationId={stackToShow.organisation_id || getCurrentOrganizationId() || ''}
              resources={stackToShow.spec.stack_resources?.map(r => ({ name: r.name || '', id: r.id || '' })) || []}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>

        {/* Metrics Tab */}
        <TabsContent value="metrics">
          {stackToShow.id ? (
            <StackMetricsTab
              stackId={stackToShow.id}
              organizationId={stackToShow.organisation_id || getCurrentOrganizationId() || ''}
              resources={stackToShow.spec.stack_resources || []}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>
      </Tabs>

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
    </div>
  );
}
