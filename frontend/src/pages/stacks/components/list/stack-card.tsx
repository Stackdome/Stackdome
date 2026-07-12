import { AlertTriangle, Box, GitBranch, GitCommitHorizontal, Layers } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { formatDistanceToNow } from "date-fns";
import { Card } from "@/components/ui/card";
import {
  statusVariant,
  statusVariantTone,
  type StatusTone,
  type StatusVariant,
} from "@/components/branded/status-variant";
import { deriveHeaderHealth, latestDeployFailed } from "@/pages/stacks/components/detail/deployments/derive";
import { cn } from "@/lib/utils";
import type { Stack } from "@/pages/stacks/types";

export type RailTone = StatusTone;

/**
 * 4px full-bleed rail across the card's top edge, shown only while a deploy
 * is in flight: a soft track with an animated amber segment (static partial
 * fill under prefers-reduced-motion). Settled states render no rail — the
 * colored status word carries the state.
 */
export function StatusRail({ tone }: { tone: RailTone }) {
  if (tone !== "deploying") return null;
  return (
    <div className="h-1 w-full overflow-hidden bg-warn-bg" role="presentation" data-rail="deploying">
      <div className="h-full w-[34%] bg-brand animate-rail-sweep" />
    </div>
  );
}

/** Mono status word shown at the header's right edge, colored to match the rail. */
export function StatusWord({ tone, children }: { tone: RailTone | "neutral"; children: string }) {
  return (
    <span
      className={cn(
        "font-mono text-[10.5px] font-medium uppercase tracking-[1px] whitespace-nowrap",
        (tone === "success") && "text-success",
        (tone === "brand" || tone === "deploying") && "text-brand",
        tone === "danger" && "text-danger",
        tone === "neutral" && "text-fg-muted",
      )}
    >
      {children}
    </span>
  );
}

export interface EndpointUrl {
  resource?: string;
  url?: string;
}

/** External-link pills for a card's public endpoints. */
export function EndpointPills({ urls }: { urls: EndpointUrl[] }) {
  const valid = urls.filter((u) => u.url);
  if (valid.length === 0) return null;
  return (
    <div className="flex flex-wrap gap-1.5">
      {valid.map((u, i) => (
        <a
          key={`${u.resource}-${u.url}-${i}`}
          href={u.url}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          className="inline-flex items-center gap-1.5 rounded-sm border border-border px-2.5 py-1 font-mono text-xs text-fg-2 transition-colors duration-120 hover:text-brand hover:border-brand focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-brand/40"
        >
          {u.resource ?? "link"}
          <span className="opacity-55">↗</span>
        </a>
      ))}
    </div>
  );
}

interface CardFooterMetaProps {
  /** Short commit SHA, rendered with a commit glyph at the left edge. */
  commit?: string | null;
  /** Middot-separated stats (absent items dropped). */
  items?: Array<string | null | undefined | false>;
  /** Relative age, uppercase at the right edge. */
  age?: string | null;
}

/** Footer meta row: commit · stats · age spread across the card width. */
export function CardFooterMeta({ commit, items = [], age }: CardFooterMetaProps) {
  const present = items.filter((i): i is string => Boolean(i));
  return (
    <div className="flex items-center justify-between gap-3 border-t border-border/60 pt-4 font-mono text-[11.5px] text-fg-muted whitespace-nowrap">
      {commit && (
        <span className="inline-flex items-center gap-1.5 tabular-nums">
          <GitCommitHorizontal className="h-3.5 w-3.5 flex-none" strokeWidth={1.6} />
          {commit}
        </span>
      )}
      {present.length > 0 && (
        <span className="inline-flex items-center gap-2 tabular-nums">
          {present.map((item, i) => (
            <span key={item} className="inline-flex items-center gap-2">
              {i > 0 && <span aria-hidden>·</span>}
              {item}
            </span>
          ))}
        </span>
      )}
      {age && <span className="uppercase tracking-[0.5px] text-right">{age}</span>}
    </div>
  );
}

export function relativeAge(timestamp?: string | null): string | null {
  if (!timestamp) return null;
  return formatDistanceToNow(new Date(timestamp), { addSuffix: true }).replace(/^about\s/, "");
}

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
 * Deployed-stack dashboard card, "Status Strip" design. Status comes from the
 * release-health rollup (headerStatus); live ingress URLs live on release
 * detail, not the list payload, so cards carry no endpoint pills.
 */
export function DeployStackCard({ stack }: { stack: Stack }) {
  const navigate = useNavigate();
  const Icon = inferStackIcon(stack);
  const { label: word, variant } = headerStatus(stack);
  const tone: RailTone | "neutral" = variant === "neutral" ? "neutral" : statusVariantTone[variant];
  const deployFailed = latestDeployFailed(stack);
  const resourceCount = stack.spec?.stack_resources?.length || 0;
  const volumeCount = stack.spec?.volumes?.length || 0;
  const age = relativeAge(stack.updated_at || stack.created_at);
  return (
    <Card
      role="link"
      tabIndex={0}
      aria-label={`${stack.name} stack`}
      onClick={() => navigate(`/stacks/${stack.id}`)}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/stacks/${stack.id}`);
      }}
      className="group flex h-[210px] w-full cursor-pointer flex-col gap-0 overflow-hidden p-0 transition-colors duration-150 hover:border-brand-border hover:bg-muted/20 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-brand/40"
    >
      <StatusRail tone={tone === "neutral" ? "success" : tone} />
      <div className="flex flex-1 flex-col gap-3.5 p-5">
        <div className="flex items-center gap-[11px]">
          <Icon className="h-[18px] w-[18px] flex-none text-brand" strokeWidth={1.6} />
          <span
            className="mr-auto truncate text-base font-medium tracking-[-0.01em] transition-colors group-hover:text-brand"
            title={stack.name}
          >
            {stack.name}
          </span>
          {deployFailed && (
            <AlertTriangle
              className="h-3 w-3 flex-none text-danger"
              aria-label="Latest deploy failed"
            >
              <title>Latest deploy failed</title>
            </AlertTriangle>
          )}
          <StatusWord tone={tone}>{word}</StatusWord>
        </div>

        <div className="mt-auto flex flex-col gap-3.5">
          <CardFooterMeta items={[`${resourceCount} res`, `${volumeCount} vol`]} age={age} />
        </div>
      </div>
    </Card>
  );
}
