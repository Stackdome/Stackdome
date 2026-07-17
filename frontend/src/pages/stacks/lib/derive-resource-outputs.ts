// Client mirror of pkg/models/output_descriptor.go::StackResourceOutputDescriptors.
// The server computes a resource's output DESCRIPTORS purely from its spec
// (host + per-port names). The values need a deployment; the names do not, so we
// derive them here for draft resources that have never been persisted. Keep this
// in sync with the Go function above.

export interface OutputSourceResource {
  ports?: ReadonlyArray<{ name?: string; exposed_to_public?: boolean }> | null;
}

export function deriveResourceOutputNames(resource: OutputSourceResource): string[] {
  const ports = (resource.ports ?? []).filter((p) => p.name);
  const multiPort = ports.length > 1;
  const key = (base: string, portName: string) => (multiPort ? `${base}.${portName}` : base);

  const names = ["host"];
  for (const port of ports) {
    names.push(key("port", port.name!), key("url", port.name!));
    if (port.exposed_to_public) {
      names.push(key("public_host", port.name!), key("public_url", port.name!));
    }
  }
  return names;
}
