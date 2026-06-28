import type { Template } from "./types";
import { tooljet } from "./tooljet/template";
import { n8n } from "./n8n/template";
import { openclaw } from "./openclaw/template";
import { grafana } from "./grafana/template";
import { immich } from "./immich/template";
import { prometheus } from "./prometheus/template";
import { gitea } from "./gitea/template";

/** Curated templates shown in the Templates Browser. Add a template by dropping a folder and listing it here. */
export const templates: Template[] = [
  tooljet,
  n8n,
  openclaw,
  grafana,
  immich,
  prometheus,
  gitea,
];

export function getTemplateById(id: string): Template | undefined {
  return templates.find((t) => t.id === id);
}
