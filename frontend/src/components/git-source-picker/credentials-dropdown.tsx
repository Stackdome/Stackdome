import { ChevronDown, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ProviderLogo } from "@/components/branded/provider-logo";
import {
  providerIdFor,
  PROVIDER_DISPLAY_NAMES,
  GIT_INTEGRATION_TYPE_GITHUB_APP,
  GIT_INTEGRATION_TYPE_CREDENTIALS,
  STATUS_INSTALLED,
  STATUS_ACTIVE,
} from "@/lib/git-integrations";
import type { GitIntegration } from "@/api/git-integrations";
import { cn } from "@/lib/utils";

/** Integrations the picker can actually use for cloning/search: installed
    GitHub Apps and active credential hosts with stored credentials. */
export function usableIntegrations(items: GitIntegration[]): GitIntegration[] {
  return items.filter((i) => {
    if (i.credentials_configured === false) return false;
    if (i.type === GIT_INTEGRATION_TYPE_GITHUB_APP) {
      return i.status === STATUS_INSTALLED || i.status === STATUS_ACTIVE;
    }
    return i.type === GIT_INTEGRATION_TYPE_CREDENTIALS && i.status === STATUS_ACTIVE;
  });
}

interface CredentialsDropdownProps {
  integrations: GitIntegration[];
  selectedId: string | null;
  onSelect: (integration: GitIntegration) => void;
  /** Opens the add-integration wizard. */
  onConnectNew: () => void;
}

export function CredentialsDropdown({
  integrations,
  selectedId,
  onSelect,
  onConnectNew,
}: CredentialsDropdownProps) {
  const selected = integrations.find((i) => i.id === selectedId) ?? null;
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button shape="flat" variant="outline" className="h-9 gap-2" aria-label="Credentials">
          {selected ? (
            <>
              <ProviderLogo providerId={providerIdFor(selected)} className="h-4 w-4" />
              <span className="font-mono text-meta">{selected.host}</span>
            </>
          ) : (
            <span className="text-meta text-muted-foreground">Credentials</span>
          )}
          <ChevronDown className="h-3 w-3 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[260px]">
        {integrations.map((integration) => (
          <DropdownMenuItem
            key={integration.id}
            onSelect={() => onSelect(integration)}
            className={cn(integration.id === selectedId && "bg-foreground/5 font-medium text-foreground")}
          >
            <ProviderLogo providerId={providerIdFor(integration)} className="h-4 w-4" />
            <span className="flex-1 text-body">{PROVIDER_DISPLAY_NAMES[providerIdFor(integration)]}</span>
            <span className="font-mono text-label text-fg-muted">{integration.host}</span>
          </DropdownMenuItem>
        ))}
        {integrations.length > 0 && <DropdownMenuSeparator />}
        {/* Deferred: opening a dialog synchronously from onSelect races the
            menu close and can wedge body pointer-events (radix-ui/primitives#1836). */}
        <DropdownMenuItem onSelect={() => setTimeout(() => onConnectNew(), 0)}>
          <Plus className="h-4 w-4" />
          Connect provider…
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
