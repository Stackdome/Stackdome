import { ResourceRow, type ResourceRowVM, type LogContext } from "./resource-row";

export interface ResourceOutcomeListProps {
  rows: ResourceRowVM[];
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
}

/**
 * "RESOURCE OUTCOME" section — one ResourceRow per resource, shared by live and historical
 * bodies (data source differs: live cluster status vs stored release outcome).
 */
export function ResourceOutcomeList({ rows, logContext, onOpenLogs }: ResourceOutcomeListProps) {
  if (rows.length === 0) return null;
  return (
    <div className="mt-4">
      <div className="mb-0.5 font-mono text-[11px] uppercase tracking-wide text-fg-muted">Resource outcome</div>
      <div className="divide-y divide-border">
        {rows.map((vm, i) => (
          <ResourceRow key={vm.name || i} vm={vm} logContext={logContext} onOpenLogs={onOpenLogs} />
        ))}
      </div>
    </div>
  );
}
