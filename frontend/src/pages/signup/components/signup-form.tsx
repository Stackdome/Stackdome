import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Loader2 } from "lucide-react";
import { useSignup } from "../hooks/use-signup";
import type { UserSignupRequest, UserSignupResponse } from "@/api/users";
import type { SignupFormData } from "../types";
import { signupSchema } from "../types";
import { setAuthSession } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { FieldLabel } from "@/pages/auth/components/auth-shell";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";
import { AlertBanner } from "@/components/branded/alert-banner";
import { FieldError } from "@/components/branded/field-error";

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
      <GitHubSignInButton />

      <form onSubmit={handleSubmit} className="space-y-3">
        {serverError && <AlertBanner>{serverError}</AlertBanner>}

        <div className="space-y-2">
          <FieldLabel htmlFor="name">Full name</FieldLabel>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Your name"
            value={formData.name || ""}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.name}
          />
          <FieldError>{errors.name}</FieldError>
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="organisationName">Organization</FieldLabel>
          <Input
            id="organisationName"
            name="organisationName"
            type="text"
            placeholder="Founder Labs"
            value={formData.organisationName || ""}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.organisationName}
          />
          <FieldError>{errors.organisationName}</FieldError>
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="email">Email</FieldLabel>
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
            disabled={isLoading}
            aria-invalid={!!errors.email}
          />
          <FieldError>{errors.email}</FieldError>
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="password" hint="min. 8 characters">
            Password
          </FieldLabel>
          <PasswordInput
            id="password"
            name="password"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.password}
          />
          <FieldError>{errors.password}</FieldError>
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="confirmPassword">Confirm password</FieldLabel>
          <PasswordInput
            id="confirmPassword"
            name="confirmPassword"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.confirmPassword}
            onChange={handleChange}
            disabled={isLoading}
            aria-invalid={!!errors.confirmPassword}
          />
          <FieldError>{errors.confirmPassword}</FieldError>
        </div>

        <Button type="submit" variant="default" className="w-full" disabled={isLoading}>
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
