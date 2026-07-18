/** Docker Hub host aliases the API normalizes between (docker.io ↔ index.docker.io). */
const DOCKER_HUB_HOSTS = new Set(["docker.io", "index.docker.io", "registry-1.docker.io"]);

/** Split an image ref into registry host and repo:tag remainder. The first
    path segment is a host only if it contains a dot or port, or is localhost —
    the same rule the Docker CLI applies. */
export function splitImageRef(ref: string): { host: string | null; remainder: string } {
  const slash = ref.indexOf("/");
  if (slash < 0) return { host: null, remainder: ref };
  const first = ref.slice(0, slash);
  const isHost = first.includes(".") || first.includes(":") || first === "localhost";
  if (!isHost) return { host: null, remainder: ref };
  return { host: first, remainder: ref.slice(slash + 1) };
}

export function joinImageRef(host: string | null, remainder: string): string {
  return host ? `${host}/${remainder}` : remainder;
}

export function dockerHostsEqual(a: string, b: string): boolean {
  const la = a.toLowerCase();
  const lb = b.toLowerCase();
  if (la === lb) return true;
  return DOCKER_HUB_HOSTS.has(la) && DOCKER_HUB_HOSTS.has(lb);
}
