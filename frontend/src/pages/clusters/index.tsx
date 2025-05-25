import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Boxes, PlusCircle, AlertCircle, Loader2 } from "lucide-react";
import { useClusters } from "./hooks/use-clusters";
import { ClusterList } from "./components/cluster-list";
import { ClusterDeleteDialog } from "./components/cluster-delete-dialog";
import AddClusterDialog from "./components/add-cluster-dialog";
import type { Cluster } from "./types";
import type { ClusterData } from "./hooks/use-clusters";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useToast } from "@/components/ui/use-toast";
import { deleteCluster, createCluster } from "@/api/clusters";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

export default function ClustersPage() {
  const { clusters, loading, error, refetch } = useClusters();
  const [deletingCluster, setDeletingCluster] = useState<Cluster | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { toast } = useToast();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  // Automatically redirect to cluster detail page if a cluster exists
  useEffect(() => {
    if (!loading && clusters.length === 1) {
      navigate(`/clusters/${clusters[0].id}`);
    }
  }, [clusters, loading, navigate]);

  // Set breadcrumb
  useEffect(() => {
    const currentPath = `/clusters`;
    setCustomLabel(currentPath, "Clusters");
    setPathLoading(currentPath, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  function handleEdit(cluster: Cluster) {
    navigate(`/clusters/${cluster.id}`);
  }

  function handleDelete(cluster: Cluster) {
    setDeletingCluster(cluster);
  }

  async function handleDeleteConfirm() {
    if (!deletingCluster?.id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      console.error('No organization selected');
      return;
    }
    setDeleteLoading(true);
    try {
      await deleteCluster(orgId, deletingCluster.id);
      refetch();
      toast({
        title: "Cluster unlinked",
        description: "The cluster has been unlinked successfully.",
        variant: "destructive",
      });
    } catch (e) {
      console.error('Failed to unlink cluster:', e);
      toast({
        title: "Error",
        description: "Failed to unlink cluster. Please try again.",
        variant: "destructive",
      });
    } finally {
      setDeleteLoading(false);
      setDeletingCluster(null);
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
      <div className="p-6">
        <header className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3 mb-1">
                <h1 className="text-2xl font-bold">Cluster management</h1>
              </div>
            </div>
          </div>
          <Separator className="mt-4" />
        </header>

        <Card className="rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl flex items-center gap-2">
              Clusters
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {clusters.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20">
                <Boxes className="h-12 w-12 mb-4 text-muted-foreground" />
                <h3 className="text-xl font-medium mb-2">No clusters configured</h3>
                <p className="text-muted-foreground mb-6">Link to your cluster to get started.</p>
                <Button onClick={() => setShowAddDialog(true)}>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Add Cluster
                </Button>
              </div>
            ) : (
              <div className="space-y-0">
                <div className="border rounded-lg">
                  <ClusterList clusters={clusters} onEdit={handleEdit} onDelete={handleDelete} />
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <AddClusterDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
          onAddCluster={handleAddCluster}
          isLoading={createLoading}
          error={createError}
        />

        <ClusterDeleteDialog
          open={!!deletingCluster}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingCluster(null)}
          loading={deleteLoading}
        />
      </div>
    </TooltipProvider>
  );
}
