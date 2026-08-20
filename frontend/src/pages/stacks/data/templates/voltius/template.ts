import type { Template } from "../types";
import icon from "./icon.png";
import stackYaml from "./stack.yaml?raw";

export const voltius: Template = {
  id: "voltius",
  name: "Voltius",
  initials: "Vo",
  icon,
  category: "Finance",
  shortDescription: "Self-hosted budget and expense sync server.",
  longDescription:
    "Voltius is an AGPLv3-licensed budget and expense tracker with a Rust/Axum API server, syncing across the desktop and mobile apps against your own Postgres database.",
  website: "https://voltius.app/",
  docs: "https://docs.voltius.app/self-hosting/",
  version: "latest",
  stackYaml,
};
