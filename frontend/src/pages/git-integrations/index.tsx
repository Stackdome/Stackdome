import { useCallback, useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, Panel } from "@/components/branded";
import { AddIntegrationWizard } from "./add-integration-wizard";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/use-toast";
import {
  listGitIntegrations, deleteGitIntegration, verifyGitIntegration,
  type GitIntegration,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { deriveRow, GIT_INTEGRATION_TYPE_GITHUB_APP } from "./lib/derive-row";
import { SummaryStrip } from "./components/summary-strip";
import { IntegrationsErrorState, IntegrationsEmptyState } from "./components/page-states";
import { IntegrationRow } from "./components/integration-row";

interface VerifyIntegrationDialogProps {
  integration: GitIntegration | null;
  onOpenChange: (open: boolean) => void;
}

function VerifyIntegrationDialog({ integration, onOpenChange }: VerifyIntegrationDialogProps) {
  const { toast } = useToast();
  const [repoUrl, setRepoUrl] = useState("");
  const [verifying, setVerifying] = useState(false);

  const integrationId = integration?.id;
  useEffect(() => {
    if (integrationId == null) return;
    setRepoUrl("");
  }, [integrationId]);

  const submit = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration?.id || !repoUrl.trim()) return;
    setVerifying(true);
    try {
      await verifyGitIntegration(orgId, integration.id, repoUrl.trim());
      toast({ title: "Verification succeeded" });
      onOpenChange(false);
    } catch (e) {
      toast({ title: "Verification failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setVerifying(false);
    }
  };

  return (
    <Dialog open={integration != null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle>Verify repository access</DialogTitle>
          <DialogDescription>
            Confirms the {integration?.host} integration can access a repository.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="verify-repo-url">Repository URL</Label>
            <Input
              id="verify-repo-url"
              placeholder="https://github.com/acme/webapp"
              value={repoUrl}
              onChange={(e) => setRepoUrl(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={verifying}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={verifying || !repoUrl.trim()}>
            Verify
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function GitIntegrationsPage() {
  const { toast } = useToast();
  const [integrations, setIntegrations] = useState<GitIntegration[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [verifying, setVerifying] = useState<GitIntegration | null>(null);
  const [removing, setRemoving] = useState<GitIntegration | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
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
    if (!orgId || !integration.id) return;
    try {
      await deleteGitIntegration(orgId, integration.id);
      toast({ title: "Integration deleted" });
      setRemoving(null);
      await refresh();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  const hasGithubApp = integrations.some((i) => i.type === GIT_INTEGRATION_TYPE_GITHUB_APP);
  const addButton = (
    <Button onClick={() => setWizardOpen(true)}>
      <Plus className="h-4 w-4" />
      Add integration
    </Button>
  );

  const rows = integrations.map((integration) => deriveRow(integration));

  return (
    <div className="space-y-6 p-6">
      <PageHeader
        title="Git integrations"
        subtitle="Grant Stackdome access to your repositories for clones, builds, and preview environments."
        actions={addButton}
      />

      {error && <IntegrationsErrorState message={error} onRetry={() => void refresh()} />}

      {!loading && !error && integrations.length === 0 && (
        <IntegrationsEmptyState onAdd={() => setWizardOpen(true)} />
      )}

      {!error && integrations.length > 0 && (
        <>
          <SummaryStrip rows={rows} />
          <Panel title="Connected providers" count={integrations.length}>
            <div className="divide-y divide-border">
              {integrations.map((integration) => (
                <IntegrationRow
                  key={integration.id}
                  integration={integration}
                  onVerify={setVerifying}
                  onRemove={setRemoving}
                  onChanged={() => void refresh()}
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

      <AlertDialog open={removing != null} onOpenChange={(o) => !o && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this integration?</AlertDialogTitle>
            <AlertDialogDescription>
              Repositories using this integration lose access for clones.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => removing && void remove(removing)}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
