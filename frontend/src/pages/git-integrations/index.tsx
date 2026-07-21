import { useCallback, useEffect, useState } from "react";
import { Loader2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, Panel } from "@/components/branded";
import { AddIntegrationWizard } from "./add-integration-wizard";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/components/ui/use-toast";
import {
  listGitIntegrations, deleteGitIntegration,
  type GitIntegration,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { GIT_INTEGRATION_TYPE_GITHUB_APP } from "./lib/derive-row";
import { IntegrationsErrorState, IntegrationsEmptyState } from "./components/page-states";
import { IntegrationRow } from "./components/integration-row";
import { VerifyIntegrationDialog } from "./components/verify-integration-dialog";
import { UpdateCredentialsDialog } from "./components/update-credentials-dialog";

export default function GitIntegrationsPage() {
  const { toast } = useToast();
  const [integrations, setIntegrations] = useState<GitIntegration[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [verifying, setVerifying] = useState<GitIntegration | null>(null);
  const [removing, setRemoving] = useState<GitIntegration | null>(null);
  const [editing, setEditing] = useState<GitIntegration | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setError("No organization selected.");
      setLoading(false);
      return;
    }
    try {
      const list = await listGitIntegrations(orgId);
      setIntegrations(list.items ?? []);
      setError(null);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const remove = async (integration: GitIntegration) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration.id) {
      toast({
        title: "Remove failed",
        description: !orgId ? "No organization selected." : "Integration is no longer available.",
        variant: "destructive",
      });
      return;
    }
    try {
      await deleteGitIntegration(orgId, integration.id);
      toast({ title: "Integration removed" });
      setRemoving(null);
      await refresh();
    } catch (e) {
      toast({ title: "Remove failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  const hasGithubApp = integrations.some((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);
  const addButton = (
    <Button onClick={() => setWizardOpen(true)}>
      <Plus className="h-4 w-4" />
      Connect provider
    </Button>
  );

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading git integrations...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <PageHeader
        eyebrow="Integrations"
        title="Git integrations"
        subtitle="Grant Stackdome access to your repositories for clones, builds, and preview environments."
        actions={addButton}
      />

      {/* Full-page error only when there's nothing to show; a failed re-fetch
          keeps the already-loaded list visible with an inline error line. */}
      {error && integrations.length === 0 && (
        <IntegrationsErrorState message={error} onRetry={() => void refresh()} />
      )}

      {!error && integrations.length === 0 && (
        <IntegrationsEmptyState onAdd={() => setWizardOpen(true)} />
      )}

      {integrations.length > 0 && (
        <>
          {error && <p className="text-sm text-destructive">Couldn&apos;t refresh integrations: {error}</p>}
          <Panel title="Connected providers" count={integrations.length}>
            <div className="divide-y divide-border">
              {integrations.map((integration) => (
                <IntegrationRow
                  key={integration.id}
                  integration={integration}
                  onVerify={setVerifying}
                  onRemove={setRemoving}
                  onUpdateCredentials={setEditing}
                />
              ))}
            </div>
          </Panel>
        </>
      )}

      <AddIntegrationWizard
        open={wizardOpen}
        onOpenChange={setWizardOpen}
        hasGithubApp={hasGithubApp}
        onCreated={() => void refresh()}
      />

      <VerifyIntegrationDialog
        integration={verifying}
        onOpenChange={(o) => !o && setVerifying(null)}
      />

      <UpdateCredentialsDialog
        integration={editing}
        onOpenChange={(o) => !o && setEditing(null)}
        onUpdated={() => void refresh()}
      />

      <AlertDialog open={removing != null} onOpenChange={(o) => !o && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove this integration?</AlertDialogTitle>
            <AlertDialogDescription>
              Repositories using this integration lose access for clones.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => removing && void remove(removing)}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
