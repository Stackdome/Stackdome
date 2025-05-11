import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { X } from "lucide-react";
import ClusterCreatePage from ".";
import { createCluster } from "@/api/clusters";
import type { ClusterFormInput } from "../../hooks/use-clusters";
import { getCurrentOrganizationId } from "@/helpers/common";

export default function ClusterCreateWrapperPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  async function handleSubmit(values: ClusterFormInput) {
    setLoading(true);
    setError(null);
    try {
      await createCluster(getCurrentOrganizationId(), values);
      setLoading(false);
      navigate("/clusters");
    } catch (e) {
      console.error("Failed to create cluster:", e);
      if (e instanceof Error) {
        setError(e.message);
      } else if (typeof e === 'object' && e !== null && 'message' in e) {
        setError(String(e.message));
      } else {
        setError("Failed to create cluster. Please check your connection and try again.");
      }
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-white z-40">
      <div className="border-b border-gray-200 p-4 flex items-center">
        <div className="flex items-center">
          <button 
            className="mr-4 text-gray-500"
            onClick={() => navigate('/clusters')}
          >
            <X size={20} />
          </button>
          <h1 className="text-lg font-medium">Create New Cluster</h1>
        </div>
      </div>
      
      <div className="flex h-[calc(100vh-64px)]">
        {/* Main Content - Full Width */}
        <div className="flex-grow p-6 overflow-y-auto">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-xl font-medium mb-6">Basic Information</h2>
            <ClusterCreatePage onSubmit={handleSubmit} loading={loading} error={error} />
          </div>
        </div>
      </div>
    </div>
  );
}
