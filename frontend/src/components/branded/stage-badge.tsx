import { cn } from "@/lib/utils";

export type FailureStage = "build" | "runtime" | "init" | "validation";

const STAGE_CLASSES: Record<FailureStage, string> = {
  build: "text-info bg-info-bg border-info-border",
  runtime: "text-warn bg-warn-bg border-warn-border",
  init: "text-warn bg-warn-bg border-warn-border",
  validation: "text-danger bg-danger-bg border-danger-border",
};

export function StageBadge({
  stage,
  className,
}: {
  stage: FailureStage;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded border px-2 py-px font-mono text-label font-semibold",
        STAGE_CLASSES[stage],
        className,
      )}
    >
      {stage}
    </span>
  );
}
