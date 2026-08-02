import type { Stack } from "@/api/stack-types";

export const stackNameTakenMessage = (name: string): string =>
  `A stack named "${name}" already exists`;

export function stackNameConflictError(params: {
  isCreate: boolean;
  name: string;
  projectId: string;
  existingStacks: Pick<Stack, "name" | "project_id">[];
}): string | undefined {
  if (!params.isCreate) return undefined;
  const trimmed = params.name.trim();
  const taken = params.existingStacks.some(
    (s) => s.project_id === params.projectId && s.name === trimmed,
  );
  return taken ? stackNameTakenMessage(trimmed) : undefined;
}
