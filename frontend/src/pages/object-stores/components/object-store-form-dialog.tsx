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
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { createObjectStore, updateObjectStore } from "@/api/object-stores";
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
};

function fromObjectStore(store: ObjectStore): ObjectStoreFormValues {
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
    };
  }
  // Phase 1 ships S3 only in the UI. Editing an Azure/GCS store falls back to the
  // S3 form shell for now — a known limitation. A follow-up phase adds full
  // Azure/GCS edit support.
  return { ...empty, name: store.name, destinationPath: store.spec.destination_path };
}

export function ObjectStoreFormDialog({ open, onOpenChange, editing, onSaved }: Props) {
  const { toast } = useToast();
  const [values, setValues] = useState<ObjectStoreFormValues>(empty);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      setErrors({});
      setValues(editing ? fromObjectStore(editing) : empty);
    }
  }, [open, editing]);

  const orgId = useMemo(() => getCurrentOrganizationId(), []);

  async function handleSubmit() {
    const parsed = objectStoreFormSchema.safeParse(values);
    if (!parsed.success) {
      const errs: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        errs[issue.path.join(".")] = issue.message;
      }
      setErrors(errs);
      return;
    }
    if (!orgId) return;

    setSubmitting(true);
    try {
      const payload = toApiPayload(parsed.data);
      if (editing?.id) {
        await updateObjectStore(orgId, editing.id, payload);
        toast({ title: "Object Store updated" });
      } else {
        await createObjectStore(orgId, payload);
        toast({ title: "Object Store created" });
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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit Object Store" : "New Object Store"}</DialogTitle>
          <DialogDescription>
            Configure a backup destination. Credentials reference an existing Secret.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="grid gap-2">
            <Label htmlFor="os-name">Name</Label>
            <Input
              id="os-name"
              value={values.name}
              onChange={(e) => setValues((v) => ({ ...v, name: e.target.value }))}
              placeholder="minio-local"
              disabled={!!editing}
            />
            {errors.name && <p className="text-xs text-danger">{errors.name}</p>}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="os-destination">Destination path</Label>
            <Input
              id="os-destination"
              value={values.destinationPath}
              onChange={(e) => setValues((v) => ({ ...v, destinationPath: e.target.value }))}
              placeholder="s3://stackdome-backups/db1"
              className="font-mono"
            />
            {errors.destinationPath && (
              <p className="text-xs text-danger">{errors.destinationPath}</p>
            )}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="os-retention">Retention</Label>
            <Input
              id="os-retention"
              value={values.retentionPolicy}
              onChange={(e) => setValues((v) => ({ ...v, retentionPolicy: e.target.value }))}
              placeholder="7d"
            />
            {errors.retentionPolicy && (
              <p className="text-xs text-danger">{errors.retentionPolicy}</p>
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
              <TabsTrigger value="azure" disabled>
                Azure (later)
              </TabsTrigger>
              <TabsTrigger value="gcs" disabled>
                GCS (later)
              </TabsTrigger>
            </TabsList>

            <TabsContent value="s3" className="flex flex-col gap-4">
              <div className="grid grid-cols-2 gap-3">
                <div className="grid gap-2">
                  <Label htmlFor="os-region">Region</Label>
                  <Input
                    id="os-region"
                    value={values.s3?.region ?? ""}
                    onChange={(e) =>
                      setValues((v) => ({ ...v, s3: { ...v.s3!, region: e.target.value } }))
                    }
                    placeholder="us-east-1"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="os-endpoint">Endpoint URL (optional)</Label>
                  <Input
                    id="os-endpoint"
                    value={values.s3?.endpointUrl ?? ""}
                    onChange={(e) =>
                      setValues((v) => ({
                        ...v,
                        s3: { ...v.s3!, endpointUrl: e.target.value },
                      }))
                    }
                    placeholder="http://localhost:9000 (for MinIO)"
                    className="font-mono"
                  />
                </div>
              </div>

              <SecretKeyPicker
                label="Access Key ID"
                helpText="A Generic secret and the key inside it that holds the access key id."
                value={values.s3?.accessKeyId ?? { secret_id: "", key: "" }}
                onChange={(next) =>
                  setValues((v) => ({ ...v, s3: { ...v.s3!, accessKeyId: next } }))
                }
                expectedKeyHint="accessKeyId"
                error={errors["s3.accessKeyId.secret_id"] || errors["s3.accessKeyId.key"]}
              />

              <SecretKeyPicker
                label="Secret Access Key"
                helpText="A Generic secret and the key inside it that holds the secret access key."
                value={values.s3?.secretAccessKey ?? { secret_id: "", key: "" }}
                onChange={(next) =>
                  setValues((v) => ({ ...v, s3: { ...v.s3!, secretAccessKey: next } }))
                }
                expectedKeyHint="secretAccessKey"
                error={
                  errors["s3.secretAccessKey.secret_id"] || errors["s3.secretAccessKey.key"]
                }
              />
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={submitting}>
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {editing ? "Save changes" : "Create Object Store"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
