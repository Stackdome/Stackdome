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

export const DEFAULT_ENV_NAMES: Record<CredField, string> = {
  host: "PG_HOST",
  port: "PG_PORT",
  username: "PG_USER",
  password: "PG_PASS",
  database: "PG_DB",
  sslmode: "PG_SSLMODE",
  connectionString: "DATABASE_URL",
  caCertificate: "PG_CA_CERT",
};

export type Preset = "postgres-conventions" | "connection-string" | "clear";

export interface PresetResult {
  selected: Set<CredField>;
  envNames: Partial<Record<CredField, string>>;
}

export function applyPreset(preset: Preset): PresetResult {
  switch (preset) {
    case "postgres-conventions":
      return {
        selected: new Set<CredField>([
          "host",
          "port",
          "username",
          "password",
          "database",
        ]),
        envNames: {
          host: "PG_HOST",
          port: "PG_PORT",
          username: "PG_USER",
          password: "PG_PASS",
          database: "PG_DB",
        },
      };
    case "connection-string":
      return {
        selected: new Set<CredField>(["connectionString"]),
        envNames: { connectionString: "DATABASE_URL" },
      };
    case "clear":
      return { selected: new Set<CredField>(), envNames: {} };
  }
}
