import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { templateToFormData } from "@/data/templates/template-to-form";
import type { Template } from "@/data/templates/types";

export interface TemplateImportState {
  isDialogOpen: boolean;
}

export interface TemplateImportActions {
  openDialog: () => void;
  closeDialog: () => void;
  /** Convert the template's preset and open the create form, prefilled. */
  useTemplate: (template: Template) => void;
}

export function useTemplateImport(): TemplateImportState & TemplateImportActions {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const navigate = useNavigate();

  const openDialog = () => setIsDialogOpen(true);
  const closeDialog = () => setIsDialogOpen(false);

  const useTemplate = (template: Template) => {
    const { data } = templateToFormData(template);
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
  };

  return { isDialogOpen, openDialog, closeDialog, useTemplate };
}
