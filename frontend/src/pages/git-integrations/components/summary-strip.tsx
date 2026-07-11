import { cn } from "@/lib/utils";
import type { RowViewModel } from "../lib/derive-row";

interface StatCardProps {
  dotClassName: string;
  count: number;
  caption: string;
}

function StatCard({ dotClassName, count, caption }: StatCardProps) {
  return (
    <div className="rounded-[10px] border border-border bg-card px-4 py-3">
      <div className="flex items-center gap-2">
        <span className={cn("h-2 w-2 rounded-full", dotClassName)} />
        <span className="text-2xl font-semibold text-foreground">{count}</span>
      </div>
      <p className="mt-1 text-xs text-muted-foreground">{caption}</p>
    </div>
  );
}

export function SummaryStrip({ rows }: { rows: RowViewModel[] }) {
  if (rows.length === 0) return null;

  const connectedCount = rows.filter((row) => row.statusKey === "connected").length;
  const attentionCount = rows.length - connectedCount;

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <StatCard dotClassName="bg-success" count={connectedCount} caption="Connected & ready" />
      <StatCard
        dotClassName={attentionCount > 0 ? "bg-warn" : "bg-muted-foreground"}
        count={attentionCount}
        caption="Needs attention"
      />
    </div>
  );
}
