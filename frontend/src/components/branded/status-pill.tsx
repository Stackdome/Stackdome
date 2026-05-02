import { cn } from "@/lib/utils";

export type StatusVariant = "ready" | "pending" | "error" | "info" | "neutral";

interface StatusPillProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: StatusVariant;
  withDot?: boolean;
}

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
    wrap: "bg-brand-bg border-brand-border text-brand",
    dot: "bg-brand",
  },
  neutral: {
    wrap: "bg-muted border-border text-muted-foreground",
    dot: "bg-muted-foreground",
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
        "inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5",
        "font-mono text-[10px] font-medium uppercase tracking-[1px]",
        s.wrap,
        className,
      )}
      {...props}
    >
      {withDot && <span className={cn("inline-block h-1.5 w-1.5 rounded-full", s.dot)} />}
      {children}
    </span>
  );
}
