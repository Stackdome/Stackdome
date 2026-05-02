import { cn } from "@/lib/utils";

export type StatusVariant = "ready" | "pending" | "error" | "info" | "neutral";

interface StatusPillProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: StatusVariant;
  withDot?: boolean;
}

// Token-driven palette: each variant maps to design-system semantic tokens.
const styles: Record<StatusVariant, { wrap: string; dot: string }> = {
  ready: {
    wrap: "bg-success-bg border-success-border text-success",
    dot: "bg-success",
  },
  pending: {
    wrap: "bg-warn-bg border-warn-border text-warn",
    dot: "bg-warn",
  },
  error: {
    wrap: "bg-danger-bg border-danger-border text-danger",
    dot: "bg-danger",
  },
  info: {
    wrap: "bg-info-bg border-info-border text-info",
    dot: "bg-info",
  },
  neutral: {
    wrap: "bg-fg-muted/15 border-fg-muted/35 text-fg-muted",
    dot: "bg-fg-muted",
  },
};

export function variantFromState(state?: string | null): StatusVariant {
  const s = (state ?? "").toLowerCase();
  if (["ready", "running", "active", "succeeded", "exposed"].includes(s)) return "ready";
  if (["pending", "deploying", "creating", "updating", "provisioning"].includes(s)) return "pending";
  if (["error", "failed", "crash", "crashloopbackoff", "unhealthy"].includes(s)) return "error";
  if (!s) return "neutral";
  return "info";
}

export function StatusPill({
  variant = "neutral",
  withDot = true,
  className,
  children,
  ...props
}: StatusPillProps) {
  const s = styles[variant];
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
      {withDot && <span className={cn("inline-block h-[7px] w-[7px] rounded-full", s.dot)} />}
      {children}
    </span>
  );
}
