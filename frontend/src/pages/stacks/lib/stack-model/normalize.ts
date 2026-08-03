import {
  DEFAULT_BUILD_CONTEXT,
  DEFAULT_DOCKERFILE_PATH,
  DEFAULT_WORKLOAD_TYPE,
  SERVER_WRITTEN_RESOURCE_FIELDS,
  SERVER_WRITTEN_VOLUME_FIELDS,
} from "./policy";

/** Drop the listed keys. The adapters call this before anything else, so no
 *  comparison downstream ever sees a server-written field. */
export function omitFields<T extends object>(value: T, fields: readonly string[]): T {
  const out = { ...value } as Record<string, unknown>;
  for (const f of fields) delete out[f];
  return out as T;
}

export function omitServerWrittenResourceFields<T extends object>(r: T): T {
  return omitFields(r, SERVER_WRITTEN_RESOURCE_FIELDS);
}

export function omitServerWrittenVolumeFields<T extends object>(v: T): T {
  return omitFields(v, SERVER_WRITTEN_VOLUME_FIELDS);
}

type GitLike = { dockerfile_path?: string; build_context?: string } & Record<string, unknown>;

/** Adopt the build-path defaults the API applies on write, so an omitted path
 *  and a spelled-out one compare equal. */
export function withGitDefaults<T extends GitLike>(git: T): T {
  return {
    ...git,
    dockerfile_path: git.dockerfile_path || DEFAULT_DOCKERFILE_PATH,
    build_context: git.build_context || DEFAULT_BUILD_CONTEXT,
  };
}

type SourceLike = { git?: GitLike; image?: unknown; volume?: unknown } | undefined;

export function normalizeSource(source: SourceLike): SourceLike {
  if (!source?.git) return source;
  return { ...source, git: withGitDefaults(source.git) };
}

export function normalizeWorkloadType(workloadType: string | undefined): string {
  return workloadType || DEFAULT_WORKLOAD_TYPE;
}
