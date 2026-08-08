import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAppConfig } from "@/hooks/use-app-config";
import { githubOAuthUrl } from "@/api/auth-github";

interface GitHubSignInButtonProps {
  inviteToken?: string;
}

// Outline, not filled — the email/password submit below is the screen's one
// filled control (it works with no OAuth provider configured; this doesn't).
export function GitHubSignInButton({ inviteToken }: GitHubSignInButtonProps) {
  const { githubOAuth } = useAppConfig();
  if (!githubOAuth) return null;

  return (
    <div>
      <Button
        type="button"
        variant="outline"
        className="w-full"
        onClick={() => window.location.assign(githubOAuthUrl(inviteToken))}
      >
        <Github className="h-4 w-4" />
        Continue with GitHub
      </Button>
      <div className="my-4 flex items-center gap-3">
        <div className="flex-1 border-t border-border" />
        <span className="text-meta text-muted-foreground">or</span>
        <div className="flex-1 border-t border-border" />
      </div>
    </div>
  );
}
