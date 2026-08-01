export const DEFAULT_DOCKERFILE_PATH = "Dockerfile";
export const DEFAULT_BUILD_CONTEXT = ".";

/**
 * The API applies these defaults on write, so a git source that omits them and
 * one that spells them out are the same config. Every side of a comparison
 * normalizes through here, or the form's defaults read as user edits: a phantom
 * autosave, a permanently dirty field, a diff nobody authored.
 */
export function withGitBuildDefaults<T extends { dockerfile_path?: string; build_context?: string }>(
  git: T,
): T {
  return {
    ...git,
    dockerfile_path: git.dockerfile_path || DEFAULT_DOCKERFILE_PATH,
    build_context: git.build_context || DEFAULT_BUILD_CONTEXT,
  };
}
