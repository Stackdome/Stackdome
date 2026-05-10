import { useEffect, useState } from "react";
import { PlusCircle, AlertCircle, Loader2, Cloud } from "lucide-react";
import { useObjectStores } from "./hooks/use-object-stores";
import { ObjectStoreList } from "./components/object-store-list";
import type { ObjectStore } from "./types";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

export default function ObjectStoresPage() {
  const { objectStores, loading, error, refetch: _refetch } = useObjectStores();
  // Dialog state is plumbed but unused until Tasks 8 and 9 wire form/delete dialogs.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [showAddDialog, setShowAddDialog] = useState(false);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [editingStore, setEditingStore] = useState<ObjectStore | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [deletingStore, setDeletingStore] = useState<ObjectStore | null>(null);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  useEffect(() => {
    const path = `/object-stores`;
    setCustomLabel(path, "Object Stores");
    setPathLoading(path, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  function handleEdit(store: ObjectStore) {
    setEditingStore(store);
    setShowAddDialog(true);
  }

  function handleDelete(store: ObjectStore) {
    setDeletingStore(store);
  }

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading object stores...</p>
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
            <Button
              onClick={() => {
                setEditingStore(null);
                setShowAddDialog(true);
              }}
            >
              <PlusCircle className="h-4 w-4" />
              New Object Store
            </Button>
          }
        />

        {error && (
          <div className="flex items-center gap-2 rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            <AlertCircle className="h-4 w-4" />
            <span>{error}</span>
          </div>
        )}

        <Panel
          title="Organization Object Stores"
          count={objectStores.length}
          bodyClassName={objectStores.length === 0 ? "p-5" : "p-0"}
        >
          {objectStores.length === 0 ? (
            <EmptyState
              icon={<Cloud className="h-8 w-8" />}
              title="No Object Stores yet"
              description="Add an S3-compatible bucket, Azure container, or GCS bucket to use as a backup destination."
              action={
                <Button onClick={() => setShowAddDialog(true)} variant="outline">
                  <PlusCircle className="h-4 w-4" />
                  New Object Store
                </Button>
              }
            />
          ) : (
            <ObjectStoreList
              objectStores={objectStores}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          )}
        </Panel>

        {/* Form and delete dialogs are wired in Tasks 8 and 9 */}
      </div>
    </TooltipProvider>
  );
}
