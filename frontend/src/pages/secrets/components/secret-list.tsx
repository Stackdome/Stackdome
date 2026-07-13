import { MoreHorizontal, Edit, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
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
import { Badge } from "@/components/ui/badge";
import type { Secret } from "../types";

interface SecretListProps {
  secrets: Secret[];
  onEdit: (secret: Secret) => void;
  onDelete: (secret: Secret) => void;
  canWrite?: (projectId?: string) => boolean;
}

function formatSecretType(type: string): string {
  switch (type) {
    case "Generic":
      return "Generic";
    case "DockerRegistry":
      return "Docker Registry";
    case "GitCredentials":
      return "Git Credentials";
    case "UsernamePassword":
      return "Username/Password";
    case "Token":
      return "Token";
    case "SSHKey":
      return "SSH Key";
    default:
      return type;
  }
}

function getSecretTypeColor(type: string): "default" | "secondary" | "destructive" | "outline" {
  switch (type) {
    case "DockerRegistry":
      return "default";
    case "GitCredentials":
      return "secondary";
    case "Token":
      return "outline";
    default:
      return "default";
  }
}

export function SecretList({ secrets, onEdit, onDelete, canWrite }: SecretListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[200px] font-semibold">Name</TableHead>
            <TableHead className="w-[140px] font-semibold">Type</TableHead>
            <TableHead className="min-w-[250px] font-semibold">Description</TableHead>
            <TableHead className="w-[120px] font-semibold">Created</TableHead>
            <TableHead className="w-[70px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {secrets.map((secret) => {
            const rowCanWrite = canWrite ? canWrite(secret.project_id) : true;
            return (
              <TableRow key={secret.id} className="hover:bg-muted/50">
                <TableCell className="font-medium py-4">
                  <div className="break-words max-w-[180px]">
                    {secret.name}
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <Badge variant={getSecretTypeColor(secret.type)} className="text-xs">
                    {formatSecretType(secret.type)}
                  </Badge>
                </TableCell>
                <TableCell className="py-4 min-w-[250px] max-w-[400px]">
                  <div className="text-sm text-muted-foreground break-words whitespace-pre-wrap">
                    {secret.description || "No description"}
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <span className="text-sm text-muted-foreground">
                    {secret.created_at ? new Date(secret.created_at).toLocaleDateString() : "Unknown"}
                  </span>
                </TableCell>
                <TableCell className="py-4">
                  {rowCanWrite && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0 hover:bg-muted">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-[160px]">
                        <DropdownMenuItem onClick={() => onEdit(secret)}>
                          <Edit className="h-4 w-4" />
                        Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-danger focus:text-danger"
                          onClick={() => onDelete(secret)}
                        >
                          <Trash2 className="h-4 w-4" />
                        Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
