import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2 } from "lucide-react";
import { useSignup } from "../hooks/use-signup";
import type { UserSignupRequest, UserSignupResponse } from "@/api/users";
import type { SignupFormData } from "../types";
import { signupSchema } from "../types";
import { setAuthSession } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { FormHead, FieldLabel, FootRow } from "@/pages/auth/components/auth-shell";

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
        setAuthSession(response.jwt_token, response.user);
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
        step="step 01 of 03 / create account"
        title="Own your runtime."
        trailing={
          <>
            Already have one?{" "}
            <Link
              to="/signin"
              className="text-foreground underline underline-offset-4 decoration-[1.5px] decoration-brand/80 hover:decoration-brand"
            >
              Sign in<span className="text-brand">.</span>
            </Link>
          </>
        }
      />

      <form onSubmit={handleSubmit} className="space-y-3">
        {serverError && (
          <div className="rounded-sm border border-[rgb(220_38_38_/_0.55)] bg-[rgb(220_38_38_/_0.12)] px-3 py-2 text-sm text-[#b91c1c] dark:text-[#fca5a5]">
            {serverError}
          </div>
        )}

        <div className="space-y-2">
          <FieldLabel htmlFor="name">full name</FieldLabel>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Jane Cooper"
            value={formData.name || ""}
            onChange={handleChange}
            aria-invalid={!!errors.name}
          />
          {errors.name && <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.name}</p>}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="organisationName">organization</FieldLabel>
          <Input
            id="organisationName"
            name="organisationName"
            type="text"
            placeholder="Acme Inc."
            value={formData.organisationName || ""}
            onChange={handleChange}
            aria-invalid={!!errors.organisationName}
          />
          {errors.organisationName && (
            <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.organisationName}</p>
          )}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="email">work email</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            autoCapitalize="none"
            autoComplete="email"
            autoCorrect="off"
            placeholder="you@company.dev"
            value={formData.email}
            onChange={handleChange}
            aria-invalid={!!errors.email}
          />
          {errors.email && <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.email}</p>}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="password" hint="min. 8 characters">
            password
          </FieldLabel>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            aria-invalid={!!errors.password}
          />
          {errors.password && (
            <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.password}</p>
          )}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="confirmPassword">confirm</FieldLabel>
          <Input
            id="confirmPassword"
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.confirmPassword}
            onChange={handleChange}
            aria-invalid={!!errors.confirmPassword}
          />
          {errors.confirmPassword && (
            <p className="text-xs text-[#b91c1c] dark:text-[#fca5a5]">{errors.confirmPassword}</p>
          )}
        </div>

        <Button
          type="submit"
          className="w-full bg-foreground text-background hover:bg-foreground/90"
          disabled={isLoading}
        >
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

      <p className="mt-4 text-xs text-muted-foreground">
        By creating an account you agree to the{" "}
        <a href="#" className="text-brand hover:underline underline-offset-4">Terms</a> and{" "}
        <a href="#" className="text-brand hover:underline underline-offset-4">Privacy Policy</a>.
      </p>

      <FootRow left="open source · apache 2.0" />
    </div>
  );
}
