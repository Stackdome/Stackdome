import { Pencil, Trash2 } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import type { ObjectStore } from "../types";

interface ObjectStoreListProps {
  objectStores: ObjectStore[];
  onEdit: (store: ObjectStore) => void;
  onDelete: (store: ObjectStore) => void;
}

function providerLabel(store: ObjectStore): string {
  const cfg = store.spec.configuration;
  if (cfg.s3_credentials) {
    return cfg.s3_credentials.endpoint_url ? "S3-compatible" : "S3";
  }
  if (cfg.azure_credentials) return "Azure";
  if (cfg.gcs_credentials) return "GCS";
  return "—";
}

function endpointLabel(store: ObjectStore): string {
  const cfg = store.spec.configuration;
  if (cfg.s3_credentials?.endpoint_url) return cfg.s3_credentials.endpoint_url;
  if (cfg.s3_credentials?.region) return cfg.s3_credentials.region;
  return "—";
}

export function ObjectStoreList({ objectStores, onEdit, onDelete }: ObjectStoreListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="font-semibold">Name</TableHead>
            <TableHead className="font-semibold">Provider</TableHead>
            <TableHead className="font-semibold">Endpoint / Region</TableHead>
            <TableHead className="font-semibold">Destination path</TableHead>
            <TableHead className="font-semibold">Retention</TableHead>
            <TableHead className="w-[120px] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {objectStores.map((store) => (
            <TableRow key={store.id} className="hover:bg-muted/50">
              <TableCell className="font-medium">{store.name}</TableCell>
              <TableCell>{providerLabel(store)}</TableCell>
              <TableCell className="font-mono text-xs">{endpointLabel(store)}</TableCell>
              <TableCell className="font-mono text-xs">{store.spec.destination_path}</TableCell>
              <TableCell>{store.spec.retention_policy}</TableCell>
              <TableCell className="text-right">
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Edit ${store.name}`}
                  onClick={() => onEdit(store)}
                >
                  <Pencil className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Delete ${store.name}`}
                  onClick={() => onDelete(store)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
