import { useCallback, useEffect, useRef, useState } from "react";
import { CircleAlert, CircleCheck, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  listInstallations,
  type GitIntegration,
  type GitInstallation,
} from "@/api/git-integrations";
import { getCurrentOrganizationId } from "@/helpers/common";
import { deriveRow, providerIdFor, GIT_INTEGRATION_TYPE_GITHUB_APP, type RowViewModel } from "../lib/derive-row";
import { RowMenu } from "./row-menu";
import { ProviderLogo } from "./provider-logo";

function statusPillClasses(tone: RowViewModel["tone"]) {
  if (tone === "ok") return "";
  return "border-warn-border bg-warn-bg text-warn";
}

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
            className="whitespace-nowrap font-medium text-brand hover:text-brand-hover"
          >
            {banner.ctaLabel}
          </a>
        ) : (
          <Button variant="ghost" size="sm" className="h-auto p-0 text-brand" disabled>
            {banner.ctaLabel}
          </Button>
        )
      ) : (
        <Button variant="ghost" size="sm" className="h-auto p-0 text-brand" onClick={onVerify}>
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

  const sync = useCallback(async () => {
    await load(true);
    onChanged();
  }, [load, onChanged]);

  return (
    <div className={cn(row.tone === "attention" && "bg-warn/[0.03]")}>
      <div className="flex items-center gap-4 px-4 py-3 hover:bg-muted/50">
        <div className="flex w-[180px] min-w-0 items-center gap-2">
          <ProviderLogo providerId={providerIdFor(integration)} className="h-5 w-5 shrink-0" />
          <div className="min-w-0">
            <p className="truncate text-[15px] font-semibold text-foreground">
              {integration.type === GIT_INTEGRATION_TYPE_GITHUB_APP ? "GitHub App" : "Git credentials"}
            </p>
            <p className="truncate font-mono text-[11.5px] text-fg-muted">{row.host}</p>
          </div>
        </div>

        <div className="w-[130px]">
          <span className="inline-flex items-center rounded-full border border-border px-2 py-0.5 text-xs text-muted-foreground">
            {row.authLabel}
          </span>
        </div>

        <div className="flex-1 min-w-0">
          {row.accessLine && <p className="truncate text-xs text-muted-foreground">{row.accessLine}</p>}
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
                statusPillClasses(row.tone),
                row.statusKey === "action_needed" && "border-danger-border bg-danger-bg text-danger",
              )}
            >
              {row.statusLabel}
            </span>
          )}
        </div>

        <RowMenu
          onVerify={() => onVerify(integration)}
          onSync={() => void sync()}
          onRemove={() => onRemove(integration)}
        />
      </div>

      {row.banner && (
        <Banner banner={row.banner} statusKey={row.statusKey} onVerify={() => onVerify(integration)} />
      )}
    </div>
  );
}
