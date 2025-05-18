import api from "./client";
import type { components } from "./types/openapi";

export type Stack = components["schemas"]["Stack"];

export async function createStack(orgId: string, input: Stack): Promise<Stack> {
  const res = await api.post(`/organizations/${orgId}/stacks`, input);
  return res.data as Stack;
}
