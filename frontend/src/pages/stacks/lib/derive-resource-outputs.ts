// Client mirror of pkg/models/output_descriptor.go::StackResourceOutputDescriptors.
// The server computes a resource's output DESCRIPTORS purely from its spec
// (host + per-port names). The values need a deployment; the names do not, so we
// derive them here for draft resources that have never been persisted. Keep this
// in sync with the Go function above.

export interface OutputSourceResource {
  ports?: ReadonlyArray<{ name?: string; exposed_to_public?: boolean }> | null;
}

export function deriveResourceOutputNames(resource: OutputSourceResource): string[] {
  const names = ["host"];
  for (const port of resource.ports ?? []) {
    if (!port.name) continue;
    names.push(`port.${port.name}`, `url.${port.name}`);
    if (port.exposed_to_public) {
      names.push(`public.${port.name}.host`, `public.${port.name}.url`);
    }
  }
  return names;
}
