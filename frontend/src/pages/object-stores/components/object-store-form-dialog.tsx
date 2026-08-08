import { useEffect, useMemo, useState } from "react";
import { Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useToast } from "@/components/ui/use-toast";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { createObjectStore, updateObjectStore } from "@/api/object-stores";
import { useResourceProjects } from "@/hooks/use-resource-projects";
import {
  objectStoreFormSchema,
  toApiPayload,
  type ObjectStoreFormValues,
} from "../schemas/form-schema";
import type { ObjectStore } from "../types";
import { SecretKeyPicker } from "./secret-key-picker";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: ObjectStore | null;
  onSaved: () => void;
};

const empty: ObjectStoreFormValues = {
  name: "",
  destinationPath: "",
  retentionPolicy: "7d",
  provider: "s3",
  s3: {
    region: "us-east-1",
    endpointUrl: "",
    accessKeyId: { secret_id: "", key: "" },
    secretAccessKey: { secret_id: "", key: "" },
  },
  azure: {
    storageAccountName: "",
    connectionString: { secret_id: "", key: "" },
  },
  gcs: {
    serviceAccountCredentials: { secret_id: "", key: "" },
  },
};

function fromObjectStore(store: ObjectStore): ObjectStoreFormValues | null {
  const cfg = store.spec.configuration;
  if (cfg.s3_credentials) {
    return {
      name: store.name,
      destinationPath: store.spec.destination_path,
      retentionPolicy: store.spec.retention_policy ?? "7d",
      provider: "s3",
      s3: {
        region: cfg.s3_credentials.region,
        endpointUrl: cfg.s3_credentials.endpoint_url ?? "",
        accessKeyId: cfg.s3_credentials.access_key_id,
        secretAccessKey: cfg.s3_credentials.secret_access_key,
      },
      azure: empty.azure,
      gcs: empty.gcs,
    };
  }
  if (cfg.azure_credentials) {
    return {
      name: store.name,
      destinationPath: store.spec.destination_path,
      retentionPolicy: store.spec.retention_policy ?? "7d",
      provider: "azure",
      s3: empty.s3,
      azure: {
        storageAccountName: cfg.azure_credentials.storage_account_name ?? "",
        connectionString: cfg.azure_credentials.connection_string,
      },
      gcs: empty.gcs,
    };
  }
  if (cfg.gcs_credentials) {
    return {
      name: store.name,
      destinationPath: store.spec.destination_path,
      retentionPolicy: store.spec.retention_policy ?? "7d",
      provider: "gcs",
      s3: empty.s3,
      azure: empty.azure,
      gcs: {
        serviceAccountCredentials: cfg.gcs_credentials.service_account_credentials,
      },
    };
  }
  return null;
}

