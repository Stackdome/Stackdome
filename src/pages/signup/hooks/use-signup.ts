import { useState } from "react";
import { signupUser } from "../../../api/users";
import type { components } from "../../../api/types/openapi";

type UserSignupRequest = components["schemas"]["UserSignupRequest"];
type User = components["schemas"]["User"];

export function useSignup() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);

  const signup = async (data: UserSignupRequest) => {
    setLoading(true);
    setError(null);
    try {
      const result = await signupUser(data);
      setUser(result);
      return result;
    } catch (err: unknown) {
      setError((err as Error)?.message || "Signup failed");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { signup, loading, error, user };
}
