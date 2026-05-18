import { useState, useRef, useEffect, useMemo } from "react";
import { Copy, Check, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { FieldShell } from "@/components/branded";
import { inviteSchema } from "../schemas/invite-schema";
import { useInvites } from "../hooks/use-invites";
import { useTeamOptions } from "../hooks/use-team-options";
import { getErrorMessage } from "@/api/client";

type Phase = "form" | "submitting" | "success-sent" | "success-failed";

interface InviteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

export function InviteDialog({ open, onOpenChange, onCreated }: InviteDialogProps) {
  const { create, submitting } = useInvites();
  const { teams } = useTeamOptions();

  // FIX 4: memoize default-team lookup
  const defaultTeam = useMemo(() => teams.find((t) => t.default_team)?.name ?? teams[0]?.name ?? "", [teams]);

  const [phase, setPhase] = useState<Phase>("form");
  const [email, setEmail] = useState("");
  const [teamName, setTeamName] = useState("");
  const [role, setRole] = useState<"Developer" | "Viewer">("Developer");
  const [emailError, setEmailError] = useState("");
  const [teamError, setTeamError] = useState("");
  const [resultToken, setResultToken] = useState("");
  const [copied, setCopied] = useState(false);
  const [localServerError, setLocalServerError] = useState<string | null>(null);

  // FIX 2: ref to track copy timeout for cleanup
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // FIX 2: cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
    };
  }, []);

  const resolvedTeam = teamName || defaultTeam;

  function resetForm() {
    setPhase("form");
    setEmail("");
    setTeamName("");
    setRole("Developer");
    setEmailError("");
    setTeamError("");
    setResultToken("");
    setCopied(false);
    setLocalServerError(null);
    // FIX 2: clear pending copy timeout on reset
    if (copyTimeoutRef.current) {
      clearTimeout(copyTimeoutRef.current);
      copyTimeoutRef.current = null;
    }
  }

  function handleOpenChange(val: boolean) {
    if (!val) resetForm();
    onOpenChange(val);
  }

  async function handleSubmit() {
    setEmailError("");
    setTeamError("");
    setLocalServerError(null);

    const parsed = inviteSchema.safeParse({
      email,
      team_name: resolvedTeam,
      role,
    });

    if (!parsed.success) {
      // FIX 5: remove redundant `as ZodError` cast
      for (const issue of parsed.error.issues) {
        if (issue.path[0] === "email") setEmailError(issue.message);
        if (issue.path[0] === "team_name") setTeamError(issue.message);
      }
      return;
    }

    setPhase("submitting");
    try {
      const res = await create(parsed.data);
      setResultToken(res.token ?? "");
      setPhase(res.invite.email_sent ? "success-sent" : "success-failed");
      onCreated();
    } catch (e) {
      // FIX 1: read message from the thrown error directly (not stale hook state)
      setLocalServerError(getErrorMessage(e));
      setPhase("form");
    }
  }

  function handleCopy() {
    void navigator.clipboard.writeText(resultToken);
    setCopied(true);
    // FIX 2: clear any existing timeout before setting a new one
    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
    copyTimeoutRef.current = setTimeout(() => {
      setCopied(false);
      copyTimeoutRef.current = null;
    }, 2000);
  }

  const isSubmitting = phase === "submitting" || submitting;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite user</DialogTitle>
        </DialogHeader>

        {phase === "form" || phase === "submitting" ? (
          <>
            {localServerError && (
              <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
                {localServerError}
              </div>
            )}

            <div className="space-y-4">
              {/* FIX 3: Email field wrapped in FieldShell */}
              <FieldShell label="Email" htmlFor="invite-email" error={emailError}>
                <Input
                  id="invite-email"
                  type="email"
                  placeholder="teammate@example.com"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value);
                    if (emailError) setEmailError("");
                  }}
                  className={emailError ? "border-danger" : ""}
                  aria-label="Email"
                />
              </FieldShell>

              {/* FIX 3: Team field wrapped in FieldShell */}
              <FieldShell label="Team" htmlFor="invite-team" error={teamError}>
                <Select
                  value={resolvedTeam}
                  onValueChange={(v) => {
                    setTeamName(v);
                    if (teamError) setTeamError("");
                  }}
                >
                  <SelectTrigger id="invite-team" className="w-full">
                    <SelectValue placeholder="Select a team" />
                  </SelectTrigger>
                  <SelectContent>
                    {teams.map((t) => (
                      <SelectItem key={t.name} value={t.name}>
                        {t.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FieldShell>

              {/* FIX 3: Role field wrapped in FieldShell */}
              <FieldShell label="Role">
                <RadioGroup
                  value={role}
                  onValueChange={(v) => setRole(v as "Developer" | "Viewer")}
                  className="grid grid-cols-2 gap-3"
                >
                  {(["Developer", "Viewer"] as const).map((r) => (
                    <label
                      key={r}
                      htmlFor={`role-${r}`}
                      className={[
                        "flex cursor-pointer items-center gap-2 rounded-md border px-3 py-2.5 text-sm transition-colors",
                        role === r
                          ? "border-brand bg-brand-bg text-brand"
                          : "border-border text-foreground hover:border-brand/50",
                      ].join(" ")}
                    >
                      <RadioGroupItem id={`role-${r}`} value={r} />
                      <span className="font-medium">{r}</span>
                    </label>
                  ))}
                </RadioGroup>
              </FieldShell>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button onClick={handleSubmit} disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
                Send invite
              </Button>
            </DialogFooter>
          </>
        ) : phase === "success-sent" ? (
          <div className="space-y-4">
            <p className="text-sm text-foreground">
              Invitation sent! Copy the one-time link below — it can only be viewed once.
            </p>
            <div className="flex items-center gap-2 rounded-md border border-border bg-muted px-3 py-2">
              <code className="flex-1 truncate font-mono text-xs text-foreground">
                {resultToken}
              </code>
              <Button variant="ghost" size="icon" onClick={handleCopy} className="shrink-0">
                {copied ? <Check className="h-4 w-4 text-brand" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Expires in 1 day. This link won&apos;t be retrievable again — copy it now.
            </p>
            <DialogFooter>
              <Button onClick={() => handleOpenChange(false)}>Done</Button>
            </DialogFooter>
          </div>
        ) : (
          /* success-failed */
          <div className="space-y-4">
            <div className="rounded-md border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
              Couldn&apos;t send the email — share this link directly with your teammate.
            </div>
            <div className="flex items-center gap-2 rounded-md border border-border bg-muted px-3 py-2">
              <code className="flex-1 truncate font-mono text-xs text-foreground">
                {resultToken}
              </code>
              <Button variant="ghost" size="icon" onClick={handleCopy} className="shrink-0">
                {copied ? <Check className="h-4 w-4 text-brand" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Expires in 1 day. This link won&apos;t be retrievable again — copy it now.
            </p>
            <DialogFooter>
              <Button onClick={() => handleOpenChange(false)}>Done</Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
