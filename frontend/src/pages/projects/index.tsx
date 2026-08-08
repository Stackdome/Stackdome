import { useState } from "react";
import { Link } from "react-router-dom";
import { format } from "date-fns";
import { Users } from "lucide-react";
import { useProjects } from "./hooks/use-projects";
import { CreateProjectDialog } from "./components/create-project-dialog";
import { RenameProjectDialog } from "./components/rename-project-dialog";
import { DeleteProjectDialog } from "./components/delete-project-dialog";
import { ProjectRowMenu } from "./components/project-row-menu";
import { PageHeader, EmptyState, StackdomeMark } from "@/components/branded";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useToast } from "@/components/ui/use-toast";
import type { Project } from "@/api/projects";

export default function ProjectsPage() {
  const { projects, loading, error, refetch, create, rename, remove, onlyDefault } = useProjects();
  const { toast } = useToast();
  const [createOpen, setCreateOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState<Project | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Project | null>(null);

  async function handleCreate(name: string) {
    const result = await create(name);
    if (result.ok) {
      toast({
        title: "Project created",
        description: `"${name}" has been created successfully.`,
        variant: "success",
      });
    }
    return result;
  }

  return (
    <div className="p-8 space-y-8">
      <PageHeader
        eyebrow="Settings"
        title="Projects"
        actionsAlign="center"
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            Create project
          </Button>
        }
      />

      {loading ? (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">Project</TableHead>
                <TableHead className="font-medium">Members</TableHead>
                <TableHead className="font-medium">Created</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 4 }).map((_, i) => (
                <TableRow key={i} className="hover:bg-transparent">
                  <TableCell className="p-2">
                    <div className="flex items-center gap-2">
                      <Skeleton className="size-5 rounded shrink-0" />
                      <Skeleton className="h-3 w-32" />
                    </div>
                  </TableCell>
                  <TableCell className="p-2"><Skeleton className="h-3 w-24" /></TableCell>
                  <TableCell className="p-2"><Skeleton className="h-3 w-20" /></TableCell>
                  <TableCell className="p-2" />
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ) : error ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="Couldn't load projects"
          description={error}
          action={
            <Button variant="outline" onClick={refetch}>
              Retry
            </Button>
          }
        />
      ) : (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className="font-medium">Project</TableHead>
                  <TableHead className="font-medium">Members</TableHead>
                  <TableHead className="font-medium">Created</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {projects.map((project) => (
                  <TableRow key={project.id}>
                    <TableCell className="p-2">
                      <div className="flex items-center gap-2">
                        <StackdomeMark size={18} />
                        <span className="font-medium text-body">{project.name}</span>
                        {project.default_project && (
                          <Badge variant="secondary" className="text-label">default</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="p-2">
                      <Link
                        to={`/settings/projects/${encodeURIComponent(project.name)}`}
                        className="text-body text-muted-foreground hover:text-foreground transition-colors"
                      >
                        Manage members
                      </Link>
                    </TableCell>
                    <TableCell className="p-2 text-body text-muted-foreground">
                      {project.created_at
                        ? format(new Date(project.created_at), "MMM d, yyyy")
                        : "—"}
                    </TableCell>
                    <TableCell className="p-2 text-right">
                      <ProjectRowMenu
                        project={project}
                        onRename={(t) => setRenameTarget(t)}
                        onDelete={(t) => setDeleteTarget(t)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {onlyDefault && (
            <EmptyState
              icon={<Users className="h-8 w-8" />}
              title="No additional projects"
              description="Create projects to organize members and control access to resources across your organization."
              action={
                <Button variant="outline" onClick={() => setCreateOpen(true)}>
                  Create project
                </Button>
              }
            />
          )}
        </>
      )}

      <CreateProjectDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreate={handleCreate}
      />

      <RenameProjectDialog
        open={!!renameTarget}
        currentName={renameTarget?.name ?? ""}
        onOpenChange={(o) => { if (!o) setRenameTarget(null); }}
        onRename={async (newName) => {
          if (!renameTarget) return { ok: false as const, error: "No project selected" };
          const oldName = renameTarget.name;
          const result = await rename(oldName, newName);
          if (result.ok) {
            toast({
              title: "Project renamed",
              description: `"${oldName}" is now "${newName}".`,
              variant: "success",
            });
            setRenameTarget(null);
          }
          return result;
        }}
      />

      <DeleteProjectDialog
        open={!!deleteTarget}
        projectName={deleteTarget?.name ?? ""}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null); }}
        onConfirm={async () => {
          if (!deleteTarget) return;
          const name = deleteTarget.name;
          const result = await remove(name);
          if (result.ok) {
            toast({
              title: "Project deleted",
              description: `"${name}" has been deleted.`,
              variant: "success",
            });
            setDeleteTarget(null);
          } else {
            toast({
              title: "Failed to delete project",
              description: result.error,
              variant: "destructive",
            });
          }
        }}
      />
    </div>
  );
}
