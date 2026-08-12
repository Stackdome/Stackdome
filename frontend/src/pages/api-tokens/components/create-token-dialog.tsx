import { useEffect, useRef, useState } from "react";
import { Copy, Check, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { useToast } from "@/components/ui/use-toast";
import { FieldShell } from "@/components/branded";
import { cn } from "@/lib/utils";
import {
  createApiToken,
  getApiTokenScopes,
  type APITokenCreateResponse,
  type ScopeList,
} from "@/api/api-tokens";
import { getErrorMessage } from "@/api/client";
import { API_BASE_URL } from "@/api/base-url";
import { copyText } from "@/lib/clipboard";
import { FULL_ACCESS, READ_ONLY, readOnlyScopes } from "../access";

const COPY_FLASH_MS = 1400;

// The CLI needs the server origin, which is the UI's own only when the API is
// same-origin or dev-proxied. An absolute VITE_API_BASE_URL carries its own.
const SERVER_URL = new URL(API_BASE_URL, window.location.origin).origin;

interface CreateTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called once the show-once token view is dismissed, so the list can refresh. */
  onCreated: () => void;
}

// Native <input type="date"> only carries a calendar day, so treat it as the
// end of that day in the user's local time so "expires on this date" reads
// naturally, then hand the API an RFC3339 timestamp.
//
// Built from numeric y/m/d rather than parsing a "YYYY-MM-DDTHH:mm:ss" string:
// a date-time string with no timezone offset is local time in every current
// engine, but some older engines read it as UTC. The numeric constructor
// can't be ambiguous either way.
function endOfDayRFC3339(date: string): string {
  const [year, month, day] = date.split("-").map(Number);
  return new Date(year, month - 1, day, 23, 59, 59).toISOString();
}

