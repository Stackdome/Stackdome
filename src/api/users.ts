import api from "./client";
import type { components } from "../api/types/openapi";

export type UserSignupRequest = components["schemas"]["UserSignupRequest"];
export type UserSignupResponse = components["schemas"]["UserSignupResponse"];
export type User = components["schemas"]["User"];
export type LoginRequest = components["schemas"]["LoginRequest"];
export type LoginResponse = components["schemas"]["LoginResponse"];

export async function signupUser(data: UserSignupRequest): Promise<UserSignupResponse> {
  const response = await api.post("/user-signup", data);
  return response.data;
}

export async function loginUser(data: LoginRequest): Promise<LoginResponse> {
  const response = await api.post("/auth/login", data);
  return response.data;
}

export async function getCurrentUser(): Promise<User> {
  const response = await api.get("/users/current");
  return response.data;
}
