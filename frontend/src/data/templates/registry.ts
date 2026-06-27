import type { Template } from "./types";
import { tooljet } from "./tooljet/template";
import { n8n } from "./n8n/template";

/** Curated templates shown in the Templates Browser. Add a template by dropping a folder and listing it here. */
export const templates: Template[] = [tooljet, n8n];

export function getTemplateById(id: string): Template | undefined {
  return templates.find((t) => t.id === id);
}
