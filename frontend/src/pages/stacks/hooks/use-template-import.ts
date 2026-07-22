import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/components/ui/use-toast";
import { templateToFormData } from "@/data/templates/template-to-form";
import { formatImportWarnings } from "@/pages/stacks/lib/import-warnings-toast";
import type { Template } from "@/data/templates/types";

export interface TemplateImportState {
  isDialogOpen: boolean;
}

export interface TemplateImportActions {
  openDialog: () => void;
  closeDialog: () => void;
  /** Convert the template's preset and navigate to the canvas editor pre-seeded with the template data. */
  useTemplate: (template: Template) => void;
}

export function useTemplateImport(): TemplateImportState & TemplateImportActions {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const navigate = useNavigate();
  const { toast } = useToast();

  const openDialog = () => setIsDialogOpen(true);
  const closeDialog = () => setIsDialogOpen(false);

  const useTemplate = (template: Template) => {
    const { data, warnings } = templateToFormData(template);
    setIsDialogOpen(false);
    navigate("/stacks/new", {
      state: {
        seed: {
          name: data.name ?? "",
          labels: data.labels ?? [],
          resources: data.spec?.stack_resources ?? [],
          volumes: data.spec?.volumes ?? [],
          linkedAddonIds: [],
        },
      },
    });
    if (warnings.length > 0) {
      const { count, description } = formatImportWarnings(warnings.map((w) => w.message));
      toast({
        title: `Template imported with ${count} warning${count === 1 ? "" : "s"}`,
        description,
        variant: "warning",
        // Give the reader time — these carry follow-up actions.
        duration: 15000,
      });
    }
  };

  return { isDialogOpen, openDialog, closeDialog, useTemplate };
}
