import { useState, useRef, useEffect, useMemo } from "react";
import { Copy, Check, Loader2, Mail, AlertCircle, Link as LinkIcon } from "lucide-react";
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

// Default badge pill shown next to default team
function DefaultPill() {
  return (
    <span className="inline-flex items-center px-1.5 py-px text-[9px] font-mono uppercase tracking-wider rounded text-brand bg-brand-bg border border-brand-border">
      DEFAULT
    </span>
  );
}

// Role card — vertically stacked button with name + radio + description
function RoleCard({
  role,
  selected,
  onSelect,
  disabled,
}: {
  role: "Developer" | "Viewer";
  selected: boolean;
  onSelect: () => void;
  disabled?: boolean;
}) {
  const description =
    role === "Developer"
      ? "Create, edit, deploy, and remove resources in this team."
      : "Read-only access to this team’s resources and configuration.";

  return (
    <button
      type="button"
      onClick={onSelect}
      disabled={disabled}
      className={[
        "flex flex-col items-start gap-1.5 rounded-md border p-3.5 text-left transition-all",
        selected
          ? "border-brand bg-brand-bg"
          : "border-border bg-card hover:border-brand/50",
        disabled ? "cursor-not-allowed opacity-60" : "cursor-pointer",
      ].join(" ")}
    >
      <div className="flex w-full items-center justify-between">
        <span
          className={[
            "text-sm font-medium font-mono",
            selected ? "text-brand" : "text-foreground",
          ].join(" ")}
        >
          {role}
        </span>
        {/* Radio dot */}
        <span
          className={[
            "h-3.5 w-3.5 rounded-full border",
            selected ? "border-brand bg-brand" : "border-border-strong bg-transparent",
          ].join(" ")}
          style={selected ? { boxShadow: "inset 0 0 0 3px var(--card)" } : undefined}
        />
      </div>
      <span className="text-[11px] text-muted-foreground leading-snug">{description}</span>
    </button>
  );
}

