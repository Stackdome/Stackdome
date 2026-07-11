import { useCallback, useEffect, useRef, useState } from "react";
import { AppWindow, CircleAlert, CircleCheck, KeyRound, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  listInstallations,
  type GitIntegration,
  type GitInstallation,
} from "@/api/git-integrations";
import { getCurrentOrganizationId } from "@/helpers/common";
import { deriveRow, providerIdFor, GIT_INTEGRATION_TYPE_GITHUB_APP, type ProviderId, type RowViewModel, type RowMeter } from "../lib/derive-row";
import { RowMenu } from "./row-menu";
import { ProviderLogo } from "./provider-logo";

function statusPillClasses(statusKey: RowViewModel["statusKey"]) {
  if (statusKey === "action_needed") return "border-danger-border bg-danger-bg text-danger";
  return "border-warn-border bg-warn-bg text-warn";
}

function meterFillClasses(fill: RowMeter["fill"]) {
  if (fill === "full") return "w-full bg-brand";
  if (fill === "partial") return "w-1/3 bg-brand";
  return "w-0 bg-brand";
}

const PROVIDER_TITLES: Record<ProviderId, string> = {
  github: "GitHub",
  gitlab: "GitLab",
  bitbucket: "Bitbucket",
  gitea: "Gitea",
  other: "Git host",
};

function Banner({
  banner,
  statusKey,
  onVerify,
}: {
  banner: NonNullable<RowViewModel["banner"]>;
  statusKey: RowViewModel["statusKey"];
  onVerify: () => void;
}) {
  const Icon = statusKey === "action_needed" ? CircleAlert : TriangleAlert;
  const toneClasses =
    statusKey === "action_needed"
      ? "border-danger-border bg-danger/[0.05] text-danger"
      : "border-warn-border bg-warn/[0.05] text-warn";

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
  onChanged,
}: {
  integration: GitIntegration;
  onVerify: (integration: GitIntegration) => void;
  onRemove: (integration: GitIntegration) => void;
  onChanged: () => void;
}) {
  const [installations, setInstallations] = useState<GitInstallation[]>([]);
  const requestSeq = useRef(0);

  const load = useCallback(
    async (refresh: boolean) => {
      const orgId = getCurrentOrganizationId();
      if (!orgId || !integration.id) return;
      const seq = ++requestSeq.current;
      try {
        const list = await listInstallations(orgId, integration.id, refresh);
        if (seq === requestSeq.current) setInstallations(list.items ?? []);
      } catch {
        // Row keeps its last-known installations on failure; the menu action can be retried.
      }
    },
    [integration.id],
  );

  useEffect(() => {
    void load(false);
  }, [load]);

  const row = deriveRow(integration, installations);
  const isGithubApp = integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP;

  const sync = useCallback(async () => {
    await load(true);
    onChanged();
  }, [load, onChanged]);

  return (
    <div className={cn(row.tone === "attention" && "bg-warn/[0.03]")}>
      <div className="flex items-center gap-4 px-4 py-3 hover:bg-muted/50">
        <div className="flex w-[180px] min-w-0 items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px] border border-border bg-card">
            <ProviderLogo providerId={providerIdFor(integration)} className="h-5 w-5 shrink-0" />
          </div>
          <div className="min-w-0">
            <p className="truncate text-[15px] font-semibold text-foreground">
              {PROVIDER_TITLES[providerIdFor(integration)]}
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
            <span className="truncate text-[11.5px] text-muted-foreground">{row.meter.left}</span>
            <span className="shrink-0 font-mono text-[11px] text-fg-muted">{row.meter.right}</span>
          </div>
          <div className="mt-1.5 h-1 w-full overflow-hidden rounded-full bg-border">
            <div className={cn("h-full rounded-full", meterFillClasses(row.meter.fill))} />
          </div>
        </div>

        <div className="w-[130px]">
          {row.statusKey === "connected" ? (
            <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <CircleCheck className="h-3.5 w-3.5 text-success" />
              Connected
            </span>
          ) : (
            <span
              className={cn(
                "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs",
                statusPillClasses(row.statusKey),
              )}
            >
              <span
                className={cn(
                  "h-1.5 w-1.5 rounded-full",
                  row.statusKey === "action_needed" ? "bg-danger" : "bg-warn",
                )}
              />
              {row.statusLabel}
            </span>
          )}
        </div>

        <RowMenu
          onVerify={isGithubApp ? undefined : () => onVerify(integration)}
          onSync={isGithubApp ? () => void sync() : undefined}
          onRemove={() => onRemove(integration)}
        />
      </div>

      {row.banner && (
        <Banner banner={row.banner} statusKey={row.statusKey} onVerify={() => onVerify(integration)} />
      )}
    </div>
  );
}
