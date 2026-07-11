import { cn } from "@/lib/utils";
import type { ReleaseEvent } from "@/api/releases";

type ReleaseEventLevel = NonNullable<ReleaseEvent["level"]>;
const DEFAULT_LEVEL: ReleaseEventLevel = "info";

const levelStyles: Record<ReleaseEventLevel, { wrap: string; glyph: string }> = {
  success: { wrap: "border-success-border bg-success-bg text-success", glyph: "✓" },
  error: { wrap: "border-danger-border bg-danger-bg text-danger", glyph: "✕" },
  warning: { wrap: "border-warn-border bg-warn-bg text-warn", glyph: "!" },
  info: { wrap: "border-info-border bg-info-bg text-info", glyph: "•" },
};

export interface ReleaseActivityFeedProps {
  events: ReleaseEvent[];
  streaming: boolean;
}

/**
 * Event log for a release: level-tinted glyph, type chip, resource tag, message,
 * links, right-aligned time. Shared by the live body (streaming) and post-mortem
 * (historical one-shot fetch).
 */
export function ReleaseActivityFeed({ events, streaming }: ReleaseActivityFeedProps) {
  if (events.length === 0 && !streaming) return null;

  return (
    <div className="overflow-hidden rounded-md border border-border">
      <div className="flex items-center gap-2 border-b border-border px-3 py-2 font-mono text-[11px] font-bold uppercase tracking-wider text-fg-muted">
        Activity
        {streaming && (
          <span className="ml-auto inline-flex items-center gap-1.5 text-success">
            <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-success" /> live
          </span>
        )}
      </div>
      {events.map((e) => {
        const s = levelStyles[e.level ?? DEFAULT_LEVEL] ?? levelStyles[DEFAULT_LEVEL];
        return (
          <div key={e.sequence} className="flex items-start gap-3 border-b border-border px-3 py-2 last:border-b-0">
            <span className={cn("mt-0.5 flex h-[22px] w-[22px] flex-none items-center justify-center rounded-full border text-[11px]", s.wrap)}>
              {s.glyph}
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-baseline gap-2 text-[13px] font-semibold">
                <span className="rounded border border-border px-1.5 font-mono text-[9.5px] font-medium uppercase tracking-wider text-fg-muted">{e.type}</span>
                {e.resource_name && <span className="rounded bg-fg-muted/10 px-1.5 font-mono text-[11.5px] text-fg-2">{e.resource_name}</span>}
                <span className="min-w-0 break-words font-normal">{e.message}</span>
              </div>
              {(e.links ?? []).map((l, i) => (
                // target is a structured key/value map (e.g. build_id, resource_name), not a URL —
                // no navigable href exists yet, so the link renders as a label, not an anchor.
                <span key={l.kind ?? i} className="block text-xs font-medium text-info">{l.label} &rarr;</span>
              ))}
            </div>
            <span className="flex-none pt-0.5 font-mono text-[10.5px] tabular-nums text-fg-muted">
              {e.occurred_at ? new Date(e.occurred_at).toLocaleTimeString() : ""}
            </span>
          </div>
        );
      })}
    </div>
  );
}
