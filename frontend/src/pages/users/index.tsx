import { useState } from "react";
import { Search, Users, UserX } from "lucide-react";
import { useUsers } from "./hooks/use-users";
import { useTeamOptions } from "./hooks/use-team-options";
import { InviteDialog } from "./components/invite-dialog";
import type { UserRowModel } from "./hooks/use-users";
import { UserRow } from "./components/user-row";
import { PendingRow } from "./components/pending-row";
import { UserRowMenu } from "./components/user-row-menu";
import { PendingRowMenu } from "./components/pending-row-menu";
import { PageHeader, EmptyState } from "@/components/branded";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tabs,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type RoleTab = "all" | "active" | "invited";

function hasActiveFilters(search: string, roleTab: RoleTab, team: string): boolean {
  return search.trim() !== "" || roleTab !== "all" || team !== "all";
}

function filterRows(
  rows: UserRowModel[],
  search: string,
  roleTab: RoleTab,
  team: string,
): UserRowModel[] {
  const q = search.trim().toLowerCase();
  return rows.filter((row) => {
    // Role tab filter
    if (roleTab === "active" && row.kind !== "active") return false;
    if (roleTab === "invited" && row.kind !== "pending") return false;

    // Search filter (name/email)
    if (q) {
      if (row.kind === "active") {
        if (!row.name.toLowerCase().includes(q) && !row.email.toLowerCase().includes(q)) {
          return false;
        }
      } else {
        if (!row.email.toLowerCase().includes(q)) return false;
      }
    }

    // Team filter
    if (team !== "all") {
      if (row.kind === "active") {
        if (!row.teams.some((t) => t.team_name === team)) return false;
      } else {
        if (row.team_name !== team) return false;
      }
    }

    return true;
  });
}

const COLUMN_HEAD_CLASS = "text-[11px] uppercase tracking-widest text-muted-foreground font-mono";

function TabCount({ count }: { count: number }) {
  return <span className="ml-1.5 font-mono text-[11px] text-muted-foreground">{count}</span>;
}

export default function UsersPage() {
  const { rows, loading, error, refetch } = useUsers();
  const { teams: teamOptions } = useTeamOptions();
  const [search, setSearch] = useState("");
  const [roleTab, setRoleTab] = useState<RoleTab>("all");
  const [team, setTeam] = useState("all");
  const [inviteOpen, setInviteOpen] = useState(false);

  const allTeams = teamOptions.map((t) => t.name).sort();
  const filtered = filterRows(rows, search, roleTab, team);
  const filtersActive = hasActiveFilters(search, roleTab, team);

  // Tab counts
  const countAll = rows.length;
  const countActive = rows.filter((r) => r.kind === "active").length;
  const countInvited = rows.filter((r) => r.kind === "pending").length;

  // Default team name for chip star
  const defaultTeamName = teamOptions.find((t) => t.default_team)?.name;

  function clearFilters() {
    setSearch("");
    setRoleTab("all");
    setTeam("all");
  }

  const tableHeader = (
    <TableRow className="hover:bg-transparent">
      <TableHead className={COLUMN_HEAD_CLASS}>User</TableHead>
      <TableHead className={COLUMN_HEAD_CLASS}>Org role</TableHead>
      <TableHead className={COLUMN_HEAD_CLASS}>Teams</TableHead>
      <TableHead className={COLUMN_HEAD_CLASS}>Last active</TableHead>
      <TableHead />
    </TableRow>
  );

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Users"
        subtitle="Everyone in this organisation. Org-admins can manage roles and team memberships."
        actions={
          <Button onClick={() => setInviteOpen(true)}>
            Invite user
          </Button>
        }
      />

      {/* States */}
      {loading ? (
        <div className="rounded-md border border-border">
          {/* Toolbar skeleton */}
          <div className="p-3 flex items-center gap-3 flex-wrap">
            <Skeleton className="h-9 w-[280px]" />
            <Skeleton className="h-9 w-[200px]" />
            <Skeleton className="h-9 w-[180px]" />
          </div>
          <div className="border-t border-border">
            <Table>
              <TableHeader>{tableHeader}</TableHeader>
              <TableBody>
                {Array.from({ length: 8 }).map((_, i) => (
                  <TableRow key={i} className="hover:bg-transparent">
                    <TableCell className="py-3.5">
                      <div className="flex items-center gap-3">
                        <Skeleton className="h-7 w-7 rounded shrink-0" />
                        <div className="space-y-1.5">
                          <Skeleton className="h-3 w-28" />
                          <Skeleton className="h-2.5 w-36" />
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="py-3.5"><Skeleton className="h-5 w-20" /></TableCell>
                    <TableCell className="py-3.5"><Skeleton className="h-5 w-40" /></TableCell>
                    <TableCell className="py-3.5"><Skeleton className="h-3 w-16" /></TableCell>
                    <TableCell className="py-3.5" />
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      ) : error ? (
        <EmptyState
          icon={<UserX className="h-8 w-8" />}
          title="Couldn't load users"
          description={error}
          action={
            <Button variant="outline" onClick={refetch}>
              Retry
            </Button>
          }
        />
      ) : filtered.length === 0 && !filtersActive ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="No users yet"
          description="Invite your first teammate to get started."
          action={
            <Button variant="outline" onClick={() => setInviteOpen(true)}>
              Invite user
            </Button>
          }
        />
      ) : (
        <div className="rounded-md border border-border">
          {/* Toolbar */}
          <div className="p-3 flex items-center gap-3 flex-wrap">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
              <Input
                placeholder="Search by name or email"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-8 h-9 w-[280px]"
              />
            </div>

            {/* Tabs */}
            <Tabs value={roleTab} onValueChange={(v) => setRoleTab(v as RoleTab)}>
              <TabsList>
                <TabsTrigger value="all">
                  All<TabCount count={countAll} />
                </TabsTrigger>
                <TabsTrigger value="active">
                  Active<TabCount count={countActive} />
                </TabsTrigger>
                <TabsTrigger value="invited">
                  Invited<TabCount count={countInvited} />
                </TabsTrigger>
              </TabsList>
            </Tabs>

            {/* Team filter */}
            <Select value={team} onValueChange={setTeam}>
              <SelectTrigger className="h-9 w-[180px]">
                <SelectValue placeholder="All teams" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All teams</SelectItem>
                {allTeams.map((t) => (
                  <SelectItem key={t} value={t}>{t}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Divider + table */}
          <div className="border-t border-border">
            {filtered.length === 0 ? (
              <div className="py-10 flex flex-col items-center gap-3">
                <Users className="h-8 w-8 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">No users match these filters</p>
                <Button variant="outline" size="sm" onClick={clearFilters}>
                  Clear filters
                </Button>
              </div>
            ) : (
              <Table>
                <TableHeader>{tableHeader}</TableHeader>
                <TableBody>
                  {filtered.map((row) =>
                    row.kind === "pending" ? (
                      <PendingRow
                        key={row.id}
                        row={row}
                        defaultTeamName={defaultTeamName}
                        actions={<PendingRowMenu row={row} onChanged={refetch} />}
                      />
                    ) : (
                      <UserRow
                        key={row.id}
                        row={row}
                        defaultTeamName={defaultTeamName}
                        actions={<UserRowMenu row={row} onChanged={refetch} />}
                      />
                    ),
                  )}
                </TableBody>
              </Table>
            )}
          </div>
        </div>
      )}

      <InviteDialog
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        onCreated={refetch}
      />
    </div>
  );
}
