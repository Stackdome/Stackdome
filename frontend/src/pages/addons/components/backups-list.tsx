import { AlertCircle } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { StatusPill, type StatusVariant } from "@/components/branded";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { PostgresBackup, PostgresBackupPhase } from "@/api/postgres-backups";

function humanSize(bytes?: number): string {
  if (!bytes) return "—";
  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  let n = bytes;
  let i = 0;
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024;
    i++;
  }
  return `${n.toFixed(n < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}

// Phase values are lowercase domain-specific strings: pending | running |
// completed | failed. Map directly to the StatusPill semantic palette rather
// than re-using `variantFromState`, which is tuned for addon/exposure states
// and would map "completed" to "info" (wrong) instead of "ready" (correct).
function phaseVariant(phase?: PostgresBackupPhase): StatusVariant {
  switch (phase) {
    case "completed":
      return "ready";
    case "failed":
      return "error";
    case "running":
    case "pending":
      return "pending";
    default:
      return "neutral";
  }
}

function fmt(ts?: string): string {
  if (!ts) return "—";
  return new Date(ts).toLocaleString();
}

// The API derives type from the CNPG Backup's owner reference, which can be
// absent for ScheduledBackup-spawned runs (reported as "manual"). CNPG names
// those runs "<cluster>-scheduled-backup-<ts>", so fall back to the name to
// distinguish them without a server round-trip.
function displayType(b: PostgresBackup): string {
  if (b.type && b.type.toLowerCase() !== "manual") return b.type;
  if (b.name && /-scheduled-backup-/.test(b.name)) return "Scheduled";
  return b.type ?? "—";
}

export function BackupsList({ backups }: { backups: PostgresBackup[] }) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Phase</TableHead>
          <TableHead>Started</TableHead>
          <TableHead>Completed</TableHead>
          <TableHead className="text-right">Size</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {backups.map((b) => (
          <TableRow key={b.id}>
            <TableCell className="font-medium">{b.name ?? b.id?.slice(0, 8)}</TableCell>
            <TableCell className="capitalize">{displayType(b)}</TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                <StatusPill variant={phaseVariant(b.phase)} pulse={false}>
                  {b.phase ?? "—"}
                </StatusPill>
                {b.phase === "failed" && b.error && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button type="button" aria-label={`Error for ${b.name ?? "backup"}`}>
                        <AlertCircle className="h-4 w-4 text-danger" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-sm font-mono text-xs">
                      {b.error}
                    </TooltipContent>
                  </Tooltip>
                )}
              </div>
            </TableCell>
            <TableCell>{fmt(b.started_at)}</TableCell>
            <TableCell>{fmt(b.completed_at)}</TableCell>
            <TableCell className="text-right font-mono">{humanSize(b.size_bytes)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
