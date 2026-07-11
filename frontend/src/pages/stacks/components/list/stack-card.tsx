import { Box, GitBranch, Layers } from "lucide-react";
import { Link } from "react-router-dom";
import { formatDistanceToNow } from "date-fns";
import { Card } from "@/components/ui/card";
import { statusVariant } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";
import type { Stack } from "@/pages/stacks/types";

export type RailTone = "success" | "brand" | "danger" | "deploying";

/**
 * 4px full-bleed status rail across the card's top edge — the primary status
 * indicator of the "Status Strip" card design. `deploying` renders a soft
 * track with an animated amber segment (static partial fill under
 * prefers-reduced-motion).
 */
export function StatusRail({ tone }: { tone: RailTone }) {
  if (tone === "deploying") {
    return (
      <div className="h-1 w-full overflow-hidden bg-warn-bg" role="presentation" data-rail="deploying">
        <div className="h-full w-[34%] bg-brand animate-rail-sweep" />
      </div>
    );
  }
  return (
    <div
      role="presentation"
      data-rail={tone}
      className={cn(
        "h-1 w-full",
        tone === "success" && "bg-success",
        tone === "brand" && "bg-brand",
        tone === "danger" && "bg-danger",
      )}
    />
  );
}

/** Mono status word shown at the header's right edge, colored to match the rail. */
export function StatusWord({ tone, children }: { tone: RailTone; children: string }) {
  return (
    <span
      className={cn(
        "font-mono text-[10.5px] font-medium uppercase tracking-[1px] whitespace-nowrap",
        (tone === "success") && "text-success",
        (tone === "brand" || tone === "deploying") && "text-brand",
        tone === "danger" && "text-danger",
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
      {valid.map((u) => (
        <a
          key={u.url}
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

/** Footer meta row: mono, middot-separated, absent items dropped. */
export function CardFooterMeta({ items }: { items: Array<string | null | undefined | false> }) {
  const present = items.filter((i): i is string => Boolean(i));
  return (
    <div className="mt-auto flex flex-wrap items-center gap-2 border-t border-border/60 pt-4 font-mono text-[11.5px] text-fg-muted whitespace-nowrap">
      {present.map((item, i) => (
        <span key={item} className="inline-flex items-center gap-2 tabular-nums">
          {i > 0 && <span aria-hidden>·</span>}
          {item}
        </span>
      ))}
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

function deployTone(state?: string | null): { tone: RailTone; word: string } {
  const v = statusVariant("stack", state);
  if (v === "ready") return { tone: "success", word: "ready" };
  if (v === "error") return { tone: "danger", word: "failed" };
  return { tone: "deploying", word: "deploying" };
}

/**
 * Deployed-stack dashboard card, "Status Strip" design. Cluster name,
 * endpoint URLs, and commit hash are not in the stacks list payload, so the
 * card keeps the middle empty and shows counts + age in the footer.
 */
export function DeployStackCard({ stack }: { stack: Stack }) {
  const Icon = inferStackIcon(stack);
  const { tone, word } = deployTone(stack.status?.state);
  const resourceCount = stack.spec?.stack_resources?.length || 0;
  const volumeCount = stack.spec?.volumes?.length || 0;
  const age = relativeAge(stack.updated_at || stack.created_at);

  return (
    <Link to={`/stacks/${stack.id}`} className="block group h-full">
      <Card className="flex h-full min-h-[210px] w-full flex-col gap-0 overflow-hidden p-0 transition-colors duration-150 hover:border-brand-border hover:bg-muted/20">
        <StatusRail tone={tone} />
        <div className="flex flex-1 flex-col gap-[18px] p-5">
          <div className="flex items-center gap-[11px]">
            <Icon className="h-[18px] w-[18px] flex-none text-fg-2" strokeWidth={1.6} />
            <span
              className="mr-auto truncate text-base font-medium tracking-[-0.01em] transition-colors group-hover:text-brand"
              title={stack.name}
            >
              {stack.name}
            </span>
            {stack.status?.state && <StatusWord tone={tone}>{word}</StatusWord>}
          </div>

          <CardFooterMeta
            items={[`${resourceCount} res`, `${volumeCount} vol`, age]}
          />
        </div>
      </Card>
    </Link>
  );
}
