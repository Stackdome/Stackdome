import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Loader2 } from "lucide-react";
import { useLogin } from "../hooks/use-login";
import type { LoginFormData } from "../types";
import { loginSchema } from "../types";
import { setAuthSession } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useCurrentUser } from "@/hooks/use-current-user";
import { FieldLabel } from "@/pages/auth/components/auth-shell";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
import { AlertBanner } from "@/components/branded/alert-banner";
import { FieldError } from "@/components/branded/field-error";

export function LoginForm() {
  const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
  const [errors, setErrors] = useState<Partial<LoginFormData>>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useLogin();
  const { refresh } = useCurrentUser();
  const navigate = useNavigate();

  const validateForm = (): boolean => {
    const result = loginSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<LoginFormData> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as keyof LoginFormData;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return false;
    }
    setErrors({});
    return true;
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof LoginFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
    setServerError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setServerError(null);
    if (!validateForm()) return;
    setIsLoading(true);
    try {
      const response = await login(formData);
      if (response && response.token && response.user) {
        setAuthSession(response.token, response.user, response.refresh_token);
      }
      // Reload the current-user context with the freshly-authenticated user so
      // role-gated nav (Settings, Clusters, Domains) reflects this session
      // immediately instead of carrying stale state until a manual reload.
      await refresh();
      navigate("/dashboard");
    } catch (err) {
      setServerError(getErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div>
      <GitHubSignInButton />

      <form onSubmit={handleSubmit} autoComplete="on" className="space-y-4">
        {serverError && <AlertBanner>{serverError}</AlertBanner>}

        <div className="space-y-2">
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="username"
            placeholder="you@company.com"
            value={formData.email}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.email}
          />
          <FieldError>{errors.email}</FieldError>
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="password">Password</FieldLabel>
          <PasswordInput
            id="password"
            name="password"
            autoComplete="current-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.password}
          />
          <FieldError>{errors.password}</FieldError>
        </div>

        <Button type="submit" variant="default" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="animate-spin h-4 w-4" />
              Signing in…
            </>
          ) : (
            <>
              Continue <span className="font-mono">→</span>
            </>
          )}
        </Button>
      </form>
    </div>
  );
}
