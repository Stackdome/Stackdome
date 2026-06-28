import type { Template } from "../types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const immich: Template = {
  id: "immich",
  name: "Immich",
  initials: "Im",
  icon,
  category: "Photos",
  shortDescription: "Self-hosted photo and video backup.",
  longDescription:
    "Immich is a high-performance, self-hosted backup solution for your photos and videos — automatic mobile backup, smart search, face recognition, and shared albums, all on your own server.",
  website: "https://immich.app/",
  docs: "https://docs.immich.app/",
  version: "v2.7.5",
  stackYaml,
};
