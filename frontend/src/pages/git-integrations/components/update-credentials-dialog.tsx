import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { updateGitIntegration, type GitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { updateCredentialsFormSchema } from "@/components/git-source-picker/credentials-form-schema";
import { providerIdFor, PROVIDER_DISPLAY_NAMES } from "@/lib/git-integrations";
import { ProviderLogo } from "@/components/branded/provider-logo";

interface UpdateCredentialsDialogProps {
  integration: GitIntegration | null;
  onOpenChange: (open: boolean) => void;
  /** Fired after a successful update so the page can refresh the list. */
  onUpdated: () => void;
}

export function UpdateCredentialsDialog({ integration, onOpenChange, onUpdated }: UpdateCredentialsDialogProps) {
  const { toast } = useToast();
  const [username, setUsername] = useState("");
  const [token, setToken] = useState("");
  const [saving, setSaving] = useState(false);
  const [tokenError, setTokenError] = useState<string | undefined>(undefined);

  const integrationId = integration?.id;
  useEffect(() => {
    if (integrationId == null) return;
    // Credentials are write-only in the API, so there is nothing to prefill.
    setUsername("");
    setToken("");
    setTokenError(undefined);
  }, [integrationId]);

  const submit = async () => {
    const parsed = updateCredentialsFormSchema.safeParse({ username, token });
    if (!parsed.success) {
      setTokenError(parsed.error.flatten().fieldErrors.token?.[0]);
      return;
    }
    setTokenError(undefined);
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration?.id) {
      toast({
        title: "Couldn't update credentials",
        description: !orgId ? "No organization selected." : "Integration is no longer available.",
        variant: "destructive",
      });
      return;
    }
    setSaving(true);
    try {
      const trimmedUsername = parsed.data.username?.trim();
      await updateGitIntegration(orgId, integration.id, {
        host: integration.host,
        auth: trimmedUsername
          ? { basic: { username: trimmedUsername, password: parsed.data.token } }
          : { token: parsed.data.token },
      });
      toast({ title: "Credentials updated", variant: "success" });
      onOpenChange(false);
      onUpdated();
    } catch (e) {
      toast({ title: "Couldn't update credentials", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={integration != null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle>Update credentials</DialogTitle>
          <DialogDescription>
            Replaces the stored credentials for this integration. Existing clones keep working once the new
            credentials are saved.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {integration && (
            <div className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 px-3 py-2">
              <ProviderLogo providerId={providerIdFor(integration)} className="h-5 w-5 shrink-0" />
              <div className="min-w-0">
                <p className="text-body font-medium text-foreground">
                  {PROVIDER_DISPLAY_NAMES[providerIdFor(integration)]}
                </p>
                <p className="truncate font-mono text-label text-fg-muted">{integration.host}</p>
              </div>
            </div>
          )}
          <FieldShell
            label="Username"
            htmlFor="update-credentials-username"
            hint="Required for providers using basic auth (e.g. Bitbucket app passwords)."
          >
            <Input
              id="update-credentials-username"
              autoComplete="off"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </FieldShell>
          <FieldShell
            label="Access token"
            htmlFor="update-credentials-token"
            required
            error={tokenError}
          >
            <Input
              id="update-credentials-token"
              type="password"
              autoComplete="new-password"
              value={token}
              onChange={(e) => {
                setToken(e.target.value);
                setTokenError(undefined);
              }}
              aria-invalid={!!tokenError}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button shape="flat" variant="ghost" onClick={() => onOpenChange(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Update credentials
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
