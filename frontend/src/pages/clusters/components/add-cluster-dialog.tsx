import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { FieldShell } from "@/components/branded";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ClusterSchema } from "../hooks/use-clusters";
import type { ClusterData } from "../hooks/use-clusters";
import { extractErrorMessage } from '@/lib/utils';

interface AddClusterDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAddCluster: (cluster: ClusterData) => void;
  isLoading?: boolean;
  error?: string | null;
}

export default function AddClusterDialog({
  open,
  onOpenChange,
  onAddCluster,
  isLoading = false,
  error = null,
}: AddClusterDialogProps) {
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

  const handleSubmit = () => {
    if (!validateForm()) return;

    // Pass the correct object for cluster_image_registry if enabled, or undefined if not
    const { cluster_image_registry, ...rest } = formData;
    onAddCluster({
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
          spec: { backend_storage_size: value },
        }
        : undefined,
    }));
  };

  // Reset form when dialog closes
  const handleOpenChange = (openState: boolean) => {
    if (!openState) {
      setFormData({
        name: "",
        cluster_url: "",
        cluster_ca_data: "",
        cluster_sa_token: "",
        cluster_image_registry: { name: "default-registry", spec: { backend_storage_size: "20Gi" } },
      });
      setErrors({});
      setShowCAData(false);
      setShowSAToken(false);
    }
    onOpenChange(openState);
  };

  const isFormValid = formData.name.trim() && formData.cluster_url.trim() &&
                     formData.cluster_ca_data.trim() && formData.cluster_sa_token.trim();

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-2xl mx-auto max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add New Cluster</DialogTitle>
          <DialogDescription>
            Connect a new Kubernetes cluster to your organization.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {error && (
            <div className="text-sm text-danger bg-danger-bg p-3 rounded-md">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <FieldShell
              label="Cluster Name"
              htmlFor="name"
              required
              error={errors.name}
            >
              <Input
                id="name"
                name="name"
                placeholder="My Production Cluster"
                value={formData.name}
                onChange={handleChange}
                className={errors.name ? "border-danger" : ""}
              />
            </FieldShell>

            <FieldShell
              label="Cluster URL"
              htmlFor="cluster_url"
              required
              error={errors.cluster_url}
            >
              <Input
                id="cluster_url"
                name="cluster_url"
                placeholder="https://k8s-api.example.com:6443"
                value={formData.cluster_url}
                onChange={handleChange}
                className={errors.cluster_url ? "border-danger" : ""}
              />
            </FieldShell>

            <FieldShell
              label="CA Certificate Data"
              htmlFor="cluster_ca_data"
              required
              hint="Base64-encoded."
              error={errors.cluster_ca_data}
            >
              <div className="relative">
                <Input
                  id="cluster_ca_data"
                  name="cluster_ca_data"
                  type={showCAData ? "text" : "password"}
                  placeholder="LS0tLS1CRUdJTi..."
                  value={formData.cluster_ca_data}
                  onChange={handleChange}
                  className={errors.cluster_ca_data ? "border-danger pr-10" : "pr-10"}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowCAData(!showCAData)}
                >
                  {showCAData ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
            </FieldShell>

            <FieldShell
              label="Service Account Token"
              htmlFor="cluster_sa_token"
              required
              error={errors.cluster_sa_token}
            >
              <div className="relative">
                <Input
                  id="cluster_sa_token"
                  name="cluster_sa_token"
                  type={showSAToken ? "text" : "password"}
                  placeholder="eyJhbGciOiJSUzI1NiIs..."
                  value={formData.cluster_sa_token}
                  onChange={handleChange}
                  className={errors.cluster_sa_token ? "border-danger pr-10" : "pr-10"}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowSAToken(!showSAToken)}
                >
                  {showSAToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
              </div>
            </FieldShell>

            <div className="space-y-3 pt-2">
              <div className="flex items-start gap-3">
                <Switch
                  id="enable-registry"
                  checked={!!formData.cluster_image_registry}
                  onCheckedChange={handleImageRegistryToggle}
                  className="mt-0.5"
                />
                <div className="space-y-1">
                  <Label htmlFor="enable-registry" className="text-[13px] font-medium text-foreground">
                    Enable Image Registry
                  </Label>
                  <p className="text-[12px] text-muted-foreground leading-relaxed">
                    Provisions a private registry for build artifacts.
                  </p>
                </div>
              </div>

              {formData.cluster_image_registry && (
                <div className="pl-11">
                  <FieldShell
                    label="Backend Storage Size"
                    htmlFor="registry-size"
                    hint="e.g. 20Gi, 100Gi"
                  >
                    <Input
                      id="registry-size"
                      placeholder="20Gi"
                      value={formData.cluster_image_registry.spec?.backend_storage_size || ""}
                      onChange={handleRegistrySizeChange}
                    />
                  </FieldShell>
                </div>
              )}
            </div>
          </div>
        </div>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" type="button">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleSubmit}
            disabled={!isFormValid || isLoading}
          >
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            Add Cluster
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
