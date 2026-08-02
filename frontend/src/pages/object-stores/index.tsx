import { useEffect, useState } from "react";
import { PlusCircle, AlertCircle, Loader2, Cloud } from "lucide-react";
import { useObjectStores } from "./hooks/use-object-stores";
import { ObjectStoreList } from "./components/object-store-list";
import { ObjectStoreFormDialog } from "./components/object-store-form-dialog";
import type { ObjectStore } from "./types";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import { getErrorMessage } from "@/api/client";
import { deleteObjectStore } from "@/api/object-stores";
import { getCurrentOrganizationId } from "@/lib/common";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useCurrentUser } from "@/hooks/use-current-user";

const IN_USE_FALLBACK = "This Object Store is in use by one or more Postgres add-ons.";

export default function ObjectStoresPage() {
  const { objectStores, loading, error, refetch } = useObjectStores();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { canWrite, canWriteAnyProject } = useCurrentUser();
  const { projectNameById } = useResourceProjects();
  const { toast } = useToast();
  const confirm = useConfirm();
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [editingStore, setEditingStore] = useState<ObjectStore | null>(null);

  async function requestDelete(store: ObjectStore) {
    if (!store.id) return;
    const ok = await confirm({
      title: "Delete object store?",
      description: (
        <>
          <span className="font-mono">{store.name}</span> will no longer be available as a backup
          destination. Existing backup files in the destination are not removed. This cannot be
          undone.
        </>
      ),
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;

    const orgId = getCurrentOrganizationId();
    const projectName = projectNameById(store.project_id);
    if (!orgId || !projectName) {
      toast({
        title: "Could not delete Object Store",
        description: orgId
          ? "Could not resolve the project for this object store."
          : "No organization selected.",
        variant: "destructive",
      });
      return;
    }

    try {
      await deleteObjectStore(orgId, projectName, store.id);
      toast({ title: "Object store deleted", variant: "success" });
      refetch();
    } catch (e: unknown) {
      toast({
        title: "Failed to delete",
        description: getErrorMessage(e) || IN_USE_FALLBACK,
        variant: "destructive",
      });
    }
  }

  useEffect(() => {
    const path = `/object-stores`;
    setCustomLabel(path, "Object Stores");
    setPathLoading(path, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading object stores...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-danger mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading Object Stores</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => refetch()}>
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title="Object Stores"
          subtitle="Backup destinations for Postgres add-ons. Supports AWS S3, S3-compatible (e.g. MinIO), Azure, and GCS."
          actions={
            canWriteAnyProject ? (
              <Button
                onClick={() => {
                  setEditingStore(null);
                  setShowAddDialog(true);
                }}
              >
                <PlusCircle className="h-4 w-4" />
                New Object Store
              </Button>
            ) : undefined
          }
        />

        {objectStores.length === 0 ? (
          <EmptyState
            icon={<Cloud className="h-8 w-8" />}
            title="No Object Stores yet"
            description="Add an S3-compatible bucket, Azure container, or GCS bucket to use as a backup destination."
            action={
              canWriteAnyProject ? (
                <Button
                  onClick={() => {
                    setEditingStore(null);
                    setShowAddDialog(true);
                  }}
                >
                  <PlusCircle className="h-4 w-4" />
                  New Object Store
                </Button>
              ) : undefined
            }
          />
        ) : (
          <Panel title="Organization Object Stores" count={objectStores.length} bodyClassName="p-0">
            <ObjectStoreList
              objectStores={objectStores}
              onEdit={(store) => {
                setEditingStore(store);
                setShowAddDialog(true);
              }}
              onDelete={(store) => void requestDelete(store)}
              canWrite={(projectId?: string) => canWrite(projectId ?? "")}
            />
          </Panel>
        )}

        <ObjectStoreFormDialog
          open={showAddDialog}
          onOpenChange={(open) => {
            setShowAddDialog(open);
            if (!open) setEditingStore(null);
          }}
          editing={editingStore}
          onSaved={() => {
            refetch();
          }}
        />
      </div>
    </TooltipProvider>
  );
}
