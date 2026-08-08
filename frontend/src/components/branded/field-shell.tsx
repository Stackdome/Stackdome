import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import { FieldError } from "./field-error";

interface FieldShellProps {
  /** Field label rendered above the input. */
  label: ReactNode;
  /** Forwarded to the underlying <Label> for input association. */
  htmlFor?: string;
  /** When true, appends a subtle mono-caps "required" tag after the label. */
  required?: boolean;
  /** Helper copy rendered below the input — replaces tooltip-on-icon for permanent affordance. */
  hint?: ReactNode;
  /** Validation error rendered below the hint. */
  error?: ReactNode;
  /** Wraps the input(s). */
  children: ReactNode;
  className?: string;
}

/**
 * Form field wrapper aligning every input to the Stackdome label/hint/error rhythm.
 * Use in place of bespoke <Label> + <Tooltip Info /> + manual <p text-danger> blocks.
 */
export function FieldShell({
  label,
  htmlFor,
  required,
  hint,
  error,
  children,
  className,
}: FieldShellProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      <Label
        htmlFor={htmlFor}
        className="text-body font-medium text-foreground"
      >
        <span>
          {label}
          {required && (
            <span className="ml-0.5 text-name font-semibold text-foreground/70 leading-none" aria-hidden>*</span>
          )}
        </span>
      </Label>
      {children}
      {hint && (
        <p className="text-meta text-muted-foreground leading-relaxed">
          {hint}
        </p>
      )}
      <FieldError>{error}</FieldError>
    </div>
  );
}
