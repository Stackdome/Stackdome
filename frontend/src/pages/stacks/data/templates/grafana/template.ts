import type { Template } from "@/pages/stacks/data/templates/types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const grafana: Template = {
  id: "grafana",
  name: "Grafana",
  initials: "Gf",
  icon,
  category: "Observability",
  shortDescription: "Dashboards and visualization for all your metrics.",
  longDescription:
    "Grafana is open-source visualization and analytics software. Query, visualize, alert on, and explore your metrics, logs, and traces wherever they're stored — no matter the underlying data source.",
  website: "https://grafana.com/",
  docs: "https://grafana.com/docs/grafana/latest/",
  version: "13.0.2",
  stackYaml,
};
