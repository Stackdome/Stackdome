import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const gitea: Template = {
  id: "gitea",
  name: "Gitea",
  initials: "Gt",
  icon,
  category: "Developer Tools",
  shortDescription: "Self-hosted Git service, lightweight and fast.",
  longDescription:
    "Gitea is a painless, self-hosted, all-in-one software development service — Git hosting, code review, issue tracking, CI/CD, packages, and wikis. A lightweight, community-managed alternative to GitHub.",
  website: "https://about.gitea.com/",
  docs: "https://docs.gitea.com/",
  version: "1.26.4",
  stackYaml,
};
