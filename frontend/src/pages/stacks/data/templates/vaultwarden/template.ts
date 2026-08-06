import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const vaultwarden: Template = {
  id: "vaultwarden",
  name: "Vaultwarden",
  initials: "Vw",
  icon,
  category: "Security",
  shortDescription: "Self-hosted Bitwarden-compatible password manager.",
  longDescription:
    "Vaultwarden is a lightweight, unofficial implementation of the Bitwarden server API written in Rust. It's fully compatible with the official Bitwarden clients and browser extensions, letting you self-host your password vault without the resource overhead of the official server.",
  website: "https://github.com/dani-garcia/vaultwarden",
  docs: "https://github.com/dani-garcia/vaultwarden/wiki",
  version: "1.37.1",
  stackYaml,
};
