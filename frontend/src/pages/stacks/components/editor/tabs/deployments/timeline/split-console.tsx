import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import { LogSnapshot } from "@/components/branded";
import { fetchLogSnapshot } from "@/api/observability";
import type { ReleaseEvent } from "@/api/releases";
import type { FailingResource, ResourceSource } from "../derive";
import { phaseTone, toneTextClass, toneDotClass, tonePillClass, ResourceFailureType, compactEventMessage } from "../derive";

export interface LogContext { orgId: string; projectName: string; stackId: string; }
export interface ResourceRowVM { name: string; phase: string; replicas?: string; msg?: string; tag?: string; failure?: FailingResource; source?: ResourceSource; }

type ReleaseEventLevel = NonNullable<ReleaseEvent["level"]>;
const DEFAULT_LEVEL: ReleaseEventLevel = "info";

const levelGlyph: Record<ReleaseEventLevel, { glyph: string; text: string }> = {
  success: { glyph: "✓", text: "text-success" },
  error: { glyph: "✕", text: "text-danger" },
  warning: { glyph: "!", text: "text-warn" },
  info: { glyph: "•", text: "text-info" },
};

function CrashLog({ ctx, name }: { ctx: LogContext; name: string }) {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    void fetchLogSnapshot(ctx.orgId, ctx.projectName, ctx.stackId, name, 50).then((l) => { if (alive) setLines(l); });
    return () => { alive = false; };
  }, [ctx.orgId, ctx.projectName, ctx.stackId, name]);
  if (!lines.length) return null;
  return <div className="mt-2.5"><LogSnapshot lines={lines} /></div>;
}

export interface SplitConsoleProps {
  rows: ResourceRowVM[];
  events: ReleaseEvent[];
  streaming: boolean;
  logContext?: LogContext;
}

/**
 * "Bright panels" split console: resource rail on the left, activity console on the
 * right. Clicking a resource filters the console to its events and pins its detail
 * (status, source, last failure) above the stream; "all resources" zooms back out.
 * Shared by the live body (streaming) and post-mortem (historical one-shot fetch).
 */
