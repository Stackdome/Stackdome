import React, { useCallback, useMemo, useRef } from "react";
import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";
import { X, GitBranch, Box, Trash2 } from "lucide-react";
import { ApiStackResourceStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { dirtyTabsForResource, isResourceDirty } from "@/pages/stacks/lib/stack-diff";
import type { z } from "zod";
import { variantFromState } from "@/components/branded";

import type { FormStackResourceData, FormEnvVarData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../hooks/use-secrets";
import type { PostgresAddon } from "@/api/addons";
import {
  StackResourceConfigurationTab,
  pickConfigurationDraft,
} from "./stack-resource-configuration-tab";
import {
  StackResourceDeploymentTab,
  pickDeploymentDraft,
} from "./stack-resource-deployment-tab";
import { StackResourceEnvironmentTab } from "./stack-resource-environment-tab";

export type AddonGroupStateMap = Map<string, "idle" | "editing-binding" | "detaching">;

interface StackResourceItemProps {
  resource: Partial<FormStackResourceData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  onChange: (index: number, updatedResource: Partial<FormStackResourceData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  volumes?: Partial<VolumeFormData>[];
  allResources?: { name: string; index: number }[];
  secrets: UseSecretsReturn;
  addons: PostgresAddon[];
  addonNameById: Map<string, string>;
  addonGroupState?: AddonGroupStateMap;
  onEditAddonBinding?: (addonId: string) => void;
  onDetachAddon?: (addonId: string) => void;
  onCancelDetachAddon?: (addonId: string) => void;
  /** Baseline snapshot of this resource. When provided, the component renders dirty visualization (modified row tints, tab dots, Modified pill) and exposes per-row reset. */
  baselineResource?: Partial<FormStackResourceData>;
  /** Reset a single env row to its baseline value. Required for the per-row reset arrow to render. */
  onDiscardEnvRow?: (envIdx: number) => void;
  /** Discard all changes for this resource. Required for the Modified pill ✕ affordance. */
  onDiscardResource?: () => void;
  /** Discard a single field on this resource by dot-path. Required for per-field reset arrows. */
  onDiscardField?: (path: string) => void;
}

function StackResourceItemImpl({
  resource,
  index,
  itemRef,
  onChange,
  onRemove,
  errors,
  volumes = [],
  allResources,
  secrets,
  addons,
  addonNameById,
  baselineResource,
  onDiscardEnvRow,
  onDiscardResource,
  onDiscardField,
}: StackResourceItemProps) {
  // Per-tab dirty bucketing → renders the small brand dot next to each tab label.
  const dirtyTabs = useMemo(
    () => (baselineResource ? dirtyTabsForResource(resource, baselineResource) : { configuration: false, deployment: false, environment: false }),
    [resource, baselineResource],
  );

  const isDirty = baselineResource ? isResourceDirty(resource, baselineResource) : false;

  // Refs to the latest props so the patch callbacks below can stay
  // referentially stable across renders. Without this, every keystroke would
  // recreate the callbacks and defeat React.memo on the per-tab children —
  // the whole point of splitting the file.
  const resourceRef = useRef(resource);
  resourceRef.current = resource;
  const indexRef = useRef(index);
  indexRef.current = index;
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  // Stable: shallow-merge a patch into the latest resource and forward.
  const onPatchResource = useCallback((patch: Partial<FormStackResourceData>) => {
    onChangeRef.current(indexRef.current, { ...resourceRef.current, ...patch });
  }, []);

  // Stable: merge into init_spec, preserving sibling fields.
  const onPatchInitSpec = useCallback(
    (patch: Partial<NonNullable<FormStackResourceData["init_spec"]>>) => {
      const r = resourceRef.current;
      onChangeRef.current(indexRef.current, {
        ...r,
        init_spec: { ...r.init_spec, ...patch },
      });
    },
    [],
  );

  // Stable: patch only command/args on execution_config, preserving env vars.
  const onPatchExecCommandArgs = useCallback(
    (patch: { command?: string[]; args?: string[] }) => {
      const r = resourceRef.current;
      onChangeRef.current(indexRef.current, {
        ...r,
        execution_config: { ...r.execution_config, ...patch },
      });
    },
    [],
  );

  // Stable: replace the env vars array, preserving the rest of execution_config.
  const onChangeEnvVars = useCallback((next: FormEnvVarData[]) => {
    const r = resourceRef.current;
    onChangeRef.current(indexRef.current, {
      ...r,
      execution_config: {
        ...r.execution_config,
        environment_variables: next,
      },
    });
  }, []);

  // Status semantics
  const statusObj = (resource.status ?? {}) as z.infer<typeof ApiStackResourceStatusSchema>;
  const statusVariant = variantFromState(statusObj.state);
  const statusDotColor =
    statusVariant === "ready" ? "bg-success"
    : statusVariant === "error" ? "bg-danger"
    : statusVariant === "pending" ? "bg-warn"
    : "bg-muted-foreground";

  // Per-tab projections. Done in useMemo so the projected object's identity
  // only changes when the tab's own fields change — which is what lets
  // React.memo skip the inactive tabs.
  const configurationDraft = useMemo(() => pickConfigurationDraft(resource), [
    resource.name,
    resource.depends_on,
    resource.sourceType,
    resource.image_spec,
    resource.build_spec,
    resource.useImageSecret,
    resource.selectedImageSecretId,
    resource.useGitSecret,
    resource.selectedGitSecretId,
    resource.gitRevisionType,
    resource.gitRevisionValue,
    resource.volume_mounts,
    resource.ports,
  ]);
  const configurationBaseline = useMemo(
    () => (baselineResource ? pickConfigurationDraft(baselineResource) : undefined),
    [baselineResource],
  );
  const deploymentDraft = useMemo(() => pickDeploymentDraft(resource), [
    resource.init_spec,
    resource.execution_config?.command,
    resource.execution_config?.args,
  ]);
  const deploymentBaseline = useMemo(
    () => (baselineResource ? pickDeploymentDraft(baselineResource) : undefined),
    [baselineResource],
  );

  const envVars = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];
  const baselineEnvVars = baselineResource?.execution_config?.environment_variables as FormEnvVarData[] | undefined;

  return (
    <TooltipProvider>
      <AccordionItem value={String(index)} className="border-t border-border first:border-t-0">
        <AccordionTrigger
          ref={itemRef}
          className="px-4 py-3 hover:bg-muted/40 data-[state=open]:bg-muted/30 rounded-t-md [&[data-state=open]]:rounded-b-none"
        >
          <div className="flex items-center gap-3 text-left flex-grow">
            <Tooltip delayDuration={300}>
              <TooltipTrigger asChild>
                <span className={`h-2 w-2 rounded-full shrink-0 ${statusDotColor}`}></span>
              </TooltipTrigger>
              <TooltipContent side="top">
                <p className="capitalize">{statusObj.state || 'Pending'}</p>
              </TooltipContent>
            </Tooltip>
            <div className="flex flex-col flex-grow min-w-0">
              <span className="font-medium flex items-center gap-2">
                {resource.name || `Resource ${index + 1}`}
              </span>
              <span className="text-sm text-muted-foreground truncate">
                {resource.sourceType === "image" ? (
                  <span className="flex items-center gap-1.5">
                    <Box className="h-3.5 w-3.5" />
                    <span>{resource.image_spec?.image || "No image specified"}</span>
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5">
                    <GitBranch className="h-3.5 w-3.5" />
                    <span>
                      {resource.build_spec?.source_context?.git_repo?.repo_url || "No repository specified"}
                      {resource.gitRevisionType && resource.gitRevisionValue && (
                        <span className="ml-1 text-xs bg-muted/50 px-1.5 py-0.5 rounded-full">
                          {resource.gitRevisionType === "branch" && "Branch: "}
                          {resource.gitRevisionType === "tag" && "Tag: "}
                          {resource.gitRevisionType === "commit" && "SHA: "}
                          {resource.gitRevisionValue}
                        </span>
                      )}
                    </span>
                  </span>
                )}
              </span>
              {errors._form && (
                <span className="text-xs text-danger mt-0.5 pl-6">{errors._form}</span>
              )}
            </div>
            {isDirty && onDiscardResource && (
              <div className="ml-auto flex items-center shrink-0 mr-2" onClick={(e) => e.stopPropagation()}>
                <span className="inline-flex items-center gap-1.5 rounded-md border border-brand-border bg-brand-bg pl-2 pr-1 py-0.5 text-[11px] font-medium text-brand">
                  Modified
                  <button
                    type="button"
                    onClick={onDiscardResource}
                    aria-label="Discard changes to this resource"
                    title="Discard changes to this resource"
                    className="inline-flex h-4 w-4 items-center justify-center rounded hover:bg-brand/15"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              </div>
            )}
          </div>
        </AccordionTrigger>
        <AccordionContent className="bg-secondary border-t border-border pb-4 pt-4 px-1">
          <div className="px-4 space-y-4">
            <Tabs defaultValue="general" className="w-full">
              <div className="mt-1 mb-3">
                <TabsList className="w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-1 px-2">
                  <TabsTrigger value="general" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Configuration
                    {dirtyTabs.configuration && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="deployment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Deployment
                    {dirtyTabs.deployment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="environment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Environment
                    {dirtyTabs.environment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                </TabsList>
              </div>

              <StackResourceConfigurationTab
                index={index}
                draft={configurationDraft}
                baseline={configurationBaseline}
                errors={errors}
                volumes={volumes}
                allResources={allResources}
                secrets={secrets}
                onDiscardField={onDiscardField}
                onPatchResource={onPatchResource}
              />

              <StackResourceDeploymentTab
                index={index}
                draft={deploymentDraft}
                baseline={deploymentBaseline}
                onPatchInitSpec={onPatchInitSpec}
                onPatchExecCommandArgs={onPatchExecCommandArgs}
                onDiscardField={onDiscardField}
              />

              <StackResourceEnvironmentTab
                index={index}
                envVars={envVars}
                baselineEnvVars={baselineEnvVars}
                errors={errors}
                secrets={secrets}
                addons={addons}
                addonNameById={addonNameById}
                onChangeEnvVars={onChangeEnvVars}
                onDiscardEnvRow={onDiscardEnvRow}
              />
            </Tabs>

            <div className="flex justify-center items-center mt-8">
              <span className="flex items-center justify-center w-full py-3 rounded-md bg-muted/70">
                <Button
                  type="button"
                  variant="ghost"
                  className="text-danger hover:text-danger hover:bg-danger-bg focus-visible:bg-danger-bg"
                  onClick={() => onRemove(index)}
                >
                  <Trash2 className="h-4 w-4 mr-1" />
                  Remove Resource
                </Button>
              </span>
            </div>
          </div>
        </AccordionContent>
      </AccordionItem>
    </TooltipProvider>
  );
}

/**
 * The form is heavy (~1500 lines of JSX). It's split into per-tab
 * subcomponents (Configuration / Deployment / Environment), each wrapped in
 * React.memo. The parent passes per-tab projected `draft` slices so a
 * keystroke that only mutates Configuration fields leaves the Deployment and
 * Environment subtrees idle (memo skips them on shallow compare).
 *
 * The outer React.memo here keeps the *whole* item idle when other resources
 * in the accordion change — Radix Accordion keeps closed items mounted.
 */
const StackResourceItem = React.memo(StackResourceItemImpl);
export default StackResourceItem;
