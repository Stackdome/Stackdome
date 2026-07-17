import type { ReactNode } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { EyebrowLabel, FieldError } from "@/components/branded";

interface LedgerSectionProps {
  /** Mono uppercase section marker, e.g. "GENERAL". */
  label: string;
  /** Right-aligned mono meta on the header rule, e.g. "1 exposed". */
  meta?: ReactNode;
  defaultOpen?: boolean;
  children: ReactNode;
  className?: string;
}

/**
 * Foldable spec-ledger section: mono marker on a hairline rule with an
 * optional right meta and a rotating chevron. Click the header to collapse.
 */
export function LedgerSection({
  label,
  meta,
  defaultOpen = true,
  children,
  className,
}: LedgerSectionProps) {
  return (
    <Collapsible defaultOpen={defaultOpen} className={className}>
      <CollapsibleTrigger className="group flex w-full cursor-pointer items-center gap-3 rounded px-1.5 pb-3 pt-5 transition-colors hover:bg-muted/30">
        <EyebrowLabel tone="muted">{label}</EyebrowLabel>
        <span className="h-px flex-1 bg-border/70" aria-hidden />
        {meta && (
          <span className="shrink-0 font-mono text-[10.5px] text-fg-muted/80">{meta}</span>
        )}
        <ChevronDown
          className="size-[15px] shrink-0 text-fg-muted transition-transform group-data-[state=open]:rotate-180"
          aria-hidden
        />
      </CollapsibleTrigger>
      <CollapsibleContent>{children}</CollapsibleContent>
    </Collapsible>
  );
}

interface LedgerDisclosureProps {
  /** Row-level label, e.g. "Advanced". */
  label: ReactNode;
  /** Mono hint shown in the control column, e.g. "push registry". */
  meta?: ReactNode;
  defaultOpen?: boolean;
  children: ReactNode;
  className?: string;
}

/**
 * Inline row-level disclosure for advanced fields: renders as a ledger row
 * whose label carries a rotating chevron. Lighter than LedgerSection — no
 * section marker, no hairline rule.
 */
export function LedgerDisclosure({
  label,
  meta,
  defaultOpen = false,
  children,
  className,
}: LedgerDisclosureProps) {
  return (
    <Collapsible defaultOpen={defaultOpen} className={className}>
      <CollapsibleTrigger className="group w-full cursor-pointer border-b border-secondary/80 py-1 text-left">
        <span className="flex w-full items-center gap-4 rounded-md px-1.5 py-2 transition-colors group-hover:bg-muted/20">
          <span className="flex w-[150px] shrink-0 items-center gap-1.5 text-[13px] text-fg-muted">
            <ChevronRight
              className="size-3.5 shrink-0 transition-transform group-data-[state=open]:rotate-90"
              aria-hidden
            />
            {label}
          </span>
          {meta && (
            <span className="font-mono text-[10.5px] text-fg-muted/70">{meta}</span>
          )}
        </span>
      </CollapsibleTrigger>
      <CollapsibleContent>{children}</CollapsibleContent>
    </Collapsible>
  );
}

interface LedgerRowProps {
  /** Left label column (fixed width so rows scan like a spec sheet). */
  label: ReactNode;
  htmlFor?: string;
  required?: boolean;
  /** Right-aligned mono hint at the row's far edge, e.g. "started first". */
  meta?: ReactNode;
  /** Helper copy under the control (use sparingly — meta is the default). */
  hint?: ReactNode;
  error?: ReactNode;
  /** Top-align the label for controls that grow (error text, textareas). */
  alignTop?: boolean;
  children: ReactNode;
  className?: string;
}

/**
 * Label-left ledger row on a hairline: 150px label column, control fills,
 * optional far-right mono meta. Hint/error render under the control.
 */
export function LedgerRow({
  label,
  htmlFor,
  required,
  meta,
  hint,
  error,
  alignTop = false,
  children,
  className,
}: LedgerRowProps) {
  return (
    <div className={cn("border-b border-secondary/80 py-1", className)}>
      <div
        className={cn(
          "flex gap-4 rounded-md px-1.5 py-1.5 transition-colors hover:bg-muted/20",
          alignTop ? "items-start" : "items-center",
        )}
      >
        <Label
          htmlFor={htmlFor}
          className={cn(
            "w-[150px] shrink-0 text-[13px] font-normal text-fg-2",
            alignTop && "pt-2.5",
          )}
        >
          <span>
            {label}
            {required && (
              <span className="ml-0.5 font-semibold text-brand/80" aria-hidden>
                *
              </span>
            )}
          </span>
        </Label>
        <div className="min-w-0 flex-1">
          {children}
          {hint && <p className="mt-1.5 text-[12px] leading-relaxed text-muted-foreground">{hint}</p>}
          <FieldError>{error}</FieldError>
        </div>
        {meta && (
          <span
            className={cn(
              "shrink-0 font-mono text-[10.5px] text-fg-muted/70",
              alignTop && "pt-3",
            )}
          >
            {meta}
          </span>
        )}
      </div>
    </div>
  );
}

interface LedgerSegmentedOption<T extends string> {
  value: T;
  label: ReactNode;
  icon?: ReactNode;
}

interface LedgerSegmentedProps<T extends string> {
  value: T;
  onValueChange: (value: T) => void;
  options: LedgerSegmentedOption<T>[];
  "aria-label"?: string;
}

/** Two-to-three-way segmented switch; the active cell earns the amber. */
export function LedgerSegmented<T extends string>({
  value,
  onValueChange,
  options,
  "aria-label": ariaLabel,
}: LedgerSegmentedProps<T>) {
  return (
    <div
      role="radiogroup"
      aria-label={ariaLabel}
      className="flex h-9 w-full overflow-hidden rounded border border-border bg-background"
    >
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          role="radio"
          aria-checked={value === option.value}
          onClick={() => onValueChange(option.value)}
          className={cn(
            "flex flex-1 cursor-pointer items-center justify-center gap-2 border-l border-border text-[13px] transition-colors first:border-l-0",
            value === option.value
              ? "bg-brand-bg text-brand"
              : "text-fg-muted hover:text-fg-2",
          )}
        >
          {option.icon}
          {option.label}
        </button>
      ))}
    </div>
  );
}
