import { useEffect, useState } from "react";
import { ChevronDown, Box, GitBranch } from "lucide-react";
import { FailureCard, LogSnapshot } from "@/components/branded";
import { fetchLogSnapshot } from "@/api/observability";
import type { FailingResource, ResourceSource } from "../derive";
import { phaseTone, toneTextClass, toneDotClass } from "../derive";

export interface LogContext { orgId: string; projectName: string; stackId: string; }
export interface ResourceRowVM { name: string; phase: string; replicas?: string; msg?: string; tag?: string; failure?: FailingResource; source?: ResourceSource; }
export interface ResourceRowProps { vm: ResourceRowVM; logContext?: LogContext; onOpenLogs?: (name: string) => void; }

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

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

export function ResourceRow({ vm, logContext, onOpenLogs }: ResourceRowProps) {
  const [open, setOpen] = useState(false);
  const tone = phaseTone(vm.phase);
  const failure = vm.failure;
  const hasMsg = !!vm.msg && vm.msg.trim().length > 0;
  // Expandable when it carries context (failure or status message) — shown inside the row, not clipped.
  const expandable = !!failure || hasMsg;
  const isRuntimeCrash = failure?.type === "runtime_crash";
  return (
    <div>
      {/* Name row + source line share one hover/click surface — the whole header is one unit. */}
      <div
        className={expandable ? "-mx-2 rounded-md px-2 transition-colors hover:bg-muted cursor-pointer" : undefined}
        onClick={expandable ? () => setOpen((o) => !o) : undefined}
      >
        <div className="flex items-baseline gap-2.5 py-2">
          <span className={`h-2 w-2 flex-none self-center rounded-full ${toneDotClass(tone)}`} />
          <span className="w-[82px] flex-none font-mono text-[12.5px] font-semibold text-foreground">{vm.name}</span>
          <span className={`min-w-0 flex-1 text-[13px] ${toneTextClass(tone)}`}>{vm.phase}</span>
          {vm.tag && <span className="self-center rounded border border-warn px-1.5 py-0.5 font-mono text-[9px] uppercase text-warn">{vm.tag}</span>}
          {failure?.exitCode != null && <span className="self-center font-mono text-[11px] text-fg-muted">exit {failure.exitCode}</span>}
          {failure?.restartCount != null && (
            <span className="self-center font-mono text-[11px] text-fg-muted">{failure.restartCount} {failure.restartCount === 1 ? "restart" : "restarts"}</span>
          )}
          {expandable && <ChevronDown className={`h-3.5 w-3.5 self-center text-fg-muted transition-transform ${open ? "rotate-180" : ""}`} />}
        </div>
        {vm.source && (
          <div className="-mt-1 pb-2 ml-[18px] flex items-center gap-1.5 font-mono text-[11px] text-fg-muted">
            {vm.source.kind === "image" ? <Box className="h-3 w-3 flex-none" /> : <GitBranch className="h-3 w-3 flex-none" />}
            <span className="truncate">{vm.source.label}</span>
          </div>
        )}
      </div>
      {open && expandable && (
        <div className="ml-[18px] mb-1 mt-1 rounded-md border border-border bg-muted p-3">
          {failure ? (
            <>
              <FailureCard
                resourceName={failure.name}
                stage={stageForCard(failure.stage)}
                reason={failure.reason}
                message={failure.message}
                exitCode={failure.exitCode}
                restartCount={failure.restartCount}
              />
              {logContext && isRuntimeCrash && <CrashLog ctx={logContext} name={failure.name} />}
              {/* Logs are a runtime concept — only offer "Open in Logs" for a crash. */}
              {onOpenLogs && isRuntimeCrash && (
                <button
                  onClick={() => onOpenLogs(vm.name)}
                  className="mt-2.5 rounded bg-primary px-3 py-1.5 font-sans text-[12px] font-medium text-primary-foreground"
                >
                  Open in Logs →
                </button>
              )}
            </>
          ) : (
            <div className="font-mono text-[11.5px] leading-relaxed text-foreground">{vm.msg}</div>
          )}
        </div>
      )}
    </div>
  );
}
