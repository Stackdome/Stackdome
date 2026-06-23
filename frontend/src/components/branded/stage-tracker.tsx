import type { ReactNode } from "react";
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

// Per the Deploy Timeline design: done/failed are solid fills with a white
// glyph; active is a small amber (brand) spinner inside a brand ring; todo is
// a hollow muted ring. Circles are 15px and connectors flex to a 340px cap.
const RING: Record<StageStatus, string> = {
  done: "border-success bg-success",
  failed: "border-danger bg-danger",
  active: "border-brand",
  todo: "border-fg-muted",
};

function glyph(status: StageStatus): ReactNode {
  if (status === "done") return <span className="text-[9px] leading-none text-white">✓</span>;
  if (status === "failed") return <span className="text-[9px] leading-none text-white">✕</span>;
  if (status === "active") return <span className="h-[7px] w-[7px] animate-spin rounded-full border-2 border-brand border-t-transparent" />;
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
        <span className={cn("font-sans text-[12px] font-medium", status === "todo" ? "text-fg-muted" : "text-foreground")}>
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
