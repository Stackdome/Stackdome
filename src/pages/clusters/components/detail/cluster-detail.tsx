import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useDeleteCluster } from "../../hooks/use-clusters";
import * as clusterApi from "@/api/clusters";
import { Button } from "@/components/ui/button";

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [cluster, setCluster] = useState<clusterApi.Cluster | null>(null);
  const [loading, setLoading] = useState(true);
  const { deleteCluster, loading: deleting, error: deleteError, success } = useDeleteCluster();

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    clusterApi.getCluster(clusterApi.getCurrentOrganizationId?.(), id)
      .then(setCluster)
      .finally(() => setLoading(false));
  }, [id]);

  const handleDelete = async () => {
    if (!id) return;
    await deleteCluster(id);
    navigate("/clusters");
  };

  if (loading) return <div>Loading...</div>;
  if (!cluster) return <div>Cluster not found.</div>;

  return (
    <div className="max-w-xl mx-auto py-8">
      <h1 className="text-2xl font-bold mb-6">Cluster Details</h1>
      <div className="mb-4">
        <div><strong>Name:</strong> {cluster.name}</div>
        <div><strong>API server URL:</strong> {cluster.cluster_url}</div>
        <div><strong>CA cert:</strong> {cluster.cluster_ca_data}</div>
        <div><strong>Service account token:</strong> {cluster.cluster_sa_token}</div>
      </div>
      {deleteError && <p className="text-red-500 text-sm">{deleteError}</p>}
      <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
        {deleting ? "Deleting..." : "Delete Cluster"}
      </Button>
    </div>
  );
}
