import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const openclaw: Template = {
  id: "openclaw",
  name: "OpenClaw",
  initials: "OC",
  icon,
  category: "AI Assistant",
  shortDescription: "Self-hosted gateway for your personal AI assistant.",
  longDescription:
    "OpenClaw is an open-source, self-hosted gateway that connects your chat apps — Discord, Slack, Telegram, WhatsApp, and more — to AI agents, giving you an always-available personal assistant you fully control.",
  website: "https://openclaw.ai/",
  docs: "https://docs.openclaw.ai/",
  version: "2026.6.10",
  stackYaml,
};
