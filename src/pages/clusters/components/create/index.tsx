import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { ClusterSchema } from "../../hooks/use-clusters";
import type { ClusterFormInput } from "../../hooks/use-clusters";

interface Props {
  onSubmit: (values: ClusterFormInput) => void;
  loading: boolean;
  error: string | null;
}

export default function ClusterCreatePage({ onSubmit, loading, error }: Props) {
  const [formData, setFormData] = useState<ClusterFormInput>({
    name: "",
    cluster_url: "",
    cluster_ca_data: "",
    cluster_sa_token: "",
  });
  const [errors, setErrors] = useState<Partial<Record<keyof ClusterFormInput, string>>>({});

  const validateForm = (): boolean => {
    const result = ClusterSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<Record<keyof ClusterFormInput, string>> = {};
      result.error.errors.forEach(err => {
        const field = err.path[0] as keyof ClusterFormInput;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return false;
    }
    setErrors({});
    return true;
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    if (errors[name as keyof ClusterFormInput]) {
      setErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <Label htmlFor="name">Cluster name</Label>
        <p className="text-gray-500 text-sm mt-1 mb-2">A descriptive name for your Kubernetes cluster</p>
        <Input 
          id="name" 
          name="name" 
          value={formData.name} 
          onChange={handleChange} 
          placeholder="e.g., Production Cluster"
          className={`${errors.name ? "border-red-500" : ""}`}
        />
        {errors.name && <p className="text-red-500 text-sm mt-1">{errors.name}</p>}
      </div>
      <div>
        <Label htmlFor="cluster_url">API server URL</Label>
        <p className="text-gray-500 text-sm mt-1 mb-2">The URL of your Kubernetes API server</p>
        <Input 
          id="cluster_url" 
          name="cluster_url" 
          value={formData.cluster_url} 
          onChange={handleChange} 
          placeholder="https://kubernetes.example.com:6443"
          className={`${errors.cluster_url ? "border-red-500" : ""}`}
        />
        {errors.cluster_url && <p className="text-red-500 text-sm mt-1">{errors.cluster_url}</p>}
      </div>
      <div>
        <Label htmlFor="cluster_ca_data">CA certificate</Label>
        <p className="text-gray-500 text-sm mt-1 mb-2">The base64-encoded certificate authority data</p>
        <Input 
          id="cluster_ca_data" 
          name="cluster_ca_data" 
          value={formData.cluster_ca_data} 
          onChange={handleChange} 
          placeholder="Base64-encoded CA certificate"
          className={`${errors.cluster_ca_data ? "border-red-500" : ""}`}
        />
        {errors.cluster_ca_data && <p className="text-red-500 text-sm mt-1">{errors.cluster_ca_data}</p>}
      </div>
      <div>
        <Label htmlFor="cluster_sa_token">Service account token</Label>
        <p className="text-gray-500 text-sm mt-1 mb-2">The service account token for authenticating with your cluster</p>
        <Input 
          id="cluster_sa_token" 
          name="cluster_sa_token" 
          value={formData.cluster_sa_token} 
          onChange={handleChange}
          type="password"
          placeholder="Service account token"
          className={`${errors.cluster_sa_token ? "border-red-500" : ""}`}
        />
        {errors.cluster_sa_token && <p className="text-red-500 text-sm mt-1">{errors.cluster_sa_token}</p>}
      </div>
      {error && (
        <div className="bg-red-50 text-red-500 p-3 rounded text-sm">
          {error}
        </div>
      )}
      <Button type="submit" disabled={loading} className="mt-4 w-full">
        {loading ? "Creating..." : "Create Cluster"}
      </Button>
    </form>
  );
}
