import { useEffect, useState } from "react";
import { ChevronDown, Box, GitBranch } from "lucide-react";
import { FailureCard, LogSnapshot } from "@/components/branded";
import { fetchLogSnapshot } from "@/api/observability";
import type { FailingResource, ResourceSource } from "../derive";
import { phaseTone, toneTextClass, toneDotClass } from "../derive";

export interface LogContext { orgId: string; teamName: string; stackId: string; }
export interface ResourceRowVM { name: string; phase: string; replicas?: string; msg?: string; tag?: string; failure?: FailingResource; source?: ResourceSource; }
export interface ResourceRowProps { vm: ResourceRowVM; logContext?: LogContext; onOpenLogs?: (name: string) => void; }

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

function CrashLog({ ctx, name }: { ctx: LogContext; name: string }) {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    void fetchLogSnapshot(ctx.orgId, ctx.teamName, ctx.stackId, name, 50).then((l) => { if (alive) setLines(l); });
    return () => { alive = false; };
  }, [ctx.orgId, ctx.teamName, ctx.stackId, name]);
  if (!lines.length) return null;
  return <div className="mt-2.5"><LogSnapshot lines={lines} /></div>;
}

export function ResourceRow({ vm, logContext, onOpenLogs }: ResourceRowProps) {
  const [open, setOpen] = useState(false);
  const tone = phaseTone(vm.phase);
  const failing = !!vm.failure;
  return (
    <div>
      <div
        className={`flex items-baseline gap-2.5 py-2 ${failing ? "cursor-pointer" : ""}`}
        onClick={failing ? () => setOpen((o) => !o) : undefined}
      >
        <span className={`h-2 w-2 flex-none self-center rounded-full ${toneDotClass(tone)}`} />
        <span className="w-[82px] flex-none font-mono text-[12.5px] font-semibold text-foreground">{vm.name}</span>
        <span className={`min-w-0 flex-1 text-[13px] ${toneTextClass(tone)}`}>{vm.phase}</span>
        {vm.tag && <span className="self-center rounded border border-warn px-1.5 py-0.5 font-mono text-[9px] uppercase text-warn">{vm.tag}</span>}
        {vm.replicas && <span className="self-center font-mono text-[11px] text-fg-muted">{vm.replicas}</span>}
        {vm.msg && <span className="max-w-[220px] self-center truncate font-mono text-[11px] text-fg-muted">{vm.msg}</span>}
        {failing && <ChevronDown className={`h-3.5 w-3.5 self-center text-fg-muted transition-transform ${open ? "rotate-180" : ""}`} />}
      </div>
      {vm.source && (
        <div className="-mt-1 mb-1 ml-[18px] flex items-center gap-1.5 font-mono text-[11px] text-fg-muted">
          {vm.source.kind === "image" ? <Box className="h-3 w-3 flex-none" /> : <GitBranch className="h-3 w-3 flex-none" />}
          <span className="truncate">{vm.source.label}</span>
        </div>
      )}
      {open && vm.failure && (
        <div className="ml-[18px] mb-1 rounded-md border border-border bg-muted p-3">
          <FailureCard
            resourceName={vm.failure.name}
            stage={stageForCard(vm.failure.stage)}
            reason={vm.failure.reason}
            message={vm.failure.message}
            exitCode={vm.failure.exitCode}
            restartCount={vm.failure.restartCount}
          />
          {logContext && vm.failure.type === "runtime_crash" && <CrashLog ctx={logContext} name={vm.failure.name} />}
          {onOpenLogs && (
            <button
              onClick={() => onOpenLogs(vm.name)}
              className="mt-2.5 rounded bg-primary px-3 py-1.5 font-sans text-[12px] font-medium text-primary-foreground"
            >
              Open in Logs →
            </button>
          )}
        </div>
      )}
    </div>
  );
}
