import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { GitPullRequest, RefreshCw, Settings2, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/components/ui/use-toast";
import {
  getPreviewEnv, deletePreviewEnv,
  PREVIEW_STACK_LABEL, PREVIEW_ID_LABEL,
  type PreviewStack,
} from "@/api/preview-envs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { SyncEnvDialog } from "@/pages/previews/components/sync-env-dialog";
import type { Label } from "@/pages/stacks/types";

interface PreviewEnvBannerProps {
  labels?: Label[];
  teamId?: string;
}

/**
 * Shown on the stack show page when the stack was created by a preview
 * environment (detected from backend-stamped labels). Hosts the env-specific
 * actions (sync / delete / configuration link) — the previews list is
 * read-only by design.
 */
export function PreviewEnvBanner({ labels, teamId }: PreviewEnvBannerProps) {
  const { teamNameById, defaultTeamName } = useResourceTeams();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [env, setEnv] = useState<PreviewStack | null>(null);
  const [syncing, setSyncing] = useState<PreviewStack | null>(null);

  const isPreview = labels?.some((l) => l.key === PREVIEW_STACK_LABEL && l.value === "true") ?? false;
  const previewId = labels?.find((l) => l.key === PREVIEW_ID_LABEL)?.value;
  const teamName = teamNameById(teamId) ?? defaultTeamName;

  useEffect(() => {
    if (!isPreview || !previewId || !teamName) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    let cancelled = false;
    getPreviewEnv(orgId, teamName, previewId)
      .then((e) => {
        if (!cancelled) setEnv(e);
      })
      .catch(() => {
        // Env record may already be deprovisioning; banner degrades to label-only.
      });
    return () => {
      cancelled = true;
    };
  }, [isPreview, previewId, teamName]);

  if (!isPreview) return null;

  const remove = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !teamName || !env?.id) return;
    try {
      await deletePreviewEnv(orgId, teamName, env.id);
      toast({ title: `Deleting PR #${env.pr_number} environment` });
      navigate("/stacks?view=previews");
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  return (
    <div className="mt-3 flex flex-wrap items-center gap-3 rounded-md border border-brand-border bg-brand-bg px-3 py-2">
      <GitPullRequest className="h-4 w-4 flex-none text-brand" />
      <span className="text-[13px] font-medium">
        Preview environment{env?.pr_number ? ` · PR #${env.pr_number}` : ""}
      </span>
      {env?.branch && <span className="font-mono text-[11px] text-muted-foreground">{env.branch}</span>}
      {env?.commit && <span className="font-mono text-[11px] text-muted-foreground">{env.commit.slice(0, 7)}</span>}
      {env?.status?.phase && <Badge variant="outline" className="text-[10px]">{env.status.phase}</Badge>}
      <span className="text-[12px] text-muted-foreground">
        Managed by preview environments — manual changes are overwritten on sync.
      </span>
      <span className="flex-1" />
      {env && (
        <>
          <Button variant="ghost" size="sm" onClick={() => setSyncing(env)}>
            <RefreshCw className="h-3.5 w-3.5" />
            Sync
          </Button>
          {env.config_id && (
            <Button variant="ghost" size="sm" asChild>
              <Link to={`/previews/${env.config_id}`}>
                <Settings2 className="h-3.5 w-3.5" />
                Configuration
              </Link>
            </Button>
          )}
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive">
                <Trash2 className="h-3.5 w-3.5" />
                Delete environment
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete PR #{env.pr_number} environment?</AlertDialogTitle>
                <AlertDialogDescription>
                  The environment&apos;s stack and resources are torn down. This cannot
                  be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={() => void remove()}>Delete</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </>
      )}
      <SyncEnvDialog
        env={syncing}
        onOpenChange={(o) => !o && setSyncing(null)}
        onSynced={() => {
          toast({ title: "Sync started" });
        }}
      />
    </div>
  );
}
