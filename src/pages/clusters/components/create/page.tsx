import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ClusterCreatePage from ".";
import { createCluster } from "@/api/clusters";
import type { ClusterFormInput } from "../../hooks/use-clusters";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useToast } from "@/components/ui/use-toast";
import { extractErrorMessage } from "@/lib/utils";

export default function ClusterCreateWrapperPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { toast } = useToast();

  async function handleSubmit(values: ClusterFormInput) {
    setLoading(true);
    setError(null);
    try {
      await createCluster(getCurrentOrganizationId(), values);
      setLoading(false);
      
      // Show success toast notification
      toast({
        title: "Success",
        description: "Cluster created successfully",
        variant: "success",
        duration: 3000,
      });
      
      navigate("/clusters");
    } catch (e) {
      console.error("Failed to create cluster:", e);
      
      const errorMessage = extractErrorMessage(
        e, 
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
    <div className="p-4 pt-0 h-full">
      <div className="flex h-[calc(100vh-64px)]">
        {/* Main Content - Full Width */}
        <div className="flex-grow p-6 overflow-y-auto">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-xl font-medium mb-6">Create New Cluster</h2>
            <ClusterCreatePage onSubmit={handleSubmit} loading={loading} error={error} onCancel={handleCancel} />
          </div>
        </div>
      </div>
    </div>
  );
}
