import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";
import type { FormEnvRow } from "@/pages/stacks/lib/connection-mapping";

/**
 * Carry a resource rename to the siblings that name it.
 *
 * Env rows and `depends_on` address a resource by name, and connections
 * serialize that name (`connection-mapping.ts:95`), so a rename that does not
 * carry its references produces a connection to a resource that does not
 * exist. The API only rejects that when it renders the release, by which point
 * the deploy has already failed.
 *
 * Untouched resources are returned by identity so the diff keeps pairing them.
 */
export function renameResourceReferences(
  resources: FormStackResourceData[],
  from: string,
  to: string,
): FormStackResourceData[] {
  if (!from || !to || from === to) return resources;

  let changed = false;
  const next = resources.map((r) => {
    const rows = (r.execution_config?.environment_variables ?? []) as FormEnvRow[];
    const renamedRows = rows.map((row) =>
      (row.from === "resource" || row.from === "resourceTemplate") && row.resourceName === from
        ? { ...row, resourceName: to }
        : row,
    );
    const deps = r.depends_on ?? [];
    const renamedDeps = deps.map((d) => (d === from ? to : d));

    const rowsChanged = renamedRows.some((row, i) => row !== rows[i]);
    const depsChanged = renamedDeps.some((d, i) => d !== deps[i]);
    if (!rowsChanged && !depsChanged) return r;

    changed = true;
    return {
      ...r,
      ...(depsChanged ? { depends_on: renamedDeps } : {}),
      ...(rowsChanged
        ? { execution_config: { ...r.execution_config, environment_variables: renamedRows } }
        : {}),
    } as FormStackResourceData;
  });

  return changed ? next : resources;
}
