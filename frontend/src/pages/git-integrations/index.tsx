import { useCallback, useEffect, useState } from "react";
import { Loader2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, Panel } from "@/components/branded";
import { AddIntegrationWizard } from "@/components/git-source-picker/add-integration-wizard";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import {
  listGitIntegrations, deleteGitIntegration,
  type GitIntegration,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useGithubConnect } from "@/hooks/use-github-connect";
import { GIT_INTEGRATION_TYPE_GITHUB_APP } from "@/lib/git-integrations";
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
  const [editing, setEditing] = useState<GitIntegration | null>(null);
  const confirm = useConfirm();
  const [wizardOpen, setWizardOpen] = useState(false);
  const { setCustomLabel } = useBreadcrumb();
  const github = useGithubConnect();

  useEffect(() => {
    setCustomLabel("/git-integrations", "Git providers");
  }, [setCustomLabel]);

  // A failed install callback lands back here with the reason in the URL
  // (the popup itself relays it to its opener and closes).
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const setupError = params.get("setup_error");
    if (!setupError) return;
    toast({ title: "GitHub App install failed", description: setupError, variant: "destructive" });
    params.delete("setup_error");
    const query = params.toString();
    window.history.replaceState(null, "", window.location.pathname + (query ? `?${query}` : ""));
  }, [toast]);

  // "Add GitHub account" runs the same connect flow as the wizard; the hook
  // polls until the new installation is bound.
  useEffect(() => {
    if (github.state === "connected") void refresh();
    if (github.state === "error" && github.error) {
      toast({ title: "GitHub App install failed", description: github.error, variant: "destructive" });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [github.state, github.error]);

  const addAccount = () => {
    void github.connect();
    toast({ title: "Finish the install in the GitHub popup", description: "Pick the account or organization to add." });
  };

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
    const ok = await confirm({
      title: "Remove this integration?",
      description: "Repositories using this integration lose access for clones.",
      confirmLabel: "Remove",
      variant: "destructive",
    });
    if (!ok) return;
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
      toast({ title: "Integration removed", variant: "success" });
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
        title="Git providers"
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
                  onRemove={(i) => void remove(i)}
                  onUpdateCredentials={setEditing}
                  onAddAccount={addAccount}
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

    </div>
  );
}
