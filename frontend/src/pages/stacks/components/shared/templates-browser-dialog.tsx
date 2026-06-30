import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { TemplatesBrowserPanel } from "@/pages/stacks/components/wizard/templates-browser-panel";
import type { Template } from "@/data/templates/types";

interface TemplatesBrowserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  templates: Template[];
  onUse: (template: Template) => void;
}

export default function TemplatesBrowserDialog({
  open,
  onOpenChange,
  templates,
  onUse,
}: TemplatesBrowserDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[1040px]">
        <DialogTitle className="sr-only">Choose a template</DialogTitle>
        <DialogDescription className="sr-only">
          Pick a curated template to prefill the new-stack form.
        </DialogDescription>
        <TemplatesBrowserPanel templates={templates} onUse={onUse} />
      </DialogContent>
    </Dialog>
  );
}
