import api from "./client";
import type { components, paths } from "./types/openapi";

export type PostgresBackup = components["schemas"]["PostgresBackup"];
export type PostgresBackupList = components["schemas"]["PostgresBackupList"];
export type PostgresBackupConfig = components["schemas"]["PostgresBackupConfig"];
export type PostgresBackupPhase = NonNullable<PostgresBackup["phase"]>;
export type PostgresBackupType = NonNullable<PostgresBackup["type"]>;

type TriggerBackupOp =
  paths["/api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/actions/backup"]["post"];

export type TriggerBackupPayload = NonNullable<
  TriggerBackupOp["requestBody"]
>["content"]["application/json"];

export type TriggerBackupResponse =
  TriggerBackupOp["responses"][202]["content"]["application/json"];

export async function listPostgresBackups(
  orgId: string,
  addonId: string,
  opts: { limit?: number; offset?: number } = {},
): Promise<PostgresBackupList> {
  const res = await api.get(
    `/organizations/${orgId}/addons/postgres/${addonId}/backups`,
    { params: opts },
  );
  return res.data as PostgresBackupList;
}

export async function triggerPostgresBackup(
  orgId: string,
  addonId: string,
  payload: TriggerBackupPayload = {},
): Promise<TriggerBackupResponse> {
  const res = await api.post(
    `/organizations/${orgId}/addons/postgres/${addonId}/actions/backup`,
    payload,
  );
  return res.data as TriggerBackupResponse;
}

export function isTerminalPhase(phase?: PostgresBackupPhase): boolean {
  return phase === "completed" || phase === "failed";
}
