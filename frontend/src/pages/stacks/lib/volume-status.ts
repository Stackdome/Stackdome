import { statusVariant, type StatusVariant } from "@/components/branded/status-variant";

const DOT: Record<StatusVariant, string> = {
  ready: "bg-success",
  pending: "bg-warn",
  error: "bg-danger",
  info: "bg-info",
  neutral: "bg-fg-muted",
};

/** Single source for the volume status dot color (list item + detail). */
export function volumeDotClass(phase?: string | null): string {
  return DOT[statusVariant("volume", phase)];
}
