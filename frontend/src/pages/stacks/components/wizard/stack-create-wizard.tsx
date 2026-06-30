import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Layers } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
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
    const ok = await compose.handleImport(yaml);
    if (ok) close();
  };

  const backToChooser = () => setPhase("chooser");

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      {/* Constant width across every step so the modal never resizes between
          phases — the composer/template steps need the wide two-column layout,
          so all steps use it and the body scrolls instead. */}
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[1000px]">
        <DialogTitle className="sr-only">New Stack</DialogTitle>
        <DialogDescription className="sr-only">Choose how to create your new stack</DialogDescription>
        {/* pr-12 reserves space so the title never sits under the dialog's
            built-in close (X), which is absolutely positioned at top-right.
            Per-step navigation lives in each path's footer (WizardFooter). */}
        <div className="flex items-center gap-3 border-b py-3.5 pl-5 pr-12">
          <span className="flex h-6 w-6 items-center justify-center text-primary">
            <Layers className="h-5 w-5" />
          </span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            New Stack
          </span>
        </div>

        {/* One stable frame for every step (fixed height, clamped on short
            screens) so the modal never grows/shrinks or re-centers between
            phases. Content scrolls inside; path panels fill it via flex. */}
        <div className="h-[620px] max-h-[82vh] overflow-hidden">
          {phase === "chooser" && (
            <div className="scrollbar-hide h-full overflow-y-auto">
              <WizardChooser
                onPickBlocks={() => setPhase("composer")}
                onPickTemplate={() => setPhase("template")}
                onPickCompose={() => setPhase("compose")}
                onPickBlank={() => {
                  navigate("/stacks/create");
                  close();
                }}
              />
            </div>
          )}
          {phase === "composer" && <BlockComposer onBack={backToChooser} onClose={close} />}
          {phase === "template" && (
            <TemplatesBrowserPanel templates={templates} onBack={backToChooser} onUse={onUseTemplate} />
          )}
          {phase === "compose" && (
            <DockerComposeImportPanel
              onImport={onImportCompose}
              isLoading={compose.isLoading}
              error={compose.error}
              onClearError={compose.clearError}
              onBack={backToChooser}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
