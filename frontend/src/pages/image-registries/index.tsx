import { useCallback, useEffect, useState } from "react";
import { Loader2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader, Panel } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import {
  listRegistryCredentials, deleteRegistryCredential,
  type RegistryCredential,
} from "@/api/registry-credentials";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { RegistriesErrorState, RegistriesEmptyState } from "./components/page-states";
import { RegistryRow } from "./components/registry-row";
import { AddRegistryDialog } from "./components/add-registry-dialog";
import { UpdateCredentialsDialog } from "./components/update-credentials-dialog";
import { VerifyRegistryDialog } from "./components/verify-registry-dialog";

export default function ImageRegistriesPage() {
  const { toast } = useToast();
  const [credentials, setCredentials] = useState<RegistryCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [editing, setEditing] = useState<RegistryCredential | null>(null);
  const [verifying, setVerifying] = useState<RegistryCredential | null>(null);
  const confirm = useConfirm();

  const refresh = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setError("No organization selected.");
      setLoading(false);
      return;
    }
    try {
      const list = await listRegistryCredentials(orgId);
      setCredentials(list.items ?? []);
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

  const remove = async (credential: RegistryCredential) => {
    const ok = await confirm({
      title: "Remove this registry?",
      description: "Stacks referencing these credentials lose pull/push access.",
      confirmLabel: "Remove",
      variant: "destructive",
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId || !credential.id) {
      toast({
        title: "Remove failed",
        description: !orgId ? "No organization selected." : "Registry is no longer available.",
        variant: "destructive",
      });
      return;
    }
    try {
      const res = await deleteRegistryCredential(orgId, credential.id);
      const affected = res.affected_stacks ?? [];
      if (affected.length > 0) {
        toast({
          title: "Registry removed",
          description: `Stacks affected: ${affected.map((s) => s.name ?? s.id).join(", ")}. Their image pulls or pushes may fail until another credential covers ${credential.host}.`,
          variant: "warning",
        });
      } else {
        toast({ title: "Registry removed", variant: "success" });
      }
      await refresh();
    } catch (e) {
      toast({ title: "Remove failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  const addButton = (
    <Button onClick={() => setAdding(true)}>
      <Plus className="h-4 w-4" />
      Add registry
    </Button>
  );

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading image registries...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <PageHeader
        eyebrow="Integrations"
        title="Image registries"
        subtitle="Store registry credentials so builds can pull private images and push artifacts."
        actions={addButton}
      />

      {/* Full-page error only when there's nothing to show; a failed re-fetch
          keeps the already-loaded list visible with an inline error line. */}
      {error && credentials.length === 0 && (
        <RegistriesErrorState message={error} onRetry={() => void refresh()} />
      )}

      {!error && credentials.length === 0 && <RegistriesEmptyState onAdd={() => setAdding(true)} />}

      {credentials.length > 0 && (
        <>
          {error && <p className="text-sm text-destructive">Couldn&apos;t refresh registries: {error}</p>}
          <Panel title="Connected registries" count={credentials.length}>
            <div className="divide-y divide-border">
              {credentials.map((credential) => (
                <RegistryRow
                  key={credential.id}
                  credential={credential}
                  onVerify={setVerifying}
                  onUpdateCredentials={setEditing}
                  onRemove={(c) => void remove(c)}
                />
              ))}
            </div>
          </Panel>
        </>
      )}

      <AddRegistryDialog open={adding} onOpenChange={setAdding} onCreated={() => void refresh()} />

      <UpdateCredentialsDialog
        credential={editing}
        onOpenChange={(o) => !o && setEditing(null)}
        onUpdated={() => void refresh()}
      />

      <VerifyRegistryDialog
        credential={verifying}
        onOpenChange={(o) => !o && setVerifying(null)}
      />

    </div>
  );
}
