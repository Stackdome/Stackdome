import { useCallback, useMemo } from "react";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { X, ScrollText, Trash2 } from "lucide-react";
import { useSecrets } from "@/pages/stacks/hooks/use-secrets";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import type {
  FormStackResourceData,
  FormVolumeExtendedData as VolumeFormData,
} from "@/pages/stacks/schemas/form-schema";
import { StackResourceConfigurationTab } from "@/pages/stacks/components/shared/stack-resource-configuration-tab";
import { StackResourceDeploymentTab } from "@/pages/stacks/components/shared/stack-resource-deployment-tab";
import { StackResourceEnvironmentTab } from "@/pages/stacks/components/shared/stack-resource-environment-tab";
import { useResourceTabProps } from "@/pages/stacks/components/shared/hooks/use-resource-tab-props";
import { NodeGlyph } from "./nodes/node-glyph";

/** Radix tab values used by the sub-tab components (they render their own TabsContent). */
const TAB_VALUE = { configuration: "general", deployment: "deployment", environment: "environment" } as const;

interface ResourceDrawerProps {
  /** Index into `session.draft.resources` of the resource being configured. */
  resourceIndex: number;
  session: UseStackEditSession;
  baselineResources: Partial<FormStackResourceData>[];
  baselineVolumes: Partial<VolumeFormData>[];
  /** Addon ids linked to the stack — filters the addon picker + drives bindings. */
  connectionAddonIds: ReadonlySet<string>;
  errors: { [field: string]: string | undefined };
  onClose: () => void;
  onRemove: (index: number) => void;
}

/**
 * Slide-in drawer for a selected canvas node. It wraps the SAME three sub-tab
 * bodies the accordion form uses (via `useResourceTabProps`), so env-var
 * grouping and per-tab dirty marks come for free. Edits flow straight into the
 * edit session; the drawer owns no stack state.
 */
export function ResourceDrawer({
  resourceIndex,
  session,
  baselineResources,
  baselineVolumes,
  connectionAddonIds,
  errors,
  onClose,
  onRemove,
}: ResourceDrawerProps) {
  const resource = session.draft.resources[resourceIndex] ?? {};
  const baselineResource = baselineResources[resourceIndex];

  const secrets = useSecrets();
  const { addons: allAddons } = usePostgresAddons();
  const addons = useMemo(
    () => allAddons.filter((a: PostgresAddon) => a.id && connectionAddonIds.has(a.id)),
    [allAddons, connectionAddonIds],
  );
  const addonNameById = useMemo(
    () =>
      new Map(
        allAddons
          .filter((a: PostgresAddon) => a.id && a.name)
          .map((a: PostgresAddon) => [a.id!, a.name!] as [string, string]),
      ),
    [allAddons],
  );
  const allResources = useMemo(
    () =>
      session.draft.resources.map((r, i) => ({
        name: r.name || `Resource ${i + 1}`,
        index: i,
        outputs: (r.outputs ?? []).map((o: { name: string }) => o.name),
      })),
    [session.draft.resources],
  );

  // Replace just this resource in the draft.
  const onChange = useCallback(
    (index: number, updated: Partial<FormStackResourceData>) => {
      session.updateResources((prev) => prev.map((r, i) => (i === index ? updated : r)));
    },
    [session],
  );

  const { dirtyTabs, isDirty, statusDotColor, statusState, configurationProps, deploymentProps, environmentProps } =
    useResourceTabProps({
      resource,
      index: resourceIndex,
      baselineResource,
      onChange,
      context: {
        errors,
        volumes: baselineVolumes,
        allResources,
        secrets,
        addons,
        addonNameById,
        onDiscardField: (path) => session.discardResourceField(resourceIndex, path),
        onDiscardEnvRow: (envIdx) => session.discardEnvRow(resourceIndex, envIdx),
      },
    });

  const defaultTab = session.openTab ? TAB_VALUE[session.openTab] : TAB_VALUE.configuration;

  const tabTriggerClass =
    "flex-none rounded-none border-b-2 border-transparent bg-transparent px-0 py-2 font-mono text-[11px] uppercase tracking-wider text-fg-muted data-[state=active]:border-brand data-[state=active]:text-foreground data-[state=active]:shadow-none";

  return (
    <aside
      className="absolute right-0 top-0 z-10 flex h-full w-[380px] flex-col border-l border-border bg-card shadow-lg"
      data-testid="resource-drawer"
    >
      {/* Header */}
      <div className="flex items-start gap-2 border-b border-border px-4 py-3">
        <span className={`mt-1.5 size-1.5 shrink-0 rounded-full ${statusDotColor}`} aria-hidden />
        <NodeGlyph glyph="service" className="mt-0.5 size-4 shrink-0 text-fg-muted" />
        <div className="min-w-0 flex-grow">
          <div className="truncate text-sm font-medium text-foreground">
            {resource.name || `Resource ${resourceIndex + 1}`}
          </div>
          <div className="font-mono text-[11px] capitalize text-muted-foreground">{statusState || "pending"}</div>
        </div>
        {isDirty && (
          <span className="shrink-0 rounded-md border border-brand-border bg-brand-bg px-1.5 py-0.5 text-[10px] font-medium text-brand">
            Modified
          </span>
        )}
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="shrink-0 rounded p-1 text-fg-muted hover:bg-muted hover:text-foreground"
        >
          <X className="size-4" />
        </button>
      </div>

      {/* Tabs */}
      <div className="flex-grow overflow-y-auto px-4 py-3">
        <Tabs defaultValue={defaultTab} className="w-full">
          <TabsList className="mb-3 h-auto w-full justify-start gap-4 rounded-none border-b border-border bg-transparent p-0">
            <TabsTrigger value={TAB_VALUE.configuration} className={tabTriggerClass}>
              Configuration
              {dirtyTabs.configuration && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
            </TabsTrigger>
            <TabsTrigger value={TAB_VALUE.deployment} className={tabTriggerClass}>
              Deployment
              {dirtyTabs.deployment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
            </TabsTrigger>
            <TabsTrigger value={TAB_VALUE.environment} className={tabTriggerClass}>
              Environment
              {dirtyTabs.environment && <span aria-hidden className="ml-1.5 inline-block size-1.5 rounded-full bg-brand" />}
            </TabsTrigger>
          </TabsList>

          <StackResourceConfigurationTab {...configurationProps} />
          <StackResourceDeploymentTab {...deploymentProps} />
          <StackResourceEnvironmentTab {...environmentProps} />
        </Tabs>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between border-t border-border px-4 py-2.5">
        <Button type="button" variant="ghost" size="sm" className="h-7 px-2 text-[12px] text-muted-foreground" disabled title="Coming soon">
          <ScrollText className="size-3.5" />
          View logs
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-[12px] text-muted-foreground/70 hover:bg-danger-bg hover:text-danger"
          onClick={() => onRemove(resourceIndex)}
        >
          <Trash2 className="size-3.5" />
          Remove resource
        </Button>
      </div>
    </aside>
  );
}
