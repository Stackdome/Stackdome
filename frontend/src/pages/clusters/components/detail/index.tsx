import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useDeleteCluster } from "../../hooks/use-clusters";
import * as clusterApi from "@/api/clusters";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Trash2, Loader2, Boxes } from "lucide-react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Switch } from "@/components/ui/switch";
import { PageHeader, Panel, FieldShell, StatusPill, EmptyState } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { statusVariant } from "@/components/branded/status-variant";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useToast } from "@/components/ui/use-toast";

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [cluster, setCluster] = useState<clusterApi.Cluster | null>(null);
  const [loading, setLoading] = useState(true);
  const { deleteCluster } = useDeleteCluster();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();
  const confirm = useConfirm();

  useEffect(() => {
    if (!id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;

    const currentPath = `/clusters/${id}`;
    setPathLoading(currentPath, true);
    setLoading(true);
    clusterApi.getCluster(orgId, id)
      .then((data) => {
        setCluster(data);
        if (data && data.name) {
          setCustomLabel(currentPath, data.name);
        }
      })
      .finally(() => {
        setPathLoading(currentPath, false);
        setLoading(false);
      });
  }, [id, setCustomLabel, setPathLoading]);

  const handleDelete = async () => {
    if (!id) return;
    const ok = await confirm({
      title: "Delete cluster?",
      description: "This action cannot be undone.",
      confirmLabel: "Delete",
      variant: "destructive",
    });
    if (!ok) return;
    try {
      await deleteCluster(id);
      toast({
        title: "Cluster deleted",
        description: "The cluster has been deleted successfully.",
        variant: "success",
        duration: 3000,
      });
      navigate("/clusters");
    } catch (err) {
      console.error("Failed to delete cluster:", err);
      toast({
        title: "Failed to delete cluster",
        description: getErrorMessage(err),
        variant: "destructive",
        duration: 5000,
      });
    }
  };

  if (loading) return (
    <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
      <Loader2 className="h-10 w-10 animate-spin text-primary" />
      <p className="mt-2 text-muted-foreground">Loading cluster details...</p>
    </div>
  );

  if (!cluster) return (
    <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-8">
      <EmptyState
        icon={<Boxes className="h-8 w-8" />}
        title="Cluster not found"
        description="The cluster you're looking for doesn't exist or you don't have access."
        action={<Button onClick={() => navigate("/clusters")}>Back to Clusters</Button>}
      />
    </div>
  );

  const registryState = cluster.cluster_image_registry?.status?.state ?? "Unknown";
  const registryLabel = (() => {
    if (registryState === "ImageRegistryRunning") return "Running";
    if (registryState === "ImageRegistryError") return "Error";
    if (registryState === "ImageRegistryPending") return "Pending";
    return registryState;
  })();

  return (
    <TooltipProvider>
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title={cluster.name}
          subtitle={`ID: ${cluster.id}`}
          actions={
            <Button
              variant="outline"
              onClick={() => void handleDelete()}
              className="border-danger-border text-danger hover:bg-danger-bg hover:text-danger"
            >
              <Trash2 className="h-4 w-4" />
              Delete Cluster
            </Button>
          }
        />

        <Panel title="Cluster Configuration">
          <div className="space-y-5 max-w-3xl">
            <FieldShell label="API Server URL">
              <Input
                value={cluster.cluster_url}
                disabled
                className="font-mono bg-muted"
              />
            </FieldShell>

            <FieldShell label="CA Certificate" hint="Encrypted at rest. Re-create the cluster to rotate.">
              <Input
                type="password"
                value="••••••••••••••••••••••"
                disabled
                readOnly
                className="font-mono bg-muted"
              />
            </FieldShell>

            <FieldShell label="Service Account Token" hint="Encrypted at rest. Re-create the cluster to rotate.">
              <Input
                type="password"
                value="••••••••••••••••••••••"
                disabled
                readOnly
                className="font-mono bg-muted"
              />
            </FieldShell>

            <div className="space-y-3 pt-2">
              <div className="flex items-start gap-3">
                <Switch
                  checked={!!cluster.cluster_image_registry}
                  disabled
                  className="mt-0.5"
                />
                <div className="space-y-1">
                  <Label className="text-body font-medium text-foreground">
                    Stackdome Image Registry
                  </Label>
                  <p className="text-meta text-muted-foreground leading-relaxed">
                    Provides a private registry inside this cluster for build artifacts.
                  </p>
                </div>
              </div>

              {cluster.cluster_image_registry && (
                <div className="grid grid-cols-2 gap-6 pl-11">
                  <div>
                    <Label className="text-body font-medium text-foreground">Registry Size</Label>
                    <p className="font-mono text-body text-muted-foreground mt-1">
                      {cluster.cluster_image_registry.spec?.backend_storage_size || "—"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-body font-medium text-foreground">Registry Status</Label>
                    <div className="mt-1.5">
                      <StatusPill variant={statusVariant("registry", registryLabel)}>{registryLabel}</StatusPill>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </Panel>
      </div>
    </TooltipProvider>
  );
}
