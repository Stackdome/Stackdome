import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Info, Eye, EyeOff } from "lucide-react";
import { ClusterSchema } from "../../hooks/use-clusters";
import type { ClusterData } from "../../hooks/use-clusters";
import { Switch } from "@/components/ui/switch";
import { extractErrorMessage } from '@/lib/utils';

interface Props {
  onSubmit: (values: ClusterData) => void;
  loading: boolean;
  error: string | null;
}

export default function ClusterCreateForm({ onSubmit, loading, error }: Props) {
  // Use a local type for form state to allow the registry object
  type ClusterFormState = Omit<ClusterData, 'cluster_image_registry'> & {
    cluster_image_registry?: {
      name: string;
      spec?: {
        backend_storage_size?: string;
      };
    };
  };
  const [formData, setFormData] = useState<ClusterFormState>({
    name: "",
    cluster_url: "",
    cluster_ca_data: "",
    cluster_sa_token: "",
    cluster_image_registry: { name: "default-registry", spec: { backend_storage_size: "20Gi" } },
  });
  const [showCAData, setShowCAData] = useState(false);
  const [showSAToken, setShowSAToken] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof ClusterData, string>>>({});

  const validateForm = (): boolean => {
    const result = ClusterSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<Record<keyof ClusterData, string>> = {};
      result.error.errors.forEach(err => {
        const field = err.path[0] as keyof ClusterData;
        fieldErrors[field] = extractErrorMessage(err, err.message);
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
    if (errors[name as keyof ClusterData]) {
      setErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    // Pass the correct object for cluster_image_registry if enabled, or undefined if not
    const { cluster_image_registry, ...rest } = formData;
    onSubmit({
      ...rest,
      ...(cluster_image_registry ? { cluster_image_registry: {
        name: cluster_image_registry.name,
        spec: { backend_storage_size: cluster_image_registry.spec?.backend_storage_size || "20Gi" }
      }} : {}),
    } as ClusterData);
  };

  // Toggle handler for image registry
  const handleImageRegistryToggle = (checked: boolean) => {
    setFormData(prev =>
      checked
        ? {
          ...prev,
          cluster_image_registry: {
            name: "default-registry",
            spec: { backend_storage_size: "20Gi" },
          },
        }
        : {
          ...prev,
          cluster_image_registry: undefined,
        }
    );
  };

  const handleRegistrySizeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setFormData(prev => ({
      ...prev,
      cluster_image_registry: prev.cluster_image_registry
        ? {
          ...prev.cluster_image_registry,
          spec: { ...prev.cluster_image_registry.spec, backend_storage_size: value },
        }
        : undefined,
    }));
    if (errors.cluster_image_registry) {
      setErrors(prev => ({ ...prev, cluster_image_registry: undefined }));
    }
  };

  return (
    <TooltipProvider>
      <form onSubmit={handleSubmit}>
        <Card className="w-full rounded-lg">
          <CardHeader>
            <CardTitle className="text-xl">Cluster Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div>
              <div className="flex items-center space-x-2">
                <Label htmlFor="name">Cluster name <span className="text-destructive">*</span></Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="h-4 w-4 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p>A descriptive name for your Kubernetes cluster</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                placeholder="Production Cluster"
                className={`mt-1 ${errors.name ? "border-danger" : ""}`}
              />
              {errors.name && <p className="text-danger text-sm mt-1">{errors.name}</p>}
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <Label htmlFor="cluster_url">API server URL <span className="text-destructive">*</span></Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="h-4 w-4 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p>The URL of your Kubernetes API server</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input
                id="cluster_url"
                name="cluster_url"
                value={formData.cluster_url}
                onChange={handleChange}
                placeholder="https://kubernetes.example.com:6443"
                className={`mt-1 ${errors.cluster_url ? "border-danger" : ""}`}
              />
              {errors.cluster_url && <p className="text-danger text-sm mt-1">{errors.cluster_url}</p>}
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <Label htmlFor="cluster_ca_data">CA certificate <span className="text-destructive">*</span></Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="h-4 w-4 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p>The base64-encoded certificate authority data</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="relative">
                <Input
                  id="cluster_ca_data"
                  name="cluster_ca_data"
                  value={formData.cluster_ca_data}
                  onChange={handleChange}
                  type={showCAData ? "text" : "password"}
                  placeholder="Base64-encoded CA certificate"
                  className={`mt-1 ${errors.cluster_ca_data ? "border-danger" : ""} pr-10`}
                />
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 flex items-center px-3 mt-1 text-muted-foreground hover:text-foreground focus:outline-none"
                  onClick={() => setShowCAData(!showCAData)}
                >
                  {showCAData ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
              {errors.cluster_ca_data && <p className="text-danger text-sm mt-1">{errors.cluster_ca_data}</p>}
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <Label htmlFor="cluster_sa_token">Service account token <span className="text-destructive">*</span></Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="h-4 w-4 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p>The service account token for authenticating with your cluster</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="relative">
                <Input
                  id="cluster_sa_token"
                  name="cluster_sa_token"
                  value={formData.cluster_sa_token}
                  onChange={handleChange}
                  type={showSAToken ? "text" : "password"}
                  placeholder="Service account token"
                  className={`mt-1 ${errors.cluster_sa_token ? "border-danger" : ""} pr-10`}
                />
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 flex items-center px-3 mt-1 text-muted-foreground hover:text-foreground focus:outline-none"
                  onClick={() => setShowSAToken(!showSAToken)}
                >
                  {showSAToken ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
              {errors.cluster_sa_token && <p className="text-danger text-sm mt-1">{errors.cluster_sa_token}</p>}
            </div>
            <div>
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <Label htmlFor="create-image-registry">Create image registry</Label>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-4 w-4 text-muted-foreground cursor-pointer" />
                    </TooltipTrigger>
                    <TooltipContent side="right">
                      <p>Creates an internal image registry for your builds</p>
                    </TooltipContent>
                  </Tooltip>
                </div>
                <Switch
                  id="create-image-registry"
                  checked={!!formData.cluster_image_registry}
                  onCheckedChange={handleImageRegistryToggle}
                />
              </div>
            </div>
            {formData.cluster_image_registry && (
              <div>
                <Label htmlFor="registry-size">Registry size <span className="text-destructive">*</span></Label>
                <Input
                  id="registry-size"
                  name="registry-size"
                  value={formData.cluster_image_registry.spec?.backend_storage_size || ""}
                  onChange={handleRegistrySizeChange}
                  placeholder="e.g. 10Gi"
                  className={`mt-1 ${errors["cluster_image_registry"] ? "border-danger" : ""}`}
                />
                {errors["cluster_image_registry"] && (
                  <p className="text-danger text-sm mt-1">{errors["cluster_image_registry"]}</p>
                )}
              </div>
            )}
            {error && (
              <div className="bg-danger-bg text-danger p-3 rounded text-sm">
                {error}
              </div>
            )}
          </CardContent>
          <CardFooter className="flex justify-end">
            <Button type="submit" disabled={loading} className="px-8">
              {loading ? "Creating..." : "Create Cluster"}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </TooltipProvider>
  );
}
