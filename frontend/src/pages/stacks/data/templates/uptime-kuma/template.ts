import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const uptimeKuma: Template = {
  id: "uptime-kuma",
  name: "Uptime Kuma",
  initials: "Uk",
  icon,
  category: "Monitoring",
  shortDescription: "Self-hosted uptime and status monitoring.",
  longDescription:
    "Uptime Kuma is a fancy, self-hosted monitoring tool that keeps watch over your websites, APIs, and services. Get notified the moment something goes down, track response times and uptime history, and publish a public status page — all from a single lightweight container.",
  website: "https://github.com/louislam/uptime-kuma",
  docs: "https://github.com/louislam/uptime-kuma/wiki",
  version: "2.5.0",
  stackYaml,
};
