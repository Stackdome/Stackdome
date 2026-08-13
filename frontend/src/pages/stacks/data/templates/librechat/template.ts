import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const librechat: Template = {
  id: "librechat",
  name: "LibreChat",
  initials: "Lc",
  icon,
  category: "AI Assistant",
  shortDescription: "Self-hosted, multi-provider AI chat UI.",
  longDescription:
    "LibreChat is an open-source, ChatGPT-style web UI that talks to OpenAI, Anthropic, Google, and many other AI providers from one place. It supports conversation history, multi-user accounts, and custom endpoints, all backed by MongoDB.",
  website: "https://www.librechat.ai/",
  docs: "https://www.librechat.ai/docs",
  version: "v0.8.7",
  stackYaml,
};
