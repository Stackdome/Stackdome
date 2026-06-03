import { Heart, Loader2, Unlock, Zap } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { SignupForm } from "@/pages/signup/components/signup-form";
import { InviteAcceptForm } from "@/pages/signup/components/invite-accept-form";
import { AuthShell, FormHead } from "@/pages/auth/components/auth-shell";
import { useInviteInfo } from "@/pages/signup/hooks/use-invite-info";
import { isUserLoggedIn, logoutAndRedirect } from "@/helpers/common";
import { Button } from "@/components/ui/button";

const CHECKLIST = [
  {
    icon: <Zap fill="currentColor" />,
    text: <>Powered by <span className="text-foreground">Kubernetes</span></>,
  },
  {
    icon: <Unlock />,
    text: <>No vendor <span className="text-foreground">lock-in</span></>,
  },
  {
    icon: <Heart fill="currentColor" />,
    text: <>Built with <span className="text-foreground">open source</span></>,
  },
];

// ------------------------------------------------------------------
// No-invite mode: existing org-creation flow
// ------------------------------------------------------------------
function DefaultSignup() {
  return (
    <AuthShell
      headlineSolid="Kickstart your"
      headlineStroke="deployment journey."
      checklist={CHECKLIST}
    >
      <SignupForm />
    </AuthShell>
  );
}

// ------------------------------------------------------------------
// Invite mode inner — handles loading / error states for the token
// ------------------------------------------------------------------
function InviteSignup({ token }: { token: string }) {
  const { state, info } = useInviteInfo(token);

  if (state === "loading") {
    return (
      <AuthShell
        headlineSolid="You've been"
        headlineStroke="invited."
        checklist={CHECKLIST}
      >
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      </AuthShell>
    );
  }

  if (state === "new-user" && info) {
    return (
      <AuthShell
        headlineSolid={`Join ${info.org_name}`}
        headlineStroke="on Stackdome."
        checklist={CHECKLIST}
      >
        <InviteAcceptForm token={token} info={info} />
      </AuthShell>
    );
  }

  // Error states: not-found, expired, revoked, already-used
  const errorCopy: Record<string, { title: string; body: string }> = {
    "not-found": {
      title: "Invite not found",
      body: "This invite link is invalid or has already been used.",
    },
    expired: {
      title: "Invite expired",
      body: "This invite link has expired. Ask your team admin to send a new one.",
    },
    revoked: {
      title: "Invite revoked",
      body: "This invite has been revoked. Ask your team admin to send a new one.",
    },
    "already-used": {
      title: "Invite already used",
      body: "This invite has already been accepted. Log in to access your account.",
    },
  };

  const copy = errorCopy[state] ?? errorCopy["not-found"];

  return (
    <AuthShell
      headlineSolid="You've been"
      headlineStroke="invited."
      checklist={CHECKLIST}
    >
      <div className="space-y-4 py-6">
        <FormHead step="invite" title={copy.title} />
        <p className="text-sm text-muted-foreground">{copy.body}</p>
        <Link
          to="/sign-in"
          className="inline-block text-sm text-foreground underline underline-offset-4 decoration-[1.5px] decoration-brand/80"
        >
          Go to sign in →
        </Link>
      </div>
    </AuthShell>
  );
}

// ------------------------------------------------------------------
// Wrong-account card (user already logged in)
// ------------------------------------------------------------------
function WrongAccountCard() {
  const currentUrl = typeof window !== "undefined" ? window.location.href : "/";
  return (
    <AuthShell
      headlineSolid="You've been"
      headlineStroke="invited."
      checklist={CHECKLIST}
    >
      <div className="space-y-4 py-6">
        <FormHead
          step="invite"
          title="Already signed in"
          trailing="Sign out to accept this invite as the invited user."
        />
        <Button
          variant="inverse"
          className="w-full"
          onClick={() => logoutAndRedirect(currentUrl)}
        >
          Sign out and accept invite
        </Button>
      </div>
    </AuthShell>
  );
}

// ------------------------------------------------------------------
// Page entry point
// ------------------------------------------------------------------
export default function Signup() {
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite_token");

  if (!inviteToken) {
    return <DefaultSignup />;
  }

  if (isUserLoggedIn()) {
    return <WrongAccountCard />;
  }

  return <InviteSignup token={inviteToken} />;
}
