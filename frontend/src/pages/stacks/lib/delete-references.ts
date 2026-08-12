import type { FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import type { ResourceArr, VolumeArr } from "@/pages/stacks/lib/stack-diff";

/** One env var that names another resource, as `resource.ENV_KEY`. */
export interface EnvRef {
  resource: string;
  envName: string;
}

export interface ResourceDependents {
  /** Resources holding a `depends_on` entry for the target. */
  dependsOn: string[];
  /** Structured rows that will be removed. */
  envRefs: EnvRef[];
  /** Literal values that merely mention the target. Flagged, never rewritten. */
  literalRefs: EnvRef[];
  /** Volumes the target mounted that nothing else mounts. */
  orphanedVolumes: string[];
}

const namesResource = (row: FormEnvRow, name: string) =>
  (row.from === "resource" || row.from === "resourceTemplate") && row.resourceName === name;

const envRowsOf = (r: ResourceArr[number]) =>
  (r?.execution_config?.environment_variables ?? []) as FormEnvRow[];

/**
 * What breaks if `name` is deleted.
 *
 * Structured rows and `depends_on` entries are repaired automatically. A
 * literal row is reported separately: `${web.host}` inside a longer string is
 * the user's own text, and the resolved value only exists server-side, so
 * there is nothing to substitute.
 */
export function findResourceDependents(
  resources: ResourceArr,
  volumes: VolumeArr,
  name: string,
): ResourceDependents {
  const dependsOn: string[] = [];
  const envRefs: EnvRef[] = [];
  const literalRefs: EnvRef[] = [];
  const literalMention = `\${${name}.`;

  for (const r of resources) {
    if (!r?.name || r.name === name) continue;

    if ((r.depends_on ?? []).includes(name)) dependsOn.push(r.name);

    for (const row of envRowsOf(r)) {
      if (namesResource(row, name)) {
        envRefs.push({ resource: r.name, envName: row.name });
      } else if (row.from === "stack" && row.value?.includes(literalMention)) {
        literalRefs.push({ resource: r.name, envName: row.name });
      }
    }
  }

  const mountedElsewhere = new Set(
    resources
      .filter((r) => r?.name !== name)
      .flatMap((r) => (r?.volume_mounts ?? []).map((m) => m.source_volume_name)),
  );
  const orphanedVolumes = volumes
    .map((v) => v?.name)
    .filter((v): v is string => !!v && !mountedElsewhere.has(v));

  return { dependsOn, envRefs, literalRefs, orphanedVolumes };
}

export function hasDependents(d: ResourceDependents): boolean {
  return d.dependsOn.length > 0 || d.envRefs.length > 0 || d.literalRefs.length > 0;
}

/**
 * Remove a resource and every structured reference to it.
 *
 * Siblings address a resource by name, and connections are derived from those
 * names at serialize time, so a delete that leaves them behind produces a
 * connection to a resource that does not exist. The API only rejects that when
 * it renders the release, by which point the deploy has already failed.
 *
 * Untouched resources are returned by identity so the diff keeps pairing them.
 */
export function deleteResourceAndReferences(resources: ResourceArr, name: string): ResourceArr {
  if (!name) return resources;

  return resources
    .filter((r) => r?.name !== name)
    .map((r) => {
      const rows = envRowsOf(r);
      const keptRows = rows.filter((row) => !namesResource(row, name));
      const deps = r?.depends_on ?? [];
      const keptDeps = deps.filter((d) => d !== name);

      const rowsChanged = keptRows.length !== rows.length;
      const depsChanged = keptDeps.length !== deps.length;
      if (!rowsChanged && !depsChanged) return r;

      return {
        ...r,
        ...(depsChanged ? { depends_on: keptDeps } : {}),
        ...(rowsChanged
          ? { execution_config: { ...r.execution_config, environment_variables: keptRows } }
          : {}),
      } as ResourceArr[number];
    });
}
