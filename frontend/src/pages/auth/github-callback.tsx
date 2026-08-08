import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { completeGitHubOAuth } from "@/api/auth-github";
import { getCurrentUser } from "@/api/users";
import { setAuthSession } from "@/lib/common";
import { useCurrentUser } from "@/hooks/use-current-user";
import { getErrorMessage } from "@/api/client";
import { AlertBanner } from "@/components/branded/alert-banner";

export default function GithubCallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { refresh } = useCurrentUser();
  const [error, setError] = useState<string | null>(null);
  const ranRef = useRef(false);

  useEffect(() => {
    if (ranRef.current) return; // StrictMode double-invoke guard: consume state once
    ranRef.current = true;

    const code = params.get("code");
    const state = params.get("state");
    if (!code || !state) {
      setError("Missing authorization code or state. Please try signing in again.");
      return;
    }

    (async () => {
      try {
        const { token, refreshToken } = await completeGitHubOAuth(code, state);
        if (!token) throw new Error("OAuth response missing access token");
        localStorage.setItem("authToken", token);
        if (refreshToken) localStorage.setItem("refreshToken", refreshToken);
        const user = await getCurrentUser();
        setAuthSession(token, user, refreshToken);
        await refresh();
        navigate("/dashboard", { replace: true });
      } catch (err) {
        setError(getErrorMessage(err));
      }
    })();
  }, [params, navigate, refresh]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-6 text-center">
      {error ? (
        <>
          <AlertBanner>{error}</AlertBanner>
          <Link to="/sign-in" className="text-body underline underline-offset-4">
            Back to sign in
          </Link>
        </>
      ) : (
        <p className="text-body text-muted-foreground">Completing GitHub sign-in…</p>
      )}
    </div>
  );
}