export function ObjectStoreFormDialog({ open, onOpenChange, editing, onSaved }: Props) {
  const { toast } = useToast();
  const { projectNameById, defaultProjectName } = useResourceProjects();
  const [values, setValues] = useState<ObjectStoreFormValues>(empty);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [unsupportedEdit, setUnsupportedEdit] = useState(false);

  useEffect(() => {
    if (open) {
      setErrors({});
      if (editing) {
        const next = fromObjectStore(editing);
        if (next === null) {
          setUnsupportedEdit(true);
          setValues(empty);
        } else {
          setUnsupportedEdit(false);
          setValues(next);
        }
      } else {
        setUnsupportedEdit(false);
        setValues(empty);
      }
    }
  }, [open, editing]);

  const orgId = useMemo(() => getCurrentOrganizationId(), []);

  function clearError(key: string) {
    setErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  async function handleSubmit() {
    const valuesForParse: ObjectStoreFormValues = {
      ...values,
      s3: values.provider === "s3" ? values.s3 : undefined,
      azure: values.provider === "azure" ? values.azure : undefined,
      gcs: values.provider === "gcs" ? values.gcs : undefined,
    };
    const parsed = objectStoreFormSchema.safeParse(valuesForParse);
    if (!parsed.success) {
      const errs: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        errs[issue.path.join(".")] = issue.message;
      }
      setErrors(errs);
      return;
    }
    if (!orgId) {
      toast({
        title: "Could not save Object Store",
        description: "No organization selected.",
        variant: "destructive",
      });
      return;
    }

    setSubmitting(true);
    try {
      const payload = toApiPayload(parsed.data);
      if (editing?.id) {
        const projectName = projectNameById(editing.project_id);
        if (!projectName) {
          toast({ title: "Failed to save Object Store", description: "Could not resolve the project for this object store.", variant: "destructive" });
          return;
        }
        await updateObjectStore(orgId, projectName, editing.id, payload);
        toast({ title: "Object store updated", variant: "success" });
      } else {
        if (!defaultProjectName) {
          toast({ title: "Failed to save Object Store", description: "You don't have a project to create object stores in.", variant: "destructive" });
          return;
        }
        await createObjectStore(orgId, defaultProjectName, payload);
        toast({ title: "Object store created", variant: "success" });
      }
      onSaved();
      onOpenChange(false);
    } catch (e) {
      toast({
        title: "Failed to save Object Store",
        description: getErrorMessage(e),
        variant: "destructive",
      });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (submitting) return;
        onOpenChange(next);
      }}
    >
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit Object Store" : "New Object Store"}</DialogTitle>
          <DialogDescription>
            Configure a backup destination. Credentials reference an existing Secret.
          </DialogDescription>
        </DialogHeader>

        {unsupportedEdit ? (
          <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-body text-danger">
            Editing Azure/GCS stores is not yet supported in the UI. Please delete and re-create the store.
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="grid gap-2">
              <Label htmlFor="os-name">
                Name <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span>
              </Label>
              <Input
                id="os-name"
                value={values.name}
                onChange={(e) => {
                  clearError("name");
                  setValues((v) => ({ ...v, name: e.target.value }));
                }}
                placeholder="minio-local"
                disabled={!!editing}
              />
              {errors.name && <p className="text-meta text-danger">{errors.name}</p>}
            </div>

            <div className="grid gap-2">
              <Label htmlFor="os-destination">
                Destination path <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span>
              </Label>
              <Input
                id="os-destination"
                value={values.destinationPath}
                onChange={(e) => {
                  clearError("destinationPath");
                  setValues((v) => ({ ...v, destinationPath: e.target.value }));
                }}
                placeholder="s3://stackdome-backups/db1"
                className="font-mono"
              />
              {errors.destinationPath && (
                <p className="text-meta text-danger">{errors.destinationPath}</p>
              )}
            </div>

            <div className="grid gap-2">
              <Label htmlFor="os-retention">
                Retention <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span>
              </Label>
              <Input
                id="os-retention"
                value={values.retentionPolicy}
                onChange={(e) => {
                  clearError("retentionPolicy");
                  setValues((v) => ({ ...v, retentionPolicy: e.target.value }));
                }}
                placeholder="7d"
              />
              {errors.retentionPolicy && (
                <p className="text-meta text-danger">{errors.retentionPolicy}</p>
              )}
            </div>

            <Tabs
              value={values.provider}
              onValueChange={(p) =>
                setValues((v) => ({ ...v, provider: p as ObjectStoreFormValues["provider"] }))
              }
            >
              <TabsList>
                <TabsTrigger value="s3">S3 / S3-compatible</TabsTrigger>
                <TabsTrigger value="azure">Azure</TabsTrigger>
                <TabsTrigger value="gcs">GCS</TabsTrigger>
              </TabsList>

              <TabsContent value="s3" className="flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-3">
                  <div className="grid gap-2">
                    <Label htmlFor="os-region">
                      Region <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span>
                    </Label>
                    <Input
                      id="os-region"
                      value={values.s3?.region ?? ""}
                      onChange={(e) => {
                        clearError("s3.region");
                        setValues((v) => ({
                          ...v,
                          s3: { ...(v.s3 ?? empty.s3!), region: e.target.value },
                        }));
                      }}
                      placeholder="us-east-1"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="os-endpoint">Endpoint URL (optional)</Label>
                    <Input
                      id="os-endpoint"
                      value={values.s3?.endpointUrl ?? ""}
                      onChange={(e) => {
                        clearError("s3.endpointUrl");
                        setValues((v) => ({
                          ...v,
                          s3: { ...(v.s3 ?? empty.s3!), endpointUrl: e.target.value },
                        }));
                      }}
                      placeholder="http://localhost:9000 (for MinIO)"
                      className="font-mono"
                    />
                  </div>
                </div>

                <SecretKeyPicker
                  label={<>Access Key ID <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span></>}
                  helpText="A Generic secret and the key inside it that holds the access key id."
                  value={values.s3?.accessKeyId ?? { secret_id: "", key: "" }}
                  onChange={(next) => {
                    clearError("s3.accessKeyId.secret_id");
                    clearError("s3.accessKeyId.key");
                    setValues((v) => ({
                      ...v,
                      s3: { ...(v.s3 ?? empty.s3!), accessKeyId: next },
                    }));
                  }}
                  expectedKeyHint="accessKeyId"
                  error={errors["s3.accessKeyId.secret_id"] || errors["s3.accessKeyId.key"]}
                />

                <SecretKeyPicker
                  label={<>Secret Access Key <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span></>}
                  helpText="A Generic secret and the key inside it that holds the secret access key."
                  value={values.s3?.secretAccessKey ?? { secret_id: "", key: "" }}
                  onChange={(next) => {
                    clearError("s3.secretAccessKey.secret_id");
                    clearError("s3.secretAccessKey.key");
                    setValues((v) => ({
                      ...v,
                      s3: { ...(v.s3 ?? empty.s3!), secretAccessKey: next },
                    }));
                  }}
                  expectedKeyHint="secretAccessKey"
                  error={
                    errors["s3.secretAccessKey.secret_id"] || errors["s3.secretAccessKey.key"]
                  }
                />
              </TabsContent>

              <TabsContent value="azure" className="flex flex-col gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="os-azure-account">
                    Storage account name (optional)
                  </Label>
                  <Input
                    id="os-azure-account"
                    value={values.azure?.storageAccountName ?? ""}
                    onChange={(e) => {
                      clearError("azure.storageAccountName");
                      setValues((v) => ({
                        ...v,
                        azure: {
                          ...(v.azure ?? empty.azure!),
                          storageAccountName: e.target.value,
                        },
                      }));
                    }}
                    placeholder="mystorageaccount"
                    className="font-mono"
                  />
                </div>

                <SecretKeyPicker
                  label={<>Connection string <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span></>}
                  helpText="A Generic secret and the key inside it that holds the Azure connection string."
                  value={values.azure?.connectionString ?? { secret_id: "", key: "" }}
                  onChange={(next) => {
                    clearError("azure.connectionString.secret_id");
                    clearError("azure.connectionString.key");
                    setValues((v) => ({
                      ...v,
                      azure: { ...(v.azure ?? empty.azure!), connectionString: next },
                    }));
                  }}
                  expectedKeyHint="connectionString"
                  error={
                    errors["azure.connectionString.secret_id"] ||
                    errors["azure.connectionString.key"]
                  }
                />
              </TabsContent>

              <TabsContent value="gcs" className="flex flex-col gap-4">
                <SecretKeyPicker
                  label={<>Service account credentials <span className="text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span></>}
                  helpText="A Generic secret and the key inside it that holds the GCS service account JSON."
                  value={
                    values.gcs?.serviceAccountCredentials ?? { secret_id: "", key: "" }
                  }
                  onChange={(next) => {
                    clearError("gcs.serviceAccountCredentials.secret_id");
                    clearError("gcs.serviceAccountCredentials.key");
                    setValues((v) => ({
                      ...v,
                      gcs: {
                        ...(v.gcs ?? empty.gcs!),
                        serviceAccountCredentials: next,
                      },
                    }));
                  }}
                  expectedKeyHint="serviceAccountKey"
                  error={
                    errors["gcs.serviceAccountCredentials.secret_id"] ||
                    errors["gcs.serviceAccountCredentials.key"]
                  }
                />
              </TabsContent>
            </Tabs>
          </div>
        )}

        <DialogFooter>
          <Button shape="flat"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || unsupportedEdit}>
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {editing ? "Save changes" : "Create Object Store"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
