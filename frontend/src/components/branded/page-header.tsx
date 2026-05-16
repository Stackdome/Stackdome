import { cn } from "@/lib/utils";
import { EyebrowLabel } from "./eyebrow-label";

interface PageHeaderProps {
  eyebrow?: React.ReactNode;
  title: React.ReactNode;
  status?: React.ReactNode;
  subtitle?: React.ReactNode;
  actions?: React.ReactNode;
  /** Vertical alignment of the actions block. Default "start" (top). */
  actionsAlign?: "start" | "center";
  className?: string;
}

export function PageHeader({ eyebrow, title, status, subtitle, actions, actionsAlign = "start", className }: PageHeaderProps) {
  return (
    <div className={cn("flex justify-between gap-6 pb-6 border-b border-border", actionsAlign === "center" ? "items-center" : "items-start", className)}>
      <div className="min-w-0 flex-1">
        {eyebrow && <EyebrowLabel className="mb-2 block">{eyebrow}</EyebrowLabel>}
        <div className="flex items-center gap-3">
          <h1 className="text-3xl font-medium tracking-tight">{title}</h1>
          {status}
        </div>
        {subtitle && (
          <p className="mt-2 font-mono text-xs text-muted-foreground tracking-[0.5px]">{subtitle}</p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  );
}
