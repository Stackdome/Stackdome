import { AlertBanner, FailureCard } from "@/components/branded";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import type { FailingResource } from "./derive";

export interface FailingResourcesAccordionProps {
  failing: FailingResource[];
  releaseMessage?: string;
}

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

export function FailingResourcesAccordion({ failing, releaseMessage }: FailingResourcesAccordionProps) {
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
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      )}
    </div>
  );
}
