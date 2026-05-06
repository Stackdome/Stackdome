import { MoreHorizontal, Trash2, Pencil } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { AddonTypeIcon } from "./addon-type-icon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { PostgresAddon } from "@/api/addons";
import { StatusPill } from "./status-pill";

interface AddonTableProps {
  addons: PostgresAddon[];
  onDelete: (addon: PostgresAddon) => void;
}

function formatStorage(size?: string): string {
  return size ?? "—";
}

function formatPlan(spec: PostgresAddon["spec"]): string {
  const cpu = spec.resources?.cpu?.limit ?? spec.resources?.cpu?.request;
  const mem = spec.resources?.memory?.limit ?? spec.resources?.memory?.request;
  if (!cpu && !mem) return "Default";
  return `${cpu ?? "—"} CPU · ${mem ?? "—"}`;
}

export function AddonTable({ addons, onDelete }: AddonTableProps) {
  const navigate = useNavigate();
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[200px] font-semibold">Name</TableHead>
            <TableHead className="w-[120px] font-semibold">Type</TableHead>
            <TableHead className="w-[120px] font-semibold">Status</TableHead>
            <TableHead className="w-[100px] font-semibold">Version</TableHead>
            <TableHead className="w-[100px] font-semibold">Storage</TableHead>
            <TableHead className="min-w-[200px] font-semibold">Compute</TableHead>
            <TableHead className="w-[120px] font-semibold">Created</TableHead>
            <TableHead className="w-[70px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {addons.map((addon) => (
            <TableRow key={addon.id} className="hover:bg-muted/50">
              <TableCell className="font-medium py-4">
                <div className="break-words max-w-[180px]">{addon.name}</div>
              </TableCell>
              <TableCell className="py-4">
                <div className="flex items-center gap-2 text-sm">
                  <AddonTypeIcon type="postgres" size={16} />
                  Postgres
                </div>
              </TableCell>
              <TableCell className="py-4">
                <StatusPill state={addon.status?.state} />
              </TableCell>
              <TableCell className="py-4 text-sm text-muted-foreground">
                PG {addon.spec.version.major}
              </TableCell>
              <TableCell className="py-4 text-sm text-muted-foreground">
                {formatStorage(addon.spec.storage.size)}
              </TableCell>
              <TableCell className="py-4 text-sm text-muted-foreground">
                {formatPlan(addon.spec)}
              </TableCell>
              <TableCell className="py-4">
                <span className="text-sm text-muted-foreground">
                  {addon.created_at
                    ? new Date(addon.created_at).toLocaleDateString()
                    : "Unknown"}
                </span>
              </TableCell>
              <TableCell className="py-4">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0 hover:bg-muted">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-[160px]">
                    <DropdownMenuItem
                      onClick={() =>
                        addon.id && navigate(`/addons/postgres/${addon.id}/edit`)
                      }
                      disabled={!addon.id}
                    >
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="text-danger focus:text-danger"
                      onClick={() => onDelete(addon)}
                    >
                      <Trash2 className="h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
