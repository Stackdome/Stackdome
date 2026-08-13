import type { ReactNode } from "react";
import { AlertCircle, Info, X } from "lucide-react";
import { cn } from "@/lib/utils";

export type AlertBannerVariant = "danger" | "notice";

export interface AlertBannerProps {
  children: ReactNode;
  variant?: AlertBannerVariant;
  action?: { label: string; onClick: () => void; disabled?: boolean };
  onDismiss?: () => void;
  className?: string;
}

const VARIANTS = {
  danger: {
    role: "alert",
    icon: AlertCircle,
    shell: "border-danger-border bg-danger-bg",
    accent: "text-danger",
    ring: "focus-visible:ring-danger",
  },
  notice: {
    role: "status",
    icon: Info,
    shell: "border-border bg-muted",
    accent: "text-muted-foreground",
    ring: "focus-visible:ring-brand",
  },
} as const;

export function AlertBanner({
  children,
  variant = "danger",
  action,
  onDismiss,
  className,
}: AlertBannerProps) {
  const { role, icon: Icon, shell, accent, ring } = VARIANTS[variant];

  return (
    <div
      role={role}
      className={cn(
        "flex items-center gap-3 rounded-lg border px-4 py-3",
        shell,
        className,
      )}
    >
      <Icon
        aria-hidden="true"
        className={cn("h-[18px] w-[18px] flex-none", accent)}
      />
      <div className="flex-1 text-[13.5px]">{children}</div>
      {action && (
        <button
          type="button"
          onClick={action.onClick}
          disabled={action.disabled}
          className={cn(
            "flex-none whitespace-nowrap text-[13px] font-semibold hover:underline focus-visible:outline-none focus-visible:ring-2 rounded disabled:opacity-50 disabled:pointer-events-none",
            accent,
            ring,
          )}
        >
          {action.label}
        </button>
      )}
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          aria-label="Dismiss"
          className={cn(
            "flex-none rounded p-1.5 opacity-80 transition-opacity hover:opacity-100 focus-visible:outline-none focus-visible:ring-2",
            accent,
            ring,
          )}
        >
          <X aria-hidden="true" className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  );
}
