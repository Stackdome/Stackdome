import { cn } from "@/lib/utils";
import { EyebrowLabel } from "./eyebrow-label";

interface PanelProps {
  title?: React.ReactNode;
  count?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
  bodyClassName?: string;
  bare?: boolean;
  /**
   * "default" — mono caps eyebrow header (legacy).
   * "soft" — sans, sentence-case header for the stack-detail redesign.
   */
  tone?: "default" | "soft";
  children: React.ReactNode;
}

/**
 * Panel composes a section: optional eyebrow header (title + count + amber action link)
 * with a hairline divider above the body content.
 */
export function Panel({ title, count, action, className, bodyClassName, bare, tone = "default", children }: PanelProps) {
  return (
    <section
      className={cn(
        !bare && "rounded-lg border border-border bg-card",
        className,
      )}
    >
      {(title || action) && (
        <header className={cn(
          "flex items-center justify-between gap-4 px-5 py-3 border-b border-border",
          bare && "px-0",
        )}>
          {tone === "soft" ? (
            <span className="text-[13px] font-semibold text-foreground">
              {title}
              {count !== undefined && count !== null && (
                <span className="ml-2 text-[13px] font-normal text-muted-foreground">· {count}</span>
              )}
            </span>
          ) : (
            <EyebrowLabel tone="muted">
              {title}
              {count !== undefined && count !== null && (
                <span className="ml-2 text-muted-foreground/60">· {count}</span>
              )}
            </EyebrowLabel>
          )}
          {action && (
            tone === "soft" ? (
              <span className="text-[12.5px] text-brand hover:text-brand-press transition-colors">
                {action}
              </span>
            ) : (
              <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-brand hover:text-brand-darker transition-colors">
                {action}
              </span>
            )
          )}
        </header>
      )}
      <div className={cn(!bare && "px-5 py-4", bodyClassName)}>{children}</div>
    </section>
  );
}
