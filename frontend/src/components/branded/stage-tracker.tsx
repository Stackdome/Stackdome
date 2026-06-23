import { cn } from "@/lib/utils";

export type StageStatus = "done" | "active" | "failed" | "todo";

export interface Stages {
  build: StageStatus;
  deploy: StageStatus;
  ready: StageStatus;
}

const ORDER: Array<{ key: keyof Stages; label: string }> = [
  { key: "build", label: "Build" },
  { key: "deploy", label: "Deploy" },
  { key: "ready", label: "Ready" },
];

const MARK: Record<StageStatus, string> = { done: "✓", active: "●", failed: "✕", todo: "" };

function dotClass(status: StageStatus): string {
  switch (status) {
    case "done": return "bg-success text-white border-success";
    case "active": return "bg-warn/15 text-warn border-warn animate-pulse";
    case "failed": return "bg-danger/10 text-danger border-danger";
    default: return "bg-muted text-muted-foreground border-border";
  }
}

export function StageTracker({ stages, className }: { stages: Stages; className?: string }) {
  return (
    <div className={cn("flex items-center gap-0", className)} role="list">
      {ORDER.map((stage, i) => {
        const status = stages[stage.key];
        return (
          <div key={stage.key} className="flex items-center" role="listitem" data-status={status}>
            <div className={cn("flex h-[18px] w-[18px] items-center justify-center rounded-full border text-[10px] font-medium", dotClass(status))}>
              {MARK[status]}
            </div>
            <span className={cn("ml-1.5 text-[12.5px]", status === "todo" ? "text-muted-foreground" : "text-foreground")}>
              {stage.label}
            </span>
            {i < ORDER.length - 1 && (
              <span className={cn("mx-2.5 h-px w-8", status === "done" ? "bg-success" : "bg-border")} />
            )}
          </div>
        );
      })}
    </div>
  );
}
