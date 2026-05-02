import { cn } from "@/lib/utils";

export type StatusVariant = "ready" | "pending" | "error" | "info" | "neutral";

interface StatusPillProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: StatusVariant;
  withDot?: boolean;
}

// Design-system status palette: tinted bg + saturated border + bright readable text.
// Hex values map to colors_and_type.css and preview/components-badges.html.
const styles: Record<StatusVariant, { wrap: string; dot: string }> = {
  ready: {
    wrap: "bg-[rgb(34_197_94_/_0.16)] border-[rgb(34_197_94_/_0.50)] text-[#22c55e] dark:text-[#86efac]",
    dot: "bg-[#22c55e]",
  },
  pending: {
    wrap: "bg-[rgb(234_179_8_/_0.16)] border-[rgb(234_179_8_/_0.55)] text-[#a16207] dark:text-[#fde047]",
    dot: "bg-[#eab308]",
  },
  error: {
    wrap: "bg-[rgb(220_38_38_/_0.18)] border-[rgb(220_38_38_/_0.65)] text-[#b91c1c] dark:text-[#fca5a5]",
    dot: "bg-[#dc2626]",
  },
  info: {
    wrap: "bg-[rgb(249_115_22_/_0.14)] border-[rgb(249_115_22_/_0.55)] text-[#c2410c] dark:text-[#fdba74]",
    dot: "bg-[#f97316]",
  },
  neutral: {
    wrap: "bg-[rgb(148_163_184_/_0.12)] border-[rgb(148_163_184_/_0.35)] text-[#475569] dark:text-[#cbd5e1]",
    dot: "bg-[#94a3b8]",
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
