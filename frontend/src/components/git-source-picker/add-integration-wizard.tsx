import { useEffect, useState } from "react";
import { CheckCircle2, Circle, GitBranch, Github, KeyRound } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { FieldShell } from "@/components/branded";
import { WizardFooter } from "@/components/wizard-footer";
import { createGitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/lib/common";
import { useGithubConnect } from "@/hooks/use-github-connect";
import { cn } from "@/lib/utils";
import { GIT_INTEGRATION_TYPE_CREDENTIALS, type ProviderId } from "@/lib/git-integrations";
import { credentialsFormSchema, type CredentialsFormValues } from "@/components/git-source-picker/credentials-form-schema";
import { ProviderLogo } from "@/components/branded/provider-logo";

type Phase = "provider" | "github" | "credentials" | "connecting" | "done";

interface Provider {
  id: ProviderId;
  label: string;
  hostPlaceholder: string;
  hostPrefill: string;
  hint: string;
}

const PROVIDERS: Provider[] = [
  { id: "github", label: "GitHub", hostPrefill: "github.com", hostPlaceholder: "github.com", hint: "Use a fine-grained personal access token with repository read access." },
  { id: "gitlab", label: "GitLab", hostPrefill: "gitlab.com", hostPlaceholder: "gitlab.com or gitlab.example.com", hint: "Use a project or personal access token with read_repository scope." },
  { id: "bitbucket", label: "Bitbucket", hostPrefill: "bitbucket.org", hostPlaceholder: "bitbucket.org", hint: "Use an app password with repository read permission. Fill in the username above — Bitbucket app passwords require basic auth." },
  { id: "gitea", label: "Gitea", hostPrefill: "", hostPlaceholder: "gitea.example.com", hint: "Use an access token with read:repository scope." },
  { id: "other", label: "Other", hostPrefill: "", hostPlaceholder: "git.example.com", hint: "Any git host reachable over HTTPS with token or basic auth." },
];

const STEPS: { phase: Phase; label: string }[] = [
  { phase: "provider", label: "Provider" },
  { phase: "github", label: "Connect" },
  { phase: "done", label: "Done" },
];

/** Maps every wizard phase onto one of the three stepper steps for highlighting. */
const STEP_FOR_PHASE: Record<Phase, Phase> = {
  provider: "provider",
  github: "github",
  credentials: "github",
  connecting: "github",
  done: "done",
};

const CONNECTING_CHECKLIST = [
  "Opening GitHub authorization…",
  "Authorizing the installation…",
  "Fetching accessible repositories…",
];

function Stepper({ phase }: { phase: Phase }) {
  const currentStep = STEP_FOR_PHASE[phase];
  return (
    <div data-testid="wizard-stepper" className="flex items-center gap-2 border-b px-5 py-3">
      {STEPS.map((step, i) => {
        const isCurrent = step.phase === currentStep;
        const isPast = STEPS.findIndex((s) => s.phase === currentStep) > i;
        return (
          <div key={step.phase} className="flex flex-1 flex-col gap-1.5">
            <div className={cn("h-[3px] rounded-full", isCurrent || isPast ? "bg-brand" : "bg-border")} />
            <span
              data-current={isCurrent}
              className={cn(
                "font-mono text-[11px] uppercase tracking-[1.5px]",
                isCurrent && "text-foreground",
                !isCurrent && isPast && "text-muted-foreground",
                !isCurrent && !isPast && "text-muted-foreground/60",
              )}
            >
              {step.label}
            </span>
          </div>
        );
      })}
    </div>
  );
}

interface AddIntegrationWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Disables the GitHub App install option with an "already connected" note. */
  hasGithubApp: boolean;
  /** Fired after a successful create (credentials POST or GitHub App connect). */
  onCreated: () => void;
}

