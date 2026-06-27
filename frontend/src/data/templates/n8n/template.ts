import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const n8n: Template = {
  id: "n8n",
  name: "n8n",
  initials: "n8",
  icon,
  category: "Workflow Automation",
  shortDescription: "Workflow automation you self-host and fully own.",
  longDescription:
    "n8n is a fair-code licensed workflow automation tool that combines AI capabilities with business process automation.",
  website: "https://n8n.io/",
  docs: "https://docs.n8n.io/",
  version: "2.27.4",
  stackYaml,
};
