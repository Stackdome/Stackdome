import { Loader2 } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { format, parseISO } from "date-fns";
import { SignupForm } from "@/pages/signup/components/signup-form";
import { InviteAcceptForm } from "@/pages/signup/components/invite-accept-form";
import { AuthShell, SwapLink } from "@/pages/auth/components/auth-shell";
import { useInviteInfo } from "@/pages/signup/hooks/use-invite-info";
import type { OrgInviteInfo } from "@/api/invites";
import { isUserLoggedIn, logoutAndRedirect } from "@/helpers/common";
import { Button } from "@/components/ui/button";

// ------------------------------------------------------------------
// No-invite mode: existing org-creation flow
// ------------------------------------------------------------------
function DefaultSignup() {
  return (
    <AuthShell
      title="Own your stack."
      sub="Create an account to start deploying."
      below={<SwapLink lead="Already have an account?" to="/sign-in" label="Log in" />}
    >
      <SignupForm />
    </AuthShell>
  );
}

function inviteSub(info: OrgInviteInfo) {
  const expires = (() => {
    if (!info.expires_at) return "";
    try {
      return format(parseISO(info.expires_at), "MMM d, yyyy");
    } catch {
      return info.expires_at;
    }
  })();
  return (
    <>
      <span className="text-foreground">{info.inviter_name}</span> invited you to the{" "}
      <span className="text-foreground">{info.project_name}</span> project
      {expires && (
        <>
          {" · "}
          <span className="text-muted-foreground/60">Expires {expires}</span>
        </>
      )}
    </>
  );
}

// ------------------------------------------------------------------
// Invite mode inner — handles loading / error states for the token
// ------------------------------------------------------------------
function InviteSignup({ token }: { token: string }) {
  const { state, info } = useInviteInfo(token);

  if (state === "loading") {
    return (
      <AuthShell title="You've been invited.">
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      </AuthShell>
    );
  }

  if (state === "new-user" && info) {
    return (
      <AuthShell title={`Join ${info.org_name}.`} sub={inviteSub(info)}>
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
      body: "This invite link has expired. Ask your project admin to send a new one.",
    },
    revoked: {
      title: "Invite revoked",
      body: "This invite has been revoked. Ask your project admin to send a new one.",
    },
    "already-used": {
      title: "Invite already used",
      body: "This invite has already been accepted. Log in to access your account.",
    },
  };

  const copy = errorCopy[state] ?? errorCopy["not-found"];

  return (
    <AuthShell title={copy.title} sub={copy.body}>
      <div className="py-2 text-center">
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
      title="Already signed in."
      sub="Sign out to accept this invite as the invited user."
    >
      <Button
        variant="inverse"
        className="w-full"
        onClick={() => logoutAndRedirect(currentUrl)}
      >
        Sign out and accept invite
      </Button>
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
