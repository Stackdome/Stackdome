import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAppConfig } from "@/hooks/use-app-config";
import { githubOAuthUrl } from "@/api/auth-github";

interface GitHubSignInButtonProps {
  inviteToken?: string;
}

export function GitHubSignInButton({ inviteToken }: GitHubSignInButtonProps) {
  const { githubOAuth } = useAppConfig();
  if (!githubOAuth) return null;

  return (
    <div>
      <div className="my-4 flex items-center gap-3">
        <div className="flex-1 border-t border-border" />
        <span className="whitespace-nowrap font-mono text-[10px] uppercase tracking-[1.5px] text-muted-foreground">
          or continue with
        </span>
        <div className="flex-1 border-t border-border" />
      </div>
      <Button
        type="button"
        variant="outline"
        className="w-full"
        onClick={() => window.location.assign(githubOAuthUrl(inviteToken))}
      >
        <Github className="h-4 w-4" />
        GitHub
      </Button>
    </div>
  );
}
