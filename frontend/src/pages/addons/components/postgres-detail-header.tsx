import { Link } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, StatusPill, type StatusVariant } from "@/components/branded";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from "@/components/ui/tooltip";
import { AddonTypeIcon } from "./addon-type-icon";
import type { PostgresAddon } from "@/api/addons";

// Map the backend's PostgresAddon status state to the StatusPill variant.
// State names per OpenAPI: Pending, Creating, Initializing, Ready, Updating,
// BackingUp, Restoring, Error, Hibernated, Fenced, Deleting.
function stateVariant(state?: string): StatusVariant {
  switch (state) {
    case "Ready":
      return "ready";
    case "Error":
      return "error";
    case "Pending":
    case "Creating":
    case "Initializing":
    case "Updating":
    case "BackingUp":
    case "Restoring":
      return "pending";
    case "Hibernated":
    case "Fenced":
    case "Deleting":
      return "neutral";
    default:
      return "info";
  }
}

export function PostgresDetailHeader({
  addon,
  onDelete,
}: {
  addon: PostgresAddon;
  onDelete: () => void;
}) {
  const state = addon.status?.state;
  // Only surface a status tooltip when the addon is in a non-healthy
  // state. When Ready, stale/non-blocking conditions (e.g. an old
  // LastBackupFailed) would otherwise show as an alarming tooltip.
  const problemCondition =
    state && state !== "Ready"
      ? addon.status?.conditions?.find(
        (c) => c.status && c.status !== "True" && (c.message || c.reason),
      )
      : undefined;
  const statusMessage =
    state && state !== "Ready"
      ? addon.status?.message ||
        (problemCondition
          ? [problemCondition.reason, problemCondition.message]
            .filter(Boolean)
            .join(": ")
          : undefined)
      : undefined;
  const statusPill = state ? (
    <StatusPill variant={stateVariant(state)}>{state}</StatusPill>
  ) : undefined;
  return (
    <div className="flex flex-col gap-3">
      <Link
        to="/addons"
        className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-3 w-3" /> All addons
      </Link>
      <PageHeader
        actionsAlign="center"
        eyebrow="Postgres add-on"
        title={
          <span className="flex items-center gap-2.5">
            <AddonTypeIcon type="postgres" size={26} className="shrink-0" />
            {addon.name}
          </span>
        }
        status={
          statusPill && statusMessage ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="cursor-help">{statusPill}</span>
              </TooltipTrigger>
              <TooltipContent side="right" className="max-w-sm text-xs">
                {statusMessage}
              </TooltipContent>
            </Tooltip>
          ) : (
            statusPill
          )
        }
        subtitle={
          addon.created_at
            ? `Created ${new Date(addon.created_at).toLocaleDateString()}`
            : "Managed PostgreSQL cluster"
        }
        actions={
          <div className="flex gap-3">
            <Link to={`/addons/postgres/${addon.id}/edit`}>
              <Button variant="outline">Edit</Button>
            </Link>
            <Button
              variant="outline"
              className="text-danger hover:border-danger hover:bg-danger/10 hover:text-danger"
              onClick={onDelete}
            >
              Delete
            </Button>
          </div>
        }
      />
    </div>
  );
}
