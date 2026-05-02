export const CRED_FIELDS = [
  "host",
  "port",
  "username",
  "password",
  "database",
  "sslmode",
  "connectionString",
  "caCertificate",
] as const;

export type CredField = (typeof CRED_FIELDS)[number];

export const CLUSTER_WIDE_FIELDS: ReadonlySet<CredField> = new Set<CredField>([
  "host",
  "port",
  "sslmode",
  "caCertificate",
]);
