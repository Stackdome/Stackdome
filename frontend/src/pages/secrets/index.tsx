import { useState, useEffect, useMemo } from "react";
import { PlusCircle, AlertCircle, Loader2, KeyRound, Search, ChevronDown } from "lucide-react";
import { useSecrets } from "./hooks/use-secrets";
import { SecretList, formatSecretType } from "./components/secret-list";
import { SecretFormDialog } from "./components/secret-form-dialog";
import type { Secret } from "./types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { cn } from "@/lib/utils";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import { deleteSecret, createSecret, updateSecret } from "@/api/secrets";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useCurrentUser } from "@/hooks/use-current-user";

type SortKey = "created" | "name";

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "created", label: "Recently created" },
  { key: "name", label: "Name (A–Z)" },
];

const ALL_TYPES = "all";

export default function SecretsPage() {
  const { secrets, loading, error, refetch } = useSecrets();
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [query, setQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>(ALL_TYPES);
  const [sortKey, setSortKey] = useState<SortKey>("created");
  const [editingSecret, setEditingSecret] = useState<Secret | null>(null);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const { toast } = useToast();
  const confirm = useConfirm();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { projectNameById, defaultProjectName } = useResourceProjects();
  const { canWrite, canWriteAnyProject } = useCurrentUser();

  // Set breadcrumb
  useEffect(() => {
    const currentPath = `/secrets`;
    setCustomLabel(currentPath, "Secrets");
    setPathLoading(currentPath, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  function handleEdit(secret: Secret) {
    setEditingSecret(secret);
    setShowAddDialog(true);
  }

  async function requestDelete(secret: Secret) {
    if (!secret.id) return;
    const ok = await confirm({
      title: "Delete secret?",
      description: `This permanently deletes “${secret.name}”. This cannot be undone.`,
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      toast({ title: "Failed to delete secret", description: "No organization selected.", variant: "destructive" });
      return;
    }
    const projectName = projectNameById(secret.project_id);
    if (!projectName) {
      toast({ title: "Failed to delete secret", description: "Could not resolve the project for this secret.", variant: "destructive" });
      return;
    }
    try {
      await deleteSecret(orgId, projectName, secret.id);
      refetch();
      toast({
        title: "Secret deleted",
        description: "The secret has been deleted successfully.",
        variant: "success",
      });
    } catch (e) {
      console.error('Failed to delete secret:', e);
      toast({
        title: "Failed to delete secret",
        description: "Failed to delete secret. Please try again.",
        variant: "destructive",
      });
    }
  }

  async function handleCreateOrUpdateSecret(secretData: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">) {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setFormError("No organization selected");
      return;
    }

    setFormLoading(true);
    setFormError(null);

    try {
      if (editingSecret?.id) {
        // Update existing secret — target the secret's own project.
        const projectName = projectNameById(editingSecret.project_id);
        if (!projectName) {
          setFormError("Could not resolve the project for this secret.");
          return;
        }
        await updateSecret(orgId, projectName, editingSecret.id, secretData);
        toast({
          title: "Secret updated",
          description: "The secret has been updated successfully.",
          variant: "success",
        });
      } else {
        // Create new secret in the user's default project.
        if (!defaultProjectName) {
          setFormError("You don't have a project to create secrets in.");
          return;
        }
        await createSecret(orgId, defaultProjectName, secretData);
        toast({
          title: "Secret created",
          description: "The secret has been created successfully.",
          variant: "success",
        });
      }
      refetch();
      setShowAddDialog(false);
      setEditingSecret(null);
    } catch (e) {
      console.error('Failed to save secret:', e);
      setFormError(getErrorMessage(e));
    } finally {
      setFormLoading(false);
    }
  }

  function handleCloseDialog() {
    setShowAddDialog(false);
    setEditingSecret(null);
    setFormError(null);
  }

  // Types present in the data drive the TYPE filter dropdown.
  const typeOptions = useMemo(() => {
    const counts = new Map<string, number>();
    for (const s of secrets) counts.set(s.type, (counts.get(s.type) ?? 0) + 1);
    return [...counts.entries()]
      .map(([type, count]) => ({ type, count }))
      .sort((a, b) => a.type.localeCompare(b.type));
  }, [secrets]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const rows = secrets.filter(
      (s) =>
        (typeFilter === ALL_TYPES || s.type === typeFilter) &&
        (q === "" || s.name?.toLowerCase().includes(q) || s.description?.toLowerCase().includes(q)),
    );
    return [...rows].sort((a, b) =>
      sortKey === "name"
        ? (a.name ?? "").localeCompare(b.name ?? "")
        : (b.created_at ?? "").localeCompare(a.created_at ?? ""),
    );
  }, [secrets, query, typeFilter, sortKey]);

  const sortLabel = SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? "";

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading secrets...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading Secrets</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => window.location.reload()}>
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title="Secrets"
          subtitle="Manage sensitive data like API keys, passwords, and certificates"
          actions={
            canWriteAnyProject ? (
              <Button onClick={() => setShowAddDialog(true)}>
                <PlusCircle className="h-4 w-4" />
                Create Secret
              </Button>
            ) : undefined
          }
        />

        {secrets.length === 0 ? (
          <EmptyState
            icon={<KeyRound className="h-8 w-8" />}
            title="No secrets yet"
            description="Create your first secret to securely store sensitive data."
            action={
              canWriteAnyProject ? (
                <Button onClick={() => setShowAddDialog(true)}>
                  <PlusCircle className="h-4 w-4" />
                  Create Secret
                </Button>
              ) : undefined
            }
          />
        ) : (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <div className="relative w-full min-w-[220px] max-w-[340px]">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Filter secrets…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="pl-9 h-9"
                />
              </div>
              <div className="ml-auto flex items-center gap-2">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                    >
                      Type:{" "}
                      <span className="text-foreground">
                        {typeFilter === ALL_TYPES ? "All" : formatSecretType(typeFilter)}
                      </span>
                      <ChevronDown className="h-3 w-3 flex-none" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent
                    align="end"
                    className="min-w-[200px]"
                    onCloseAutoFocus={(e) => e.preventDefault()}
                  >
                    <DropdownMenuItem
                      onSelect={() => setTypeFilter(ALL_TYPES)}
                      className={cn(
                        "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                        typeFilter === ALL_TYPES && "text-brand"
                      )}
                    >
                      <span>All</span>
                      <span className="tabular-nums opacity-80">{secrets.length}</span>
                    </DropdownMenuItem>
                    {typeOptions.map((o) => (
                      <DropdownMenuItem
                        key={o.type}
                        onSelect={() => setTypeFilter(o.type)}
                        className={cn(
                          "justify-between font-mono text-[11px] uppercase tracking-[1.5px]",
                          typeFilter === o.type && "text-brand"
                        )}
                      >
                        <span>{formatSecretType(o.type)}</span>
                        <span className="tabular-nums opacity-80">{o.count}</span>
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 h-8 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground hover:bg-muted/50"
                    >
                      Sort: <span className="text-foreground">{sortLabel}</span>
                      <ChevronDown className="h-3 w-3 flex-none" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent
                    align="end"
                    className="min-w-[200px]"
                    onCloseAutoFocus={(e) => e.preventDefault()}
                  >
                    {SORT_OPTIONS.map((o) => (
                      <DropdownMenuItem
                        key={o.key}
                        onSelect={() => setSortKey(o.key)}
                        className={cn(
                          "font-mono text-[11px] uppercase tracking-[1.5px]",
                          sortKey === o.key && "text-brand"
                        )}
                      >
                        {o.label}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            {filtered.length === 0 ? (
              <EmptyState
                icon={<Search className="h-8 w-8" />}
                title="No secrets match"
                description="Try a different search or type filter."
              />
            ) : (
              <Panel title="Organization Secrets" count={filtered.length} bodyClassName="p-0">
                <SecretList
                  secrets={filtered}
                  onEdit={handleEdit}
                  onDelete={requestDelete}
                  canWrite={(projectId?: string) => canWrite(projectId ?? "")}
                />
              </Panel>
            )}
          </div>
        )}

        <SecretFormDialog
          open={showAddDialog}
          onOpenChange={handleCloseDialog}
          onSubmit={handleCreateOrUpdateSecret}
          isLoading={formLoading}
          error={formError}
          editingSecret={editingSecret}
        />
      </div>
    </TooltipProvider>
  );
}
