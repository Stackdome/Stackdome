import { cn } from "@/lib/utils";

export type StatusVariant = "ready" | "pending" | "error" | "info" | "neutral";

interface StatusPillProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: StatusVariant;
  withDot?: boolean;
  // Force the dot heartbeat on/off, overriding the variant default. Use `false`
  // for historical/snapshot data (e.g. past backup runs) where the row is not a
  // live, transitioning resource and motion would falsely imply activity.
  pulse?: boolean;
}

// Token-driven palette: each variant maps to design-system semantic tokens.
// `pulse` toggles the dot heartbeat — on for live/transitioning states, off
// for terminal (error) and inert (neutral) states where motion would mislead.
const styles: Record<StatusVariant, { wrap: string; dot: string; pulse: boolean }> = {
  ready: {
    wrap: "bg-success-bg border-success-border text-success",
    dot: "bg-success",
    pulse: true,
  },
  pending: {
    wrap: "bg-warn-bg border-warn-border text-warn",
    dot: "bg-warn",
    pulse: true,
  },
  error: {
    wrap: "bg-danger-bg border-danger-border text-danger",
    dot: "bg-danger",
    pulse: false,
  },
  info: {
    wrap: "bg-info-bg border-info-border text-info",
    dot: "bg-info",
    pulse: true,
  },
  neutral: {
    wrap: "bg-fg-muted/15 border-fg-muted/35 text-fg-muted",
    dot: "bg-fg-muted",
    pulse: false,
  },
};

export function variantFromState(state?: string | null): StatusVariant {
  const s = (state ?? "").toLowerCase();
  if (["ready", "running", "active", "succeeded", "exposed", "released"].includes(s)) return "ready";
  if (["pending", "deploying", "creating", "updating", "provisioning", "inprogress"].includes(s)) return "pending";
  if (["error", "failed", "crash", "crashloopbackoff", "unhealthy"].includes(s)) return "error";
  if (["superseded", "cancelled"].includes(s)) return "neutral";
  if (!s) return "neutral";
  return "info";
}

export function StatusPill({
  variant = "neutral",
  withDot = true,
  pulse,
  className,
  children,
  ...props
}: StatusPillProps) {
  const s = styles[variant];
  const animate = pulse ?? s.pulse;
  return (
    <span
      className={cn(
        // Pill: full-rounded, 1px border, mono uppercase bold per design system
        "inline-flex items-center gap-[7px] rounded-full border px-2.5 py-1 leading-none",
        "font-mono text-[11px] font-bold uppercase tracking-[0.08em]",
        s.wrap,
        className,
      )}
      {...props}
    >
      {withDot && (
        <span
          className={cn("inline-block h-[7px] w-[7px] rounded-full", s.dot, animate && "animate-pulse")}
        />
      )}
      {children}
    </span>
  );
}
