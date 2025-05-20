import api from "./client"; // Changed to default import
import type { components } from "./types/openapi";
import type { StackList } from "@/pages/stacks/types";

export type Stack = components["schemas"]["Stack"];
export type StackResource = components["schemas"]["StackResource"];
export type StackResourceList = components["schemas"]["StackResourceList"];

export async function getStacksByOrg(orgId: string): Promise<StackList> {
  const response = await api.get<StackList>(
    `/organizations/${orgId}/stacks`
  );
  return response.data;
}

export async function createStack(orgId: string, input: Stack): Promise<Stack> {
  const response = await api.post(`/organizations/${orgId}/stacks`, input);
  return response.data;
}

export async function getStackById(orgId: string, stackId: string): Promise<Stack> {
  const response = await api.get<Stack>(`/organizations/${orgId}/stacks/${stackId}`);
  return response.data;
}

