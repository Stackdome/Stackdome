import { clsx, type ClassValue } from "clsx"
import { extendTailwindMerge } from "tailwind-merge"
import type { components } from "@/api/types/openapi"
import type { ZodIssue } from "zod";

/**
 * The product type scale is named by job (`text-body`, `text-meta`, …) rather
 * than by size. tailwind-merge cannot know that: it classifies any unfamiliar
 * `text-*` utility as a text COLOUR, so `cn("text-body", "text-fg-2")` used to
 * drop the size entirely and the element fell back to the inherited 16px.
 *
 * Teaching it the font-size group fixes that at the root — without this, every
 * component that merges a size and a colour through `cn` silently loses its
 * size.
 */
const TEXT_SCALE = ["label", "meta", "body", "name", "title", "head"] as const

const twMerge = extendTailwindMerge({
  extend: {
    classGroups: {
      "font-size": [{ text: [...TEXT_SCALE] }],
    },
  },
})

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function extractApiErrorMessage(
  error: Partial<components["schemas"]["Error"]>,
  fallbackMessage = "An error occurred. Please try again."
): string {
  if (!error) return fallbackMessage;
  if (error.reason) return `Error: ${error.reason}`;

  if (error.code || error.kind) {
    const kind = error.kind || "Error";
    const code = error.code ? ` (${error.code})` : "";
    return `${kind}${code}`;
  }

  return fallbackMessage;
}

function isApiError(error: unknown): error is Partial<components["schemas"]["Error"]> {
  return (
    typeof error === "object" &&
    error !== null &&
    ("reason" in error && "code" in error && "kind" in error)
  );
}

function isZodIssue(error: unknown): error is ZodIssue {
  return (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    "path" in error &&
    typeof (error as { message: unknown }).message === "string" &&
    Array.isArray((error as { path: unknown }).path)
  );
}

export function extractErrorMessage(
  error: Error | Partial<components["schemas"]["Error"]> | ZodIssue,
  fallbackMessage = "An error occurred. Please try again."
): string {
  if (isApiError(error)) return extractApiErrorMessage(error, fallbackMessage);
  if (isZodIssue(error)) return error.message;
  if (error instanceof Error) console.error(error.message);

  return fallbackMessage;
}
