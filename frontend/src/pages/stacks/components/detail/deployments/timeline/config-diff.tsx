import type { ResourceDiff, DiffRow } from "../release-snapshot-diff";

function Row({ row }: { row: DiffRow }) {
  return (
    <div className="flex items-start gap-2.5 px-3 pb-2 pt-1">
      <span className="w-[150px] flex-none font-mono text-[11px] text-fg-muted">{row.key}</span>
      <span className="flex flex-wrap items-center gap-1.5 font-mono text-[11px]">
        {row.kind === "added" && <span className="text-success">— → {row.to}</span>}
        {row.kind === "removed" && <span className="text-danger line-through">{row.from}</span>}
        {row.kind === "changed" && (
          <>
            <span className="text-danger line-through opacity-80">{row.from}</span>
            <span className="text-fg-muted">→</span>
            <span className="text-success">{row.to}</span>
          </>
        )}
      </span>
    </div>
  );
}

export interface ConfigDiffProps {
  diffs: ResourceDiff[];
  prevSeq?: number;
}

export function ConfigDiff({ diffs }: ConfigDiffProps) {
  if (!diffs.length)
    return <div className="text-[12.5px] text-fg-muted">Initial release — nothing to compare.</div>;

  return (
    <div className="space-y-2.5">
      {diffs.map((d) => {
        const dot =
          d.change === "added" ? "bg-success" : d.change === "removed" ? "bg-fg-muted" : "bg-warn";
        return (
          <div
            key={d.name}
            className={`overflow-hidden rounded-md border border-border ${d.change === "removed" ? "opacity-80" : ""}`}
          >
            <div className="flex items-center gap-2.5 bg-muted px-3 py-2.5">
              <span className={`h-[7px] w-[7px] flex-none rounded-full ${dot}`} />
              <span
                className={`font-mono text-[12.5px] font-semibold text-foreground ${d.change === "removed" ? "line-through" : ""}`}
              >
                {d.name}
              </span>
              {d.change === "added" && (
                <span className="rounded border border-success px-1.5 py-0.5 font-mono text-[9px] text-success">
                  ADDED
                </span>
              )}
              {d.change === "removed" && (
                <span className="rounded border border-fg-muted px-1.5 py-0.5 font-mono text-[9px] text-fg-muted">
                  REMOVED
                </span>
              )}
            </div>
            {d.change === "removed" ? (
              <div className="flex items-start gap-2.5 border-t border-border px-3 py-2.5 text-[12px] text-fg-muted">
                <span className="flex-none">−</span>
                <span>{d.note}</span>
              </div>
            ) : (
              d.sections.map((sec, si) => (
                <div key={si} className="border-t border-border">
                  <div className="px-3 pb-0.5 pt-2 font-mono text-[9px] uppercase tracking-wide text-fg-muted">
                    {sec.kind}
                  </div>
                  {sec.rows.map((row, ri) => (
                    <Row key={ri} row={row} />
                  ))}
                </div>
              ))
            )}
          </div>
        );
      })}
    </div>
  );
}
