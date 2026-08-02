import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2 } from "lucide-react";
import { useSignup } from "../hooks/use-signup";
import { inviteAcceptSchema } from "../types";
import type { InviteAcceptFormData } from "../types";
import type { OrgInviteInfo } from "@/api/invites";
import { setAuthSession } from "@/lib/common";
import { isErrorStatus, getErrorMessage } from "@/api/client";
import { FieldLabel } from "@/pages/auth/components/auth-shell";
import { GitHubSignInButton } from "@/components/auth/github-sign-in-button";

type Phase = "form" | "accepting" | "accepted" | "existing-user";

interface InviteAcceptFormProps {
  token: string;
  info: OrgInviteInfo;
}

export function InviteAcceptForm({ token }: InviteAcceptFormProps) {
  const [formData, setFormData] = useState<InviteAcceptFormData>({
    name: "",
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState<Partial<InviteAcceptFormData>>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [phase, setPhase] = useState<Phase>("form");
  const { signup } = useSignup();
  const navigate = useNavigate();
  const navTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => () => { if (navTimerRef.current) clearTimeout(navTimerRef.current); }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name as keyof InviteAcceptFormData]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }));
    }
  };

  const runSignup = () => {
    const result = inviteAcceptSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<InviteAcceptFormData> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as keyof InviteAcceptFormData;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    setServerError(null);
    setPhase("accepting");

    signup(
      { name: formData.name, email: formData.email, password: formData.password },
      token,
    ).then(
      (response) => {
        if (response?.jwt_token && response?.user) {
          setAuthSession(response.jwt_token, response.user, response.refresh_token);
        }
        setPhase("accepted");
        navTimerRef.current = setTimeout(() => navigate("/"), 1200);
      },
      (err: unknown) => {
        if (isErrorStatus(err, 409)) {
          setPhase("existing-user");
        } else {
          setServerError(getErrorMessage(err));
          setPhase("form");
        }
      },
    );
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    runSignup();
  };

  if (phase === "accepted") {
    return (
      <div className="space-y-4 text-center py-8">
        <p className="text-2xl font-semibold">You&apos;re in!</p>
        <p className="text-sm text-muted-foreground">Taking you to Stackdome…</p>
      </div>
    );
  }

  if (phase === "existing-user") {
    return (
      <div className="space-y-4 py-8">
        <p className="text-sm text-muted-foreground">
          You already have an account with this email address.{" "}
          <Link to="/sign-in" className="text-foreground underline underline-offset-4 decoration-[1.5px] decoration-brand/80">
            Log in
          </Link>{" "}
          to accept this invite.
        </p>
      </div>
    );
  }

  return (
    <div>
      <GitHubSignInButton inviteToken={token} />

      <form onSubmit={handleSubmit} className="space-y-3">
        {serverError && (
          <div className="rounded-2xl border border-danger-border bg-danger-bg px-4 py-2 text-sm text-danger">
            {serverError}
          </div>
        )}

        <div className="space-y-2">
          <FieldLabel htmlFor="invite-name">Full name</FieldLabel>
          <Input
            id="invite-name"
            name="name"
            type="text"
            placeholder="Your name"
            value={formData.name}
            onChange={handleChange}
            aria-invalid={!!errors.name}
          />
          {errors.name && <p className="text-xs text-danger">{errors.name}</p>}
        </div>

        <div className="space-y-2">
          <FieldLabel htmlFor="invite-email">Email</FieldLabel>
          <Input
            id="invite-email"
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
          <FieldLabel htmlFor="invite-password" hint="min. 8 characters">
            Password
          </FieldLabel>
          <Input
            id="invite-password"
            name="password"
            type="password"
            autoComplete="new-password"
            placeholder="••••••••••••"
            value={formData.password}
            onChange={handleChange}
            aria-invalid={!!errors.password}
          />
          {errors.password && <p className="text-xs text-danger">{errors.password}</p>}
        </div>

        <Button
          type="submit"
          variant="outline"
          className="w-full"
          disabled={phase === "accepting"}
        >
          {phase === "accepting" ? (
            <>
              <Loader2 className="animate-spin h-4 w-4" />
              Creating account…
            </>
          ) : (
            "Create account and join"
          )}
        </Button>
      </form>
    </div>
  );
}
