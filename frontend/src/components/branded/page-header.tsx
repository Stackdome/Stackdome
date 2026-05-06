import { cn } from "@/lib/utils";
import { EyebrowLabel } from "./eyebrow-label";

interface PageHeaderProps {
  eyebrow?: React.ReactNode;
  title: React.ReactNode;
  status?: React.ReactNode;
  subtitle?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
}

export function PageHeader({ eyebrow, title, status, subtitle, actions, className }: PageHeaderProps) {
  return (
    <div className={cn("flex items-start justify-between gap-6 pb-6 border-b border-border", className)}>
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
