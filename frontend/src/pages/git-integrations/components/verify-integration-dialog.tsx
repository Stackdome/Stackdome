import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/use-toast";
import { verifyGitIntegration, type GitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";

interface VerifyIntegrationDialogProps {
  integration: GitIntegration | null;
  onOpenChange: (open: boolean) => void;
}

export function VerifyIntegrationDialog({ integration, onOpenChange }: VerifyIntegrationDialogProps) {
  const { toast } = useToast();
  const [repoUrl, setRepoUrl] = useState("");
  const [verifying, setVerifying] = useState(false);

  const integrationId = integration?.id;
  useEffect(() => {
    if (integrationId == null) return;
    setRepoUrl("");
  }, [integrationId]);

  const submit = async () => {
    if (!repoUrl.trim()) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration?.id) {
      toast({
        title: "Couldn't verify repository access",
        description: !orgId ? "No organization selected." : "Integration is no longer available.",
        variant: "destructive",
      });
      return;
    }
    setVerifying(true);
    try {
      await verifyGitIntegration(orgId, integration.id, repoUrl.trim());
      toast({ title: "Repository access verified" });
      onOpenChange(false);
    } catch (e) {
      toast({ title: "Couldn't verify repository access", description: getErrorMessage(e), variant: "destructive" });
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
