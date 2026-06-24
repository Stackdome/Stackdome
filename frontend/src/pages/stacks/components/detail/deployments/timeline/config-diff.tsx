import type { SnapshotDiff, ItemDiff, ResourceDiff, DiffRow } from "../release-snapshot-diff";

function Row({ row }: { row: DiffRow }) {
  return (
    <div className="flex items-start gap-2.5 px-3 pb-2 pt-1">
      <span className="w-[150px] flex-none font-mono text-[11px] text-fg-muted">{row.key}</span>
      <span className="flex flex-wrap items-center gap-1.5 font-mono text-[11px]">
        {row.kind === "added" && <span className="text-success">— → {row.to}</span>}
        {row.kind === "removed" && <span className="text-danger">{row.from}</span>}
        {row.kind === "changed" && (
          <>
            <span className="text-danger opacity-80">{row.from}</span>
            <span className="text-fg-muted">→</span>
            <span className="text-success">{row.to}</span>
          </>
        )}
      </span>
    </div>
  );
}

function CardHead({ name, change }: { name: string; change: "added" | "removed" | "modified" }) {
  const dot = change === "added" ? "bg-success" : change === "removed" ? "bg-danger" : "bg-brand";
  return (
    <div className="flex items-center gap-2.5 bg-muted px-3 py-2.5">
      <span className={`h-[7px] w-[7px] flex-none rounded-full ${dot}`} />
      <span className="font-mono text-[12.5px] font-semibold text-foreground">{name}</span>
      {change === "added" && <span className="inline-flex items-center rounded-md border border-success-border bg-success-bg px-2 py-0.5 text-[11px] font-medium text-success">Added</span>}
      {change === "removed" && <span className="inline-flex items-center rounded-md border border-danger-border bg-danger-bg px-2 py-0.5 text-[11px] font-medium text-danger">Removed</span>}
      {change === "modified" && <span className="inline-flex items-center rounded-md border border-brand-border bg-brand-bg px-2 py-0.5 text-[11px] font-medium text-brand">Modified</span>}
    </div>
  );
}

function Note({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2.5 border-t border-border px-3 py-2.5 text-[12px] text-fg-muted">
      <span className="flex-none">−</span>
      <span>{children}</span>
    </div>
  );
}

function ResourceCard({ d }: { d: ResourceDiff }) {
  return (
    <div className={`overflow-hidden rounded-md border border-border ${d.change === "removed" ? "opacity-80" : ""}`}>
      <CardHead name={d.name} change={d.change} />
      {d.note && <Note>{d.note}</Note>}
      {d.sections.map((sec, si) => (
        <div key={si} className="border-t border-border">
          <div className="px-3 pb-0.5 pt-2 font-mono text-[9px] uppercase tracking-wide text-fg-muted">{sec.kind}</div>
          {sec.rows.map((row, ri) => <Row key={ri} row={row} />)}
        </div>
      ))}
    </div>
  );
}

function ItemCard({ item }: { item: ItemDiff }) {
  return (
    <div className={`overflow-hidden rounded-md border border-border ${item.change === "removed" ? "opacity-80" : ""}`}>
      <CardHead name={item.name} change={item.change} />
      {item.note && <Note>{item.note}</Note>}
      {item.rows.length > 0 && <div className="border-t border-border py-1">{item.rows.map((row, ri) => <Row key={ri} row={row} />)}</div>}
    </div>
  );
}

function Group({ label, items }: { label: string; items: ItemDiff[] }) {
  if (!items.length) return null;
  return (
    <div className="space-y-2.5">
      <div className="font-mono text-[9px] uppercase tracking-wide text-fg-muted">{label}</div>
      {items.map((it) => <ItemCard key={it.name} item={it} />)}
    </div>
  );
}

export interface ConfigDiffProps {
  diff: SnapshotDiff;
  hasPrev: boolean;
  prevSeq?: number;
}

export function ConfigDiff({ diff, hasPrev, prevSeq }: ConfigDiffProps) {
  if (!diff.resources.length && !diff.volumes.length && !diff.connections.length) {
    return (
      <div className="text-[12.5px] text-fg-muted">
        {hasPrev ? `No configuration changes since #${prevSeq ?? "previous"}.` : "Initial release — nothing to compare."}
      </div>
    );
  }

  return (
    <div className="space-y-2.5">
      {diff.resources.map((d) => <ResourceCard key={d.name} d={d} />)}
      <Group label="Volumes" items={diff.volumes} />
      <Group label="Connections" items={diff.connections} />
    </div>
  );
}
