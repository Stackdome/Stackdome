import type { Template } from "@/pages/stacks/data/templates/types";
import { tooljet } from "@/pages/stacks/data/templates/tooljet/template";
import { n8n } from "@/pages/stacks/data/templates/n8n/template";
import { openclaw } from "@/pages/stacks/data/templates/openclaw/template";
import { grafana } from "@/pages/stacks/data/templates/grafana/template";
import { immich } from "@/pages/stacks/data/templates/immich/template";
import { prometheus } from "@/pages/stacks/data/templates/prometheus/template";
import { gitea } from "@/pages/stacks/data/templates/gitea/template";
import { vaultwarden } from "@/pages/stacks/data/templates/vaultwarden/template";
import { uptimeKuma } from "@/pages/stacks/data/templates/uptime-kuma/template";
import { librechat } from "@/pages/stacks/data/templates/librechat/template";
import { twenty } from "@/pages/stacks/data/templates/twenty/template";
import { umami } from "@/pages/stacks/data/templates/umami/template";

/** Curated templates shown in the Templates Browser. Add a template by dropping a folder and listing it here. */
export const templates: Template[] = [
  tooljet,
  n8n,
  openclaw,
  grafana,
  immich,
  prometheus,
  gitea,
  vaultwarden,
  uptimeKuma,
  librechat,
  twenty,
  umami,
];

export function getTemplateById(id: string): Template | undefined {
  return templates.find((t) => t.id === id);
}
