import { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Box, Cpu, MemoryStick, AlertCircle, Clock } from "lucide-react";
import { EmptyState, StatusPill, type StatusVariant } from "@/components/branded";
import { cn } from "@/lib/utils";
import { useMetricsStream } from "./use-metrics-stream";
import type { StackResource } from "@/pages/stacks/types";
import type { ReleaseLiveStatus } from "@/api/releases";
import type { SseStreamStatus } from "@/api/observability";
import { isResourceReady } from "@/pages/stacks/lib/resource-readiness";
import { convertToDisplayMetrics } from "./utils";

function connectionStatusInfo(status: SseStreamStatus): { variant: StatusVariant; label: string } {
  switch (status) {
    case 'connecting':
      return { variant: 'pending', label: 'Connecting' };
    case 'connected':
      return { variant: 'ready', label: 'Live' };
    case 'reconnecting':
      return { variant: 'pending', label: 'Reconnecting' };
    case 'disconnected':
      return { variant: 'neutral', label: 'Disconnected' };
    default:
      return { variant: 'neutral', label: 'Unknown' };
  }
}

/** Parse a usage string ("180", "180m", "312Mi") to a bare number. */
function toNumber(v: string | number | undefined): number {
  if (v == null) return 0;
  const n = parseFloat(String(v).replace(/[^0-9.]/g, ''));
  return Number.isFinite(n) ? n : 0;
}

const HISTORY = 16;

/** A tiny end-aligned bar sparkline. Heights scale to the window max. */
function Sparkline({ data, className }: { data: number[]; className?: string }) {
  const max = Math.max(1, ...data);
  const bars = [...Array(HISTORY)].map((_, i) => data[data.length - HISTORY + i] ?? null);
  return (
    <div className="flex h-10 items-end gap-[3px]">
      {bars.map((v, i) => (
        <span
          key={i}
          className={cn("w-[5px] rounded-[1px]", v == null ? "bg-border/40" : className)}
          style={{ height: v == null ? "10%" : `${Math.max(6, Math.round((v / max) * 100))}%` }}
        />
      ))}
    </div>
  );
}

/** A labelled usage bar (fill width is relative to the peer max for that metric). */
function MetricBar({ label, value, pct, fill }: { label: string; value: string; pct: number; fill: string }) {
  return (
    <div>
      <div className="flex items-center justify-between">
        <span className="text-[11.5px] text-fg-2">{label}</span>
        <span className="font-mono text-[12px] text-foreground">{value}</span>
      </div>
      <div className="mt-1.5 h-[5px] overflow-hidden rounded-[3px] bg-muted">
        <span className={cn("block h-full rounded-[3px]", fill)} style={{ width: `${Math.max(2, Math.min(100, pct))}%` }} />
      </div>
    </div>
  );
}

interface MetricsTabProps {
  stackId: string;
  organizationId: string;
  resources: StackResource[];
  /** Live rollout status per resource name (release live_status); streams only
   *  open for Ready resources, since the backend rejects the rest pre-stream. */
  liveStatusResources?: ReleaseLiveStatus["resources"];
}

