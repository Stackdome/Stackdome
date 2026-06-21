import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Loader2 } from "lucide-react";
import { useSignup } from "../hooks/use-signup";
import type { UserSignupRequest, UserSignupResponse } from "@/api/users";
import type { SignupFormData } from "../types";
import { signupSchema } from "../types";
import { setAuthSession } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { FormHead, FieldLabel } from "@/pages/auth/components/auth-shell";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";

export function SignupForm() {
  const [formData, setFormData] = useState<SignupFormData>({
    email: "",
    password: "",
    confirmPassword: "",
    name: "",
    organisationName: "",
  });
  const [errors, setErrors] = useState<Partial<SignupFormData>>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { signup } = useSignup();
  const navigate = useNavigate();

  const validateForm = (): boolean => {
    const result = signupSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<SignupFormData> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as keyof SignupFormData;
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
    if (errors[name as keyof SignupFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
    setServerError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    setIsLoading(true);
    setErrors({});
    setServerError(null);
    try {
      const payload: UserSignupRequest = {
        name: formData.name,
        email: formData.email,
        password: formData.password,
        organisation: { name: formData.organisationName },
      };
      const response: UserSignupResponse = await signup(payload);
      if (response && response.jwt_token && response.user) {
        setAuthSession(response.jwt_token, response.user, response.refresh_token);
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
        step="create account"
        title="Own your stack."
        trailing={
          <>
            Already have one?{" "}
            <Link to="/sign-in" className="text-foreground">
              <span className="underline underline-offset-4 decoration-[1.5px] decoration-brand/80 hover:decoration-brand">
                Sign in
              </span>
              <span className="text-brand">.</span>
            </Link>
          </>
        }
      />

      <form onSubmit={handleSubmit} className="space-y-3">
        <GitHubSignInButton />

        {serverError && (
          <div className="rounded-sm border border-danger-border bg-danger-bg px-3 py-2 text-sm text-danger">
            {serverError}
          </div>
        )}

        <div className="space-y-2">
          <FieldLabel htmlFor="name">full name</FieldLabel>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Your name"
            value={formData.name || ""}
            onChange={handleChange}
            aria-invalid={!!errors.name}
          />
          {errors.name && <p className="text-xs text-danger">{errors.name}</p>}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="organisationName">organization</FieldLabel>
          <Input
            id="organisationName"
            name="organisationName"
            type="text"
            placeholder="Founder Labs"
            value={formData.organisationName || ""}
            onChange={handleChange}
            aria-invalid={!!errors.organisationName}
          />
          {errors.organisationName && (
            <p className="text-xs text-danger">{errors.organisationName}</p>
          )}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="email">email</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            autoCapitalize="none"
            autoComplete="email"
            autoCorrect="off"
            placeholder="you@company.com"
            value={formData.email}
            onChange={handleChange}
            aria-invalid={!!errors.email}
          />
          {errors.email && <p className="text-xs text-danger">{errors.email}</p>}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="password" hint="min. 8 characters">
            password
          </FieldLabel>
          <PasswordInput
            id="password"
            name="password"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            aria-invalid={!!errors.password}
          />
          {errors.password && (
            <p className="text-xs text-danger">{errors.password}</p>
          )}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="confirmPassword">confirm</FieldLabel>
          <PasswordInput
            id="confirmPassword"
            name="confirmPassword"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.confirmPassword}
            onChange={handleChange}
            aria-invalid={!!errors.confirmPassword}
          />
          {errors.confirmPassword && (
            <p className="text-xs text-danger">{errors.confirmPassword}</p>
          )}
        </div>

        <Button type="submit" variant="inverse" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="animate-spin h-4 w-4" />
              Creating account…
            </>
          ) : (
            <>
              Create account <span className="font-mono">→</span>
            </>
          )}
        </Button>
      </form>
    </div>
  );
}
