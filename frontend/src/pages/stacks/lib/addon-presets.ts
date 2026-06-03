// Postgres addon output accessors, exactly as the backend exposes them
// (pkg/models/output_descriptor.go). These are the `value.output` strings
// written into an addon env connection's mappings.
export const ADDON_OUTPUT_FIELDS = [
  "host",
  "port",
  "database",
  "username",
  "password",
  "sslmode",
  "ca_certificate",
  "url",
] as const;

export type AddonOutputField = (typeof ADDON_OUTPUT_FIELDS)[number];

// Fields that are identical across every database in the addon (cluster-wide),
// shown with a "cluster" hint in pickers.
export const CLUSTER_WIDE_FIELDS: ReadonlySet<AddonOutputField> = new Set<AddonOutputField>([
  "host",
  "port",
  "sslmode",
  "ca_certificate",
]);
