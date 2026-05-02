import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { isUserLoggedIn } from "@/helpers/common";

import { LoginForm } from "@/pages/login/components/login-form";
import { AuthShell } from "@/pages/auth/components/auth-shell";

export default function Login() {
  const navigate = useNavigate();
  useEffect(() => {
    if (isUserLoggedIn()) {
      navigate("/dashboard");
    }
  }, [navigate]);

  return (
    <AuthShell
      marker={{ code: "01 / AUTH", expr: "session.init()" }}
      headlineSolid="Deploy. Own."
      headlineStroke="Scale."
      sub="The runtime you depend on should be a runtime you own. Sign in to your control plane."
      stageStatus="live"
      meta={[
        { label: "Region", value: "us-east-1" },
        { label: "Version", value: "v0.1.4" },
        { label: "Status", value: "● operational", tone: "brand" },
      ]}
    >
      <LoginForm />
    </AuthShell>
  );
}
