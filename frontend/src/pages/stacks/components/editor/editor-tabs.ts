export const EDITOR_TABS = {
  architecture: "architecture",
  deployments: "deployments",
  logs: "logs",
  metrics: "metrics",
} as const;

export type EditorTabId = (typeof EDITOR_TABS)[keyof typeof EDITOR_TABS];
