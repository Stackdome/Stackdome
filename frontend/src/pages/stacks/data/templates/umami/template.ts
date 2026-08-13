import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const umami: Template = {
  id: "umami",
  name: "Umami",
  initials: "Um",
  icon,
  category: "Analytics",
  shortDescription: "Privacy-focused, open-source web analytics.",
  longDescription:
    "Umami is a simple, fast, privacy-focused alternative to Google Analytics. It gives you clear, easy-to-understand website traffic data without tracking or storing any personal information about visitors, and without cookie banners.",
  website: "https://umami.is/",
  docs: "https://docs.umami.is/",
  version: "3.2.0",
  stackYaml,
};
