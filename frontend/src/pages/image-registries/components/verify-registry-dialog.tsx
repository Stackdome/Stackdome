import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { verifyRegistryCredential, type RegistryCredential } from "@/api/registry-credentials";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { verifyRegistrySchema } from "../lib/form-schemas";

interface VerifyRegistryDialogProps {
  credential: RegistryCredential | null;
  onOpenChange: (open: boolean) => void;
}

export function VerifyRegistryDialog({ credential, onOpenChange }: VerifyRegistryDialogProps) {
  const { toast } = useToast();
  const [repository, setRepository] = useState("");
  const [verifying, setVerifying] = useState(false);
  const [fieldError, setFieldError] = useState<string | undefined>(undefined);

  const credentialId = credential?.id;
  useEffect(() => {
    if (credentialId == null) return;
    setRepository("");
    setFieldError(undefined);
  }, [credentialId]);

  const submit = async () => {
    const parsed = verifyRegistrySchema.safeParse({ repository });
    if (!parsed.success) {
      setFieldError(parsed.error.flatten().fieldErrors.repository?.[0]);
      return;
    }
    setFieldError(undefined);
    const orgId = getCurrentOrganizationId();
    if (!orgId || !credential?.id) {
      toast({
        title: "Couldn't verify registry access",
        description: !orgId ? "No organization selected." : "Registry is no longer available.",
        variant: "destructive",
      });
      return;
    }
    setVerifying(true);
    try {
      await verifyRegistryCredential(orgId, credential.id, parsed.data.repository);
      toast({ title: "Registry access verified" });
      onOpenChange(false);
    } catch (e) {
      toast({ title: "Couldn't verify registry access", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setVerifying(false);
    }
  };

  return (
    <Dialog open={credential != null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle>Verify registry access</DialogTitle>
          <DialogDescription>
            Confirms the {credential?.host} credentials can access a repository.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <FieldShell
            label="Repository"
            htmlFor="verify-repository"
            required
            error={fieldError}
            hint="Repository path on this registry, e.g. acme/app."
          >
            <Input
              id="verify-repository"
              placeholder="acme/app"
              value={repository}
              onChange={(e) => {
                setRepository(e.target.value);
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
