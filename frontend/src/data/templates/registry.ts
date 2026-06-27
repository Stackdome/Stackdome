import type { Template } from "./types";
import { tooljet } from "./tooljet/template";

/** Curated templates shown in the Templates Browser. Add a template by dropping a folder and listing it here. */
export const templates: Template[] = [tooljet];

export function getTemplateById(id: string): Template | undefined {
  return templates.find((t) => t.id === id);
}
