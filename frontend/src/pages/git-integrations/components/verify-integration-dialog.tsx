import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { verifyGitIntegration, type GitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { verifyIntegrationFormSchema } from "@/pages/git-integrations/lib/form-schemas";

interface VerifyIntegrationDialogProps {
  integration: GitIntegration | null;
  onOpenChange: (open: boolean) => void;
}

export function VerifyIntegrationDialog({ integration, onOpenChange }: VerifyIntegrationDialogProps) {
  const { toast } = useToast();
  const [repoUrl, setRepoUrl] = useState("");
  const [verifying, setVerifying] = useState(false);
  const [fieldError, setFieldError] = useState<string | undefined>(undefined);

  const integrationId = integration?.id;
  useEffect(() => {
    if (integrationId == null) return;
    setRepoUrl("");
    setFieldError(undefined);
  }, [integrationId]);

  const submit = async () => {
    const parsed = verifyIntegrationFormSchema.safeParse({ repoUrl });
    if (!parsed.success) {
      setFieldError(parsed.error.flatten().fieldErrors.repoUrl?.[0]);
      return;
    }
    setFieldError(undefined);
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
      await verifyGitIntegration(orgId, integration.id, parsed.data.repoUrl);
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
          <FieldShell label="Repository URL" htmlFor="verify-repo-url" required error={fieldError}>
            <Input
              id="verify-repo-url"
              placeholder="https://github.com/acme/webapp"
              value={repoUrl}
              onChange={(e) => {
                setRepoUrl(e.target.value);
                setFieldError(undefined);
              }}
              aria-invalid={!!fieldError}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={verifying}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={verifying}>
            Verify
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
