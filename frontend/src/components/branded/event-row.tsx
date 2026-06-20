import { cn } from "@/lib/utils";

export interface EventRowProps {
  ok: boolean;
  kind: string;
  title: string;
  sub?: string;
  when: string;
  duration?: string;
}

export function EventRow({ ok, kind, title, sub, when, duration }: EventRowProps) {
  return (
    <div className="flex items-center gap-3.5 border-b border-border px-4 py-3 last:border-b-0">
      <span
        aria-hidden="true"
        className={cn(
          "flex h-[26px] w-[26px] flex-none items-center justify-center rounded-full border text-xs",
          ok
            ? "border-success-border bg-success-bg text-success"
            : "border-danger-border bg-danger-bg text-danger",
        )}
      >
        {ok ? "✓" : "✕"}
      </span>
      <div className="min-w-0 flex-1">
        <div className="text-[13.5px] font-semibold">
          <span className="mr-2 rounded border border-border px-1.5 py-px align-[2px] font-mono text-[10px] font-medium uppercase tracking-wider text-fg-muted">
            {kind}
          </span>
          {title}
        </div>
        {sub && <div className="mt-px text-xs text-fg-muted">{sub}</div>}
      </div>
      <div className="flex-none text-right text-xs text-fg-muted">
        <div className="font-mono">{when}</div>
        {duration && <div>{duration}</div>}
      </div>
    </div>
  );
}
