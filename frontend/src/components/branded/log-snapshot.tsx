import { cn } from "@/lib/utils";

export interface LogSnapshotProps {
  lines: string[];
  label?: string;
  className?: string;
}

export function LogSnapshot({ lines, label, className }: LogSnapshotProps) {
  if (lines.length === 0) return null;
  return (
    <div className={cn("mt-3", className)}>
      <div className="mb-1.5 flex items-center gap-2 font-mono text-[10.5px] uppercase tracking-wider text-fg-muted">
        <span>{label ?? `log snapshot · last ${lines.length} lines`}</span>
        <span className="normal-case tracking-normal">(read-only)</span>
      </div>
      <pre className="overflow-x-auto whitespace-pre-wrap rounded-md bg-foreground/95 px-3.5 py-3 font-mono text-xs leading-relaxed text-background">
        {lines.join("\n")}
      </pre>
    </div>
  );
}
