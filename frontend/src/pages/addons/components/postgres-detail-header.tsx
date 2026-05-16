import { Link } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, StatusPill, type StatusVariant } from "@/components/branded";
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
  return (
    <div className="flex flex-col gap-3">
      <Link
        to="/addons"
        className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-3 w-3" /> All addons
      </Link>
      <PageHeader
        eyebrow="Postgres add-on"
        title={addon.name}
        subtitle={addon.status?.message || "Managed PostgreSQL cluster"}
        actions={
          <div className="flex gap-3">
            <Link to={`/addons/postgres/${addon.id}/edit`}>
              <Button variant="outline">Edit configuration</Button>
            </Link>
            <Button variant="outline" className="text-danger" onClick={onDelete}>
              Delete
            </Button>
          </div>
        }
      />
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        {state && (
          <StatusPill variant={stateVariant(state)}>{state}</StatusPill>
        )}
        {addon.created_at && (
          <span>
            Created {new Date(addon.created_at).toLocaleString()}
          </span>
        )}
      </div>
    </div>
  );
}
