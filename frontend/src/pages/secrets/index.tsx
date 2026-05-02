import { useState, useEffect } from "react";
import { PlusCircle, AlertCircle, Loader2, KeyRound } from "lucide-react";
import { useSecrets } from "./hooks/use-secrets";
import { SecretList } from "./components/secret-list";
import { SecretDeleteDialog } from "./components/secret-delete-dialog";
import { SecretFormDialog } from "./components/secret-form-dialog";
import type { Secret } from "./types";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { deleteSecret, createSecret, updateSecret } from "@/api/secrets";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

export default function SecretsPage() {
  const { secrets, loading, error, refetch } = useSecrets();
  const [deletingSecret, setDeletingSecret] = useState<Secret | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [editingSecret, setEditingSecret] = useState<Secret | null>(null);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const { toast } = useToast();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

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

  function handleDelete(secret: Secret) {
    setDeletingSecret(secret);
  }

  async function handleDeleteConfirm() {
    if (!deletingSecret?.id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      console.error('No organization selected');
      return;
    }
    setDeleteLoading(true);
    try {
      await deleteSecret(orgId, deletingSecret.id);
      refetch();
      toast({
        title: "Secret deleted",
        description: "The secret has been deleted successfully.",
        variant: "destructive",
      });
    } catch (e) {
      console.error('Failed to delete secret:', e);
      toast({
        title: "Error",
        description: "Failed to delete secret. Please try again.",
        variant: "destructive",
      });
    } finally {
      setDeleteLoading(false);
      setDeletingSecret(null);
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
        // Update existing secret
        await updateSecret(orgId, editingSecret.id, secretData);
        toast({
          title: "Secret updated",
          description: "The secret has been updated successfully.",
        });
      } else {
        // Create new secret
        await createSecret(orgId, secretData);
        toast({
          title: "Secret created",
          description: "The secret has been created successfully.",
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
            <Button onClick={() => setShowAddDialog(true)} className="bg-brand text-white hover:bg-brand-darker">
              <PlusCircle className="h-4 w-4" />
              Create Secret
            </Button>
          }
        />

        <Panel
          title="Organization Secrets"
          count={secrets.length}
          bodyClassName={secrets.length === 0 ? "p-5" : "p-0"}
        >
          {secrets.length === 0 ? (
            <EmptyState
              icon={<KeyRound className="h-8 w-8" />}
              title="No secrets yet"
              description="Create your first secret to securely store sensitive data."
              action={
                <Button onClick={() => setShowAddDialog(true)} variant="outline">
                  <PlusCircle className="h-4 w-4" />
                  Create Secret
                </Button>
              }
            />
          ) : (
            <SecretList secrets={secrets} onEdit={handleEdit} onDelete={handleDelete} />
          )}
        </Panel>

        <SecretFormDialog
          open={showAddDialog}
          onOpenChange={handleCloseDialog}
          onSubmit={handleCreateOrUpdateSecret}
          isLoading={formLoading}
          error={formError}
          editingSecret={editingSecret}
        />

        <SecretDeleteDialog
          open={!!deletingSecret}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingSecret(null)}
          loading={deleteLoading}
          secretName={deletingSecret?.name}
        />
      </div>
    </TooltipProvider>
  );
}
