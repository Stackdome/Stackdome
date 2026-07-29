import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { Check, Copy, Download, Eye, EyeOff, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Panel, EmptyState, StatusPill } from "@/components/branded";
import { getErrorMessage } from "@/api/client";
import {
  getPostgresCredentials,
  type PostgresAddon,
  type PostgresCredentials,
} from "@/api/addons";
import { getCurrentOrganizationId } from "@/helpers/common";
import { ADDON_OUTPUT_FIELDS } from "@/pages/stacks/lib/addon-presets";
import { copyText } from "@/lib/clipboard";
import { cn } from "@/lib/utils";

const COPY_FLASH_MS = 1400;
const MASK = "••••••••••••";

interface CredentialState {
  loading: boolean;
  error: string | null;
  credentials: PostgresCredentials | null;
  superuser: boolean;
  revealed: boolean;
}

const INITIAL: CredentialState = {
  loading: true,
  error: null,
  credentials: null,
  superuser: false,
  revealed: false,
};

function CopyButton({ value, label }: { value: string; label: string }) {
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(null);
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current); }, []);

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      aria-label={label}
      className="size-7 text-muted-foreground hover:text-foreground"
      onClick={() => {
        void copyText(value).then(() => {
          setCopied(true);
          if (timer.current) clearTimeout(timer.current);
          timer.current = setTimeout(() => setCopied(false), COPY_FLASH_MS);
        });
      }}
    >
      {copied ? <Check className="size-3.5 text-success" /> : <Copy className="size-3.5" />}
    </Button>
  );
}

function ValueRow({
  label,
  display,
  copyValue,
  copyLabel,
  action,
  className,
}: {
  label: ReactNode;
  display: ReactNode;
  copyValue?: string;
  copyLabel?: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex min-w-0 flex-col gap-1", className)}>
      <span className="text-[12px] text-muted-foreground">{label}</span>
      <div className="flex min-w-0 items-start gap-1">
        <span className="min-w-0 flex-1 break-all pt-1 font-mono text-[13px] leading-5 text-foreground">
          {display}
        </span>
        {action}
        {copyValue !== undefined && <CopyButton value={copyValue} label={copyLabel ?? "Copy"} />}
      </div>
    </div>
  );
}

function maskPassword(url: string, password: string): string {
  if (!password) return url;
  return url.split(password).join(MASK).split(encodeURIComponent(password)).join(MASK);
}

