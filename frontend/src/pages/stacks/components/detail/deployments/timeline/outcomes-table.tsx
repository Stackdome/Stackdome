import { phaseTone, toneTextClass } from "../derive";

export interface OutcomeRow { phase?: string; ready_replicas?: number; replicas?: number; message?: string; }
export interface OutcomesTableProps { outcomes: Record<string, OutcomeRow>; }

export function OutcomesTable({ outcomes }: OutcomesTableProps) {
  const rows = Object.entries(outcomes);
  return (
    <div>
      <div className="grid grid-cols-[1fr_90px_60px_1fr] gap-2 pb-1.5 font-mono text-[10px] uppercase tracking-wide text-fg-muted">
        <span>Resource</span><span>Phase</span><span>Repl.</span><span>Message</span>
      </div>
      {rows.map(([name, o]) => {
        const tone = phaseTone(o.phase ?? "");
        return (
          <div key={name} className="grid grid-cols-[1fr_90px_60px_1fr] items-center gap-2 border-t border-border py-2">
            <span className="font-mono text-[12px] font-semibold text-foreground">{name}</span>
            <span className={`text-[12.5px] ${toneTextClass(tone)}`}>{o.phase}</span>
            <span className="font-mono text-[11px] text-fg-muted">{o.ready_replicas ?? 0}/{o.replicas ?? 0}</span>
            <span className={`font-mono text-[11px] ${tone === "err" ? "text-danger" : "text-fg-muted"}`}>{o.message || "—"}</span>
          </div>
        );
      })}
    </div>
  );
}
