import React from "react";
import { TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Info } from "lucide-react";
import { DirtyField } from "@/pages/stacks/components/shared/dirty-field";

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
        <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Pre-Deployment step</h3>
        <div className="grid gap-4 max-w-3xl">
          <div>
            <div className="flex items-center gap-1 mb-2">
              <Label htmlFor={`init-command-${index}`} className="text-sm font-medium">
                Init Command
              </Label>
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                </TooltipTrigger>
                <TooltipContent side="top">Pre-deployment init command (comma separated)</TooltipContent>
              </Tooltip>
            </div>
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
          </div>
          <div>
            <div className="flex items-center gap-1 mb-2">
              <Label htmlFor={`init-args-${index}`} className="text-sm font-medium">
                Init Arguments
              </Label>
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                </TooltipTrigger>
                <TooltipContent side="top">Pre-deployment arguments (comma separated)</TooltipContent>
              </Tooltip>
            </div>
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
          </div>
        </div>
      </div>
      <Separator className="my-4" />
      {/* Post-Deploy Section (Execution) */}
      <div>
        <h3 className="text-xs font-semibold text-muted-foreground mb-2.5">Main container step</h3>
        <div className="grid gap-4 max-w-3xl">
          <div>
            <div className="flex items-center gap-1 mb-2">
              <Label htmlFor={`exec-command-${index}`} className="text-sm font-medium">
                Command
              </Label>
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                </TooltipTrigger>
                <TooltipContent side="top">Container runtime command (comma separated)</TooltipContent>
              </Tooltip>
            </div>
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
          </div>
          <div>
            <div className="flex items-center gap-1 mb-2">
              <Label htmlFor={`exec-args-${index}`} className="text-sm font-medium">
                Arguments
              </Label>
              <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                </TooltipTrigger>
                <TooltipContent side="top">Container runtime arguments (comma separated)</TooltipContent>
              </Tooltip>
            </div>
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
          </div>
        </div>
      </div>
    </TabsContent>
  );
}

export const StackResourceDeploymentTab = React.memo(StackResourceDeploymentTabImpl);
