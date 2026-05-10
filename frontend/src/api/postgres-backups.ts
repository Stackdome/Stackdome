import api from "./client";
import type { components } from "./types/openapi";

export type PostgresBackup = components["schemas"]["PostgresBackup"];
export type PostgresBackupList = components["schemas"]["PostgresBackupList"];
export type PostgresBackupConfig = components["schemas"]["PostgresBackupConfig"];
export type PostgresBackupPhase = NonNullable<PostgresBackup["phase"]>;
export type PostgresBackupType = NonNullable<PostgresBackup["type"]>;

export type TriggerBackupPayload = {
  name?: string;
  description?: string;
};

export async function listPostgresBackups(
  orgId: string,
  addonId: string,
): Promise<PostgresBackupList> {
  const res = await api.get(`/organizations/${orgId}/addons/postgres/${addonId}/backups`);
  return res.data as PostgresBackupList;
}

export async function triggerPostgresBackup(
  orgId: string,
  addonId: string,
  payload: TriggerBackupPayload = {},
): Promise<PostgresBackup> {
  const res = await api.post(
    `/organizations/${orgId}/addons/postgres/${addonId}/actions/backup`,
    payload,
  );
  return res.data as PostgresBackup;
}

export function isTerminalPhase(phase?: PostgresBackupPhase): boolean {
  return phase === "completed" || phase === "failed";
}
