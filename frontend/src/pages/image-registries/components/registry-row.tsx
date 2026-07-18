import type { RegistryCredential } from "@/api/registry-credentials";
import { providerIdForHost, PURPOSE_LABELS, PURPOSE_BOTH, REGISTRY_PROVIDERS } from "../lib/providers";
import { ProviderLogo } from "./provider-logo";
import { RowMenu } from "./row-menu";

export function RegistryRow({
  credential,
  onVerify,
  onUpdateCredentials,
  onRemove,
}: {
  credential: RegistryCredential;
  onVerify: (credential: RegistryCredential) => void;
  onUpdateCredentials: (credential: RegistryCredential) => void;
  onRemove: (credential: RegistryCredential) => void;
}) {
  const providerId = providerIdForHost(credential.host);
  const providerLabel = REGISTRY_PROVIDERS.find((p) => p.id === providerId)?.label ?? "Registry";

  return (
    <div className="flex items-center gap-4 px-4 py-3 hover:bg-muted/50">
      <div className="flex w-[220px] min-w-0 items-center gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card">
          <ProviderLogo providerId={providerId} className="h-6 w-6" />
        </div>
        <div className="min-w-0">
          <p className="truncate text-[15px] font-medium text-foreground">{providerLabel}</p>
          <p className="truncate font-mono text-[11.5px] text-fg-muted">{credential.host}</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <span className="inline-flex rounded-full border border-border bg-card px-2 py-1 text-xs text-fg-muted">
          {PURPOSE_LABELS[credential.purpose ?? PURPOSE_BOTH]}
        </span>
      </div>

      <RowMenu
        onVerify={() => onVerify(credential)}
        onUpdateCredentials={() => onUpdateCredentials(credential)}
        onRemove={() => onRemove(credential)}
      />
    </div>
  );
}
