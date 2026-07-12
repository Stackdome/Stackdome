import { Ellipsis, GitPullRequest, RefreshCw, Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  previewStatusVariant,
  statusVariantLabel,
  statusVariantTone,
} from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";
import {
  StatusRail,
  StatusWord,
  EndpointPills,
  CardFooterMeta,
  relativeAge,
  type RailTone,
} from "@/pages/stacks/components/list/stack-card";
import type { PreviewStack, PreviewPhase } from "@/api/preview-envs";

interface PreviewEnvCardProps {
  env: PreviewStack;
  /** Repository configuration name shown in the identity block. */
  configName?: string;
  onSync?: (env: PreviewStack) => void;
  onDelete?: (env: PreviewStack) => void;
}

function previewTone(phase: PreviewPhase | undefined): { tone: RailTone; word: string } {
  const v = previewStatusVariant(phase);
  return { tone: statusVariantTone[v], word: statusVariantLabel[v] };
}

/**
 * Preview environment card, "Status Strip" design. The card navigates to the
 * underlying stack; sync/delete live in the kebab menu. Not wrapped in a Link
 * because the endpoint pills (and the kebab trigger) inside are real
 * interactive elements — nested <a>/<button> is invalid HTML.
 */
export function PreviewEnvCard({ env, configName, onSync, onDelete }: PreviewEnvCardProps) {
  const navigate = useNavigate();
  const phase = env.status?.phase;
  const variant = previewStatusVariant(phase);
  const { tone, word } = previewTone(phase);
  const urls = env.status?.outputs?.urls ?? [];
  const age = relativeAge(env.updated_at || env.created_at);
  const clickable = Boolean(env.stack_id);
  const reason = env.status?.reason;
  const failed = variant === "error" && Boolean(reason);
  const stackfileHint = failed && /stackfile/i.test(reason ?? "");
  const menuDisabled = phase === "Deleting";

  const goToStack = () => navigate(`/stacks/${env.stack_id}`);

  return (
    <Card
      role={clickable ? "link" : undefined}
      tabIndex={clickable ? 0 : undefined}
      aria-label={`PR #${env.pr_number} preview environment`}
      onClick={clickable ? goToStack : undefined}
      onKeyDown={
        clickable
          ? (e) => {
            if (e.key === "Enter") goToStack();
          }
          : undefined
      }
      className={cn(
        "group flex h-[210px] w-full flex-col gap-0 overflow-hidden p-0 transition-colors duration-150 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-brand/40",
        clickable && "cursor-pointer hover:border-brand-border hover:bg-muted/20",
      )}
    >
      <StatusRail tone={tone} />
      <div className="flex flex-1 flex-col gap-3.5 p-5">
        <div className="flex items-center gap-[11px]">
          <GitPullRequest className="h-[18px] w-[18px] flex-none text-brand" strokeWidth={1.6} />
          <span className="mr-auto truncate text-base font-medium tracking-[-0.01em] transition-colors group-hover:text-brand">
            PR #{env.pr_number}
          </span>
          <StatusWord tone={tone}>{word}</StatusWord>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Actions for PR #${env.pr_number}`}
                className="h-7 w-7 flex-none"
                onClick={(e) => e.stopPropagation()}
              >
                <Ellipsis className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              className="w-[160px]"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Both items open a downstream dialog. Radix's DropdownMenu→Dialog
                  composition races the menu's close (which resets
                  document.body.style.pointerEvents) against the dialog's mount, and
                  can leave pointer-events "none" on body forever if the dialog is
                  cancelled. Deferring the callback until after the menu has fully
                  closed avoids the race. See https://github.com/radix-ui/primitives/issues/1836 */}
              <DropdownMenuItem
                disabled={menuDisabled}
                onSelect={() => setTimeout(() => onSync?.(env), 0)}
              >
                <RefreshCw className="h-4 w-4" />
                Sync
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                disabled={menuDisabled}
                onSelect={() => setTimeout(() => onDelete?.(env), 0)}
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Identity: repo / branch rows */}
        <div className="grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-1.5">
          {(configName ?? env.name) && (
            <>
              <span className="font-mono text-[9px] uppercase tracking-[1.2px] text-fg-muted">repo</span>
              <span className="truncate font-mono text-[11px] text-fg-2">{configName ?? env.name}</span>
            </>
          )}
          {env.branch && (
            <>
              <span className="font-mono text-[9px] uppercase tracking-[1.2px] text-fg-muted">branch</span>
              <span className="truncate font-mono text-[11px] text-brand">{env.branch}</span>
            </>
          )}
        </div>

        {/* Bottom-anchored group: pills (or the failed-reason strip) sit just
            above the footer in every card variant. */}
        <div className="mt-auto flex flex-col gap-3.5">
          {failed ? (
            <div className="rounded-sm bg-danger-bg px-2 py-1.5 text-xs text-danger line-clamp-2">
              <span className="font-semibold">{reason}:</span> {env.status?.message}
              {stackfileHint && (
                <span className="block">Check the stackfile path in Settings.</span>
              )}
            </div>
          ) : (
            <EndpointPills urls={urls} />
          )}
          <CardFooterMeta commit={env.commit ? env.commit.slice(0, 7) : null} age={age} />
        </div>
      </div>
    </Card>
  );
}
