import { useEffect, useState } from "react";
import { isAxiosError } from "axios";
import { Container } from "lucide-react";
import {
  Dialog, DialogContent, DialogDescription, DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { cn } from "@/lib/utils";
import { createRegistryCredential, type RegistryCredentialPurpose } from "@/api/registry-credentials";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { createRegistrySchema } from "../lib/form-schemas";
import {
  REGISTRY_PROVIDERS, PURPOSE_BOTH, PURPOSE_PULL, PURPOSE_PUSH, PURPOSE_LABELS,
  type RegistryProvider,
} from "../lib/providers";
import { ProviderLogo } from "./provider-logo";

interface AddRegistryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

interface FieldErrors {
  host?: string;
  username?: string;
  password?: string;
}

export function AddRegistryDialog({ open, onOpenChange, onCreated }: AddRegistryDialogProps) {
  const { toast } = useToast();
  const [provider, setProvider] = useState<RegistryProvider | null>(null);
  const [host, setHost] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [purpose, setPurpose] = useState<RegistryCredentialPurpose>(PURPOSE_BOTH);
  const [saving, setSaving] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  useEffect(() => {
    if (!open) return;
    setProvider(null);
    setHost("");
    setUsername("");
    setPassword("");
    setPurpose(PURPOSE_BOTH);
    setFieldErrors({});
  }, [open]);

  const pickProvider = (p: RegistryProvider) => {
    setProvider(p);
    setHost(p.hostPrefill);
    setFieldErrors({});
  };

  const submit = async () => {
    const parsed = createRegistrySchema.safeParse({ host, username, password });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({ host: flat.host?.[0], username: flat.username?.[0], password: flat.password?.[0] });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      toast({ title: "Couldn't add registry", description: "No organization selected.", variant: "destructive" });
      return;
    }
    setSaving(true);
    try {
      await createRegistryCredential(orgId, { ...parsed.data, purpose });
      toast({ title: "Registry added" });
      onOpenChange(false);
      onCreated();
    } catch (e) {
      // 409 = host+purpose duplicate; anchor it to the host field instead of a toast.
      if (isAxiosError(e) && e.response?.status === 409) {
        setFieldErrors({ host: "Credentials for this registry and purpose already exist." });
      } else {
        toast({ title: "Couldn't add registry", description: getErrorMessage(e), variant: "destructive" });
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[540px]">
        <DialogTitle className="sr-only">Add image registry</DialogTitle>
        <DialogDescription className="sr-only">
          Store pull/push credentials so builds can use private registries.
        </DialogDescription>
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-brand">
            <Container className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Add image registry
          </span>
        </div>

        {/* Tile step centers inside the min height; the credentials step grows
            with its fields up to the viewport cap instead of clipping. */}
        <div className="flex max-h-[80vh] min-h-[440px] flex-col overflow-hidden">
          {provider == null ? (
            <div className="flex flex-1 overflow-y-auto p-8">
              {/* my-auto centers when the content is shorter than the shell but
                  top-aligns (instead of clipping) once it overflows. */}
              <div className="my-auto w-full">
                <div className="mb-7 text-center">
                  <h2 className="mb-2 text-2xl font-medium tracking-tight">
                    Where do your images live?
                  </h2>
                  <p className="text-sm text-muted-foreground">Pick a registry.</p>
                </div>
                <div className="grid grid-cols-2 gap-2.5">
                  {REGISTRY_PROVIDERS.map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => pickProvider(p)}
                      className={cn(
                        "flex min-h-[76px] items-start gap-3 rounded-md border bg-card p-4 text-left transition-colors",
                        "hover:border-brand focus-visible:border-brand focus-visible:outline-none",
                      )}
                    >
                      <span className="flex h-9 w-9 flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                        <ProviderLogo providerId={p.id} className="h-[18px] w-[18px]" />
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="mb-0.5 block text-sm font-medium text-foreground">{p.label}</span>
                        <span className="block text-xs text-muted-foreground">
                          {p.hostPrefill || "Custom host"}
                        </span>
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <>
              <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-8 [&>*]:shrink-0">
                <div className="mb-1 text-center">
                  <h2 className="mb-2 text-2xl font-medium tracking-tight">
                    Connect {provider.id === "other" ? "your registry" : provider.label}
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    Stackdome stores the credentials encrypted and uses them only for image pulls and pushes.
                  </p>
                </div>
                <FieldShell label="Host" htmlFor="registry-host" required error={fieldErrors.host}>
                  <Input
                    id="registry-host"
                    placeholder={provider.hostPlaceholder}
                    value={host}
                    onChange={(e) => {
                      setHost(e.target.value);
                      setFieldErrors((prev) => ({ ...prev, host: undefined }));
                    }}
                    aria-invalid={!!fieldErrors.host}
                  />
                </FieldShell>
                <FieldShell label="Username" htmlFor="registry-username" required error={fieldErrors.username}>
                  <Input
                    id="registry-username"
                    autoComplete="off"
                    value={username}
                    onChange={(e) => {
                      setUsername(e.target.value);
                      setFieldErrors((prev) => ({ ...prev, username: undefined }));
                    }}
                    aria-invalid={!!fieldErrors.username}
                  />
                </FieldShell>
                <FieldShell
                  label="Password"
                  htmlFor="registry-password"
                  required
                  hint={provider.hint}
                  error={fieldErrors.password}
                >
                  <Input
                    id="registry-password"
                    type="password"
                    autoComplete="new-password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      setFieldErrors((prev) => ({ ...prev, password: undefined }));
                    }}
                    aria-invalid={!!fieldErrors.password}
                  />
                </FieldShell>
                <FieldShell label="Purpose" htmlFor="registry-purpose">
                  <Select value={purpose} onValueChange={(v) => setPurpose(v as RegistryCredentialPurpose)}>
                    <SelectTrigger id="registry-purpose" className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={PURPOSE_BOTH}>{PURPOSE_LABELS[PURPOSE_BOTH]}</SelectItem>
                      <SelectItem value={PURPOSE_PULL}>{PURPOSE_LABELS[PURPOSE_PULL]}</SelectItem>
                      <SelectItem value={PURPOSE_PUSH}>{PURPOSE_LABELS[PURPOSE_PUSH]}</SelectItem>
                    </SelectContent>
                  </Select>
                </FieldShell>
              </div>
              <WizardFooter
                onBack={() => setProvider(null)}
                onContinue={() => void submit()}
                continueLabel="Add registry"
                loading={saving}
              />
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
