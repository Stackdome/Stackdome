import { useCallback, useEffect, useRef, useState } from "react";
import { Github, RefreshCw, ShieldCheck, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/use-toast";
import {
  listGitIntegrations, deleteGitIntegration, listInstallations, verifyGitIntegration,
  type GitIntegration, type GitInstallation,
} from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useGithubConnect } from "@/pages/previews/hooks/use-github-connect";

function IntegrationInstallations({ integration }: { integration: GitIntegration }) {
  const { toast } = useToast();
  const [installs, setInstalls] = useState<GitInstallation[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const requestSeq = useRef(0);

  const load = useCallback(async (refresh: boolean) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration.id) return;
    const seq = ++requestSeq.current;
    setRefreshing(true);
    try {
      const list = await listInstallations(orgId, integration.id, refresh);
      if (seq === requestSeq.current) setInstalls(list.items ?? []);
    } catch (e) {
      if (seq === requestSeq.current) {
        toast({ title: "Failed to load installations", description: getErrorMessage(e), variant: "destructive" });
      }
    } finally {
      if (seq === requestSeq.current) setRefreshing(false);
    }
  }, [integration.id, toast]);

  useEffect(() => {
    void load(false);
  }, [load]);

  return (
    <div className="mt-2 space-y-1 border-l pl-4">
      {installs.map((i) => (
        <p key={i.id} className="text-xs text-muted-foreground">
          <span className="font-mono">{i.account_login}</span> · {i.repository_selection} repositories
        </p>
      ))}
      {installs.length === 0 && (
        <p className="text-xs text-muted-foreground">No installations yet.</p>
      )}
      <Button variant="ghost" size="sm" onClick={() => void load(true)} disabled={refreshing}>
        <RefreshCw className="h-3.5 w-3.5" />
        Refresh
      </Button>
    </div>
  );
}

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
  const { state: connectState, error: connectError, connect } = useGithubConnect();
  const [integrations, setIntegrations] = useState<GitIntegration[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [verifying, setVerifying] = useState<GitIntegration | null>(null);

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

  useEffect(() => {
    if (connectState === "connected") void refresh();
  }, [connectState, refresh]);

  const remove = async (integration: GitIntegration) => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration.id) return;
    try {
      await deleteGitIntegration(orgId, integration.id);
      toast({ title: "Integration deleted" });
      await refresh();
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  const hasGithubApp = integrations.some((i) => i.type === "github_app");

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Git integrations</h1>
          <p className="text-sm text-muted-foreground">
            Access to your repositories for preview environments and builds.
          </p>
        </div>
        {!loading && !hasGithubApp && (
          <Button onClick={() => void connect()}>
            <Github className="h-4 w-4" />
            Connect GitHub
          </Button>
        )}
      </div>

      {connectState === "waiting" && (
        <p className="text-sm text-muted-foreground">
          Waiting for the GitHub App installation to finish in the popup…
        </p>
      )}
      {connectError && <p className="text-sm text-destructive">{connectError}</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}
      {loading && <p className="text-sm text-muted-foreground">Loading…</p>}

      <div className="space-y-3">
        {integrations.map((integration) => (
          <div key={integration.id} className="rounded-lg border p-4">
            <div className="flex items-center gap-3">
              <span className="font-mono text-sm">{integration.host}</span>
              <Badge variant="outline">{integration.type}</Badge>
              <Badge variant={integration.status === "pending_install" ? "secondary" : "default"}>
                {integration.status}
              </Badge>
              {integration.credentials_configured && (
                <Badge variant="outline" className="text-muted-foreground">credentials set</Badge>
              )}
              <span className="flex-1" />
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Verify ${integration.host} integration`}
                onClick={() => setVerifying(integration)}
              >
                <ShieldCheck className="h-4 w-4" />
              </Button>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="ghost" size="icon" aria-label={`Delete ${integration.host} integration`}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete this integration?</AlertDialogTitle>
                    <AlertDialogDescription>
                      Repositories using this integration lose access for clones.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={() => void remove(integration)}>Delete</AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
            {integration.type === "github_app" && (
              <IntegrationInstallations integration={integration} />
            )}
          </div>
        ))}
        {!loading && integrations.length === 0 && (
          <p className="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground">
            No git integrations yet. Connect GitHub to get started.
          </p>
        )}
      </div>

      <VerifyIntegrationDialog
        integration={verifying}
        onOpenChange={(o) => !o && setVerifying(null)}
      />
    </div>
  );
}
