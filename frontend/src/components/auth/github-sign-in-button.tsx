import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAppConfig } from "@/hooks/use-app-config";
import { githubOAuthUrl } from "@/api/auth-github";

interface GitHubSignInButtonProps {
  inviteToken?: string;
}

// GitHub is the filled primary; the email submit below it is a ghost of the
// same geometry, so the screen keeps exactly one primary action.
export function GitHubSignInButton({ inviteToken }: GitHubSignInButtonProps) {
  const { githubOAuth } = useAppConfig();
  if (!githubOAuth) return null;

  return (
    <div>
      <Button
        type="button"
        variant="inverse"
        className="w-full"
        onClick={() => window.location.assign(githubOAuthUrl(inviteToken))}
      >
        <Github className="h-4 w-4" />
        Continue with GitHub
      </Button>
      <div className="my-4 flex items-center gap-3">
        <div className="flex-1 border-t border-border" />
        <span className="text-xs text-muted-foreground">or</span>
        <div className="flex-1 border-t border-border" />
      </div>
    </div>
  );
}
