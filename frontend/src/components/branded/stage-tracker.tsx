import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export type StageStatus = "done" | "active" | "failed" | "todo" | "skipped";

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

// `skipped` (e.g. image-only stack, no build step) is a solid muted fill so it
// reads as inert, not pending — unlike `todo`'s hollow ring.
const RING: Record<StageStatus, string> = {
  done: "border-success bg-success",
  failed: "border-danger bg-danger",
  active: "border-warn",
  skipped: "border-fg-muted bg-fg-muted",
  todo: "border-fg-muted",
};

function glyph(status: StageStatus): ReactNode {
  if (status === "done") return <span className="text-[9px] leading-none text-white">✓</span>;
  if (status === "failed") return <span className="text-[9px] leading-none text-white">✕</span>;
  if (status === "active") return <span className="h-[7px] w-[7px] animate-spin rounded-full border-2 border-warn border-t-transparent" />;
  return null;
}

export function StageTracker({ stages, className }: { stages: Stages; className?: string }) {
  const items: ReactNode[] = [];
  ORDER.forEach((stage, i) => {
    const status = stages[stage.key];
    items.push(
      <div key={stage.key} className="flex items-center gap-1.5" role="listitem" data-status={status}>
        <span className={cn("box-border flex h-[15px] w-[15px] flex-none items-center justify-center rounded-full border-[1.5px]", RING[status])}>
          {glyph(status)}
        </span>
        <span className={cn("font-sans text-[12px] font-medium", status === "todo" || status === "skipped" ? "text-fg-muted" : "text-foreground")}>
          {stage.label}
        </span>
      </div>,
    );
    if (i < ORDER.length - 1) {
      items.push(
        <span
          key={`l${i}`}
          className={cn("mx-[9px] h-[1.5px] min-w-[18px] flex-1", status === "done" ? "bg-success" : "bg-border")}
        />,
      );
    }
  });
  return (
    <div className={cn("flex max-w-[340px] items-center", className)} role="list">
      {items}
    </div>
  );
}
