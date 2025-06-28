import { useState, useEffect } from "react";
import { PlusCircle, AlertCircle, Loader2, KeyRound } from "lucide-react";
import { useSecrets } from "./hooks/use-secrets";
import { SecretList } from "./components/secret-list";
import { SecretDeleteDialog } from "./components/secret-delete-dialog";
import { SecretFormDialog } from "./components/secret-form-dialog";
import type { Secret } from "./types";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TooltipProvider } from "@/components/ui/tooltip";
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
      <div className="p-6">
        <header className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3 mb-1">
                <h1 className="text-2xl font-bold">Secrets management</h1>
              </div>
              <p className="text-muted-foreground">
                Manage sensitive data like API keys, passwords, and certificates securely.
              </p>
            </div>
            <Button onClick={() => setShowAddDialog(true)}>
              <PlusCircle className="mr-2 h-4 w-4" />
              Create Secret
            </Button>
          </div>
          <Separator className="mt-4" />
        </header>

        <Card className="rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl flex items-center gap-2">
              <KeyRound className="h-5 w-5" />
              Organization Secrets
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {secrets.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20">
                <KeyRound className="h-12 w-12 mb-4 text-muted-foreground" />
                <h3 className="text-xl font-medium mb-2">No secrets found</h3>
                <p className="text-muted-foreground mb-6">Create your first secret to securely store sensitive data.</p>
                <Button onClick={() => setShowAddDialog(true)}>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Create Secret
                </Button>
              </div>
            ) : (
              <div className="space-y-0">
                <div className="border rounded-lg">
                  <SecretList secrets={secrets} onEdit={handleEdit} onDelete={handleDelete} />
                </div>
              </div>
            )}
          </CardContent>
        </Card>

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