// Local, not toISOString(): the picker's floor has to agree with the local
// end-of-day above, or today is greyed out for anyone west of UTC.
function todayLocal(): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${now.getFullYear()}-${month}-${day}`;
}

const cliLoginCommand = (token: string) => `stackdome login --url ${SERVER_URL} --token ${token}`;

function CopyBlock({
  label,
  copyLabel,
  value,
  copied,
  onCopy,
}: {
  label: string;
  copyLabel: string;
  value: string;
  copied: boolean;
  onCopy: () => void;
}) {
  return (
    <div className="min-w-0 space-y-1.5">
      <span className="text-[12px] text-muted-foreground select-none">{label}</span>
      <div className="relative rounded-md border border-border bg-muted/40 py-2 pr-11 pl-3">
        <p className="min-w-0 font-mono text-[13px] leading-5 break-all text-foreground">{value}</p>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="absolute top-1.5 right-1.5 size-7 text-muted-foreground transition-transform hover:text-foreground active:scale-95"
          onClick={onCopy}
          aria-label={copied ? "Copied" : copyLabel}
        >
          {copied ? <Check className="size-3.5 text-success" /> : <Copy className="size-3.5" />}
        </Button>
      </div>
    </div>
  );
}

export function CreateTokenDialog({ open, onOpenChange, onCreated }: CreateTokenDialogProps) {
  const { toast } = useToast();
  const [name, setName] = useState("");
  const [expiry, setExpiry] = useState("");
  const [preset, setPreset] = useState(READ_ONLY);
  const [scopes, setScopes] = useState<ScopeList | null>(null);
  const [scopesError, setScopesError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<APITokenCreateResponse | null>(null);
  const [copied, setCopied] = useState<string | null>(null);
  const copyTimer = useRef<ReturnType<typeof setTimeout>>(null);

  useEffect(() => () => { if (copyTimer.current) clearTimeout(copyTimer.current); }, []);

  useEffect(() => {
    if (!open) return;
    setName("");
    setExpiry("");
    setPreset(READ_ONLY);
    setError(null);
    setScopes(null);
    setScopesError(null);
    setCopied(null);
    getApiTokenScopes()
      .then((res) => {
        setScopes(res);
        setScopesError(null);
      })
      .catch((e) => setScopesError(getErrorMessage(e)));
  }, [open]);

  const grantedScopes = !scopes
    ? []
    : preset === FULL_ACCESS
      ? [scopes.full_access_scope].filter((s): s is string => !!s)
      : readOnlyScopes(scopes);

  async function handleCreate() {
    setError(null);
    if (!name.trim()) {
      setError("Name is required.");
      return;
    }
    if (expiry && new Date(endOfDayRFC3339(expiry)) <= new Date()) {
      setError("Expiry must be in the future.");
      return;
    }
    // Scopes come from the server contract. Never guess a full-access scope
    // when that contract couldn't be read.
    if (!scopes) {
      setError("Scopes haven't loaded yet.");
      return;
    }
    if (grantedScopes.length === 0) {
      setError("The server didn't provide any scopes for this access level.");
      return;
    }

    setCreating(true);
    try {
      const res = await createApiToken({
        name: name.trim(),
        scopes: grantedScopes,
        expires_at: expiry ? endOfDayRFC3339(expiry) : undefined,
      });
      setCreated(res);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setCreating(false);
    }
  }

  async function handleCopy(field: string, value: string) {
    try {
      await copyText(value);
    } catch (e) {
      toast({ title: "Copy failed", description: getErrorMessage(e), variant: "destructive" });
      return;
    }
    setCopied(field);
    if (copyTimer.current) clearTimeout(copyTimer.current);
    copyTimer.current = setTimeout(() => setCopied(null), COPY_FLASH_MS);
  }

  const rawToken = created?.token;

  function handleClose() {
    const wasCreated = created !== null;
    onOpenChange(false);
    setCreated(null);
    if (wasCreated) onCreated();
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !next && handleClose()}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-[560px]">
        {created ? (
          <>
            <DialogHeader>
              <DialogTitle>Token created</DialogTitle>
              <DialogDescription>
                {rawToken
                  ? "Copy it now. You won't be able to see it again."
                  : "The token was created, but the server did not return it. Revoke it and create another."}
              </DialogDescription>
            </DialogHeader>
            <div className="min-w-0 space-y-4 py-2">
              {rawToken && (
                <>
                  <CopyBlock
                    label="Token"
                    copyLabel="Copy token"
                    value={rawToken}
                    copied={copied === "token"}
                    onCopy={() => handleCopy("token", rawToken)}
                  />
                  <CopyBlock
                    label="Log in with the CLI"
                    copyLabel="Copy CLI login command"
                    value={cliLoginCommand(rawToken)}
                    copied={copied === "cli"}
                    onCopy={() => handleCopy("cli", cliLoginCommand(rawToken))}
                  />
                </>
              )}
            </div>
            <DialogFooter>
              <Button onClick={handleClose}>Done</Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Create token</DialogTitle>
              <DialogDescription>
                Tokens act on your behalf. Pick what the caller is for.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-5 py-2">
              {error && <div className="text-sm text-danger bg-danger-bg p-3 rounded-md">{error}</div>}

              <div className="grid grid-cols-2 gap-4">
                <FieldShell label="Name" htmlFor="token-name" required>
                  <Input
                    id="token-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. ci, agent"
                  />
                </FieldShell>

                <FieldShell label="Expires" htmlFor="token-expiry" hint="Blank never expires.">
                  <Input
                    id="token-expiry"
                    type="date"
                    min={todayLocal()}
                    value={expiry}
                    onChange={(e) => setExpiry(e.target.value)}
                  />
                </FieldShell>
              </div>

              <FieldShell label="Access" required>
                {scopesError && <p className="text-xs text-danger">{scopesError}</p>}
                {!scopesError && !scopes && (
                  <p className="text-xs text-muted-foreground">Loading available scopes…</p>
                )}
                {scopes && (
                  <RadioGroup value={preset} onValueChange={setPreset} className="grid grid-cols-2 gap-3">
                    {[
                      { id: READ_ONLY, title: "Read-only", description: "Observe everything, change nothing." },
                      { id: FULL_ACCESS, title: "Full access", description: "Everything you can do, including deploys." },
                    ].map((option) => {
                      const active = preset === option.id;
                      return (
                        <Label
                          key={option.id}
                          htmlFor={`preset-${option.id}`}
                          className={cn(
                            "flex cursor-pointer flex-col items-start gap-1 rounded-md border p-3 font-normal",
                            active ? "border-brand-border bg-brand-bg" : "border-border hover:bg-muted/50",
                          )}
                        >
                          <div className="flex w-full items-center justify-between gap-2">
                            <span className="text-sm font-medium text-foreground">{option.title}</span>
                            <RadioGroupItem id={`preset-${option.id}`} value={option.id} />
                          </div>
                          <span className="text-xs text-muted-foreground">{option.description}</span>
                        </Label>
                      );
                    })}
                  </RadioGroup>
                )}
              </FieldShell>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={handleClose} disabled={creating}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={creating || !scopes}>
                {creating && <Loader2 className="h-4 w-4 animate-spin" />}
                Create
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
