import { AlertTriangle, ChevronRight, X } from "lucide-react";
import { cn } from "@/lib/utils";
import type { EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";

export interface ValidationBannerItem {
  /** Human label for where the error lives, e.g. "api" or "Stack settings". */
  label: string;
  /** Backend message, shown verbatim. */
  message: string;
  /** Resource index to open on click; absent for stack-level errors. */
  resourceIndex?: number;
  /** Resource-drawer tab holding the offending field (defaults to configuration). */
  tab?: EditSessionTab;
}

interface ValidationBannerProps {
  items: ValidationBannerItem[];
  onJump?: (resourceIndex: number, tab: EditSessionTab) => void;
  onDismiss: () => void;
}

export function ValidationBanner({ items, onJump, onDismiss }: ValidationBannerProps) {
  if (items.length === 0) return null;

  return (
    <div className="mb-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3.5 py-3 text-destructive">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <AlertTriangle className="size-4 shrink-0" />
          <span className="text-sm font-semibold">
            {items.length} validation {items.length === 1 ? "error" : "errors"}
          </span>
        </div>
        <button
          type="button"
          aria-label="Dismiss validation errors"
          onClick={onDismiss}
          className="shrink-0 rounded p-0.5 text-destructive/60 transition-colors hover:bg-destructive/10 hover:text-destructive"
        >
          <X className="size-4" />
        </button>
      </div>
      <ul className="mt-2 flex flex-col gap-0.5">
        {items.map((item, i) => {
          const jumpable = item.resourceIndex !== undefined && onJump !== undefined;
          return (
            <li key={i}>
              <button
                type="button"
                disabled={!jumpable}
                onClick={jumpable ? () => onJump!(item.resourceIndex!, item.tab ?? "configuration") : undefined}
                className={cn(
                  "group flex w-full items-start gap-1.5 rounded px-1.5 py-1 text-left text-sm leading-snug text-destructive/90 transition-colors",
                  jumpable ? "cursor-pointer hover:bg-destructive/10 hover:text-destructive" : "cursor-default",
                )}
              >
                {jumpable && (
                  <ChevronRight className="mt-0.5 size-3.5 shrink-0 opacity-50 transition-opacity group-hover:opacity-100" />
                )}
                <span className={cn(!jumpable && "pl-[1.25rem]")}>
                  <span className="font-medium">{item.label}</span>
                  <span className="text-destructive/50"> — </span>
                  {item.message}
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
