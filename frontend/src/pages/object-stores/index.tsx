import { useEffect } from "react";
import { PlusCircle, AlertCircle, Loader2, Cloud } from "lucide-react";
import { useObjectStores } from "./hooks/use-object-stores";
import { ObjectStoreList } from "./components/object-store-list";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

export default function ObjectStoresPage() {
  const { objectStores, loading, error, refetch } = useObjectStores();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

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
            <Button onClick={() => {}}>
              <PlusCircle className="h-4 w-4" />
              New Object Store
            </Button>
          }
        />

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
                <Button onClick={() => {}} variant="outline">
                  <PlusCircle className="h-4 w-4" />
                  New Object Store
                </Button>
              }
            />
          ) : (
            <ObjectStoreList
              objectStores={objectStores}
              onEdit={() => {}}
              onDelete={() => {}}
            />
          )}
        </Panel>

        {/* Form and delete dialogs are wired in Tasks 8 and 9 */}
      </div>
    </TooltipProvider>
  );
}
