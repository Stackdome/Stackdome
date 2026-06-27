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
    "A drag-and-drop builder for internal tools — connect to your databases and APIs and ship admin panels, dashboards and CRUD apps without a frontend team.",
  website: "https://tooljet.com",
  docs: "https://docs.tooljet.com",
  version: "ee-lts-latest",
  stackYaml,
};
