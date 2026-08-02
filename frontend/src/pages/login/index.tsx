import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { isUserLoggedIn } from "@/lib/common";

import { LoginForm } from "@/pages/login/components/login-form";
import { AuthShell, SwapLink } from "@/pages/auth/components/auth-shell";

export default function Login() {
  const navigate = useNavigate();
  useEffect(() => {
    if (isUserLoggedIn()) {
      navigate("/dashboard");
    }
  }, [navigate]);

  return (
    <AuthShell
      title="Welcome back."
      sub="Sign in to manage your stacks."
      below={<SwapLink lead="New to Stackdome?" to="/sign-up" label="Create an account" />}
    >
      <LoginForm />
    </AuthShell>
  );
}
