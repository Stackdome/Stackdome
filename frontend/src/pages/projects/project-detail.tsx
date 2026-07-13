import { useEffect, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { format } from "date-fns";
import { Users, ChevronLeft } from "lucide-react";
import { useProjectMembers } from "./hooks/use-project-members";
import { MemberRowMenu } from "./components/member-row-menu";
import { AddMemberDialog } from "./components/add-member-dialog";
import { RenameProjectDialog } from "./components/rename-project-dialog";
import { PageHeader, EmptyState, StackdomeMark } from "@/components/branded";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useToast } from "@/components/ui/use-toast";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getProject, renameProject } from "@/api/projects";
import { getErrorMessage } from "@/api/client";
import type { Project } from "@/api/projects";

export default function ProjectDetailPage() {
  const { projectName } = useParams<{ projectName: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();

  const [renameOpen, setRenameOpen] = useState(false);

  const [project, setProject] = useState<Project | null>(null);
  const [projectLoading, setProjectLoading] = useState(false);
  const [projectError, setProjectError] = useState<string | null>(null);

  const [addOpen, setAddOpen] = useState(false);
  const [search, setSearch] = useState("");

  const { members, loading, error, refetch, addMember, changeRole, removeMember } =
    useProjectMembers(projectName ?? "");

  // Fetch single project for the header
  useEffect(() => {
    if (!projectName) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    setProjectLoading(true);
    setProjectError(null);
    getProject(orgId, projectName)
      .then(setProject)
      .catch((e) => setProjectError(getErrorMessage(e)))
      .finally(() => setProjectLoading(false));
  }, [projectName]);

  // Filtered members by search
  const filteredMembers = search.trim()
    ? members.filter((m) => {
      const q = search.toLowerCase();
      const name = (m.user?.name ?? "").toLowerCase();
      const email = (m.user?.email ?? "").toLowerCase();
      return name.includes(q) || email.includes(q);
    })
    : members;

  const existingIds = members.map((m) => m.user_id ?? m.user?.id ?? "").filter(Boolean);

  async function handleAddMember(userId: string, role: "Developer" | "Viewer") {
    const result = await addMember(userId, role);
    if (result.ok) {
      toast({ title: "Member added", description: "The user has been added to the project.", variant: "success" });
    } else {
      toast({ title: "Failed to add member", description: result.error, variant: "destructive" });
    }
    return result;
  }

  function getInitials(name?: string, email?: string): string {
    if (name) {
      const parts = name.trim().split(" ");
      return parts.length >= 2
        ? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
        : name.slice(0, 2).toUpperCase();
    }
    return email ? email.slice(0, 2).toUpperCase() : "??";
  }

  const headerTitle = projectLoading ? (
    <Skeleton className="h-8 w-40" />
  ) : project ? (
    <span className="flex items-center gap-3">
      <StackdomeMark size={28} />
      {project.name}
      {project.default_project && (
        <Badge variant="secondary" className="text-xs">default</Badge>
      )}
    </span>
  ) : (
    (projectName ?? "")
  );

  return (
    <div className="p-8 space-y-8">
      {/* Back nav */}
      <Link
        to="/settings/projects"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ChevronLeft className="size-4" />
        Projects
      </Link>

      <PageHeader
        eyebrow="Settings / Projects"
        title={headerTitle}
        actionsAlign="center"
        actions={
          <>
            <Button
              variant="outline"
              disabled={project?.default_project ?? true}
              title={project?.default_project ? "Default project cannot be renamed" : undefined}
              onClick={() => setRenameOpen(true)}
            >
              Rename
            </Button>
            <Button onClick={() => setAddOpen(true)}>
              Add member
            </Button>
          </>
        }
      />

      {projectError && (
        <p className="text-sm text-destructive">{projectError}</p>
      )}

      {/* Search input — only shown when there are members to search */}
      {!loading && !error && members.length > 0 && (
        <Input
          placeholder="Search members…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-sm"
        />
      )}

      {loading ? (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">User</TableHead>
                <TableHead className="font-medium">Org role</TableHead>
                <TableHead className="font-medium">Project role</TableHead>
                <TableHead className="font-medium">Joined</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i} className="hover:bg-transparent">
                  <td className="p-2">
                    <div className="flex items-center gap-3">
                      <Skeleton className="size-7 rounded-full shrink-0" />
                      <div className="space-y-1.5">
                        <Skeleton className="h-3 w-28" />
                        <Skeleton className="h-2.5 w-36" />
                      </div>
                    </div>
                  </td>
                  <td className="p-2"><Skeleton className="h-5 w-20" /></td>
                  <td className="p-2"><Skeleton className="h-5 w-20" /></td>
                  <td className="p-2"><Skeleton className="h-3 w-20" /></td>
                  <td className="p-2" />
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ) : error ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="Couldn't load members"
          description={error}
          action={
            <Button variant="outline" onClick={refetch}>
              Retry
            </Button>
          }
        />
      ) : members.length === 0 ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="No members yet"
          description="Add people to this project to get started."
          action={
            <Button onClick={() => setAddOpen(true)}>
              Add first member
            </Button>
          }
        />
      ) : filteredMembers.length === 0 ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="No members match"
          description={`No members found for "${search}".`}
          dashed={false}
        />
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">User</TableHead>
                <TableHead className="font-medium">Org role</TableHead>
                <TableHead className="font-medium">Project role</TableHead>
                <TableHead className="font-medium">Joined</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMembers.map((m) => {
                const name = m.user?.name;
                const email = m.user?.email;
                const orgRole = m.user?.role as string | undefined;
                const projectRole = m.role as "Developer" | "Viewer" | undefined;
                return (
                  <TableRow key={m.id}>
                    <TableCell className="p-2">
                      <div className="flex items-center gap-3">
                        <Avatar className="size-7 shrink-0">
                          <AvatarFallback className="text-[10px]">
                            {getInitials(name, email)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="text-sm font-medium leading-tight">
                            {name ?? email ?? m.user_id}
                          </p>
                          {name && email && (
                            <p className="text-xs text-muted-foreground">{email}</p>
                          )}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="p-2">
                      {orgRole ? (
                        <Badge variant="outline">{orgRole}</Badge>
                      ) : (
                        <span className="text-muted-foreground text-sm">—</span>
                      )}
                    </TableCell>
                    <TableCell className="p-2">
                      {projectRole ? (
                        <Badge variant="secondary">{projectRole}</Badge>
                      ) : (
                        <span className="text-muted-foreground text-sm">—</span>
                      )}
                    </TableCell>
                    <TableCell className="p-2 text-sm text-muted-foreground">
                      {m.created_at
                        ? format(new Date(m.created_at), "MMM d, yyyy")
                        : "—"}
                    </TableCell>
                    <TableCell className="p-2 text-right">
                      <MemberRowMenu
                        membershipId={m.id ?? ""}
                        currentRole={projectRole}
                        memberName={name ?? email}
                        onChangeRole={changeRole}
                        onRemove={removeMember}
                      />
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}

      <AddMemberDialog
        open={addOpen}
        onOpenChange={setAddOpen}
        onAdd={handleAddMember}
        existingMemberUserIds={existingIds}
      />

      <RenameProjectDialog
        open={renameOpen}
        currentName={project?.name ?? ""}
        onOpenChange={setRenameOpen}
        onRename={async (newName) => {
          const orgId = getCurrentOrganizationId();
          if (!orgId || !project) return { ok: false as const, error: "No project loaded" };
          const oldName = project.name;
          try {
            await renameProject(orgId, oldName, { name: newName });
            toast({
              title: "Project renamed",
              description: `"${oldName}" is now "${newName}".`,
              variant: "success",
            });
            setRenameOpen(false);
            // The detail route is keyed by project name — move to the new slug.
            void navigate(`/settings/projects/${encodeURIComponent(newName)}`, { replace: true });
            return { ok: true as const };
          } catch (e) {
            return { ok: false as const, error: getErrorMessage(e) };
          }
        }}
      />
    </div>
  );
}
