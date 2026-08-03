import { AlertTriangle, Box, Ellipsis, GitBranch, Layers, Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { statusVariant, statusVariantTone, type StatusVariant } from "@/components/branded/status-variant";
import {
  StatusRail,
  CardMetaGrid,
  CardFooterMeta,
  relativeAge,
  absoluteAge,
  type RailTone,
} from "@/components/branded/entity-card";
import { deriveHeaderHealth, latestDeployFailed } from "@/pages/stacks/components/editor/tabs/deployments/derive";
import type { Stack } from "@/api/stack-types";

function inferStackIcon(stack: Stack) {
  // Pick an icon based on the first resource's source type
  const first = stack.spec?.stack_resources?.[0];
  if (!first) return Layers;
  if (first.source?.git) return GitBranch;
  if (first.source?.image) return Box;
  return Layers;
}

/** Card status: "Deleting" overrides health; no derivable health (never
 *  deployed, or only cancelled/superseded attempts) reads "Not deployed";
 *  otherwise the release health rollup drives it — mirrors the stack detail
 *  page's header. */

export function headerStatus(stack: Stack): { label: string; variant: StatusVariant } {
  if (stack.lifecycle === "deleting") return { label: "Deleting", variant: "pending" };
  const health = deriveHeaderHealth(stack);
  if (!health) return { label: "Not deployed", variant: "neutral" };
  return { label: health, variant: statusVariant("health", health) };
}

/**
 * Deployed-stack dashboard card, "card grammar" layout: header (icon · name ·
 * kebab), META grid, FOOTER (status · age). Status comes from the
 * release-health rollup (headerStatus). Live ingress URLs live on release
 * detail, not the list payload, so cards carry no endpoint pills (deliberate —
 * see the release-centric API thinning; inline `endpoints` rollup is a
 * possible future backend change).
 */
export function DeployStackCard({ stack, onDelete }: { stack: Stack; onDelete?: (stack: Stack) => void }) {
  const navigate = useNavigate();
  const Icon = inferStackIcon(stack);
  const { label: word, variant } = headerStatus(stack);
  const tone: RailTone | "neutral" = variant === "neutral" ? "neutral" : statusVariantTone[variant];
  const deployFailed = latestDeployFailed(stack);
  const resourceCount = stack.spec?.stack_resources?.length || 0;
  const volumeCount = stack.spec?.volumes?.length || 0;
  const branch = stack.spec?.stack_resources?.find((r) => r.source?.git)?.source?.git?.branch;
  const age = relativeAge(stack.updated_at || stack.created_at);
  const menuDisabled = stack.lifecycle === "deleting";
  return (
    <Card
      role="link"
      tabIndex={0}
      aria-label={`${stack.name} stack`}
      onClick={() => navigate(`/stacks/${stack.id}`)}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/stacks/${stack.id}`);
      }}
      className="group flex h-[210px] w-full cursor-pointer flex-col gap-0 overflow-hidden p-0 transition-colors duration-150 hover:bg-foreground/[0.04] focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2"
    >
      {tone !== "neutral" && <StatusRail tone={tone} />}
      <div className="flex flex-1 flex-col gap-3.5 p-5">
        <div className="flex items-center gap-[11px]">
          <Icon className="h-[18px] w-[18px] flex-none text-fg-2" strokeWidth={1.6} />
          <span
            className="mr-auto truncate text-base font-medium tracking-[-0.01em]"
            title={stack.name}
          >
            {stack.name}
          </span>
          {onDelete && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Actions for ${stack.name}`}
                  className="flex-none"
                  onClick={(e) => e.stopPropagation()}
                >
                  <Ellipsis className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                side="right"
                align="start"
                className="w-[160px]"
                onClick={(e) => e.stopPropagation()}
              >
                {/* No deferral needed: the confirm service defers its own open
                    a tick past the menu close (radix-ui/primitives#1836). */}
                <DropdownMenuItem
                  variant="destructive"
                  disabled={menuDisabled}
                  onSelect={() => onDelete(stack)}
                >
                  <Trash2 className="h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>

        <div className="mt-auto flex flex-col gap-3.5">
          <CardMetaGrid
            rows={[
              branch ? { label: "branch", value: branch } : null,
              { label: "resources", value: String(resourceCount) },
              { label: "volumes", value: String(volumeCount) },
            ]}
          />
          <CardFooterMeta
            tone={tone}
            word={word}
            age={age}
            ageTitle={absoluteAge(stack.updated_at || stack.created_at)}
            alert={
              deployFailed ? (
                <AlertTriangle className="h-3 w-3 flex-none text-danger" aria-label="Latest deploy failed" />
              ) : undefined
            }
          />
        </div>
      </div>
    </Card>
  );
}
