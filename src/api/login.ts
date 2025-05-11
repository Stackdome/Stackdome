import api from "./client";
import type { components } from "../api/types/openapi";

export type LoginRequest = components["schemas"]["LoginRequest"];
export type LoginResponse = components["schemas"]["LoginResponse"];

export async function loginUser(data: LoginRequest): Promise<LoginResponse> {
  const response = await api.post("/auth/login", data);
  return response.data;
}