export function MetricsTab({ stackId, organizationId, resources, liveStatusResources }: MetricsTabProps) {
  const readyNames = useMemo(
    () =>
      resources
        .map((r) => r.name)
        .filter((name): name is string => !!name && isResourceReady(liveStatusResources?.[name]?.state)),
    [resources, liveStatusResources],
  );

  const { stackMetrics, resourceMetrics, connectionStatus, error } = useMetricsStream({
    stackId,
    organizationId,
    resourceNames: readyNames,
    enabled: readyNames.length > 0,
  });

  // Rolling window of stack-level samples for the summary sparklines.
  const [cpuHist, setCpuHist] = useState<number[]>([]);
  const [memHist, setMemHist] = useState<number[]>([]);
  const lastTs = useRef<string | undefined>(undefined);
  useEffect(() => {
    const ts = stackMetrics?.timestamp;
    if (!stackMetrics || ts === lastTs.current) return;
    lastTs.current = ts;
    setCpuHist((h) => [...h, toNumber(stackMetrics.cpu_usage)].slice(-HISTORY));
    setMemHist((h) => [...h, toNumber(stackMetrics.memory_usage)].slice(-HISTORY));
  }, [stackMetrics]);

  // Every resource gets a card: metrics when streaming, a placeholder while
  // not ready or before the first sample lands.
  const resourceCards = resources
    .filter((r): r is StackResource & { name: string } => !!r.name)
    .map((r) => {
      const metrics = resourceMetrics.get(r.name);
      return {
        resourceName: r.name,
        ready: readyNames.includes(r.name),
        state: liveStatusResources?.[r.name]?.state,
        metrics,
        displayMetrics: metrics ? convertToDisplayMetrics(metrics) : null,
      };
    });

  // Peer maxima → relative bar widths (no per-resource limit is available).
  const withMetrics = resourceCards.filter((r) => r.metrics);
  const cpuMax = Math.max(1, ...withMetrics.map((r) => toNumber(r.metrics!.cpu_usage)));
  const memMax = Math.max(1, ...withMetrics.map((r) => toNumber(r.metrics!.memory_usage)));

  const statusInfo =
    readyNames.length > 0
      ? connectionStatusInfo(connectionStatus)
      : { variant: 'neutral' as StatusVariant, label: 'Waiting' };
  const updatedAt = stackMetrics?.timestamp ? new Date(stackMetrics.timestamp).toLocaleTimeString() : null;

  return (
    <div className="mx-auto max-w-[1000px] px-[30px] py-[26px]">
      {/* Header */}
      <div className="mb-[18px] flex items-center gap-3">
        <h2 className="text-[18px] font-medium tracking-[-0.01em] text-foreground">Stack metrics</h2>
        <StatusPill variant={statusInfo.variant}>{statusInfo.label}</StatusPill>
        <div className="flex-1" />
        {updatedAt && <span className="font-mono text-[11px] text-fg-muted">updated {updatedAt}</span>}
      </div>

      {error && (
        <Alert variant="destructive" className="mb-4">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Summary cards */}
      <div className="mb-[18px] grid grid-cols-1 gap-3.5 md:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-[18px]">
          <div className="flex items-center gap-2 font-mono text-[11px] font-medium uppercase tracking-[1.5px] text-muted-foreground">
            <Cpu className="size-4 text-brand" />
            Stack CPU
          </div>
          <div className="mt-3 flex items-end justify-between gap-4">
            <div>
              <div className="text-[30px] font-medium leading-none tracking-[-0.02em] text-foreground">
                {stackMetrics?.cpu_usage ? `${toNumber(stackMetrics.cpu_usage)}m` : '—'}
              </div>
              <div className="mt-1 font-mono text-[11px] text-fg-muted">millicores</div>
            </div>
            <Sparkline data={cpuHist} className="bg-brand" />
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-[18px]">
          <div className="flex items-center gap-2 font-mono text-[11px] font-medium uppercase tracking-[1.5px] text-muted-foreground">
            <MemoryStick className="size-4 text-brand" />
            Stack memory
          </div>
          <div className="mt-3 flex items-end justify-between gap-4">
            <div>
              <div className="text-[30px] font-medium leading-none tracking-[-0.02em] text-foreground">
                {stackMetrics?.memory_usage ? `${toNumber(stackMetrics.memory_usage)}` : '—'}
                {stackMetrics?.memory_usage && <span className="ml-1 text-[15px] text-fg-muted">MiB</span>}
              </div>
              <div className="mt-1 font-mono text-[11px] text-fg-muted">mebibytes</div>
            </div>
            <Sparkline data={memHist} className="bg-fg-2" />
          </div>
        </div>
      </div>

      {/* Per-resource */}
      <div className="mb-3 font-mono text-[11px] font-medium uppercase tracking-[1.5px] text-fg-muted">Per resource</div>
      {resourceCards.length === 0 ? (
        <EmptyState
          icon={<Box className="h-6 w-6" />}
          title="No resource metrics yet"
          description="Metrics will appear once the stack starts emitting data."
        />
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
          {resourceCards.map((r) => (
            <div key={r.resourceName} className="rounded-md border border-border bg-card p-[14px_15px]">
              <div className="mb-3 flex items-center gap-2.5">
                <span
                  className={cn("size-2 shrink-0 rounded-full", r.ready ? "bg-success" : "bg-border")}
                  aria-hidden
                />
                <Box className="size-[15px] shrink-0 text-fg-muted" aria-hidden />
                <span className="flex-1 truncate text-sm font-medium text-foreground">{r.resourceName}</span>
                <span
                  className={cn(
                    "font-mono text-[9px] uppercase tracking-[0.12em]",
                    r.ready ? "text-success" : "text-fg-muted",
                  )}
                >
                  {r.ready ? 'Ready' : r.state ?? 'Pending'}
                </span>
              </div>
              {r.metrics && r.displayMetrics ? (
                <div className="space-y-2.5">
                  <MetricBar
                    label="CPU"
                    value={r.displayMetrics.cpu}
                    pct={(toNumber(r.metrics.cpu_usage) / cpuMax) * 100}
                    fill="bg-brand"
                  />
                  <MetricBar
                    label="Memory"
                    value={r.displayMetrics.memory}
                    pct={(toNumber(r.metrics.memory_usage) / memMax) * 100}
                    fill="bg-fg-2"
                  />
                </div>
              ) : (
                <div className="flex items-center gap-2 py-3 text-[12px] text-fg-muted">
                  <Clock className="size-3.5" aria-hidden />
                  {r.ready ? 'Waiting for data' : 'Waiting for resource'}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
