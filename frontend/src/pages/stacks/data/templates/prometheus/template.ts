import type { Template } from "@/pages/stacks/data/templates/types";
import icon from "./icon.svg";
import stackYaml from "./stack.yaml?raw";

export const prometheus: Template = {
  id: "prometheus",
  name: "Prometheus",
  initials: "Pr",
  icon,
  category: "Observability",
  shortDescription: "Metrics collection and monitoring with PromQL.",
  longDescription:
    "Prometheus is an open-source systems monitoring and alerting toolkit. It scrapes and stores metrics as time series, queries them with the powerful PromQL language, and triggers alerts — with no external dependencies.",
  website: "https://prometheus.io/",
  docs: "https://prometheus.io/docs/",
  version: "v3.12.0",
  stackYaml,
};
