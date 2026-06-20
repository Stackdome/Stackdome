import { useEffect, useState } from "react";
import { AlertBanner, FailureCard, LogSnapshot } from "@/components/branded";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import { fetchLogSnapshot } from "@/api/observability";
import type { FailingResource } from "./derive";

export interface LogContext { orgId: string; teamName: string; stackId: string; }

export interface FailingResourcesAccordionProps {
  failing: FailingResource[];
  releaseMessage?: string;
  logContext?: LogContext;
}

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

function CrashLog({ ctx, resourceName }: { ctx: LogContext; resourceName: string }) {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    void fetchLogSnapshot(ctx.orgId, ctx.teamName, ctx.stackId, resourceName, 50)
      .then((l) => { if (alive) setLines(l); });
    return () => { alive = false; };
  }, [ctx.orgId, ctx.teamName, ctx.stackId, resourceName]);
  if (lines.length === 0) return null; // best-effort; pod may be unreachable (#98)
  return <LogSnapshot lines={lines} />;
}

export function FailingResourcesAccordion({ failing, releaseMessage, logContext }: FailingResourcesAccordionProps) {
  return (
    <div className="space-y-3">
      {releaseMessage && failing.length === 0 && (
        <AlertBanner>{releaseMessage}</AlertBanner>
      )}
      {failing.length > 0 && (
        <Accordion type="single" collapsible className="space-y-2">
          {failing.map((f) => (
            <AccordionItem key={f.name} value={f.name} className="rounded-md border border-danger-border">
              <AccordionTrigger className="px-3 py-2 text-[13px]">
                <span className="flex items-center gap-2">
                  <span className="font-medium text-foreground">{f.name}</span>
                  <span className="text-danger">{f.reason}</span>
                </span>
              </AccordionTrigger>
              <AccordionContent className="px-3 pb-3">
                <FailureCard
                  resourceName={f.name}
                  stage={stageForCard(f.stage)}
                  reason={f.reason}
                  message={f.message}
                  exitCode={f.exitCode}
                  restartCount={f.restartCount}
                />
                {logContext && f.type === "runtime_crash" && <CrashLog ctx={logContext} resourceName={f.name} />}
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      )}
    </div>
  );
}
