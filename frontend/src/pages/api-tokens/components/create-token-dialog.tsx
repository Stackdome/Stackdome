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

// Native <input type="date"> only carries a calendar day — treat it as the
// end of that day in the user's local time so "expires on this date" reads
// naturally, then hand the API an RFC3339 timestamp.
//
// Built from numeric y/m/d rather than parsing a "YYYY-MM-DDTHH:mm:ss" string:
// a date-time string with no timezone offset is local time in every current
// engine, but some older engines read it as UTC — the numeric constructor
// can't be ambiguous either way.
function endOfDayRFC3339(date: string): string {
  const [year, month, day] = date.split("-").map(Number);
  return new Date(year, month - 1, day, 23, 59, 59).toISOString();
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
  const [copied, setCopied] = useState(false);
  const copyTimer = useRef<ReturnType<typeof setTimeout>>(null);

  useEffect(() => () => { if (copyTimer.current) clearTimeout(copyTimer.current); }, []);

  useEffect(() => {
    if (!open) return;
    setName("");
    setExpiry("");
    setPreset(READ_ONLY);
    setError(null);
    setScopesError(null);
    setCopied(false);
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
    // Scopes come from the server contract — never guess a full-access scope
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

  async function handleCopy() {
    if (!created?.token) return;
    try {
      await copyText(created.token);
    } catch (e) {
      toast({ title: "Copy failed", description: getErrorMessage(e), variant: "destructive" });
      return;
    }
    setCopied(true);
    if (copyTimer.current) clearTimeout(copyTimer.current);
    copyTimer.current = setTimeout(() => setCopied(false), COPY_FLASH_MS);
  }

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
              <DialogDescription>You won&apos;t see this again — copy it now.</DialogDescription>
            </DialogHeader>
            <div className="min-w-0 space-y-4 py-2">
              <div className="min-w-0 space-y-1.5">
                <Label>Token</Label>
                <div className="flex items-center gap-2">
                  <pre className="min-w-0 flex-1 overflow-x-auto rounded-md border border-border bg-muted/50 px-3 py-2 font-mono text-xs">
                    {created.token}
                  </pre>
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={handleCopy}
                    aria-label={copied ? "Copied" : "Copy token"}
                  >
                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="token-quickstart">Quick start</Label>
                <Input
                  id="token-quickstart"
                  readOnly
                  value={`stackdome login --url ${SERVER_URL} --token ${created.token}`}
                  onFocus={(e) => e.currentTarget.select()}
                  className="font-mono text-xs"
                />
              </div>
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
                    min={new Date().toISOString().slice(0, 10)}
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
              <Button variant="outline" onClick={() => onOpenChange(false)} disabled={creating}>
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
