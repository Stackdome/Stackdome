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
    "A source-available workflow automation tool — connect 400+ apps and APIs with a node-based editor and run every automation on your own infrastructure.",
  website: "https://n8n.io/",
  docs: "https://docs.n8n.io/",
  version: "2.27.4",
  stackYaml,
};