function downloadCertificate(database: string, certificate: string) {
  const url = URL.createObjectURL(new Blob([certificate], { type: "application/x-pem-file" }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `${database}-ca.crt`;
  anchor.click();
  URL.revokeObjectURL(url);
}

function CredentialsBlock({
  database,
  state,
  superuserAllowed,
  onScopeChange,
  onReveal,
  onRetry,
}: {
  database: string;
  state: CredentialState;
  superuserAllowed: boolean;
  onScopeChange: (superuser: boolean) => void;
  onReveal: () => void;
  onRetry: () => void;
}) {
  if (state.loading) {
    return (
      <div className="flex items-center gap-2 rounded-md border border-border bg-muted/25 px-4 py-3 text-[13px] text-muted-foreground">
        <Loader2 className="size-3.5 animate-spin" />
        Reading credentials…
      </div>
    );
  }

  if (state.error) {
    return (
      <div className="flex items-center justify-between gap-4 rounded-md border border-danger-border bg-danger-bg px-4 py-3">
        <span className="text-[13px] text-danger">{state.error}</span>
        <Button type="button" variant="outline" size="sm" onClick={onRetry}>
          Retry
        </Button>
      </div>
    );
  }

  const credentials = state.credentials;
  if (!credentials) return null;

  const password = credentials.password ?? "";
  const url = credentials.connectionString ?? "";
  const certificate = credentials.caCertificate ?? "";

  return (
    <div className="flex flex-col gap-4 rounded-md border border-border bg-muted/25 px-4 py-3.5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="text-[12.5px] text-muted-foreground">
          Credentials for <span className="font-mono text-foreground">{database}</span>
        </span>
        {superuserAllowed && (
          <div className="flex items-center gap-1">
            <Button
              type="button"
              variant={state.superuser ? "outline" : "secondary"}
              size="sm"
              onClick={() => onScopeChange(false)}
            >
              Owner
            </Button>
            <Button
              type="button"
              variant={state.superuser ? "secondary" : "outline"}
              size="sm"
              onClick={() => onScopeChange(true)}
            >
              Superuser
            </Button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <ValueRow
          label="Username"
          display={credentials.username ?? "—"}
          copyValue={credentials.username ?? ""}
          copyLabel="Copy username"
        />
        <ValueRow
          label="Password"
          display={state.revealed ? password : MASK}
          copyValue={password}
          copyLabel="Copy password"
          action={
            <Button
              type="button"
              variant="ghost"
              size="icon"
              aria-label={state.revealed ? "Hide password" : "Reveal password"}
              className="size-7 text-muted-foreground hover:text-foreground"
              onClick={onReveal}
            >
              {state.revealed ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
            </Button>
          }
        />
      </div>

      <ValueRow
        label="Connection URL"
        display={state.revealed ? url : maskPassword(url, password)}
        copyValue={url}
        copyLabel="Copy connection URL"
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <ValueRow label="SSL mode" display={credentials.sslMode ?? "—"} />
        {certificate && (
          <ValueRow
            label="CA certificate"
            display={<span className="text-muted-foreground">PEM certificate</span>}
            copyValue={certificate}
            copyLabel="Copy CA certificate"
            action={
              <Button
                type="button"
                variant="ghost"
                size="icon"
                aria-label="Download CA certificate"
                className="size-7 text-muted-foreground hover:text-foreground"
                onClick={() => downloadCertificate(database, certificate)}
              >
                <Download className="size-3.5" />
              </Button>
            }
          />
        )}
      </div>

      <p className="text-[12px] text-muted-foreground">
        Copy always copies the real value — only the text on screen is masked.
      </p>
    </div>
  );
}

/**
 * Connection details for a Postgres add-on: the in-cluster endpoint from the
 * add-on status, plus per-database credentials read on demand.
 */
export function PostgresConnectionPanel({
  addon,
  projectName,
}: {
  addon: PostgresAddon;
  projectName: string;
}) {
  const [states, setStates] = useState<Record<string, CredentialState>>({});
  const info = addon.status?.connection_info;
  const databases = info?.databases ?? [];
  const appUserSecrets = info?.credentials?.app_user_secrets ?? {};
  const superuserAllowed = !!addon.spec.configuration?.enable_superuser_access;

  const load = useCallback(
    async (database: string, superuser: boolean) => {
      const orgId = getCurrentOrganizationId();
      if (!orgId || !addon.id) return;
      setStates((prev) => ({ ...prev, [database]: { ...INITIAL, superuser } }));
      try {
        const credentials = await getPostgresCredentials(
          orgId,
          projectName,
          addon.id,
          database,
          superuser,
        );
        setStates((prev) => ({
          ...prev,
          [database]: { loading: false, error: null, credentials, superuser, revealed: false },
        }));
      } catch (e) {
        setStates((prev) => ({
          ...prev,
          [database]: {
            loading: false,
            error: getErrorMessage(e),
            credentials: null,
            superuser,
            revealed: false,
          },
        }));
      }
    },
    [addon.id, projectName],
  );

  const hide = useCallback((database: string) => {
    setStates((prev) => {
      const next = { ...prev };
      delete next[database];
      return next;
    });
  }, []);

  const toggleReveal = useCallback((database: string) => {
    setStates((prev) => {
      const current = prev[database];
      if (!current) return prev;
      return { ...prev, [database]: { ...current, revealed: !current.revealed } };
    });
  }, []);

  const body = !info || databases.length === 0 ? (
    <EmptyState
      title="No connection details yet"
      description="Connection details appear once the database is ready."
    />
  ) : (
    <div className="flex flex-col gap-6">
      <div className="grid max-w-3xl grid-cols-1 gap-5 sm:grid-cols-2">
        <ValueRow
          label="Host"
          display={info.host ?? "—"}
          copyValue={info.host ?? ""}
          copyLabel="Copy host"
        />
        <ValueRow label="Port" display={info.port ?? 5432} />
      </div>

      <p className="max-w-2xl text-[12px] text-muted-foreground">
        Reachable from stacks running on this cluster. The database is not exposed to the
        internet — connect from a stack, or port-forward for a local session.
      </p>

      <div className="flex flex-col gap-3 border-t border-border pt-5">
        <h3 className="text-sm font-semibold text-foreground">Databases</h3>
        {databases.map((db) => {
          const name = db.name ?? "";
          const state = states[name];
          return (
            <div key={name} className="flex flex-col gap-3 border-b border-border pb-3 last:border-b-0 last:pb-0">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex flex-col gap-0.5">
                  <span className="font-mono text-[13px] text-foreground">{name}</span>
                  <span className="text-[12px] text-muted-foreground">
                    Owner <span className="font-mono">{db.owner ?? "—"}</span>
                    {appUserSecrets[name] && (
                      <>
                        {" · secret "}
                        <span className="font-mono">{appUserSecrets[name]}</span>
                      </>
                    )}
                  </span>
                </div>
                {state ? (
                  <Button type="button" variant="outline" size="sm" onClick={() => hide(name)}>
                    <EyeOff className="size-3.5" />
                    Hide
                  </Button>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void load(name, false)}
                  >
                    <Eye className="size-3.5" />
                    Show credentials
                  </Button>
                )}
              </div>
              {state && (
                <CredentialsBlock
                  database={name}
                  state={state}
                  superuserAllowed={superuserAllowed}
                  onScopeChange={(superuser) => void load(name, superuser)}
                  onReveal={() => toggleReveal(name)}
                  onRetry={() => void load(name, state.superuser)}
                />
              )}
            </div>
          );
        })}
      </div>

      <div className="flex flex-col gap-3 border-t border-border pt-5">
        <h3 className="text-sm font-semibold text-foreground">Use from a stack</h3>
        <p className="max-w-2xl text-[12px] leading-relaxed text-muted-foreground">
          On a stack resource, add an env connection from this add-on, pick the database and
          credential scope, then map any of these outputs to environment variable names. The
          values are injected as secret references when the stack is released — nothing is
          baked into the image, and rotating the add-on password reaches the app on its next
          release.
        </p>
        <div className="flex flex-wrap gap-1.5">
          {ADDON_OUTPUT_FIELDS.map((field) => (
            <span
              key={field}
              className="rounded-sm border border-border bg-muted/40 px-1.5 py-0.5 font-mono text-[11.5px] text-muted-foreground"
            >
              {field}
            </span>
          ))}
        </div>
      </div>
    </div>
  );

  return (
    <Panel
      title="Connection"
      action={<StatusPill variant="neutral" withDot={false}>In-cluster only</StatusPill>}
    >
      {body}
    </Panel>
  );
}
