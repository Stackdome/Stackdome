import type { ReactNode } from "react";

const MAX_SHOWN = 3;

/**
 * Compact toast body for import conversion warnings: dedupes repeated
 * messages (the converter emits one per service), keeps only the leading
 * sentence of each, and lists at most MAX_SHOWN with a "+N more" tail.
 */
export function formatImportWarnings(messages: string[]): { count: number; description: ReactNode } {
  const unique = [...new Set(messages.map((m) => m.split(". ")[0].replace(/\.$/, "")))];
  const shown = unique.slice(0, MAX_SHOWN);
  const rest = unique.length - shown.length;
  return {
    count: unique.length,
    description: (
      <ul className="mt-1 list-disc space-y-0.5 pl-4">
        {shown.map((m) => (
          <li key={m}>{m}</li>
        ))}
        {rest > 0 && <li className="list-none pl-0 text-muted-foreground">+ {rest} more</li>}
      </ul>
    ),
  };
}