export function AddIntegrationWizard({ open, onOpenChange, hasGithubApp, onCreated }: AddIntegrationWizardProps) {
  const github = useGithubConnect();
  const [phase, setPhase] = useState<Phase>("provider");
  const [provider, setProvider] = useState<Provider>(PROVIDERS[0]);
  const [host, setHost] = useState("");
  const [username, setUsername] = useState("");
  const [token, setToken] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof CredentialsFormValues, string>>>({});

  const reset = () => {
    setPhase("provider");
    setHost("");
    setUsername("");
    setToken("");
    setSubmitError(null);
    setSubmitting(false);
    setFieldErrors({});
  };

  const close = () => {
    onOpenChange(false);
    reset();
  };

  const pickProvider = (p: Provider) => {
    setProvider(p);
    setHost(p.hostPrefill);
    setUsername("");
    setToken("");
    setSubmitError(null);
    setFieldErrors({});
    setPhase(p.id === "github" ? "github" : "credentials");
  };

  const submitCredentials = async () => {
    const parsed = credentialsFormSchema.safeParse({ host, username, token });
    if (!parsed.success) {
      const flat = parsed.error.flatten().fieldErrors;
      setFieldErrors({ host: flat.host?.[0], token: flat.token?.[0] });
      return;
    }
    setFieldErrors({});
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      const trimmedUsername = parsed.data.username?.trim();
      await createGitIntegration(orgId, {
        host: parsed.data.host,
        type: GIT_INTEGRATION_TYPE_CREDENTIALS,
        auth: trimmedUsername
          ? { basic: { username: trimmedUsername, password: parsed.data.token } }
          : { token: parsed.data.token },
      });
      onCreated();
      setPhase("done");
    } catch (e) {
      setSubmitError(getErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  const startAppInstall = () => {
    setPhase("connecting");
    void github.connect();
  };

  // Surface GitHub connect completion to the parent even if the wizard was
  // closed while the popup finished in the background, so the parent's list
  // refresh isn't dropped. The done phase itself only renders while open.
  useEffect(() => {
    if (github.state === "connected") {
      onCreated();
      if (open) setPhase("done");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [github.state]);

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[540px]">
        <DialogTitle className="sr-only">Connect provider</DialogTitle>
        <DialogDescription className="sr-only">
          Connect a git provider so Stackdome can clone repositories and trigger preview environments
        </DialogDescription>
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-brand">
            <GitBranch className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Connect provider
          </span>
        </div>

        <Stepper phase={phase} />

        <div className="flex h-[480px] max-h-[80vh] flex-col overflow-hidden">
          {phase === "provider" && (
            <div className="flex flex-1 items-center justify-center overflow-y-auto p-8">
              <div className="w-full">
                <div className="mb-7 text-center">
                  <h2 className="mb-2 text-2xl font-medium tracking-tight">
                    Where does your code live?
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    Pick a provider. You can add more later.
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-2.5">
                  {PROVIDERS.map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => pickProvider(p)}
                      className={cn(
                        "flex min-h-[76px] items-start gap-3 rounded-md border bg-card p-4 text-left transition-colors",
                        "hover:border-brand focus-visible:border-brand focus-visible:outline-none",
                      )}
                    >
                      <span className="flex h-9 w-9 flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                        <ProviderLogo providerId={p.id} className="h-[18px] w-[18px]" />
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="mb-0.5 block text-sm font-medium text-foreground">{p.label}</span>
                        <span className="block text-xs text-muted-foreground">
                          {p.id === "github" ? "App install or access token" : "Access token"}
                        </span>
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {phase === "github" && (
            <>
              <div className="flex flex-1 flex-col justify-center gap-3 overflow-y-auto p-8">
                <div className="mb-4 text-center">
                  <h2 className="mb-2 text-2xl font-medium tracking-tight">Connect GitHub</h2>
                  <p className="text-sm text-muted-foreground">
                    Install the GitHub App, or paste an access token.
                  </p>
                </div>
                <button
                  type="button"
                  disabled={hasGithubApp}
                  onClick={startAppInstall}
                  className={cn(
                    "flex flex-col items-start gap-1.5 rounded-md border bg-card p-4 text-left",
                    "hover:border-brand focus-visible:border-brand focus-visible:outline-none",
                    "disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:border-border",
                  )}
                >
                  <span className="flex items-center gap-2 text-sm font-medium">
                    <Github className="h-4 w-4" />
                    Install GitHub App
                    <span className="rounded-sm bg-brand-bg px-1.5 py-0.5 font-mono text-[11px] uppercase tracking-[1.5px] text-brand">
                      Recommended
                    </span>
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {hasGithubApp
                      ? "Already connected. Manage installations from the integration card."
                      : "Fine-grained access you manage on GitHub. Webhooks keep previews in sync. No tokens to rotate."}
                  </span>
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setHost(provider.hostPrefill);
                    setSubmitError(null);
                    setPhase("credentials");
                  }}
                  className={cn(
                    "flex flex-col items-start gap-1.5 rounded-md border bg-card p-4 text-left",
                    "hover:border-brand focus-visible:border-brand focus-visible:outline-none",
                  )}
                >
                  <span className="flex items-center gap-2 text-sm font-medium">
                    <KeyRound className="h-4 w-4" />
                    Use an access token
                  </span>
                  <span className="text-xs text-muted-foreground">{provider.hint}</span>
                </button>
              </div>
              <WizardFooter
                onBack={() => setPhase("provider")}
                onContinue={close}
                continueLabel="Done"
              />
            </>
          )}

          {phase === "connecting" && (
            <div className="flex flex-1 flex-col items-center justify-center gap-4 overflow-y-auto p-5 text-center">
              <span className="flex h-12 w-12 items-center justify-center rounded-full border bg-card text-brand">
                <Github className="h-6 w-6" />
              </span>
              <div className="space-y-1">
                <h3 className="text-sm font-medium">Installing the GitHub App</h3>
                <p className="text-xs text-muted-foreground">
                  Finish the installation in the GitHub popup. We&apos;ll pick it up here.
                </p>
              </div>

              {github.error ? (
                <div className="w-full space-y-2 rounded-md border border-destructive/40 bg-destructive/10 p-4 text-left">
                  <p className="text-sm font-medium text-destructive">Couldn&apos;t connect to GitHub</p>
                  <p className="text-xs text-muted-foreground">{github.error}</p>
                  <Button type="button" variant="outline" size="sm" onClick={() => void github.connect()}>
                    Retry
                  </Button>
                </div>
              ) : (
                <>
                  <ul className="mx-auto w-full max-w-[320px] space-y-2.5 rounded-md border bg-card p-4 text-left">
                    {CONNECTING_CHECKLIST.map((step, i) => {
                      // Popup already opened when this phase renders, so step 0
                      // is done while waiting; everything completes on connect.
                      const status: "done" | "active" | "pending" =
                        github.state === "connected" ? "done" : i === 0 ? "done" : i === 1 ? "active" : "pending";
                      return (
                        <li
                          key={step}
                          className={cn(
                            "flex items-center gap-2.5 text-xs",
                            status === "done" && "text-muted-foreground",
                            status === "active" && "text-foreground",
                            status === "pending" && "text-muted-foreground/50",
                          )}
                        >
                          {status === "done" && <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-success" />}
                          {status === "active" && (
                            <span className="flex h-3.5 w-3.5 shrink-0 items-center justify-center motion-safe:animate-pulse">
                              <span className="h-2 w-2 rounded-full bg-brand" />
                            </span>
                          )}
                          {status === "pending" && <Circle className="h-3.5 w-3.5 shrink-0 text-border" />}
                          {step}
                        </li>
                      );
                    })}
                  </ul>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void github.checkAgain()}
                  >
                    Check again
                  </Button>
                </>
              )}
            </div>
          )}

          {phase === "credentials" && (
            <>
              <div className="flex flex-1 flex-col justify-center gap-4 overflow-y-auto p-8">
                <div className="mb-2 text-center">
                  <h2 className="mb-2 text-2xl font-medium tracking-tight">
                    Connect {provider.id === "other" ? "your git host" : provider.label}
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    Stackdome stores the credentials encrypted and uses them only for clones.
                  </p>
                </div>
                <FieldShell
                  label="Host"
                  htmlFor="integration-host"
                  required
                  error={fieldErrors.host}
                >
                  <Input
                    id="integration-host"
                    placeholder={provider.hostPlaceholder}
                    value={host}
                    onChange={(e) => {
                      setHost(e.target.value);
                      setFieldErrors((prev) => ({ ...prev, host: undefined }));
                    }}
                    aria-invalid={!!fieldErrors.host}
                  />
                </FieldShell>
                <FieldShell
                  label="Username"
                  htmlFor="integration-username"
                  hint="Required for providers using basic auth (e.g. Bitbucket app passwords)."
                >
                  <Input
                    id="integration-username"
                    autoComplete="off"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                </FieldShell>
                <FieldShell
                  label="Access token"
                  htmlFor="integration-token"
                  required
                  hint={provider.hint}
                  error={fieldErrors.token}
                >
                  <Input
                    id="integration-token"
                    type="password"
                    autoComplete="off"
                    value={token}
                    onChange={(e) => {
                      setToken(e.target.value);
                      setFieldErrors((prev) => ({ ...prev, token: undefined }));
                    }}
                    aria-invalid={!!fieldErrors.token}
                  />
                </FieldShell>
                {submitError && (
                  <div className="space-y-1 rounded-md border border-destructive/40 bg-destructive/10 p-3">
                    <p className="text-sm font-medium text-destructive">Couldn&apos;t verify the token</p>
                    <p className="text-xs text-muted-foreground">{submitError}</p>
                  </div>
                )}
              </div>
              <WizardFooter
                onBack={() => setPhase(provider.id === "github" ? "github" : "provider")}
                onContinue={() => void submitCredentials()}
                continueLabel="Connect"
                loading={submitting}
              />
            </>
          )}

          {phase === "done" && (
            <div className="flex flex-1 flex-col items-center justify-center gap-4 overflow-y-auto p-5 text-center">
              <span className="flex h-14 w-14 items-center justify-center rounded-full bg-success/10 text-success">
                <CheckCircle2 className="h-8 w-8" />
              </span>
              <div className="space-y-1">
                <h3 className="text-sm font-medium">
                  {provider.id === "github" ? "GitHub App installed" : `${provider.label} connected`}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {provider.id === "github"
                    ? "Stackdome can now clone repositories from your installed accounts. Webhooks will keep preview environments in sync."
                    : `Stackdome can now clone repositories on ${host} using your access token.`}
                </p>
              </div>
              <Button type="button" onClick={close}>
                Done
              </Button>
            </div>
          )}

        </div>
      </DialogContent>
    </Dialog>
  );
}