export function SplitConsole({ rows, events, streaming, logContext }: SplitConsoleProps) {
  const [selected, setSelected] = useState<string | null>(null);
  if (rows.length === 0 && events.length === 0 && !streaming) return null;

  const selectedRow = rows.find((r) => r.name === selected);
  const visible = selectedRow ? events.filter((e) => e.resource_name === selectedRow.name) : events;
  const readyCount = rows.filter((r) => phaseTone(r.phase) === "ok").length;
  const failure = selectedRow?.failure;
  const detailMsg = failure ? (failure.message ?? failure.reason) : selectedRow?.msg;
  const isRuntimeCrash = failure?.type === ResourceFailureType.Runtime;

  return (
    <div className="mt-4 overflow-hidden rounded-lg border border-border bg-card">
      <div className="flex items-stretch">
        <div className="w-64 flex-none border-r border-border px-2.5 pb-3.5 pt-3">
          <div className="flex items-baseline px-2 pb-2.5">
            <span className="font-mono text-[10px] uppercase tracking-[1.2px] text-fg-muted">Resources</span>
            <span className="ml-auto font-mono text-[10px] text-fg-muted">{readyCount}/{rows.length} ready</span>
          </div>
          <button
            onClick={() => setSelected(null)}
            className={cn(
              "flex w-full items-center gap-2 rounded border px-2.5 py-[7px] text-left hover:bg-muted",
              selected === null ? "border-border bg-muted" : "border-transparent",
            )}
          >
            <span className="h-[7px] w-[7px] flex-none rounded-full border-[1.5px] border-fg-muted" />
            <span className="whitespace-nowrap font-mono text-[12px] text-fg-2">all resources</span>
            <span className="ml-auto font-mono text-[10.5px] text-fg-muted">{events.length}</span>
          </button>
          {rows.map((vm) => {
            const tone = phaseTone(vm.phase);
            return (
              <button
                key={vm.name}
                onClick={() => setSelected(vm.name)}
                className={cn(
                  "mt-0.5 flex w-full items-center gap-2 rounded border px-2.5 py-[7px] text-left hover:bg-muted",
                  selected === vm.name ? "border-brand bg-background" : "border-transparent",
                )}
              >
                <span className={cn("h-[7px] w-[7px] flex-none rounded-full", toneDotClass(tone))} />
                <span className="min-w-0 truncate font-mono text-[12px] font-medium text-foreground">{vm.name}</span>
                <span className={cn("ml-auto flex-none rounded-full border px-[7px] py-px font-mono text-[9px] font-medium uppercase tracking-[0.06em]", tonePillClass(tone))}>
                  {vm.phase}
                </span>
              </button>
            );
          })}
        </div>

        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3">
            <span className="font-mono text-[10px] uppercase tracking-[1.2px] text-fg-muted">Activity</span>
            <span className="whitespace-nowrap font-mono text-[10.5px] text-fg-2">· {selectedRow ? selectedRow.name : "all resources"}</span>
            {streaming && (
              <span className="ml-auto inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-success">
                <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-success" /> live
              </span>
            )}
          </div>

          {selectedRow && (
            <div className="border-b border-border bg-muted px-4 py-3">
              <div className="flex items-center gap-2.5">
                <span className={cn("h-[7px] w-[7px] flex-none rounded-full", toneDotClass(phaseTone(selectedRow.phase)))} />
                <span className="font-mono text-[12.5px] font-semibold text-foreground">{selectedRow.name}</span>
                <span className={cn("text-[12px] font-medium", toneTextClass(phaseTone(selectedRow.phase)))}>{selectedRow.phase}</span>
                {selectedRow.replicas && <span className="font-mono text-[11px] text-fg-muted">{selectedRow.replicas}</span>}
                {selectedRow.tag && <span className="rounded border border-warn px-1.5 py-0.5 font-mono text-[9px] uppercase text-warn">{selectedRow.tag}</span>}
                {failure?.exitCode != null && <span className="font-mono text-[11px] text-fg-muted">exit {failure.exitCode}</span>}
                {failure?.restartCount != null && (
                  <span className="font-mono text-[11px] text-fg-muted">{failure.restartCount} {failure.restartCount === 1 ? "restart" : "restarts"}</span>
                )}
                {selectedRow.source && (
                  <span className="ml-auto min-w-0 truncate font-mono text-[10.5px] text-fg-muted">▢ {selectedRow.source.label}</span>
                )}
              </div>
              {detailMsg && <div className={cn("mt-1.5 font-mono text-[11px] leading-relaxed", failure ? "text-danger" : "text-foreground")}>{detailMsg}</div>}
              {/* Crash-log snapshot is a one-shot follow=false read — the only log surface
                  that captures a crashing container's output (the Logs tab is live-only). */}
              {logContext && isRuntimeCrash && failure && <CrashLog ctx={logContext} name={failure.name} />}
            </div>
          )}

          <div className="py-1.5">
            {visible.length === 0 && (
              <div className="px-4 py-2 text-[12.5px] text-fg-muted">No activity yet</div>
            )}
            {visible.map((e) => {
              const lv = levelGlyph[e.level ?? DEFAULT_LEVEL] ?? levelGlyph[DEFAULT_LEVEL];
              return (
                <div key={e.sequence} className="flex items-start gap-2.5 px-4 py-[5px] hover:bg-muted">
                  <span className="w-14 flex-none pt-0.5 font-mono text-[10.5px] tabular-nums text-fg-muted">
                    {e.occurred_at ? new Date(e.occurred_at).toLocaleTimeString() : ""}
                  </span>
                  <span className={cn("w-3 flex-none text-center font-mono text-[11px]", lv.text)}>{lv.glyph}</span>
                  <span className="w-[90px] flex-none truncate pt-0.5 font-mono text-[10.5px] text-fg-muted">{e.resource_name || "release"}</span>
                  <span className="min-w-0 flex-1 break-words text-[12.5px] leading-[1.45] text-fg-2">
                    {compactEventMessage(e)}
                    {(e.links ?? []).map((l, i) => (
                      // target is a structured key/value map (e.g. build_id), not a URL —
                      // no navigable href exists yet, so the link renders as a label, not an anchor.
                      <span key={`${l.kind ?? "link"}-${i}`} className="block text-xs font-medium text-info">{l.label} &rarr;</span>
                    ))}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
