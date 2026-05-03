import { cn } from "@/lib/utils";
import type { ButtonHTMLAttributes, HTMLAttributes } from "react";

/**
 * Local soft-label primitives, scoped to the stack-detail page redesign.
 * The global EyebrowLabel keeps its mono-caps styling for sidebar / auth /
 * other panels — these soft variants render sans + sentence-case.
 */

export function SectionLabel({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "text-[13px] font-semibold text-foreground tracking-normal",
        className,
      )}
      {...props}
    />
  );
}

export function SectionCount({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn("text-[13px] font-normal text-muted-foreground", className)}
      {...props}
    />
  );
}

export function SectionAction({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      className={cn(
        "text-[12.5px] text-brand hover:text-brand-press transition-colors",
        className,
      )}
      {...props}
    />
  );
}

export function SoftLabel({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "text-[11.5px] font-medium text-muted-foreground tracking-normal",
        className,
      )}
      {...props}
    />
  );
}
