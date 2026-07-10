import { Github, Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGithubConnect } from "@/pages/previews/hooks/use-github-connect";
import { useEffect } from "react";

interface ConnectPhaseProps {
  onConnected: (integrationId: string | null) => void;
  onCancel: () => void;
  onSkip?: () => void;
}

export function ConnectPhase({ onConnected, onCancel, onSkip }: ConnectPhaseProps) {
  const { state, error, connect, checkAgain, integrationId } = useGithubConnect();

  useEffect(() => {
    if (state === "connected") onConnected(integrationId);
  }, [state, integrationId, onConnected]);

  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 px-8 text-center">
      <Github className="h-10 w-10 text-muted-foreground" />
      <div>
        <h3 className="text-base font-semibold">Connect GitHub</h3>
        <p className="mt-1 max-w-md text-sm text-muted-foreground">
          Install the Stackdome GitHub App to grant repository access. A popup
          will walk you through creating and installing the app.
        </p>
      </div>

      {state === "idle" || state === "error" ? (
        <Button onClick={() => void connect()}>
          <Github className="h-4 w-4" />
          Connect GitHub
        </Button>
      ) : null}

      {state === "waiting" && (
        <div className="flex flex-col items-center gap-3">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Waiting for installation to finish in the popup…
          </div>
          <Button variant="outline" size="sm" onClick={() => void checkAgain()}>
            <RefreshCw className="h-4 w-4" />
            I&apos;ve installed it — check again
          </Button>
        </div>
      )}

      {error && <p className="text-sm text-destructive">{error}</p>}

      {onSkip && (
        <Button variant="link" size="sm" className="text-muted-foreground" onClick={onSkip}>
          Skip — use a public repository URL instead
        </Button>
      )}

      <Button variant="ghost" size="sm" onClick={onCancel}>
        Cancel
      </Button>
    </div>
  );
}
