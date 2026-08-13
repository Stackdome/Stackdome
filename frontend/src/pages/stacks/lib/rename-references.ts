import type { FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import type { ResourceArr } from "@/pages/stacks/lib/stack-diff";

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
  resources: ResourceArr,
  from: string,
  to: string,
): ResourceArr {
  if (!from || !to || from === to) return resources;
  return renameResourceReferencesByMap(resources, new Map([[from, to]]));
}

/**
 * Carry several renames at once.
 *
 * Applying the single-pair rename in a loop corrupts overlapping sets: renaming
 * `a` to `a-2` first would then catch a sibling already named `a-2`. Every
 * lookup here reads the original name, so the substitution is simultaneous.
 */
export function renameResourceReferencesByMap(
  resources: ResourceArr,
  renames: Map<string, string>,
): ResourceArr {
  if (renames.size === 0) return resources;

  let changed = false;
  const next = resources.map((r) => {
    const rows = (r.execution_config?.environment_variables ?? []) as FormEnvRow[];
    const renamedRows = rows.map((row) =>
      (row.from === "resource" || row.from === "resourceTemplate") && renames.has(row.resourceName)
        ? { ...row, resourceName: renames.get(row.resourceName) as string }
        : row,
    );
    const deps = r.depends_on ?? [];
    const renamedDeps = deps.map((d) => renames.get(d) ?? d);

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
    } as ResourceArr[number];
  });

  return changed ? next : resources;
}