// Server error banner — shown at top of form body
function ServerErrorBanner({
  message,
  onDismiss,
}: {
  message: string;
  onDismiss: () => void;
}) {
  return (
    <div className="flex items-start gap-2.5 rounded-md border border-danger-border bg-danger-bg px-3.5 py-3">
      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-danger" />
      <div className="grow">
        <p className="text-sm font-medium text-foreground">
          We couldn&apos;t create the invitation
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground leading-relaxed">{message}</p>
      </div>
      <button
        type="button"
        onClick={onDismiss}
        aria-label="Dismiss error"
        className="ml-1 shrink-0 text-muted-foreground hover:text-foreground transition-colors"
      >
        <span aria-hidden className="text-sm leading-none">&times;</span>
      </button>
    </div>
  );
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
  const [resultEmail, setResultEmail] = useState("");
  const [resultEmailError, setResultEmailError] = useState<string | null>(null);
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
  const resolvedDefaultTeam = teams.find((t) => t.default_team)?.name ?? "";
  const isDefaultTeamSelected = resolvedTeam === resolvedDefaultTeam;

  function resetForm() {
    setPhase("form");
    setEmail("");
    setTeamName("");
    setRole("Developer");
    setEmailError("");
    setTeamError("");
    setResultToken("");
    setResultEmail("");
    setResultEmailError(null);
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
      setResultEmail(email);
      // email_error not in OpenAPI schema — use generic fallback when email_sent=false
      setResultEmailError((res.invite as Record<string, unknown>).email_error as string ?? null);
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
  const isSuccess = phase === "success-sent" || phase === "success-failed";

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        {/* ── Header ── */}
        <DialogHeader className="flex-row items-start gap-3.5 space-y-0 pb-0">
          {/* Avatar icon */}
          <div
            className={[
              "flex h-9 w-9 shrink-0 items-center justify-center rounded-md border",
              isSuccess
                ? "border-success-border bg-success-bg text-success"
                : "border-brand-border bg-brand-bg text-brand",
            ].join(" ")}
          >
            {isSuccess ? (
              <Check className="h-[18px] w-[18px]" />
            ) : (
              <Mail className="h-[18px] w-[18px]" />
            )}
          </div>
          <div className="grow">
            <DialogTitle className="text-base">
              {isSuccess ? "Invitation created" : "Invite user"}
            </DialogTitle>
            <p className="mt-1 text-xs text-muted-foreground leading-snug">
              {isSuccess
                ? "Share the one-time link below — it won’t be retrievable again."
                : "They’ll receive an email with a one-time link to join this organisation."}
            </p>
          </div>
        </DialogHeader>

        {/* ── Form / Submitting phase ── */}
        {(phase === "form" || phase === "submitting") && (
          <>
            <div className="space-y-4">
              {/* Server error banner */}
              {localServerError && (
                <ServerErrorBanner
                  message={localServerError}
                  onDismiss={() => setLocalServerError(null)}
                />
              )}

              {/* Email field */}
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
                  disabled={isSubmitting}
                />
                {!emailError && (
                  <p className="text-[11px] text-muted-foreground">
                    A one-time invite link will be sent to this address.
                  </p>
                )}
              </FieldShell>

              {/* Team field */}
              <FieldShell label="Team" htmlFor="invite-team" error={teamError}>
                <Select
                  value={resolvedTeam}
                  onValueChange={(v) => {
                    setTeamName(v);
                    if (teamError) setTeamError("");
                  }}
                  disabled={isSubmitting}
                >
                  <SelectTrigger id="invite-team" className="w-full">
                    <SelectValue placeholder="Select a team">
                      {resolvedTeam && (
                        <span className="flex items-center gap-1.5">
                          <span className="font-mono text-sm">{resolvedTeam}</span>
                          {isDefaultTeamSelected && <DefaultPill />}
                        </span>
                      )}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {teams.map((t) => (
                      <SelectItem key={t.name} value={t.name}>
                        <span className="flex items-center gap-1.5">
                          <span className="font-mono">{t.name}</span>
                          {t.default_team && <DefaultPill />}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-[11px] text-muted-foreground leading-snug">
                  The invite is scoped to one team. The workspace default team is preselected
                  — change it if they should land somewhere else.
                </p>
              </FieldShell>

              {/* Role field */}
              <FieldShell label="Role on this team">
                <div className="grid grid-cols-2 gap-2.5">
                  {(["Developer", "Viewer"] as const).map((r) => (
                    <RoleCard
                      key={r}
                      role={r}
                      selected={role === r}
                      onSelect={() => setRole(r)}
                      disabled={isSubmitting}
                    />
                  ))}
                </div>
              </FieldShell>
            </div>

            <DialogFooter className="items-center">
              {/* Left: "Will send to" hint */}
              <span className="mr-auto text-xs text-muted-foreground">
                {email ? (
                  <>
                    Will send to{" "}
                    <code className="font-mono text-foreground">{email}</code>
                  </>
                ) : (
                  <span className="invisible">placeholder</span>
                )}
              </span>
              <Button
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button onClick={handleSubmit} disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
                {isSubmitting ? "Sending invitation…" : "Send invitation"}
              </Button>
            </DialogFooter>
          </>
        )}

        {/* ── Success phases ── */}
        {isSuccess && (
          <>
            <div className="space-y-4">
              {/* Email delivery banner */}
              <div
                className={[
                  "flex items-start gap-2.5 rounded-md border px-3.5 py-3",
                  phase === "success-sent"
                    ? "border-success-border bg-success-bg"
                    : "border-warn-border bg-warn-bg",
                ].join(" ")}
              >
                {phase === "success-sent" ? (
                  <Check className="mt-0.5 h-4 w-4 shrink-0 text-success" />
                ) : (
                  <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-warn" />
                )}
                <div>
                  <p
                    className={[
                      "text-sm font-medium",
                      phase === "success-sent" ? "text-success" : "text-warn",
                    ].join(" ")}
                  >
                    {phase === "success-sent" ? (
                      <>
                        We emailed{" "}
                        <code className="font-mono">{resultEmail}</code>
                      </>
                    ) : (
                      "Email delivery failed — share the link manually"
                    )}
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground leading-relaxed">
                    {phase === "success-sent"
                      ? "If they don’t see it within a few minutes, share the link below directly."
                      : `${resultEmailError ?? "Couldn’t send the email"}. The invitation is still valid — they just won’t get a notification.`}
                  </p>
                </div>
              </div>

              {/* One-time link block */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium">One-time invite link</span>
                  <span className="font-mono text-[10px] uppercase tracking-widest text-brand">
                    SHOWN ONCE
                  </span>
                </div>

                {/* Link box */}
                <div className="flex items-center gap-2 rounded-md border border-border bg-input px-3 py-2">
                  <LinkIcon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                  <code className="grow truncate font-mono text-xs text-foreground">
                    {resultToken}
                  </code>
                  <Button variant="ghost" size="icon" onClick={handleCopy} className="shrink-0 h-7 w-7">
                    {copied ? (
                      <Check className="h-3.5 w-3.5 text-brand" />
                    ) : (
                      <Copy className="h-3.5 w-3.5" />
                    )}
                  </Button>
                </div>

                <p className="text-[11px] text-muted-foreground leading-relaxed">
                  Expires in{" "}
                  <code className="font-mono text-foreground">1 day</code>.
                  {" "}This link won&apos;t be retrievable again — copy it now if you need to
                  share manually.
                </p>
              </div>
            </div>

            <DialogFooter className="items-center">
              {/* Left: token prefix */}
              <span className="mr-auto font-mono text-xs text-muted-foreground">
                invite &middot; {resultToken.slice(0, 12)}&hellip;
              </span>
              <Button
                variant="ghost"
                onClick={resetForm}
              >
                Invite another
              </Button>
              <Button onClick={() => handleOpenChange(false)}>Done</Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
