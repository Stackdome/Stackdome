import { SignupForm } from "@/pages/signup/components/signup-form";
import { AuthShell } from "@/pages/auth/components/auth-shell";

export default function Signup() {
  return (
    <AuthShell
      headlineSolid="Kickstart your delightful deployment journey."
      tagline={
        <>
          Be the captain. Just ship it<span className="text-brand">.</span>
        </>
      }
      checklist={[
        <>Powered by <span className="text-foreground">open source</span></>,
        <>Built on top of <span className="text-foreground">Kubernetes</span></>,
        <>No vendor lock-in. <span className="text-foreground">Ever.</span></>,
      ]}
    >
      <SignupForm />
    </AuthShell>
  );
}
