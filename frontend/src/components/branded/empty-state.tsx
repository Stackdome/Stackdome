import { cn } from "@/lib/utils";

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: React.ReactNode;
  description?: React.ReactNode;
  action?: React.ReactNode;
  dashed?: boolean;
  className?: string;
}

export function EmptyState({ icon, title, description, action, dashed = true, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-2 rounded-md py-14 px-6 text-center min-h-[260px]",
        dashed ? "border border-dashed border-border" : "border border-border",
        className,
      )}
    >
      {icon && <div className="text-muted-foreground/60 mb-1">{icon}</div>}
      <div className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
        {title}
      </div>
      {description && (
        <p className="text-xs text-muted-foreground/80">{description}</p>
      )}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}
