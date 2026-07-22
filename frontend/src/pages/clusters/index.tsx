import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Boxes, PlusCircle, AlertCircle, Loader2 } from "lucide-react";
import { useClusters } from "./hooks/use-clusters";
import { ClusterList } from "./components/cluster-list";
import AddClusterDialog from "./components/add-cluster-dialog";
import type { Cluster } from "./types";
import type { ClusterData } from "./hooks/use-clusters";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState, LimitedAction } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import { deleteCluster, createCluster } from "@/api/clusters";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

export default function ClustersPage() {
  const { clusters, loading, error, refetch } = useClusters();
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  // Set breadcrumb
  useEffect(() => {
    const currentPath = `/clusters`;
    setCustomLabel(currentPath, "Clusters");
    setPathLoading(currentPath, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  function handleEdit(cluster: Cluster) {
    navigate(`/clusters/${cluster.id}`);
  }

  async function handleDelete(cluster: Cluster) {
    if (!cluster.id) return;
    const ok = await confirm({
      title: "Delete cluster?",
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      console.error('No organization selected');
      return;
    }
    try {
      await deleteCluster(orgId, cluster.id);
      refetch();
      toast({
        title: "Cluster unlinked",
        description: "The cluster has been unlinked successfully.",
        variant: "success",
      });
    } catch (e) {
      console.error('Failed to unlink cluster:', e);
      toast({
        title: "Failed to unlink cluster",
        description: "Failed to unlink cluster. Please try again.",
        variant: "destructive",
      });
    }
  }

  async function handleAddCluster(clusterData: ClusterData) {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setCreateError("No organization selected");
      return;
    }

    setCreateLoading(true);
    setCreateError(null);

    try {
      await createCluster(orgId, clusterData);
      refetch();
      setShowAddDialog(false);
      toast({
        title: "Cluster added",
        description: "The cluster has been added successfully.",
        variant: "success",
      });
    } catch (e) {
      console.error('Failed to create cluster:', e);
      setCreateError(getErrorMessage(e));
    } finally {
      setCreateLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading clusters...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading Clusters</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => window.location.reload()}>
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
          title="Clusters"
          subtitle="Compute targets for your stacks"
          actions={
            <LimitedAction
              limitReached={clusters.length >= 1}
              limitMessage="Currently only one cluster is supported."
            >
              <Button onClick={() => setShowAddDialog(true)}>
                <PlusCircle className="h-4 w-4" />
                Add Cluster
              </Button>
            </LimitedAction>
          }
        />

        {clusters.length === 0 ? (
          <EmptyState
            icon={<Boxes className="h-8 w-8" />}
            title="No clusters configured"
            description="Link to your cluster to get started."
            action={
              <Button onClick={() => setShowAddDialog(true)}>
                <PlusCircle className="h-4 w-4" />
                Add Cluster
              </Button>
            }
          />
        ) : (
          <Panel title="All Clusters" count={clusters.length} bodyClassName="p-0">
            <ClusterList clusters={clusters} onEdit={handleEdit} onDelete={(cluster) => void handleDelete(cluster)} />
          </Panel>
        )}

        <AddClusterDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
          onAddCluster={handleAddCluster}
          isLoading={createLoading}
          error={createError}
        />
      </div>
    </TooltipProvider>
  );
}
