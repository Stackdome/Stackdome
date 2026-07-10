import { useCallback, useMemo } from "react";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { X, ScrollText, Trash2 } from "lucide-react";
import { useSecrets } from "@/pages/stacks/hooks/use-secrets";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import type { PostgresAddon } from "@/api/addons";
import type { UseStackEditSession, EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";
import { StackResourceConfigurationTab } from "@/pages/stacks/components/shared/stack-resource-configuration-tab";
import { StackResourceDeploymentTab } from "@/pages/stacks/components/shared/stack-resource-deployment-tab";
import { StackResourceEnvironmentTab } from "@/pages/stacks/components/shared/stack-resource-environment-tab";
import { useResourceTabProps } from "@/pages/stacks/components/shared/hooks/use-resource-tab-props";
import { nodePresentation } from "@/pages/stacks/lib/canvas/node-presentation";
import { NodeGlyph } from "./nodes/node-glyph";

/** Radix tab values used by the sub-tab components (they render their own TabsContent). */
const TAB_VALUE = { configuration: "general", deployment: "deployment", environment: "environment" } as const;
const TAB_FROM_VALUE: Record<string, EditSessionTab> = {
  general: "configuration",
  deployment: "deployment",
  environment: "environment",
};

interface ResourceDrawerProps {
  /** Index into `session.draft.resources` of the resource being configured. */
  resourceIndex: number;
  session: UseStackEditSession;
  baselineResources: Partial<FormStackResourceData>[];
  /** Server-computed resource outputs keyed by resource name. The draft copy of a
   *  newly-added resource has no outputs, so the env-var output pickers read from
   *  this server-truth map (matched by name) instead. */
  serverOutputsByName?: ReadonlyMap<string, string[]>;
  /** Addon ids linked to the stack — filters the addon picker + drives bindings. */
  connectionAddonIds: ReadonlySet<string>;
  errors: { [field: string]: string | undefined };
  onClose: () => void;
  onRemove: (index: number) => void;
  /** Open the stack's Logs view (from the footer "View logs"). */
  onViewLogs?: () => void;
  /** Push a volume's drawer onto the floating drawer stack. */
  onOpenVolume?: (name: string) => void;
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
  serverOutputsByName,
  connectionAddonIds,
  errors,
  onClose,
  onRemove,
  onViewLogs,
  onOpenVolume,
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
  // Keep the draft's resource name/index/order (preserves unsaved edits &
  // ordering) but source outputs from the server-truth map by name — a
  // resource added on the canvas has none in its draft copy until saved.
  const allResources = useMemo(
    () =>
      session.draft.resources.map((r, i) => ({
        name: r.name || `Resource ${i + 1}`,
        index: i,
        outputs:
          serverOutputsByName?.get(r.name ?? "") ??
          (r.outputs ?? []).map((o: { name: string }) => o.name),
      })),
    [session.draft.resources, serverOutputsByName],
  );

  // Replace just this resource in the draft.
  const onChange = useCallback(
    (index: number, updated: Partial<FormStackResourceData>) => {
      session.updateResources((prev) => prev.map((r, i) => (i === index ? updated : r)));
    },
    [session],
  );

  const { dirtyTabs, isDirty, statusDotColor, configurationProps, deploymentProps, environmentProps } =
    useResourceTabProps({
      resource,
      index: resourceIndex,
      baselineResource,
      onChange,
      context: {
        errors,
        // Draft volumes, not baseline: an inline-added volume must be pickable
        // (and resolvable by existing mount rows) before the next autosave
        // cycle advances the baseline.
        volumes: session.draft.volumes,
        allResources,
        serverOutputsByName,
        secrets,
        addons,
        addonNameById,
        onDiscardField: (path) => session.discardResourceField(resourceIndex, path),
        onDiscardEnvRow: (envIdx) => session.discardEnvRow(resourceIndex, envIdx),
        mountsReadOnly: true,
        onOpenVolume,
      },
    });

  // Controlled tab: driven by session.openTab so a banner "jump to error" can
  // switch to the tab holding the offending field even while the drawer is open.
  const activeTab = session.openTab ? TAB_VALUE[session.openTab] : TAB_VALUE.configuration;

  // Kind glyph + summary sub-line, derived the same way the node card is.
  const pres = useMemo(
    () =>
      nodePresentation({
        isAddon: false,
        image: resource.source?.image?.ref,
        hasBuild: !!resource.source?.git,
        ports: (resource.ports ?? []).map((p) => ({
          number: p.number,
          protocol: p.protocol,
          exposedToPublic: p.exposed_to_public,
        })),
      }),
    [resource.source?.image?.ref, resource.source?.git, resource.ports],
  );

  // "N changes" counts the dirty sub-tabs (config / deployment / environment).
  const changeCount = [dirtyTabs.configuration, dirtyTabs.deployment, dirtyTabs.environment].filter(Boolean).length;

  const tabTriggerClass =
    "flex-none rounded-none border-b-[1.5px] border-transparent bg-transparent px-[13px] py-3 text-sm font-medium text-fg-muted data-[state=active]:border-brand data-[state=active]:text-foreground data-[state=active]:shadow-none";

  return (
    <div
      className="flex h-full w-full flex-col bg-background"
      data-testid="resource-drawer"
    >
      {/* Header */}
      <div className="flex flex-none items-center gap-3 border-b border-border px-4 py-[15px]">
        <span className={`size-[9px] shrink-0 rounded-full ${statusDotColor}`} aria-hidden />
        <NodeGlyph glyph={pres.glyph} className="size-[19px] shrink-0 text-brand" />
        <div className="min-w-0 flex-1 leading-tight">
          <div className="truncate text-base font-medium text-foreground">
            {resource.name || `Resource ${resourceIndex + 1}`}
          </div>
          <div className="truncate font-mono text-[11px] text-fg-muted">{pres.summary}</div>
        </div>
        {isDirty ? (
          <span className="flex shrink-0 items-center gap-1 rounded-md border border-brand pl-2 pr-1 py-0.5 text-[11px] font-medium text-brand">
            {changeCount === 1 ? "1 change" : `${changeCount} changes`}
            <button
              type="button"
              onClick={() => session.discardResource(resourceIndex)}
              aria-label="Discard changes to this resource"
              title="Discard changes"
              className="flex size-4 items-center justify-center rounded-sm hover:bg-brand hover:text-background"
            >
              <X className="size-3" />
            </button>
          </span>
        ) : (
          <span className="shrink-0 font-mono text-[9px] uppercase tracking-[0.12em] text-fg-muted">
            {pres.kindLabel}
          </span>
        )}
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="shrink-0 rounded p-1 text-fg-muted hover:bg-muted hover:text-foreground"
        >
          <X className="size-[18px]" />
        </button>
      </div>

      {/* Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={(v) => session.setOpenTab(TAB_FROM_VALUE[v] ?? "configuration")}
        className="flex min-h-0 flex-1 flex-col"
      >
        <TabsList className="h-auto w-full flex-none justify-start gap-1 rounded-none border-b border-border bg-transparent p-0 px-1">
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

        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
          <StackResourceConfigurationTab {...configurationProps} />
          <StackResourceDeploymentTab {...deploymentProps} />
          <StackResourceEnvironmentTab {...environmentProps} />
        </div>
      </Tabs>

      {/* Footer */}
      <div className="flex flex-none items-center justify-between border-t border-border px-4 py-[11px]">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 px-2 text-[12.5px] text-fg-muted hover:bg-brand-bg hover:text-brand"
          disabled={!onViewLogs}
          onClick={onViewLogs}
        >
          <ScrollText className="size-3.5" />
          View logs
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 px-2 text-[12.5px] text-danger hover:bg-danger-bg hover:text-danger"
          onClick={() => onRemove(resourceIndex)}
        >
          <Trash2 className="size-3.5" />
          Remove resource
        </Button>
      </div>
    </div>
  );
}
