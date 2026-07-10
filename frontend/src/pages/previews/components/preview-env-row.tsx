import { Link } from "react-router-dom";
import { ExternalLink, Layers, RefreshCw, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { PreviewStack, PreviewPhase } from "@/api/preview-envs";

interface PreviewEnvRowProps {
  env: PreviewStack;
  onSync: (env: PreviewStack) => void;
  onDelete: (env: PreviewStack) => void;
}

function phaseVariant(phase: PreviewPhase | undefined): "default" | "secondary" | "destructive" | "outline" {
  switch (phase) {
    case "Ready":
      return "default";
    case "Failed":
      return "destructive";
    default:
      return "secondary";
  }
}

export function PreviewEnvRow({ env, onSync, onDelete }: PreviewEnvRowProps) {
  const phase = env.status?.phase;
  const urls = env.status?.outputs?.urls ?? [];
  const failed = phase === "Failed";
  const stackfileHint = failed && /stackfile/i.test(env.status?.reason ?? "");

  return (
    <div className="rounded-md border px-3 py-2.5">
      <div className="flex items-center gap-3">
        <span className="font-mono text-xs font-semibold">PR #{env.pr_number}</span>
        <span className="truncate text-xs text-muted-foreground">{env.branch}</span>
        {env.commit && (
          <span className="font-mono text-[11px] text-muted-foreground">{env.commit.slice(0, 7)}</span>
        )}
        <Badge variant={phaseVariant(phase)}>{phase ?? "Unknown"}</Badge>
        {env.source && (
          <Badge variant="outline" className="text-[10px]">{env.source}</Badge>
        )}
        <span className="flex-1" />
        {urls.map((u) => (
          <a
            key={u.url}
            href={u.url}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
          >
            <ExternalLink className="h-3 w-3" />
            {u.resource}
          </a>
        ))}
        {env.stack_id && (
          <Button variant="ghost" size="icon" asChild>
            <Link to={`/stacks/${env.stack_id}`} aria-label={`View stack for PR #${env.pr_number}`}>
              <Layers className="h-4 w-4" />
            </Link>
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Sync PR #${env.pr_number}`}
          onClick={() => onSync(env)}
          disabled={phase === "Deleting"}
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Delete PR #${env.pr_number}`}
          onClick={() => onDelete(env)}
          disabled={phase === "Deleting"}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
      {failed && (
        <div className="mt-2 rounded-sm bg-destructive/10 px-2 py-1.5 text-xs text-destructive">
          <span className="font-semibold">{env.status?.reason}:</span> {env.status?.message}
          {stackfileHint && (
            <span className="block text-muted-foreground">
              Check the stackfile path in the repository configuration.
            </span>
          )}
        </div>
      )}
    </div>
  );
}
