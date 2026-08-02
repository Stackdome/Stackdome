import type { ReactNode } from "react";
import { AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AlertBannerProps {
  children: ReactNode;
  action?: { label: string; onClick: () => void; disabled?: boolean };
  className?: string;
}

export function AlertBanner({ children, action, className }: AlertBannerProps) {
  return (
    <div
      role="alert"
      className={cn(
        "flex items-center gap-3 rounded-[14px] border border-danger-border bg-danger-bg px-4 py-3",
        className,
      )}
    >
      <AlertCircle
        aria-hidden="true"
        className="h-[18px] w-[18px] flex-none text-danger"
      />
      <div className="flex-1 text-[13.5px]">{children}</div>
      {action && (
        <button
          type="button"
          onClick={action.onClick}
          disabled={action.disabled}
          // Deliberate exception to the --ring focus convention: this action sits
          // inside a danger-tinted banner, so its outline matches --danger instead.
          className="flex-none whitespace-nowrap text-[13px] font-semibold text-danger hover:underline focus-visible:outline-2 focus-visible:outline-danger focus-visible:outline-offset-2 rounded disabled:opacity-50 disabled:pointer-events-none"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
