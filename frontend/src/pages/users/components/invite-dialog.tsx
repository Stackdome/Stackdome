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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FieldShell, AlertBanner } from "@/components/branded";
import { inviteSchema } from "../schemas/invite-schema";
import { useInvites } from "../hooks/use-invites";
import { useProjectOptions } from "../hooks/use-project-options";
import { getErrorMessage } from "@/api/client";

type Phase = "form" | "submitting" | "success-sent" | "success-failed";

interface InviteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}

// Default badge pill shown next to default project
function DefaultPill() {
  return (
    <span className="inline-flex items-center px-1.5 py-px text-[9px] font-mono uppercase tracking-wider rounded text-fg-2 bg-foreground/5 border border-border">
      DEFAULT
    </span>
  );
}

// Role card — labeled radio card with name + description. Selected state is an
// ink tint (rubric: active/selected = ink tint, never brand orange).
function RoleCard({
  role,
  selected,
  disabled,
}: {
  role: "Developer" | "Viewer";
  selected: boolean;
  disabled?: boolean;
}) {
  const description =
    role === "Developer"
      ? "Create, edit, deploy, and remove resources in this project."
      : "Read-only access to this project’s resources and configuration.";
  const inputId = `invite-role-${role}`;

  return (
    <label
      htmlFor={inputId}
      className={[
        "flex flex-col items-start gap-1.5 rounded-md border p-3.5 text-left transition-colors",
        selected
          ? "border-border-strong bg-foreground/5"
          : "border-border bg-card hover:bg-muted/30",
        disabled ? "cursor-not-allowed opacity-60" : "cursor-pointer",
      ].join(" ")}
    >
      <div className="flex w-full items-center justify-between">
        <span className="text-sm font-medium font-mono text-foreground">{role}</span>
        <RadioGroupItem id={inputId} value={role} disabled={disabled} />
      </div>
      <span className="text-[11px] text-muted-foreground leading-snug">{description}</span>
    </label>
  );
}

export function InviteDialog({ open, onOpenChange, onCreated }: InviteDialogProps) {
  const { create, submitting } = useInvites();
  const { projects } = useProjectOptions();

  // FIX 4: memoize default-project lookup
  const defaultProject = useMemo(() => projects.find((t) => t.default_project)?.name ?? projects[0]?.name ?? "", [projects]);

  const [phase, setPhase] = useState<Phase>("form");
  const [email, setEmail] = useState("");
  const [projectName, setProjectName] = useState("");
  const [role, setRole] = useState<"Developer" | "Viewer">("Developer");
  const [emailError, setEmailError] = useState("");
  const [projectError, setProjectError] = useState("");
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

  const resolvedProject = projectName || defaultProject;
  const resolvedDefaultProject = projects.find((t) => t.default_project)?.name ?? "";
  const isDefaultProjectSelected = resolvedProject === resolvedDefaultProject;

  function resetForm() {
    setPhase("form");
    setEmail("");
    setProjectName("");
    setRole("Developer");
    setEmailError("");
    setProjectError("");
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
    setProjectError("");
    setLocalServerError(null);

    const parsed = inviteSchema.safeParse({
      email,
      project_name: resolvedProject,
      role,
    });

    if (!parsed.success) {
      // FIX 5: remove redundant `as ZodError` cast
      for (const issue of parsed.error.issues) {
        if (issue.path[0] === "email") setEmailError(issue.message);
        if (issue.path[0] === "project_name") setProjectError(issue.message);
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
    const urlToCopy = resultToken
      ? `${window.location.origin}/sign-up?invite_token=${resultToken}`
      : resultToken;
    void navigator.clipboard.writeText(urlToCopy);
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
  const inviteUrl = resultToken
    ? `${window.location.origin}/sign-up?invite_token=${resultToken}`
    : "";

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
                : "border-border bg-foreground/5 text-foreground",
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
            <div className="min-w-0 space-y-4">
              {/* Server error banner */}
              {localServerError && (
                <AlertBanner
                  action={{ label: "Dismiss", onClick: () => setLocalServerError(null) }}
                >
                  We couldn&apos;t create the invitation — {localServerError}
                </AlertBanner>
              )}

              {/* Email field */}
              <FieldShell label="Email" htmlFor="invite-email" error={emailError} required>
                <Input
                  id="invite-email"
                  type="email"
                  placeholder="projectmate@example.com"
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

              {/* Project field */}
              <FieldShell label="Project" htmlFor="invite-project" error={projectError} required>
                <Select
                  value={resolvedProject}
                  onValueChange={(v) => {
                    setProjectName(v);
                    if (projectError) setProjectError("");
                  }}
                  disabled={isSubmitting}
                >
                  <SelectTrigger id="invite-project" className="w-full">
                    <SelectValue placeholder="Select a project">
                      {resolvedProject && (
                        <span className="flex items-center gap-1.5">
                          {isDefaultProjectSelected && resolvedProject === "default" ? (
                            <DefaultPill />
                          ) : (
                            <>
                              <span className="font-mono text-sm">{resolvedProject}</span>
                              {isDefaultProjectSelected && <DefaultPill />}
                            </>
                          )}
                        </span>
                      )}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {projects.map((t) => (
                      <SelectItem key={t.name} value={t.name}>
                        <span className="flex items-center gap-1.5">
                          {t.default_project && t.name === "default" ? (
                            <DefaultPill />
                          ) : (
                            <>
                              <span className="font-mono text-sm">{t.name}</span>
                              {t.default_project && <DefaultPill />}
                            </>
                          )}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-[11px] text-muted-foreground leading-snug">
                  The invite is scoped to one project. The workspace default project is preselected
                  — change it if they should land somewhere else.
                </p>
              </FieldShell>

              {/* Role field */}
              <FieldShell label="Role on this project" required>
                <RadioGroup
                  value={role}
                  onValueChange={(v) => setRole(v as "Developer" | "Viewer")}
                  className="grid grid-cols-2 gap-2.5"
                >
                  {(["Developer", "Viewer"] as const).map((r) => (
                    <RoleCard
                      key={r}
                      role={r}
                      selected={role === r}
                      disabled={isSubmitting}
                    />
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
                {isSubmitting ? "Sending invitation…" : "Send invitation"}
              </Button>
            </DialogFooter>
          </>
        )}

        {/* ── Success phases ── */}
        {isSuccess && (
          <>
            <div className="min-w-0 space-y-4">
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
                <div className="min-w-0 grow">
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
                  <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                    SHOWN ONCE
                  </span>
                </div>

                {/* Link box */}
                <div className="flex items-center gap-2 rounded-md border border-border bg-input px-3 py-2">
                  <LinkIcon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                  <code className="min-w-0 grow truncate font-mono text-xs text-foreground">
                    {inviteUrl}
                  </code>
                  <Button variant="ghost" size="icon" onClick={handleCopy} className="shrink-0" aria-label="Copy invite link">
                    {copied ? (
                      <Check className="h-3.5 w-3.5 text-success" />
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
              {/* Left: invited email */}
              <span className="mr-auto min-w-0 truncate font-mono text-xs text-muted-foreground">
                {resultEmail}
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
