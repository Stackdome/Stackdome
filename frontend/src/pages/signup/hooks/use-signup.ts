import { useState } from "react";
import * as userApi from "@/api/users";
import type { components } from "@/api/types/openapi";
import { getErrorMessage } from "@/api/client";

type UserSignupRequest = components["schemas"]["UserSignupRequest"];
type UserSignupResponse = components["schemas"]["UserSignupResponse"];

export function useSignup() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [user, setUser] = useState<UserSignupResponse | null>(null);

  const signup = async (data: UserSignupRequest, inviteToken?: string) => {
    setLoading(true);
    setError(null);
    try {
      const payload: UserSignupRequest = inviteToken
        ? { name: data.name, email: data.email, password: data.password, invite_token: inviteToken }
        : data;
      const result = await userApi.signupUser(payload);
      setUser(result);
      return result;
    } catch (err: unknown) {
      setError(getErrorMessage(err));
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { signup, loading, error, user };
}
