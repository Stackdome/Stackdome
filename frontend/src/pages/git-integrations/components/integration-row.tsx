import { useCallback, useEffect, useRef, useState } from "react";
import { AppWindow, CircleAlert, CircleCheck, KeyRound, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StatusPill } from "@/components/branded";
import { cn } from "@/lib/utils";
import {
  listInstallations,
  type GitIntegration,
  type GitInstallation,
} from "@/api/git-integrations";
import { getCurrentOrganizationId } from "@/lib/common";
import { deriveRow, providerIdFor, GIT_INTEGRATION_TYPE_GITHUB_APP, PROVIDER_DISPLAY_NAMES, type RowViewModel } from "@/lib/git-integrations";
import { RowMenu } from "./row-menu";
import { ProviderLogo } from "@/components/branded/provider-logo";

function Banner({
  banner,
  statusKey,
  onVerify,
  onUpdateCredentials,
}: {
  banner: NonNullable<RowViewModel["banner"]>;
  statusKey: RowViewModel["statusKey"];
  onVerify: () => void;
  onUpdateCredentials?: () => void;
}) {
  const Icon = statusKey === "action_needed" ? CircleAlert : TriangleAlert;
  const toneClasses =
    statusKey === "action_needed"
      ? "border-danger-border bg-danger-bg text-danger"
      : "border-warn-border bg-warn-bg text-warn";

  return (
    <div className={cn("flex items-center gap-2 border-t px-4 py-2 text-xs", toneClasses)}>
      <Icon className="h-3.5 w-3.5 shrink-0" />
      <span className="flex-1 text-foreground/80">{banner.message}</span>
      {statusKey === "needs_setup" ? (
        banner.ctaHref ? (
          <a
            href={banner.ctaHref}
            target="_blank"
            rel="noreferrer"
            className="whitespace-nowrap rounded-md border border-brand-border px-2.5 py-1 font-medium text-brand hover:bg-brand-bg-hover"
          >
            {banner.ctaLabel}
          </a>
        ) : (
          <Button
            variant="outline"
            size="sm"
            className="h-auto whitespace-nowrap rounded-md border-brand-border px-2.5 py-1 text-brand"
            disabled
          >
            {banner.ctaLabel}
          </Button>
        )
      ) : statusKey === "action_needed" ? (
        // github_app rows can't be PUT-updated: message without a CTA.
        onUpdateCredentials && (
          <Button
            variant="outline"
            size="sm"
            className="h-auto whitespace-nowrap rounded-md border-brand-border px-2.5 py-1 text-brand hover:bg-brand-bg-hover"
            onClick={onUpdateCredentials}
          >
            {banner.ctaLabel}
          </Button>
        )
      ) : (
        <Button
          variant="outline"
          size="sm"
          className="h-auto whitespace-nowrap rounded-md border-brand-border px-2.5 py-1 text-brand hover:bg-brand-bg-hover"
          onClick={onVerify}
        >
          {banner.ctaLabel}
        </Button>
      )}
    </div>
  );
}

export function IntegrationRow({
  integration,
  onVerify,
  onRemove,
  onUpdateCredentials,
  onAddAccount,
}: {
  integration: GitIntegration;
  onVerify: (integration: GitIntegration) => void;
  onRemove: (integration: GitIntegration) => void;
  /** Opens the update-credentials dialog for this row (credentials-type only). */
  onUpdateCredentials?: (integration: GitIntegration) => void;
  /** Starts the install flow for another GitHub account (github_app only). */
  onAddAccount?: (integration: GitIntegration) => void;
}) {
  const [installations, setInstallations] = useState<GitInstallation[]>([]);
  const requestSeq = useRef(0);

  const isGithubApp = integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP;

  const load = useCallback(async () => {
    // Only GitHub App integrations have installations; a credentials row would
    // burn a GitHub refresh call for an always-empty list.
    if (!isGithubApp) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId || !integration.id) return;
    const seq = ++requestSeq.current;
    try {
      // refresh=true re-lists installations from GitHub, so state lost to a
      // missed webhook (backend downtime, local dev without a public URL)
      // self-heals on every page visit — no manual sync action needed.
      const list = await listInstallations(orgId, integration.id, true);
      if (seq === requestSeq.current) setInstallations(list.items ?? []);
    } catch {
      // Row keeps its last-known installations on failure; reload retries.
    }
  }, [integration.id, isGithubApp]);

  useEffect(() => {
    void load();
  }, [load]);

  const row = deriveRow(integration, installations);

  return (
    <div className={cn(row.tone === "attention" && "bg-warn-bg/50")}>
      <div className="flex items-center gap-4 px-4 py-3 hover:bg-muted/50">
        <div className="flex w-[180px] min-w-0 items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card">
            <ProviderLogo providerId={providerIdFor(integration)} className="h-5 w-5 shrink-0" />
          </div>
          <div className="min-w-0">
            <p className="truncate text-[15px] font-medium text-foreground">
              {PROVIDER_DISPLAY_NAMES[providerIdFor(integration)]}
            </p>
            <p className="truncate font-mono text-[11.5px] text-fg-muted">{row.host}</p>
          </div>
        </div>

        <div className="w-[130px]">
          <span className="inline-flex items-center gap-1.5 rounded-full border border-border px-2 py-0.5 text-xs text-muted-foreground">
            {isGithubApp ? <AppWindow className="h-3 w-3" /> : <KeyRound className="h-3 w-3" />}
            {row.authLabel}
          </span>
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-baseline justify-between gap-2">
            <span className="truncate text-[11.5px] text-muted-foreground">{row.access.label}</span>
            <span className="shrink-0 font-mono text-[11px] text-fg-muted">{row.access.hint}</span>
          </div>
        </div>

        <div className="w-[130px]">
          {row.statusKey === "connected" ? (
            <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <CircleCheck className="h-3.5 w-3.5 text-success" />
              Connected
            </span>
          ) : (
            <StatusPill variant={row.statusKey === "action_needed" ? "error" : "pending"}>
              {row.statusLabel}
            </StatusPill>
          )}
        </div>

        <RowMenu
          onVerify={isGithubApp ? undefined : () => onVerify(integration)}
          onUpdateCredentials={
            isGithubApp || !onUpdateCredentials ? undefined : () => onUpdateCredentials(integration)
          }
          onAddAccount={
            // BYO rows already have the direct manageUrl link; only platform
            // rows (no install_url) need the state-minting flow.
            isGithubApp && !integration.install_url && onAddAccount
              ? () => onAddAccount(integration)
              : undefined
          }
          manageUrl={isGithubApp ? integration.install_url : undefined}
          onRemove={() => onRemove(integration)}
        />
      </div>

      {row.banner && (
        <Banner
          banner={row.banner}
          statusKey={row.statusKey}
          onVerify={() => onVerify(integration)}
          onUpdateCredentials={
            isGithubApp || !onUpdateCredentials ? undefined : () => onUpdateCredentials(integration)
          }
        />
      )}
    </div>
  );
}
