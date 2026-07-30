import React from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { DirtyField } from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/dirty-field";
import {
  LedgerRow,
  LedgerSection,
} from "@/pages/stacks/components/editor/tabs/architecture/drawer-tabs/ledger";

import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

type Resource = Partial<FormStackResourceData>;

interface StackResourceDeploymentTabProps {
  index: number;
  /** Projected slice of `resource` containing only the fields this tab reads. */
  draft: DeploymentDraft;
  baseline: DeploymentDraft | undefined;
  /** Patch the resource's `init_spec`. Identity must be stable across renders. */
  onPatchInitSpec: (patch: Partial<NonNullable<FormStackResourceData["init_spec"]>>) => void;
  /** Patch the command/args fields of `execution_config`, preserving other
   *  nested keys (notably `environment_variables`). Identity must be stable. */
  onPatchExecCommandArgs: (patch: { command?: string; args?: string }) => void;
  onDiscardField?: (path: string) => void;
}

/** Subset of FormStackResourceData read by the Deployment tab. Notably we
 *  strip out execution_config.environment_variables — those drive the
 *  Environment tab and would invalidate this tab's memo on every env-var
 *  keystroke if included here. */
export interface DeploymentDraft {
  init_spec?: Resource["init_spec"];
  /** execution_config WITHOUT environment_variables. Shape preserved so that
   *  <DirtyField path="execution_config.command"> resolves correctly. */
  execution_config?: Omit<NonNullable<Resource["execution_config"]>, "environment_variables">;
}

export function pickDeploymentDraft(resource: Resource): DeploymentDraft {
  const ec = resource.execution_config;

  const stripped = ec ? (() => { const { environment_variables, ...rest } = ec; return rest; })() : undefined;
  return {
    init_spec: resource.init_spec,
    execution_config: stripped,
  };
}

function StackResourceDeploymentTabImpl({
  index,
  draft,
  baseline,
  onPatchInitSpec,
  onPatchExecCommandArgs,
  onDiscardField,
}: StackResourceDeploymentTabProps) {

  return (
    <TabsContent value="deployment" className="pt-1">
      <LedgerSection label="Pre-deployment step" meta="runs before the main container">
        <LedgerRow
          label="Init command"
          htmlFor={`init-command-${index}`}
          alignTop
          hint="Type as in a terminal; quotes group arguments. Not run in a shell — $VARS won't expand."
        >
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="init_spec.command"
            compact
            onReset={onDiscardField ? () => onDiscardField("init_spec.command") : undefined}
          >
            <Input
              id={`init-command-${index}`}
              value={draft.init_spec?.command ?? ""}
              onChange={(e) => onPatchInitSpec({ command: e.target.value })}
              placeholder="e.g., sh /scripts/init.sh"
              className="h-9 font-mono text-[12.5px]"
            />
          </DirtyField>
        </LedgerRow>
        <LedgerRow label="Init arguments" htmlFor={`init-args-${index}`}>
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="init_spec.args"
            compact
            onReset={onDiscardField ? () => onDiscardField("init_spec.args") : undefined}
          >
            <Input
              id={`init-args-${index}`}
              value={draft.init_spec?.args ?? ""}
              onChange={(e) => onPatchInitSpec({ args: e.target.value })}
              placeholder="e.g., arg1 arg2 arg3"
              className="h-9 font-mono text-[12.5px]"
            />
          </DirtyField>
        </LedgerRow>
      </LedgerSection>

      <LedgerSection label="Main container step">
        <LedgerRow
          label="Command"
          htmlFor={`exec-command-${index}`}
          alignTop
          hint="Overrides the container's default ENTRYPOINT. Type as in a terminal; quotes group arguments."
        >
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="execution_config.command"
            compact
            onReset={onDiscardField ? () => onDiscardField("execution_config.command") : undefined}
          >
            <Input
              id={`exec-command-${index}`}
              value={draft.execution_config?.command ?? ""}
              onChange={(e) => onPatchExecCommandArgs({ command: e.target.value })}
              placeholder="e.g., node server.js"
              className="h-9 font-mono text-[12.5px]"
            />
          </DirtyField>
        </LedgerRow>
        <LedgerRow label="Arguments" htmlFor={`exec-args-${index}`}>
          <DirtyField
            draft={draft}
            baseline={baseline}
            path="execution_config.args"
            compact
            onReset={onDiscardField ? () => onDiscardField("execution_config.args") : undefined}
          >
            <Input
              id={`exec-args-${index}`}
              value={draft.execution_config?.args ?? ""}
              onChange={(e) => onPatchExecCommandArgs({ args: e.target.value })}
              placeholder="e.g., --port=3000 --verbose"
              className="h-9 font-mono text-[12.5px]"
            />
          </DirtyField>
        </LedgerRow>
      </LedgerSection>
    </TabsContent>
  );
}

export const StackResourceDeploymentTab = React.memo(StackResourceDeploymentTabImpl);
