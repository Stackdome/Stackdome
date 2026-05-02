import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Heart, Unlock, Zap } from "lucide-react";
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
      headlineSolid="Deploy. Own."
      headlineStroke="Scale."
      checklist={[
        {
          icon: <Heart fill="currentColor" />,
          text: <>Built with <span className="text-foreground">open source</span></>,
        },
        {
          icon: <Zap fill="currentColor" />,
          text: <>Powered by <span className="text-foreground">Kubernetes</span></>,
        },
        {
          icon: <Unlock />,
          text: <>No vendor <span className="text-foreground">lock-in</span></>,
        },
      ]}
    >
      <LoginForm />
    </AuthShell>
  );
}
