import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2 } from "lucide-react";
import { useLogin } from "../hooks/use-login";
import type { LoginFormData } from "../types";
import { loginSchema } from "../types";
import { setAuthSession } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { FormHead, FieldLabel, FootRow } from "@/pages/auth/components/auth-shell";

export function LoginForm() {
  const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
  const [errors, setErrors] = useState<Partial<LoginFormData>>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useLogin();
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
        setAuthSession(response.token, response.user);
      }
      navigate("/dashboard");
    } catch (err) {
      setServerError(getErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div>
      <FormHead
        step="step 01 / sign in"
        title="Welcome back."
        trailing={
          <>
            No account yet?{" "}
            <Link to="/sign-up" className="text-brand hover:underline underline-offset-4">
              Start here →
            </Link>
          </>
        }
      />

      <form onSubmit={handleSubmit} autoComplete="on" className="space-y-4">
        {serverError && (
          <div className="rounded-sm border border-[rgb(220_38_38_/_0.55)] bg-[rgb(220_38_38_/_0.12)] px-3 py-2 text-sm text-[#b91c1c] dark:text-[#fca5a5]">
            {serverError}
          </div>
        )}

        <div className="space-y-2">
          <FieldLabel htmlFor="email" number="→">email</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="username"
            placeholder="you@company.dev"
            value={formData.email}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.email}
          />
          {errors.email && (
            <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.email}</p>
          )}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="password" number="→">password</FieldLabel>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.password}
          />
          {errors.password && (
            <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.password}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
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

      <FootRow left="secure / tls 1.3" right="encrypted at rest" />
    </div>
  );
}
