import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ClusterCreateForm from "./cluster-create-form";
import { createCluster } from "@/api/clusters";
import type { ClusterData } from "../../hooks/use-clusters";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useToast } from "@/components/ui/use-toast";
import { extractErrorMessage } from "@/lib/utils";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";

export default function ClusterCreatePage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { toast } = useToast();

  async function handleSubmit(values: ClusterData) {
    setLoading(true);
    setError(null);
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setLoading(false);
      const errorMessage = "No organization selected";
      setError(errorMessage);
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
      return;
    }
    try {
      // Always send cluster_image_registry as an object or omit it
      const apiPayload: Omit<ClusterData, 'id'> & { cluster_image_registry?: { name: string; spec: { backend_storage_size: string } } } = { ...values };
      if (typeof values.cluster_image_registry === 'boolean') {
        if (values.cluster_image_registry) {
          apiPayload.cluster_image_registry = { name: "default-registry", spec: { backend_storage_size: "20Gi" } };
        } else {
          delete apiPayload.cluster_image_registry;
        }
      }
      await createCluster(orgId, apiPayload);
      setLoading(false);

      // Show success toast notification
      toast({
        title: "Success",
        description: "Cluster created successfully",
        variant: "success",
        duration: 3000,
      });

      navigate("/clusters");
    } catch (e: unknown) {
      console.error("Failed to create cluster:", e);

      const errorMessage = extractErrorMessage(
        e as Error,
        "Failed to create cluster. Please check your connection and try again."
      );

      setError(errorMessage);

      // Show error toast notification
      toast({
        title: "Failed to create cluster",
        description: errorMessage,
        variant: "destructive",
        duration: 5000,
      });

      setLoading(false);
    }
  }

  function handleCancel() {
    navigate("/clusters");
  }

  return (
    <div className="p-6">
      <header className="mb-6">
        <div className="flex justify-between items-center">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold">Create New Cluster</h1>
            </div>
            <div className="flex items-center gap-4 text-muted-foreground text-sm mb-1">
              <span>Configure your Kubernetes cluster connection</span>
            </div>
          </div>
          <div className="flex gap-3">
            <Button variant="outline" size="lg" onClick={handleCancel}>
              <X className="h-4 w-4" />
              Cancel
            </Button>
          </div>
        </div>
        <Separator className="mt-4" />
      </header>

      <div className="max-w-3xl">
        <ClusterCreateForm
          onSubmit={handleSubmit}
          loading={loading}
          error={error}
        />
      </div>
    </div>
  );
}
