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
    "Build enterprise apps, AI agents, and workflows in minutes, not months. Just describe what you need in natural language.",
  website: "https://tooljet.com/",
  docs: "https://docs.tooljet.com/",
  version: "ee-lts-latest",
  stackYaml,
};
