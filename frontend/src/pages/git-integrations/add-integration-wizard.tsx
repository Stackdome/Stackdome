import { useEffect, useState } from "react";
import { CheckCircle2, Circle, GitBranch, Github, KeyRound, Loader2 } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { createGitIntegration, type GitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useGithubConnect } from "@/pages/previews/hooks/use-github-connect";
import { cn } from "@/lib/utils";

type Phase = "provider" | "github" | "credentials" | "connecting" | "done";

type ProviderId = "github" | "gitlab" | "bitbucket" | "gitea" | "other";

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

const GIT_INTEGRATION_TYPE_CREDENTIALS: GitIntegration["type"] = "git_credentials";

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
            <div className={cn("h-[3px] rounded-full", isCurrent || isPast ? "bg-primary" : "bg-border")} />
            <span
              data-current={isCurrent}
              className={cn(
                "font-mono text-[10px] uppercase tracking-[1.5px]",
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

  const reset = () => {
    setPhase("provider");
    setHost("");
    setUsername("");
    setToken("");
    setSubmitError(null);
    setSubmitting(false);
  };

  const close = () => {
    onOpenChange(false);
    reset();
  };

  const pickProvider = (p: Provider) => {
    setProvider(p);
    setHost(p.hostPrefill);
    setUsername("");
    setSubmitError(null);
    setPhase(p.id === "github" ? "github" : "credentials");
  };

  const submitCredentials = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !host.trim() || !token.trim()) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      const trimmedUsername = username.trim();
      await createGitIntegration(orgId, {
        host: host.trim(),
        type: GIT_INTEGRATION_TYPE_CREDENTIALS,
        auth: trimmedUsername
          ? { basic: { username: trimmedUsername, password: token.trim() } }
          : { token: token.trim() },
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
        <DialogTitle className="sr-only">Add git integration</DialogTitle>
        <DialogDescription className="sr-only">
          Connect a git provider so Stackdome can clone repositories and trigger preview environments
        </DialogDescription>
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-primary">
            <GitBranch className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Add git integration
          </span>
        </div>

        <Stepper phase={phase} />

        <div className="flex h-[480px] max-h-[80vh] flex-col overflow-hidden">
          {phase === "provider" && (
            <div className="grid flex-1 grid-cols-2 content-start gap-3 overflow-y-auto p-5 sm:grid-cols-3">
              {PROVIDERS.map((p) => (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => pickProvider(p)}
                  className={cn(
                    "flex flex-col items-start gap-2 rounded-md border bg-card p-4 text-left",
                    "hover:border-primary focus-visible:border-primary focus-visible:outline-none",
                  )}
                >
                  {p.id === "github"
                    ? <Github className="h-5 w-5 text-muted-foreground" />
                    : <GitBranch className="h-5 w-5 text-muted-foreground" />}
                  <span className="text-sm font-medium">{p.label}</span>
                  <span className="text-xs text-muted-foreground">
                    {p.id === "github" ? "App install or access token" : "Access token"}
                  </span>
                </button>
              ))}
            </div>
          )}

          {phase === "github" && (
            <>
              <div className="flex flex-1 flex-col gap-3 overflow-y-auto p-5">
                <button
                  type="button"
                  disabled={hasGithubApp}
                  onClick={startAppInstall}
                  className={cn(
                    "flex flex-col items-start gap-1.5 rounded-md border bg-card p-4 text-left",
                    "hover:border-primary focus-visible:border-primary focus-visible:outline-none",
                    "disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:border-border",
                  )}
                >
                  <span className="flex items-center gap-2 text-sm font-medium">
                    <Github className="h-4 w-4" />
                    Install GitHub App
                    <span className="rounded-sm bg-primary/10 px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-primary">
                      Recommended
                    </span>
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {hasGithubApp
                      ? "Already connected — manage installations from the integration card."
                      : "Fine-grained access you manage on GitHub. Webhooks keep previews in sync — no tokens to rotate."}
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
                    "hover:border-primary focus-visible:border-primary focus-visible:outline-none",
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
              <span className="flex h-12 w-12 items-center justify-center rounded-full border bg-card text-primary">
                <Github className="h-6 w-6" />
              </span>
              <div className="space-y-1">
                <h3 className="text-sm font-medium">Installing the GitHub App</h3>
                <p className="text-xs text-muted-foreground">
                  Finish the installation in the GitHub popup — we&apos;ll pick it up here.
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
                  <ul className="w-full space-y-2 text-left">
                    {CONNECTING_CHECKLIST.map((step) => (
                      <li key={step} className="flex items-center gap-2 text-xs text-muted-foreground">
                        {github.state === "connected"
                          ? <CheckCircle2 className="h-3.5 w-3.5 text-success" />
                          : <Circle className="h-3.5 w-3.5 animate-pulse text-primary" />}
                        {step}
                      </li>
                    ))}
                  </ul>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    Waiting for the GitHub App installation to finish…
                  </div>
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
              <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-5">
                <div className="space-y-1.5">
                  <Label htmlFor="integration-host">Host</Label>
                  <Input
                    id="integration-host"
                    placeholder={provider.hostPlaceholder}
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="integration-username">Username</Label>
                  <Input
                    id="integration-username"
                    autoComplete="off"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Required for providers using basic auth (e.g. Bitbucket app passwords).
                  </p>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="integration-token">Access token</Label>
                  <Input
                    id="integration-token"
                    type="password"
                    autoComplete="off"
                    value={token}
                    onChange={(e) => setToken(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">{provider.hint}</p>
                </div>
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
                continueDisabled={!host.trim() || !token.trim()}
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
