import { GitPullRequest } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { statusVariant } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";
import {
  StatusRail,
  StatusWord,
  EndpointPills,
  CardFooterMeta,
  relativeAge,
  type RailTone,
} from "./stack-card";
import type { PreviewStack, PreviewPhase } from "@/api/preview-envs";

interface PreviewEnvCardProps {
  env: PreviewStack;
  /** Repository configuration name shown in the identity block. */
  configName?: string;
}

function previewTone(phase: PreviewPhase | undefined): { tone: RailTone; word: string } {
  const v = statusVariant("preview", phase);
  if (v === "ready") return { tone: "success", word: "ready" };
  if (v === "error") return { tone: "danger", word: "failed" };
  // Provisioning / Deploying / Deleting / not yet reported → in flight
  return { tone: "deploying", word: (phase ?? "deploying").toLowerCase() };
}

/**
 * Read-only preview environment card, "Status Strip" design. The card
 * navigates to the underlying stack; env management (sync/delete) lives on
 * the stack show page. Not wrapped in a Link because the endpoint pills
 * inside are real anchors — nested <a> is invalid HTML.
 */
export function PreviewEnvCard({ env, configName }: PreviewEnvCardProps) {
  const navigate = useNavigate();
  const { tone, word } = previewTone(env.status?.phase);
  const urls = env.status?.outputs?.urls ?? [];
  const age = relativeAge(env.updated_at || env.created_at);
  const clickable = Boolean(env.stack_id);

  return (
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
        "group flex h-full min-h-[180px] w-full flex-col gap-0 overflow-hidden p-0 transition-colors duration-150",
        clickable && "cursor-pointer hover:border-brand-border hover:bg-muted/20",
      )}
    >
      <StatusRail tone={tone} />
      <div className="flex flex-1 flex-col gap-[18px] p-5">
        <div className="flex items-center gap-[11px]">
          <GitPullRequest className="h-[18px] w-[18px] flex-none text-brand" strokeWidth={1.6} />
          <span className="mr-auto truncate text-base font-medium tracking-[-0.01em] transition-colors group-hover:text-brand">
            PR #{env.pr_number}
          </span>
          <StatusWord tone={tone}>{word}</StatusWord>
        </div>

        {/* Identity: repo / branch rows */}
        <div className="grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-[9px]">
          {(configName ?? env.name) && (
            <>
              <span className="font-mono text-[9.5px] uppercase tracking-[1.2px] text-fg-muted">repo</span>
              <span className="truncate font-mono text-xs text-fg-2">{configName ?? env.name}</span>
            </>
          )}
          {env.branch && (
            <>
              <span className="font-mono text-[9.5px] uppercase tracking-[1.2px] text-fg-muted">branch</span>
              <span className="truncate font-mono text-xs text-brand">{env.branch}</span>
            </>
          )}
        </div>

        <EndpointPills urls={urls} />

        <CardFooterMeta items={[env.commit ? env.commit.slice(0, 7) : null, age]} />
      </div>
    </Card>
  );
}
