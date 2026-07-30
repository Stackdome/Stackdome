import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const tooljet: Template = {
  id: "tooljet",
  name: "ToolJet",
  initials: "Tj",
  icon,
  category: "Dev Tools",
  shortDescription:
    "Low-code platform for building internal tools and dashboards.",
  longDescription:
    "Build enterprise apps, AI agents, and workflows in minutes, not months. Ships with workflow workers, the built-in ToolJet Database, and an observability stack (Grafana, Tempo, Prometheus) with a ready-made monitoring dashboard.",
  website: "https://tooljet.com/",
  docs: "https://docs.tooljet.com/",
  version: "v3.20.189-lts",
  stackYaml,
};
