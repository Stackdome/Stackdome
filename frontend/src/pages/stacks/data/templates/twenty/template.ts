import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const twenty: Template = {
  id: "twenty",
  name: "Twenty",
  initials: "Tw",
  icon,
  category: "CRM",
  shortDescription: "Open-source CRM to manage companies, people, and deals.",
  longDescription:
    "A modern, open-source alternative to Salesforce. Track companies, contacts, and opportunities in a fully customizable workspace, with a worker process handling background jobs like marketplace app syncs.",
  website: "https://twenty.com/",
  docs: "https://docs.twenty.com/",
  version: "v2.26.0",
  stackYaml,
};
