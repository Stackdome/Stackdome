import { Ellipsis, GitCommitHorizontal, GitPullRequest, RefreshCw, Trash2 } from "lucide-react";
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
  EndpointPills,
  CardFooterMeta,
  CardMetaGrid,
  relativeAge,
  absoluteAge,
  type RailTone,
} from "@/pages/stacks/components/list/stack-card";
import type { PreviewStack, PreviewPhase } from "@/api/preview-envs";

interface PreviewEnvCardProps {
  env: PreviewStack;
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
export function PreviewEnvCard({ env, onSync, onDelete }: PreviewEnvCardProps) {
  const navigate = useNavigate();
  const phase = env.status?.phase;
  const { tone, word } = previewTone(phase);
  const urls = env.status?.outputs?.urls ?? [];
  const age = relativeAge(env.updated_at || env.created_at);
  const clickable = Boolean(env.stack_id);
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
              side="right"
              align="start"
              className="w-[160px]"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Deferred so the menu finishes closing before the dialog mounts — see radix-ui/primitives#1836 (canonical note in row-menu.tsx). */}
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

        <CardMetaGrid
          rows={[
            env.branch ? { label: "branch", value: env.branch } : null,
            env.commit
              ? {
                label: "commit",
                value: (
                  <span className="inline-flex items-center gap-1.5 tabular-nums">
                    <GitCommitHorizontal className="h-3.5 w-3.5 flex-none" strokeWidth={1.6} />
                    {env.commit.slice(0, 7)}
                  </span>
                ),
              }
              : null,
          ]}
        />

        {/* Bottom-anchored group: pills sit just above the footer. They reveal
            on hover/focus so the resting card stays quiet; the slot is always
            laid out, so nothing jumps. Failure details live on the stack
            detail page, not the card — the FAILED status word is the signal. */}
        <div className="mt-auto flex flex-col gap-3.5">
          <div className="opacity-0 transition-opacity duration-150 group-hover:opacity-100 group-focus-within:opacity-100 focus-within:opacity-100">
            <EndpointPills urls={urls} />
          </div>
          <CardFooterMeta
            tone={tone}
            word={word}
            age={age}
            ageTitle={absoluteAge(env.updated_at || env.created_at)}
          />
        </div>
      </div>
    </Card>
  );
}
