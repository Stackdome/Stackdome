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
  const hasLinks = urls.some((u) => u.url);
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
        "group flex h-[210px] w-full flex-col gap-0 overflow-hidden p-0 transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2",
        clickable && "cursor-pointer hover:bg-foreground/[0.04]",
      )}
    >
      <StatusRail tone={tone} />
      <div className="flex flex-1 flex-col gap-3.5 p-5">
        <div className="flex items-center gap-[11px]">
          <GitPullRequest className="h-[18px] w-[18px] flex-none text-fg-2" strokeWidth={1.6} />
          <span className="mr-auto truncate text-base font-medium tracking-[-0.01em]">
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
              {onSync && (
                <DropdownMenuItem
                  disabled={menuDisabled}
                  onSelect={() => setTimeout(() => onSync(env), 0)}
                >
                  <RefreshCw className="h-4 w-4" />
                  Sync
                </DropdownMenuItem>
              )}
              {/* No deferral needed: the confirm service defers its own open a
                  tick past the menu close. onSync above keeps its setTimeout —
                  it opens a plain Dialog directly. */}
              {onDelete && (
                <DropdownMenuItem
                  variant="destructive"
                  disabled={menuDisabled}
                  onSelect={() => onDelete(env)}
                >
                  <Trash2 className="h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Bottom-anchored group above the footer. The meta grid and the
            endpoint pills share one slot: at rest the meta shows; on
            hover/focus they crossfade so the pills take its place. The swap
            only arms when there are links, so failed/pending cards keep their
            meta on hover. Failure details live on the stack detail page, not
            the card — the FAILED status word is the signal. */}
        <div className="mt-auto flex flex-col gap-3.5">
          <div className="relative">
            <div
              className={cn(
                hasLinks &&
                  "transition-opacity duration-150 group-hover:opacity-0 group-focus-within:opacity-0",
              )}
            >
              <CardMetaGrid
                rows={[
                  env.branch ? { label: "branch", value: env.branch } : null,
                  env.commit
                    ? { label: "commit", value: <span className="tabular-nums">{env.commit.slice(0, 7)}</span> }
                    : null,
                ]}
              />
            </div>
            {hasLinks && (
              <div className="pointer-events-none absolute inset-x-0 bottom-0 flex items-end opacity-0 transition-opacity duration-150 group-hover:pointer-events-auto group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:opacity-100">
                <EndpointPills urls={urls} />
              </div>
            )}
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
