import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { format } from "date-fns";
import { Users, ChevronLeft } from "lucide-react";
import { useTeamMembers } from "./hooks/use-team-members";
import { MemberRowMenu } from "./components/member-row-menu";
import { AddMemberDialog } from "./components/add-member-dialog";
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
import { getTeam } from "@/api/teams";
import { getErrorMessage } from "@/api/client";
import type { Team } from "@/api/teams";

export default function TeamDetailPage() {
  const { teamName } = useParams<{ teamName: string }>();
  const { toast } = useToast();

  const [team, setTeam] = useState<Team | null>(null);
  const [teamLoading, setTeamLoading] = useState(false);
  const [teamError, setTeamError] = useState<string | null>(null);

  const [addOpen, setAddOpen] = useState(false);
  const [search, setSearch] = useState("");

  const { members, loading, error, refetch, addMember, changeRole, removeMember } =
    useTeamMembers(teamName ?? "");

  // Fetch single team for the header
  useEffect(() => {
    if (!teamName) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    setTeamLoading(true);
    setTeamError(null);
    getTeam(orgId, teamName)
      .then(setTeam)
      .catch((e) => setTeamError(getErrorMessage(e)))
      .finally(() => setTeamLoading(false));
  }, [teamName]);

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
      toast({ title: "Member added", description: "The user has been added to the team." });
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

  const headerTitle = teamLoading ? (
    <Skeleton className="h-8 w-40" />
  ) : team ? (
    <span className="flex items-center gap-3">
      <StackdomeMark size={28} />
      {team.name}
      {team.default_team && (
        <Badge variant="secondary" className="text-xs">default</Badge>
      )}
    </span>
  ) : (
    (teamName ?? "")
  );

  return (
    <div className="p-8 space-y-8">
      {/* Back nav */}
      <Link
        to="/settings/teams"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ChevronLeft className="size-4" />
        Teams
      </Link>

      <PageHeader
        eyebrow="Settings / Teams"
        title={headerTitle}
        actionsAlign="center"
        actions={
          <>
            <Button
              variant="outline"
              disabled={team?.default_team ?? true}
              title={team?.default_team ? "Default team cannot be renamed" : undefined}
            >
              Rename
            </Button>
            <Button onClick={() => setAddOpen(true)}>
              Add member
            </Button>
          </>
        }
      />

      {teamError && (
        <p className="text-sm text-destructive">{teamError}</p>
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
                <TableHead className="font-medium">Team role</TableHead>
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
          description="Add people to this team to get started."
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
                <TableHead className="font-medium">Team role</TableHead>
                <TableHead className="font-medium">Joined</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMembers.map((m) => {
                const name = m.user?.name;
                const email = m.user?.email;
                const orgRole = m.user?.role as string | undefined;
                const teamRole = m.role as "Developer" | "Viewer" | undefined;
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
                      {teamRole ? (
                        <Badge variant="secondary">{teamRole}</Badge>
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
                        currentRole={teamRole}
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
    </div>
  );
}
