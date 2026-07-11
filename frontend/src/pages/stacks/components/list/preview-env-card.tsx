import { ExternalLink, GitPullRequest } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import type { PreviewStack, PreviewPhase } from "@/api/preview-envs";

interface PreviewEnvCardProps {
  env: PreviewStack;
  /** Repository configuration name shown as the card subtitle. */
  configName?: string;
}

function phaseDotClass(phase: PreviewPhase | undefined): string {
  switch (phase) {
    case "Ready":
      return "bg-success";
    case "Failed":
      return "bg-danger";
    default:
      // Provisioning / Deploying / Deleting / not yet reported
      return "bg-warn";
  }
}

/**
 * Read-only preview environment card for the stacks dashboard. The card
 * navigates to the underlying stack; env management (sync/delete) lives on
 * the stack show page. Not wrapped in a Link because the URL outputs inside
 * are real anchors — nested <a> is invalid HTML.
 */
export function PreviewEnvCard({ env, configName }: PreviewEnvCardProps) {
  const navigate = useNavigate();
  const phase = env.status?.phase;
  const urls = env.status?.outputs?.urls ?? [];
  const updatedAt = env.updated_at || env.created_at;
  const clickable = Boolean(env.stack_id);

  return (
    <TooltipProvider>
      <Card
        role={clickable ? "link" : undefined}
        tabIndex={clickable ? 0 : undefined}
        aria-label={`PR #${env.pr_number} preview environment`}
        onClick={clickable ? () => navigate(`/stacks/${env.stack_id}`) : undefined}
        onKeyDown={
          clickable
            ? (e) => {
              if (e.key === "Enter") navigate(`/stacks/${env.stack_id}`);
            }
            : undefined
        }
        className={cn(
          "group flex flex-col w-full h-full min-h-[180px] p-5 gap-5 transition-colors duration-150",
          clickable && "cursor-pointer hover:border-brand-border hover:bg-muted/20",
        )}
      >
        {/* Header: icon + PR title + phase dot */}
        <div className="flex items-start gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-brand-border bg-brand-bg text-brand">
            <GitPullRequest className="h-[18px] w-[18px]" />
          </span>
          <div className="flex-1 min-w-0 flex items-start justify-between gap-2">
            <div className="min-w-0">
              <span className="block font-medium text-base leading-tight group-hover:text-brand transition-colors truncate">
                PR #{env.pr_number}
              </span>
              <span className="block truncate font-mono text-[11px] text-muted-foreground mt-0.5">
                {configName ?? env.name}{env.branch ? ` · ${env.branch}` : ""}
              </span>
            </div>
            <Tooltip>
              <TooltipTrigger asChild>
                <span
                  className={cn("mt-1.5 inline-block h-2 w-2 shrink-0 rounded-full", phaseDotClass(phase))}
                  aria-label={phase ?? "Unknown"}
                />
              </TooltipTrigger>
              <TooltipContent side="top">
                <span className="font-mono text-[11px] uppercase tracking-[1.5px]">{phase ?? "Unknown"}</span>
              </TooltipContent>
            </Tooltip>
          </div>
        </div>

        {/* URLs */}
        {urls.length > 0 && (
          <div className="flex flex-wrap gap-x-3 gap-y-1">
            {urls.map((u) => (
              <a
                key={u.url}
                href={u.url}
                target="_blank"
                rel="noreferrer"
                onClick={(e) => e.stopPropagation()}
                className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
              >
                <ExternalLink className="h-3 w-3" />
                {u.resource}
              </a>
            ))}
          </div>
        )}

        {/* Footer: commit on left, relative time on right */}
        <div className="flex items-end justify-between gap-2 mt-auto pt-4 border-t border-border/60 font-mono text-[11px] text-muted-foreground whitespace-nowrap">
          <span className="tabular-nums">{env.commit ? env.commit.slice(0, 7) : "—"}</span>
          <span className="uppercase tracking-[0.5px] text-right">
            {updatedAt
              ? formatDistanceToNow(new Date(updatedAt), { addSuffix: true }).replace(/^about\s/, "")
              : "—"}
          </span>
        </div>
      </Card>
    </TooltipProvider>
  );
}
