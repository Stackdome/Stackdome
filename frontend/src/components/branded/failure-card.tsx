import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { StageBadge, type FailureStage } from "./stage-badge";
import type { components } from "@/api/types/openapi";

type Condition = components["schemas"]["Condition"];

export interface FailureCardProps {
  resourceName: string;
  stage: FailureStage;
  reason: string;
  message?: string;
  exitCode?: number;
  restartCount?: number;
  revision?: string;
  conditions?: Condition[];
  hint?: string;
  /** action area, e.g. "Open in Logs →" button */
  actions?: ReactNode;
  /** extra content, e.g. <LogSnapshot> */
  children?: ReactNode;
  className?: string;
}

export function FailureCard({
  resourceName,
  stage,
  reason,
  message,
  exitCode,
  restartCount,
  revision,
  conditions,
  hint,
  actions,
  children,
  className,
}: FailureCardProps) {
  return (
    <div
      className={cn(
        "rounded-md border border-danger-border bg-danger-bg p-4",
        className,
      )}
    >
      <div className="flex flex-wrap items-center gap-2.5">
        <span className="font-mono text-[13px] font-semibold">
          {resourceName}
        </span>
        <StageBadge stage={stage} />
        <span className="text-[13px] font-semibold text-danger">{reason}</span>
        <span className="ml-auto flex gap-3.5 font-mono text-[11.5px] text-fg-muted">
          {exitCode !== undefined && (
            <span>
              exit <b>{exitCode}</b>
            </span>
          )}
          {restartCount !== undefined && (
            <span>
              restarts <b>{restartCount}</b>
            </span>
          )}
          {revision && (
            <span>
              rev <b>{revision.slice(0, 7)}</b>
            </span>
          )}
        </span>
      </div>
      {message && (
        <div className="mt-2.5 rounded-md border border-dashed border-danger-border bg-danger/5 px-3 py-2 font-mono text-[12.5px] text-danger">
          {message}
        </div>
      )}
      {conditions && conditions.length > 0 && (
        <div className="mt-2.5 grid gap-1 border-t border-danger-border pt-2.5">
          {conditions.map((c, i) => (
            <div key={i} className="flex items-baseline gap-2.5 text-xs">
              <span className="w-28 flex-none font-mono text-[11px] text-fg-muted">
                {c.type}
              </span>
              <span className="w-32 flex-none font-semibold">{c.reason}</span>
              <span className="flex-1 text-fg-muted">{c.message}</span>
            </div>
          ))}
        </div>
      )}
      {children}
      {(hint || actions) && (
        <div className="mt-3 flex items-center gap-2.5">
          {actions}
          {hint && <span className="text-[11.5px] text-fg-muted">{hint}</span>}
        </div>
      )}
    </div>
  );
}
