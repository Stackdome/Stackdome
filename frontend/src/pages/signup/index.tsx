import { SignupForm } from "@/pages/signup/components/signup-form";
import { AuthShell } from "@/pages/auth/components/auth-shell";

export default function Signup() {
  return (
    <AuthShell
      marker={{ code: "02 / NEW", expr: "workspace.create()" }}
      headlineSolid="Three layers."
      headlineStroke="One runtime."
      sub="Spin up your first cluster in under a minute. No credit card. Open source under Apache 2.0."
      stageStatus="building"
      checklist={[
        "Free during preview · no credit card",
        "Open source · self-host or managed",
        "14 regions worldwide · low-latency edge",
      ]}
    >
      <SignupForm />
    </AuthShell>
  );
}
