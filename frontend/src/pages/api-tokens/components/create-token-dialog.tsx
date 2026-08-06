import { useEffect, useState } from "react";
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
import { Switch } from "@/components/ui/switch";
import { FieldShell } from "@/components/branded";
import {
  createApiToken,
  getApiTokenScopes,
  type APITokenCreateResponse,
  type ScopeList,
} from "@/api/api-tokens";
import { getErrorMessage } from "@/api/client";

interface CreateTokenDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called once the show-once token view is dismissed, so the list can refresh. */
  onCreated: () => void;
}

function scopeKey(resource: string, action: string): string {
  return `${resource}:${action}`;
}

// Native <input type="date"> only carries a calendar day — treat it as the
// end of that day in the user's local time so "expires on this date" reads
// naturally, then hand the API an RFC3339 timestamp.
function endOfDayRFC3339(date: string): string {
  return new Date(`${date}T23:59:59`).toISOString();
}

export function CreateTokenDialog({ open, onOpenChange, onCreated }: CreateTokenDialogProps) {
  const [name, setName] = useState("");
  const [expiry, setExpiry] = useState("");
  const [fullAccess, setFullAccess] = useState(true);
  const [selectedScopes, setSelectedScopes] = useState<Set<string>>(new Set());
  const [scopes, setScopes] = useState<ScopeList | null>(null);
  const [scopesError, setScopesError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<APITokenCreateResponse | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!open) return;
    setName("");
    setExpiry("");
    setFullAccess(true);
    setSelectedScopes(new Set());
    setError(null);
    setCreated(null);
    setCopied(false);
    getApiTokenScopes()
      .then(setScopes)
      .catch((e) => setScopesError(getErrorMessage(e)));
  }, [open]);

  function toggleScope(resource: string, action: string) {
    const key = scopeKey(resource, action);
    setSelectedScopes((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }

  async function handleCreate() {
    setError(null);
    if (!name.trim()) {
      setError("Name is required.");
      return;
    }
    const requestedScopes = fullAccess
      ? [scopes?.full_access_scope ?? "*"]
      : [...selectedScopes];
    if (requestedScopes.length === 0) {
      setError("Select at least one scope.");
      return;
    }

    setCreating(true);
    try {
      const res = await createApiToken({
        name: name.trim(),
        scopes: requestedScopes,
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
    await navigator.clipboard.writeText(created.token);
    setCopied(true);
  }

  function handleClose() {
    const wasCreated = created !== null;
    onOpenChange(false);
    if (wasCreated) onCreated();
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !next && handleClose()}>
      <DialogContent className="sm:max-w-[500px]">
        {created ? (
          <>
            <DialogHeader>
              <DialogTitle>Token created</DialogTitle>
              <DialogDescription>You won&apos;t see this again — copy it now.</DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-1.5">
                <Label>Token</Label>
                <div className="flex items-center gap-2">
                  <pre className="flex-1 overflow-x-auto rounded-md border border-border bg-muted/50 px-3 py-2 font-mono text-xs">
                    {created.token}
                  </pre>
                  <Button type="button" variant="outline" size="icon" onClick={handleCopy} title="Copy token">
                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="token-quickstart">Quick start</Label>
                <Input
                  id="token-quickstart"
                  readOnly
                  value={`stackdome login --url ${window.location.origin} --token ${created.token}`}
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
                Issue an API token for scripts, CI, or agents to authenticate as you.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-5 py-2">
              {error && <div className="text-sm text-danger bg-danger-bg p-3 rounded-md">{error}</div>}

              <FieldShell label="Name" htmlFor="token-name" required>
                <Input
                  id="token-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. ci, agent"
                />
              </FieldShell>

              <FieldShell label="Expires" htmlFor="token-expiry" hint="Leave blank for a token that never expires.">
                <Input
                  id="token-expiry"
                  type="date"
                  value={expiry}
                  onChange={(e) => setExpiry(e.target.value)}
                />
              </FieldShell>

              <FieldShell label="Scopes" required>
                <div className="space-y-3 rounded-md border border-border p-3">
                  <div className="flex items-center justify-between gap-2">
                    <Label htmlFor="full-access" className="font-normal">
                      Full access
                    </Label>
                    <Switch id="full-access" checked={fullAccess} onCheckedChange={setFullAccess} />
                  </div>
                  {scopesError && <p className="text-xs text-danger">{scopesError}</p>}
                  {!fullAccess && (
                    <div className="space-y-2 border-t border-border pt-3">
                      {scopes?.items?.map((resource) =>
                        resource.resource
                          ? resource.actions?.map((action) => {
                            const key = scopeKey(resource.resource!, action);
                            return (
                              <div key={key} className="flex items-center justify-between gap-2">
                                <Label htmlFor={key} className="font-mono text-xs font-normal">
                                  {key}
                                </Label>
                                <Switch
                                  id={key}
                                  checked={selectedScopes.has(key)}
                                  onCheckedChange={() => toggleScope(resource.resource!, action)}
                                />
                              </div>
                            );
                          })
                          : null,
                      )}
                    </div>
                  )}
                </div>
              </FieldShell>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)} disabled={creating}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={creating}>
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
