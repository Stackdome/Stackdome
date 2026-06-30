import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Layers } from "lucide-react";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { templates } from "@/data/templates/registry";
import { useTemplateImport } from "@/pages/stacks/hooks/use-template-import";
import { useDockerComposeImport } from "@/pages/stacks/hooks/use-docker-compose-import";
import { WizardChooser } from "./wizard-chooser";
import { BlockComposer } from "./block-composer";
import { TemplatesBrowserPanel } from "./templates-browser-panel";
import { DockerComposeImportPanel } from "./docker-compose-import-panel";
import type { Template } from "@/data/templates/types";

type Phase = "chooser" | "composer" | "template" | "compose";

interface StackCreateWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function StackCreateWizard({ open, onOpenChange }: StackCreateWizardProps) {
  const navigate = useNavigate();
  const [phase, setPhase] = useState<Phase>("chooser");
  const tpl = useTemplateImport();
  const compose = useDockerComposeImport();

  const close = () => {
    onOpenChange(false);
    setPhase("chooser"); // reset for next open
  };

  const onUseTemplate = (t: Template) => {
    tpl.useTemplate(t);
    close();
  };

  const onImportCompose = async (yaml: string) => {
    await compose.handleImport(yaml);
    close();
  };

  const wide = phase === "composer" || phase === "template";

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent
        className={`block gap-0 overflow-hidden p-0 ${wide ? "sm:max-w-[1000px]" : "sm:max-w-[640px]"}`}
      >
        <DialogTitle className="sr-only">New Stack</DialogTitle>
        <div className="flex items-center gap-3 border-b px-5 py-3.5">
          <span className="flex h-6 w-6 items-center justify-center text-primary">
            <Layers className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            New Stack
          </span>
        </div>

        <div className="max-h-[80vh] overflow-y-auto">
          {phase === "chooser" && (
            <WizardChooser
              onPickBlocks={() => setPhase("composer")}
              onPickTemplate={() => setPhase("template")}
              onPickCompose={() => setPhase("compose")}
              onPickBlank={() => {
                navigate("/stacks/create");
                close();
              }}
            />
          )}
          {phase === "composer" && (
            <BlockComposer onBack={() => setPhase("chooser")} onClose={close} />
          )}
          {phase === "template" && (
            <div className="h-[70vh]">
              <TemplatesBrowserPanel templates={templates} onUse={onUseTemplate} />
            </div>
          )}
          {phase === "compose" && (
            <DockerComposeImportPanel
              onImport={onImportCompose}
              isLoading={compose.isLoading}
              error={compose.error}
              onClearError={compose.clearError}
              onCancel={() => setPhase("chooser")}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
