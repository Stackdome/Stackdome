import { useState } from "react";
import { Link } from "react-router-dom";
import { format } from "date-fns";
import { Users } from "lucide-react";
import { useTeams } from "./hooks/use-teams";
import { CreateTeamDialog } from "./components/create-team-dialog";
import { DeleteTeamDialog } from "./components/delete-team-dialog";
import { TeamRowMenu } from "./components/team-row-menu";
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
import type { Team } from "@/api/teams";

export default function TeamsPage() {
  const { teams, loading, error, refetch, create, remove, onlyDefault } = useTeams();
  const { toast } = useToast();
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Team | null>(null);

  async function handleCreate(name: string) {
    const result = await create(name);
    if (result.ok) {
      toast({
        title: "Team created",
        description: `"${name}" has been created successfully.`,
      });
    }
    return result;
  }

  return (
    <div className="p-8 space-y-8">
      <PageHeader
        eyebrow="Settings"
        title="Teams"
        actionsAlign="center"
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            Create team
          </Button>
        }
      />

      {loading ? (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">Team</TableHead>
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
          title="Couldn't load teams"
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
                  <TableHead className="font-medium">Team</TableHead>
                  <TableHead className="font-medium">Members</TableHead>
                  <TableHead className="font-medium">Created</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {teams.map((team) => (
                  <TableRow key={team.id}>
                    <TableCell className="p-2">
                      <div className="flex items-center gap-2">
                        <StackdomeMark size={18} />
                        <span className="font-medium text-sm">{team.name}</span>
                        {team.default_team && (
                          <Badge variant="secondary" className="text-[10px]">default</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="p-2">
                      <Link
                        to={`/settings/teams/${encodeURIComponent(team.name)}`}
                        className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                      >
                        Manage members
                      </Link>
                    </TableCell>
                    <TableCell className="p-2 text-sm text-muted-foreground">
                      {team.created_at
                        ? format(new Date(team.created_at), "MMM d, yyyy")
                        : "—"}
                    </TableCell>
                    <TableCell className="p-2 text-right">
                      <TeamRowMenu team={team} onDelete={(t) => setDeleteTarget(t)} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {onlyDefault && (
            <EmptyState
              icon={<Users className="h-8 w-8" />}
              title="No additional teams"
              description="Create teams to organize members and control access to resources across your organization."
              action={
                <Button variant="outline" onClick={() => setCreateOpen(true)}>
                  Create team
                </Button>
              }
            />
          )}
        </>
      )}

      <CreateTeamDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreate={handleCreate}
      />

      <DeleteTeamDialog
        open={!!deleteTarget}
        teamName={deleteTarget?.name ?? ""}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null); }}
        onConfirm={async () => {
          if (!deleteTarget) return;
          const name = deleteTarget.name;
          const result = await remove(name);
          if (result.ok) {
            toast({
              title: "Team deleted",
              description: `"${name}" has been deleted.`,
            });
            setDeleteTarget(null);
          } else {
            toast({
              title: "Failed to delete team",
              description: result.error,
              variant: "destructive",
            });
          }
        }}
      />
    </div>
  );
}
