import { Heart, Unlock, Zap } from "lucide-react";
import { SignupForm } from "@/pages/signup/components/signup-form";
import { AuthShell } from "@/pages/auth/components/auth-shell";

export default function Signup() {
  return (
    <AuthShell
      headlineSolid="Kickstart your"
      headlineStroke="deployment journey."
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
      <SignupForm />
    </AuthShell>
  );
}
