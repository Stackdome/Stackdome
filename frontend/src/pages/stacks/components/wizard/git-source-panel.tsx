import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { WizardFooter } from "./wizard-footer";
import { GitSourcePicker } from "@/components/git-source-picker/git-source-picker";
import type { PickedRepo } from "@/components/git-source-picker/types";
import { buildGitSeed, defaultServiceName } from "@/pages/stacks/lib/git-source-seed";

type Step = "pick" | "form";

interface GitSourcePanelProps {
  onBack: () => void;
  onClose: () => void;
}

export function GitSourcePanel({ onBack, onClose }: GitSourcePanelProps) {
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>("pick");
  const [repo, setRepo] = useState<PickedRepo | null>(null);
  const [serviceName, setServiceName] = useState("");
  const [branch, setBranch] = useState("");
  const [dockerfilePath, setDockerfilePath] = useState("Dockerfile");
  const [buildContext, setBuildContext] = useState(".");
  const [port, setPort] = useState("");
  const [exposePublic, setExposePublic] = useState(true);

  const toForm = () => {
    if (!repo) return;
    setServiceName(defaultServiceName(repo));
    setBranch(repo.defaultBranch || "main");
    setStep("form");
  };

  const openInEditor = () => {
    if (!repo) return;
    const portNumber = Number.parseInt(port, 10);
    navigate("/stacks/new", {
      state: {
        seed: buildGitSeed(repo, {
          serviceName,
          branch,
          dockerfilePath,
          buildContext,
          port: portNumber,
          exposePublic,
        }),
      },
    });
    onClose();
  };

  const portValid = /^\d+$/.test(port) && Number(port) > 0 && Number(port) < 65536;

  if (step === "pick") {
    return (
      <div className="flex h-full flex-col">
        <div className="flex-1 overflow-y-auto p-6">
          <div className="mx-auto max-w-[640px] space-y-4">
            <div className="text-center">
              <h3 className="text-lg font-medium">Where does your code live?</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Pick a repository — the stack builds straight from it.
              </p>
            </div>
            <GitSourcePicker value={repo} onChange={setRepo} />
          </div>
        </div>
        <WizardFooter
          onBack={onBack}
          onContinue={toForm}
          continueDisabled={repo == null}
          hint={repo ? repo.fullName : "Pick a repository to continue"}
        />
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-y-auto p-6">
        <div className="mx-auto max-w-[520px] space-y-5">
          <div>
            <h3 className="font-mono text-xs">{repo?.fullName}</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              The image is built from this repository&apos;s Dockerfile on every deploy.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="git-service-name">Service name</Label>
            <Input
              id="git-service-name"
              value={serviceName}
              onChange={(e) => setServiceName(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="git-branch">Branch</Label>
              <Input id="git-branch" value={branch} onChange={(e) => setBranch(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="git-port">Port</Label>
              <Input
                id="git-port"
                inputMode="numeric"
                placeholder="3000"
                value={port}
                onChange={(e) => setPort(e.target.value)}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="git-dockerfile">Dockerfile path</Label>
              <Input
                id="git-dockerfile"
                value={dockerfilePath}
                onChange={(e) => setDockerfilePath(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="git-context">Build context</Label>
              <Input
                id="git-context"
                value={buildContext}
                onChange={(e) => setBuildContext(e.target.value)}
              />
            </div>
          </div>
          <div className="flex items-center justify-between rounded-md border px-3 py-2.5">
            <div>
              <p className="text-sm font-medium">Expose publicly</p>
              <p className="text-xs text-muted-foreground">Route external traffic to this port.</p>
            </div>
            <Switch checked={exposePublic} onCheckedChange={setExposePublic} aria-label="Expose publicly" />
          </div>
        </div>
      </div>
      <WizardFooter
        onBack={() => setStep("pick")}
        onContinue={openInEditor}
        continueLabel="Open in editor"
        continueDisabled={serviceName.trim() === "" || branch.trim() === "" || !portValid}
        hint="You can add env vars and more services on the canvas"
      />
    </div>
  );
}
