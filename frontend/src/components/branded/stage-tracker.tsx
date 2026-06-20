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
    case "done": return "bg-brand text-brand-foreground border-brand";
    case "active": return "bg-brand/10 text-brand border-brand animate-pulse";
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
            <div className={cn("flex h-5 w-5 items-center justify-center rounded-full border text-[11px] font-medium", dotClass(status))}>
              {MARK[status]}
            </div>
            <span className={cn("ml-1.5 text-[13px]", status === "todo" ? "text-muted-foreground" : "text-foreground")}>
              {stage.label}
            </span>
            {i < ORDER.length - 1 && (
              <span className={cn("mx-3 h-px w-10", status === "done" ? "bg-brand" : "bg-border")} />
            )}
          </div>
        );
      })}
    </div>
  );
}
