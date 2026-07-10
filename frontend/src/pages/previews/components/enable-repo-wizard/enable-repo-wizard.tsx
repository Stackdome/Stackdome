import { useEffect, useState } from "react";
import { GitPullRequest } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { listGitIntegrations } from "@/api/git-integrations";
import { getCurrentOrganizationId } from "@/helpers/common";
import { ConnectPhase } from "./connect-phase";
import { RepoPickerPhase } from "./repo-picker-phase";
import { ConfigurePhase } from "./configure-phase";

type Phase = "connect" | "pick" | "configure";

const CONNECTED_STATUSES = new Set(["installed", "active"]);

export interface PickedRepo {
  /** e.g. "acme/webapp" */
  fullName: string;
  cloneUrl: string;
  defaultBranch: string;
  /** null when the user typed a URL manually (no discovery available) */
  integrationId: string | null;
}

interface EnableRepoWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called with the created config's id after a successful create. */
  onCreated: (configId: string) => void;
}

export function EnableRepoWizard({ open, onOpenChange, onCreated }: EnableRepoWizardProps) {
  const [phase, setPhase] = useState<Phase>("connect");
  const [integrationId, setIntegrationId] = useState<string | null>(null);
  const [repo, setRepo] = useState<PickedRepo | null>(null);
  const [checkedIntegrations, setCheckedIntegrations] = useState(false);

  // On open: skip connect when a usable GitHub App integration already exists.
  useEffect(() => {
    if (!open) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    let cancelled = false;
    setCheckedIntegrations(false);
    listGitIntegrations(orgId)
      .then((list) => {
        if (cancelled) return;
        const connected = (list.items ?? []).find(
          (i) => i.type === "github_app" && CONNECTED_STATUSES.has(i.status ?? ""),
        );
        if (connected) {
          setIntegrationId(connected.id ?? null);
          setPhase("pick");
        } else {
          setPhase("connect");
        }
        setCheckedIntegrations(true);
      })
      .catch(() => {
        if (!cancelled) {
          setPhase("connect");
          setCheckedIntegrations(true);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  const close = () => {
    onOpenChange(false);
    setPhase("connect");
    setRepo(null);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[760px]">
        <DialogTitle className="sr-only">Enable repository</DialogTitle>
        <DialogDescription className="sr-only">
          Connect GitHub and enable preview environments on a repository
        </DialogDescription>
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-primary">
            <GitPullRequest className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Enable repository
          </span>
        </div>

        <div className="h-[520px] max-h-[80vh] overflow-hidden">
          {phase === "connect" && checkedIntegrations && (
            <ConnectPhase
              onConnected={(id) => {
                setIntegrationId(id);
                setPhase("pick");
              }}
              onCancel={close}
              onSkip={() => {
                setIntegrationId(null);
                setPhase("pick");
              }}
            />
          )}
          {phase === "pick" && (
            <RepoPickerPhase
              integrationId={integrationId}
              onPicked={(r) => {
                setRepo(r);
                setPhase("configure");
              }}
              onBack={() => setPhase("connect")}
            />
          )}
          {phase === "configure" && repo && (
            <ConfigurePhase
              repo={repo}
              onCreated={(configId) => {
                onCreated(configId);
                close();
              }}
              onBack={() => setPhase("pick")}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Referenced by later tasks; used from Task 10 onward.
export type { Phase };
