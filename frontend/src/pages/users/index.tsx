import { useState } from "react";
import { Users, UserX } from "lucide-react";
import { useUsers } from "./hooks/use-users";
import type { UserRowModel } from "./hooks/use-users";
import { UserRow } from "./components/user-row";
import { PendingRow } from "./components/pending-row";
import { PageHeader, EmptyState } from "@/components/branded";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
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

function extractTeams(rows: UserRowModel[]): string[] {
  const set = new Set<string>();
  for (const row of rows) {
    if (row.kind === "active") {
      row.teams.forEach((t) => { if (t.team_name) set.add(t.team_name); });
    } else if (row.team_name) {
      set.add(row.team_name);
    }
  }
  return Array.from(set).sort();
}

export default function UsersPage() {
  const { rows, loading, error, refetch } = useUsers();
  const [search, setSearch] = useState("");
  const [roleTab, setRoleTab] = useState<RoleTab>("all");
  const [team, setTeam] = useState("all");

  const allTeams = extractTeams(rows);
  const filtered = filterRows(rows, search, roleTab, team);
  const filtersActive = hasActiveFilters(search, roleTab, team);

  function clearFilters() {
    setSearch("");
    setRoleTab("all");
    setTeam("all");
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        eyebrow="Settings"
        title="Users"
        actions={
          <Button onClick={() => {}}>
            Invite user
          </Button>
        }
      />

      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-3">
        <Input
          placeholder="Search by name or email…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="h-9 w-[260px]"
        />
        <Tabs value={roleTab} onValueChange={(v) => setRoleTab(v as RoleTab)}>
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="active">Active</TabsTrigger>
            <TabsTrigger value="invited">Invited</TabsTrigger>
          </TabsList>
        </Tabs>
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

      {/* States */}
      {loading ? (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">User</TableHead>
                <TableHead className="font-medium">Role</TableHead>
                <TableHead className="font-medium">Teams</TableHead>
                <TableHead className="font-medium">Last active</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 8 }).map((_, i) => (
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
                  <td className="p-2"><Skeleton className="h-5 w-40" /></td>
                  <td className="p-2"><Skeleton className="h-3 w-16" /></td>
                  <td className="p-2" />
                </TableRow>
              ))}
            </TableBody>
          </Table>
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
            <Button variant="outline" onClick={() => {}}>
              Invite user
            </Button>
          }
        />
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={<Users className="h-8 w-8" />}
          title="No users match these filters"
          action={
            <Button variant="outline" onClick={clearFilters}>
              Clear filters
            </Button>
          }
        />
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="font-medium">User</TableHead>
                <TableHead className="font-medium">Role</TableHead>
                <TableHead className="font-medium">Teams</TableHead>
                <TableHead className="font-medium">Last active</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((row) =>
                row.kind === "pending" ? (
                  <PendingRow key={row.id} row={row} />
                ) : (
                  <UserRow key={row.id} row={row} />
                ),
              )}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
