import api from "./client";
import type { components } from "../api/types/openapi";

export type AppConfigResponse = components["schemas"]["AppConfigResponse"];

export async function getAppConfig(): Promise<AppConfigResponse> {
  const response = await api.get("/config");
  return response.data;
}
