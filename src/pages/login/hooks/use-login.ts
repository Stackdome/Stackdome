import { useState } from "react";
import { loginUser } from "../../../api/users";
import type { LoginRequest, LoginResponse } from "../../../api/users";

export function useLogin() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [user, setUser] = useState<LoginResponse | null>(null);

  const login = async (data: LoginRequest) => {
    setLoading(true);
    setError(null);
    try {
      const result = await loginUser(data);
      setUser(result);
      return result;
    } catch (err: unknown) {
      setError((err as Error)?.message || "Login failed");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { login, loading, error, user };
}
