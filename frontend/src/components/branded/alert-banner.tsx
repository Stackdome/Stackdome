import type { ReactNode } from "react";
import { AlertCircle, Info, TriangleAlert } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * `danger` — it failed. `blocking` — it will fail unless you deal with this
 * first. `info` — we did something to your input and you should know.
 *
 * A tone is one hue at three alphas (§7): the opaque hue for the glyph and the
 * action, ~10–12% for the fill, ~40–50% for the edge. The message itself stays
 * ink, so the banner says its severity once.
 */
export type AlertBannerTone = "danger" | "blocking" | "info";

const TONES = {
  danger: {
    icon: AlertCircle,
    box: "border-danger-border bg-danger-bg",
    ink: "text-danger",
    action: "text-danger focus-visible:outline-danger",
  },
  blocking: {
    icon: TriangleAlert,
    box: "border-warn-border bg-warn-bg",
    ink: "text-warn",
    action: "text-warn focus-visible:outline-warn",
  },
  info: {
    icon: Info,
    box: "border-info-border bg-info-bg",
    ink: "text-info",
    action: "text-info focus-visible:outline-info",
  },
} as const;

export interface AlertBannerProps {
  children: ReactNode;
  /** Defaults to `danger` — every caller that predates tones is a failure. */
  tone?: AlertBannerTone;
  action?: { label: string; onClick: () => void; disabled?: boolean };
  className?: string;
}

export function AlertBanner({ children, tone = "danger", action, className }: AlertBannerProps) {
  const { icon: Glyph, box, ink, action: actionInk } = TONES[tone];

  return (
    <div
      role="alert"
      className={cn("flex items-center gap-3 rounded-md border px-4 py-3", box, className)}
    >
      <Glyph aria-hidden="true" className={cn("h-[18px] w-[18px] flex-none", ink)} />
      <div className="flex-1 text-body">{children}</div>
      {action && (
        <button
          type="button"
          onClick={action.onClick}
          disabled={action.disabled}
          // Deliberate exception to the --ring focus convention: this action sits
          // inside a tinted banner, so its outline matches the tone instead.
          className={cn(
            "flex-none whitespace-nowrap rounded text-body font-semibold hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 disabled:pointer-events-none disabled:opacity-50",
            actionInk,
          )}
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
