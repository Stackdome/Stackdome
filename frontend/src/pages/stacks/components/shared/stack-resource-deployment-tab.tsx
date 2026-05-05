import React from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { DirtyField } from "@/pages/stacks/components/shared/dirty-field";
import { FieldShell } from "@/components/branded";

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
  onPatchExecCommandArgs: (patch: { command?: string[]; args?: string[] }) => void;
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
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
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
    <TabsContent value="deployment" className="pt-4 space-y-6">
      {/* Pre-Deploy Section (Init) */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Pre-Deployment step</h3>
        <div className="grid gap-5 max-w-3xl">
          <FieldShell
            label="Init Command"
            htmlFor={`init-command-${index}`}
            hint="Runs before the main container starts. Comma-separate the executable and its segments."
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="init_spec.command"
              onReset={onDiscardField ? () => onDiscardField("init_spec.command") : undefined}
            >
              <Input
                id={`init-command-${index}`}
                value={draft.init_spec?.command?.join(",") || ""}
                onChange={(e) =>
                  onPatchInitSpec({
                    command: e.target.value.split(",").map((s) => s.trim()).filter(Boolean),
                  })
                }
                placeholder="e.g., sh,/scripts/init.sh"
              />
            </DirtyField>
          </FieldShell>
          <FieldShell
            label="Init Arguments"
            htmlFor={`init-args-${index}`}
            hint="Comma-separated arguments passed to the init command."
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="init_spec.args"
              onReset={onDiscardField ? () => onDiscardField("init_spec.args") : undefined}
            >
              <Input
                id={`init-args-${index}`}
                value={draft.init_spec?.args?.join(",") || ""}
                onChange={(e) =>
                  onPatchInitSpec({
                    args: e.target.value.split(",").map((s) => s.trim()).filter(Boolean),
                  })
                }
                placeholder="e.g., arg1,arg2,arg3"
              />
            </DirtyField>
          </FieldShell>
        </div>
      </div>
      <Separator className="my-6" />
      {/* Post-Deploy Section (Execution) */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Main container step</h3>
        <div className="grid gap-5 max-w-3xl">
          <FieldShell
            label="Command"
            htmlFor={`exec-command-${index}`}
            hint="Overrides the container's default ENTRYPOINT. Comma-separate the executable and its segments."
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="execution_config.command"
              onReset={onDiscardField ? () => onDiscardField("execution_config.command") : undefined}
            >
              <Input
                id={`exec-command-${index}`}
                value={draft.execution_config?.command?.join(",") || ""}
                onChange={(e) =>
                  onPatchExecCommandArgs({
                    command: e.target.value.split(",").map((s) => s.trim()).filter(Boolean),
                  })
                }
                placeholder="e.g., node,server.js"
              />
            </DirtyField>
          </FieldShell>
          <FieldShell
            label="Arguments"
            htmlFor={`exec-args-${index}`}
            hint="Comma-separated arguments passed to the command."
          >
            <DirtyField
              draft={draft}
              baseline={baseline}
              path="execution_config.args"
              onReset={onDiscardField ? () => onDiscardField("execution_config.args") : undefined}
            >
              <Input
                id={`exec-args-${index}`}
                value={draft.execution_config?.args?.join(",") || ""}
                onChange={(e) =>
                  onPatchExecCommandArgs({
                    args: e.target.value.split(",").map((s) => s.trim()).filter(Boolean),
                  })
                }
                placeholder="e.g., --port=3000,--verbose"
              />
            </DirtyField>
          </FieldShell>
        </div>
      </div>
    </TabsContent>
  );
}

export const StackResourceDeploymentTab = React.memo(StackResourceDeploymentTabImpl);
