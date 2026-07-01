import React, { useCallback, useMemo, useRef } from "react";
import type { z } from "zod";
import { ApiStackResourceStatusSchema } from "@/pages/stacks/schemas/api-schema";
import { variantFromState } from "@/components/branded";
import {
  dirtyTabsForResource,
  isResourceDirty,
  type ResourceDirtyTabs,
} from "@/pages/stacks/lib/stack-diff";
import type {
  FormStackResourceData,
  FormEnvVarData,
  FormVolumeExtendedData as VolumeFormData,
} from "@/pages/stacks/schemas/form-schema";
import type { UseSecretsReturn } from "../../../hooks/use-secrets";
import type { PostgresAddon } from "@/api/addons";
import { StackResourceConfigurationTab, pickConfigurationDraft } from "../stack-resource-configuration-tab";
import { StackResourceDeploymentTab, pickDeploymentDraft } from "../stack-resource-deployment-tab";
import { StackResourceEnvironmentTab } from "../stack-resource-environment-tab";

/**
 * Page-level context the three resource sub-tabs need but that does not vary per
 * keystroke (secrets, addons, sibling resources, discard callbacks).
 */
export interface ResourceTabContext {
  errors: { [field: string]: string | undefined };
  volumes?: Partial<VolumeFormData>[];
  allResources?: { name: string; index: number; outputs: string[] }[];
  secrets: UseSecretsReturn;
  addons: PostgresAddon[];
  addonNameById: Map<string, string>;
  onDiscardField?: (path: string) => void;
  onDiscardEnvRow?: (envIdx: number) => void;
  /** When provided, the canvas drawer offers inline volume creation. */
  onCreateVolume?: (input: { name: string; size: string; targetPath: string }) => void;
}

type ConfigurationProps = React.ComponentProps<typeof StackResourceConfigurationTab>;
type DeploymentProps = React.ComponentProps<typeof StackResourceDeploymentTab>;
type EnvironmentProps = React.ComponentProps<typeof StackResourceEnvironmentTab>;

export interface ResourceTabProps {
  dirtyTabs: ResourceDirtyTabs;
  isDirty: boolean;
  /** Tailwind class for the status dot (`bg-success` etc.). */
  statusDotColor: string;
  statusState: string | undefined;
  configurationProps: ConfigurationProps;
  deploymentProps: DeploymentProps;
  environmentProps: EnvironmentProps;
}

/**
 * Assembles the props for the Configuration / Deployment / Environment sub-tabs
 * of a single stack resource, plus its dirty/status derivations.
 *
 * Lifted verbatim out of `stack-resource-item.tsx` so both the accordion form
 * and the canvas drawer drive the SAME wiring (DRY) and the SAME env-var
 * grouping. The patch callbacks are kept referentially stable via refs — that
 * stability is what lets `React.memo` skip the inactive tabs on each keystroke.
 */
export function useResourceTabProps(args: {
  resource: Partial<FormStackResourceData>;
  index: number;
  baselineResource?: Partial<FormStackResourceData>;
  onChange: (index: number, updated: Partial<FormStackResourceData>) => void;
  context: ResourceTabContext;
}): ResourceTabProps {
  const { resource, index, baselineResource, onChange, context } = args;

  const dirtyTabs = useMemo(
    () =>
      baselineResource
        ? dirtyTabsForResource(resource, baselineResource)
        : { configuration: false, deployment: false, environment: false },
    [resource, baselineResource],
  );
  const isDirty = baselineResource ? isResourceDirty(resource, baselineResource) : false;

  // Refs keep the patch callbacks stable across renders without going stale.
  const resourceRef = useRef(resource);
  resourceRef.current = resource;
  const indexRef = useRef(index);
  indexRef.current = index;
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const onPatchResource = useCallback((patch: Partial<FormStackResourceData>) => {
    onChangeRef.current(indexRef.current, { ...resourceRef.current, ...patch });
  }, []);

  const onPatchInitSpec = useCallback((patch: Partial<NonNullable<FormStackResourceData["init_spec"]>>) => {
    const r = resourceRef.current;
    onChangeRef.current(indexRef.current, { ...r, init_spec: { ...r.init_spec, ...patch } });
  }, []);

  const onPatchExecCommandArgs = useCallback((patch: { command?: string[]; args?: string[] }) => {
    const r = resourceRef.current;
    onChangeRef.current(indexRef.current, { ...r, execution_config: { ...r.execution_config, ...patch } });
  }, []);

  const onChangeEnvVars = useCallback((next: FormEnvVarData[]) => {
    const r = resourceRef.current;
    onChangeRef.current(indexRef.current, {
      ...r,
      execution_config: { ...r.execution_config, environment_variables: next },
    });
  }, []);

  const statusObj = (resource.status ?? {}) as z.infer<typeof ApiStackResourceStatusSchema>;
  const statusVariant = variantFromState(statusObj.state);
  const statusDotColor =
    statusVariant === "ready"
      ? "bg-success"
      : statusVariant === "error"
        ? "bg-danger"
        : statusVariant === "pending"
          ? "bg-warn"
          : "bg-muted-foreground";

  const configurationDraft = useMemo(
    () => pickConfigurationDraft(resource),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [
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
    ],
  );
  const configurationBaseline = useMemo(
    () => (baselineResource ? pickConfigurationDraft(baselineResource) : undefined),
    [baselineResource],
  );
  const deploymentDraft = useMemo(
    () => pickDeploymentDraft(resource),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [resource.init_spec, resource.execution_config?.command, resource.execution_config?.args],
  );
  const deploymentBaseline = useMemo(
    () => (baselineResource ? pickDeploymentDraft(baselineResource) : undefined),
    [baselineResource],
  );

  const envVars = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];
  const baselineEnvVars = baselineResource?.execution_config?.environment_variables as
    | FormEnvVarData[]
    | undefined;

  const thisResourceName = resource.name ?? `Resource ${index + 1}`;
  const resourceOptions = useMemo(
    () =>
      (context.allResources ?? [])
        .filter((r) => r.name !== thisResourceName)
        .map((r) => ({ name: r.name, outputs: r.outputs })),
    [context.allResources, thisResourceName],
  );
  const selfOutputs = useMemo(() => (resource.outputs ?? []).map((o: { name: string }) => o.name), [resource.outputs]);

  return {
    dirtyTabs,
    isDirty,
    statusDotColor,
    statusState: statusObj.state,
    configurationProps: {
      index,
      draft: configurationDraft,
      baseline: configurationBaseline,
      errors: context.errors,
      volumes: context.volumes ?? [],
      allResources: context.allResources,
      secrets: context.secrets,
      onDiscardField: context.onDiscardField,
      onPatchResource,
      onCreateVolume: context.onCreateVolume,
    },
    deploymentProps: {
      index,
      draft: deploymentDraft,
      baseline: deploymentBaseline,
      onPatchInitSpec,
      onPatchExecCommandArgs,
      onDiscardField: context.onDiscardField,
    },
    environmentProps: {
      index,
      envVars,
      baselineEnvVars,
      errors: context.errors,
      resourceOptions,
      selfOutputs,
      secrets: context.secrets.secrets,
      secretsLoading: context.secrets.isLoading,
      addons: context.addons,
      addonNameById: context.addonNameById,
      onChangeEnvVars,
      onDiscardEnvRow: context.onDiscardEnvRow,
    },
  };
}
