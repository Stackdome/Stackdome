import type { z } from "zod";

/** Turn draft-deploy zod issues into user-facing rows: the stack-name issue is
 * split out for the inline title input; everything else becomes a labelled,
 * deduped message list for the toast. */
export function formatDraftValidationIssues(
  issues: z.ZodIssue[],
  resources: ReadonlyArray<{ name?: string }>,
  volumes: ReadonlyArray<{ name?: string }> | undefined,
): { nameError?: string; messages: string[] } {
  const messages: string[] = [];
  let nameError: string | undefined;

  for (const issue of issues) {
    const [scope0, scope1, idx, ...fieldPath] = issue.path;
    const field = fieldPath.join(".");
    if (scope0 === "name") {
      nameError = issue.message;
    } else if (scope0 === "spec" && scope1 === "stack_resources" && typeof idx === "number") {
      const label = resources[idx]?.name?.trim() || `Resource ${idx + 1}`;
      messages.push(field ? `${label}: ${issue.message} (${field})` : `${label}: ${issue.message}`);
    } else if (scope0 === "spec" && scope1 === "volumes" && typeof idx === "number") {
      const label = volumes?.[idx]?.name?.trim() || `Volume ${idx + 1}`;
      messages.push(field ? `${label}: ${issue.message} (${field})` : `${label}: ${issue.message}`);
    } else {
      messages.push(issue.path.length ? `${issue.path.join(".")}: ${issue.message}` : issue.message);
    }
  }

  return { nameError, messages: [...new Set(messages)] };
}
