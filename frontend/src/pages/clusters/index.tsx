import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Boxes , PlusCircle } from "lucide-react";
import { useClusters } from "./hooks/use-clusters";
import { ClusterList } from "./components/cluster-list";
import { ClusterDeleteDialog } from "./components/cluster-delete-dialog";
import type { Cluster } from "./types";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { deleteCluster } from "@/api/clusters";
import { getCurrentOrganizationId } from "@/helpers/common";

export default function ClustersPage() {
  const { clusters, loading, error, refetch } = useClusters();
  const [deletingCluster, setDeletingCluster] = useState<Cluster | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const navigate = useNavigate();

  const showBlankSlate = clusters.length === 0;

  // Automatically redirect to cluster detail page if a cluster exists
  useEffect(() => {
    if (!loading && clusters.length === 1) {
      navigate(`/clusters/${clusters[0].id}`);
    }
  }, [clusters, loading, navigate]);

  function handleCreate() {
    navigate('/clusters/create');
  }

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
    } catch (e) {
      console.error('Failed to delete cluster:', e);
    } finally {
      setDeleteLoading(false);
      setDeletingCluster(null);
    }
  }

  return (
    <div className="flex flex-1 flex-col p-4 pt-0 h-full">
      <header className="flex h-16 shrink-0 items-center gap-2">
        <div className="flex items-center gap-2 px-4">
          <Separator orientation="vertical" className="mr-2 h-4" />
        </div>
      </header>
      {showBlankSlate ? (
        <div className="flex flex-col items-center justify-center h-[80vh] text-center">
          <div className="flex flex-col items-center max-w-md">
            <div className="rounded-full bg-primary/10 p-4 mb-4">
              <Boxes className="h-8 w-8 text-primary" />
            </div>
            <h2 className="text-2xl font-bold mb-2">No clusters created yet</h2>
            <p className="text-muted-foreground mb-6">
              Create your first cluster to get started.
            </p>
            <Button onClick={handleCreate}>
              <PlusCircle className="mr-2 h-4 w-4" />
              New Cluster
            </Button>
          </div>
        </div>
      ) : (
        <>
          {error && <div className="text-destructive mb-4">{error}</div>}
          <ClusterList clusters={clusters} onEdit={handleEdit} onDelete={handleDelete} />
        </>
      )}
      <ClusterDeleteDialog
        open={!!deletingCluster}
        onConfirm={handleDeleteConfirm}
        onCancel={() => setDeletingCluster(null)}
        loading={deleteLoading}
      />
    </div>
  );
}
