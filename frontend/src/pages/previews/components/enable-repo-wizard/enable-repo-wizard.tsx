import { useState } from "react";
import { GitPullRequest } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { WizardFooter } from "@/pages/stacks/components/wizard/wizard-footer";
import { GitSourcePicker } from "@/components/git-source-picker/git-source-picker";
import type { PickedRepo } from "@/components/git-source-picker/types";
import { ConfigurePhase } from "./configure-phase";

// Canonical home of PickedRepo moved to the shared picker; re-exported here so
// existing importers (configure-phase, page tests) keep working.
export type { PickedRepo } from "@/components/git-source-picker/types";

type Phase = "pick" | "configure";

const PR_AUTOMATION_HINT =
  "PR automation requires a connected provider. Public URLs support manually created previews.";

interface EnableRepoWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called with the created config's id after a successful create. */
  onCreated: (configId: string) => void;
}

export function EnableRepoWizard({ open, onOpenChange, onCreated }: EnableRepoWizardProps) {
  const [phase, setPhase] = useState<Phase>("pick");
  const [repo, setRepo] = useState<PickedRepo | null>(null);

  const close = () => {
    onOpenChange(false);
    setPhase("pick");
    setRepo(null);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[760px]">
        <DialogTitle className="sr-only">Enable repository</DialogTitle>
        <DialogDescription className="sr-only">
          Pick a repository and enable preview environments on it
        </DialogDescription>
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-brand">
            <GitPullRequest className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Enable repository
          </span>
        </div>

        <div className="h-[520px] max-h-[80vh] overflow-hidden">
          {phase === "pick" && (
            <div className="flex h-full flex-col">
              <div className="flex-1 overflow-y-auto p-6">
                <GitSourcePicker value={repo} onChange={setRepo} publicUrlHint={PR_AUTOMATION_HINT} />
              </div>
              <WizardFooter
                onBack={close}
                onContinue={() => setPhase("configure")}
                continueDisabled={repo == null}
                hint={repo ? repo.fullName : "Pick a repository to continue"}
              />
            </div>
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
