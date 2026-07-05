import React from "react";
import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";
import { X, GitBranch, Box, Trash2 } from "lucide-react";

import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../hooks/use-secrets";
import type { PostgresAddon } from "@/api/addons";
import { StackResourceConfigurationTab } from "./stack-resource-configuration-tab";
import { StackResourceDeploymentTab } from "./stack-resource-deployment-tab";
import { StackResourceEnvironmentTab } from "./stack-resource-environment-tab";
import { useResourceTabProps } from "./hooks/use-resource-tab-props";

export type AddonGroupStateMap = Map<string, "idle" | "editing-binding" | "detaching">;

interface StackResourceItemProps {
  resource: Partial<FormStackResourceData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  onChange: (index: number, updatedResource: Partial<FormStackResourceData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  volumes?: Partial<VolumeFormData>[];
  allResources?: { name: string; index: number; outputs: string[] }[];
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
  const {
    dirtyTabs,
    isDirty,
    statusDotColor,
    statusState,
    configurationProps,
    deploymentProps,
    environmentProps,
  } = useResourceTabProps({
    resource,
    index,
    baselineResource,
    onChange,
    context: { errors, volumes, allResources, secrets, addons, addonNameById, onDiscardField, onDiscardEnvRow },
  });

  return (
    <TooltipProvider>
      <AccordionItem
        value={String(index)}
        className="relative border-t border-border first:border-t-0"
        style={isDirty ? { boxShadow: "inset 3px 0 0 var(--brand)" } : undefined}
      >
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
                <p className="capitalize">{statusState || 'Pending'}</p>
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
                    <span>{resource.source?.image?.ref || "No image specified"}</span>
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5">
                    <GitBranch className="h-3.5 w-3.5" />
                    <span>
                      {resource.source?.git?.repo_url || "No repository specified"}
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
          </div>
        </AccordionTrigger>
        {/* "Modified" badge + discard sits OUTSIDE the AccordionTrigger button —
            a nested interactive <button> inside the trigger's <button> is invalid
            HTML and throws a hydration warning. */}
        {isDirty && onDiscardResource && (
          <div
            className="absolute right-11 top-[18px] flex items-center"
            onClick={(e) => e.stopPropagation()}
          >
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
        <AccordionContent className="bg-background dark:bg-secondary border-t border-border pb-4 pt-4 px-1">
          <div className="px-4 space-y-4">
            <Tabs defaultValue="general" className="w-full">
              <div className="mt-1 mb-3">
                <TabsList className="w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-1 px-2">
                  <TabsTrigger value="general" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Configuration
                    {dirtyTabs.configuration && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="deployment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Deployment
                    {dirtyTabs.deployment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                  <TabsTrigger value="environment" className="flex-none rounded-t-md rounded-b-none border border-transparent px-4 py-2 text-[13px] text-muted-foreground hover:text-foreground -mb-px data-[state=active]:bg-background dark:data-[state=active]:bg-secondary data-[state=active]:border-border data-[state=active]:border-b-transparent data-[state=active]:text-foreground data-[state=active]:font-medium data-[state=active]:shadow-none">
                    Environment
                    {dirtyTabs.environment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
                  </TabsTrigger>
                </TabsList>
              </div>

              <StackResourceConfigurationTab {...configurationProps} />
              <StackResourceDeploymentTab {...deploymentProps} />
              <StackResourceEnvironmentTab {...environmentProps} />
            </Tabs>

            <div className="mt-8 pt-3 border-t border-border flex justify-end">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-[12.5px] text-muted-foreground/70 hover:text-danger hover:bg-danger-bg"
                onClick={() => onRemove(index)}
              >
                <Trash2 className="h-3.5 w-3.5" />
                Remove resource
              </Button>
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
 * React.memo. `useResourceTabProps` assembles their props with referentially
 * stable callbacks so a keystroke that only mutates Configuration fields leaves
 * the Deployment and Environment subtrees idle (memo skips them).
 *
 * The outer React.memo here keeps the *whole* item idle when other resources
 * in the accordion change — Radix Accordion keeps closed items mounted.
 */
const StackResourceItem = React.memo(StackResourceItemImpl);
export default StackResourceItem;
