import { useEffect, useState } from "react";
import { GitBranch, Github, KeyRound, Loader2 } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { createGitIntegration, type GitIntegration } from "@/api/git-integrations";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useGithubConnect } from "@/pages/previews/hooks/use-github-connect";
import { cn } from "@/lib/utils";

type Phase = "provider" | "github" | "credentials";

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
  { id: "bitbucket", label: "Bitbucket", hostPrefill: "bitbucket.org", hostPlaceholder: "bitbucket.org", hint: "Use an app password with repository read permission." },
  { id: "gitea", label: "Gitea", hostPrefill: "", hostPlaceholder: "gitea.example.com", hint: "Use an access token with read:repository scope." },
  { id: "other", label: "Other", hostPrefill: "", hostPlaceholder: "git.example.com", hint: "Any git host reachable over HTTPS with token or basic auth." },
];

const GIT_INTEGRATION_TYPE_CREDENTIALS: GitIntegration["type"] = "git_credentials";

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
  const [token, setToken] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const reset = () => {
    setPhase("provider");
    setHost("");
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
    setSubmitError(null);
    setPhase(p.id === "github" ? "github" : "credentials");
  };

  const submitCredentials = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !host.trim() || !token.trim()) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      await createGitIntegration(orgId, {
        host: host.trim(),
        type: GIT_INTEGRATION_TYPE_CREDENTIALS,
        auth: { token: token.trim() },
      });
      onCreated();
      close();
    } catch (e) {
      setSubmitError(getErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  const startAppInstall = () => {
    void github.connect().then(() => {
      // connect() resolves after the popup opens; completion is signalled by state.
    });
  };

  // Surface GitHub connect completion to the parent, then close.
  useEffect(() => {
    if (github.state === "connected" && open) {
      onCreated();
      close();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [github.state]);

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[760px]">
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
                  disabled={hasGithubApp || github.state === "waiting"}
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
                      : "Grants repository access via a GitHub App installation. Webhooks keep previews in sync."}
                  </span>
                  {github.state === "waiting" && (
                    <span className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                      Waiting for the GitHub App installation to finish in the popup…
                    </span>
                  )}
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
                {github.error && <p className="text-sm text-destructive">{github.error}</p>}
              </div>
              <WizardFooter
                onBack={() => setPhase("provider")}
                onContinue={close}
                continueLabel="Done"
                loading={github.state === "waiting"}
              />
            </>
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
                {submitError && <p className="text-sm text-destructive">{submitError}</p>}
              </div>
              <WizardFooter
                onBack={() => setPhase(provider.id === "github" ? "github" : "provider")}
                onContinue={() => void submitCredentials()}
                continueLabel="Add integration"
                continueDisabled={!host.trim() || !token.trim()}
                loading={submitting}
              />
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
