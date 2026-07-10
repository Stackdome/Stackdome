import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/components/ui/use-toast";
import {
  getPreviewConfig, updatePreviewConfig, deletePreviewConfig,
  type StackPreviewConfig,
} from "@/api/preview-configs";
import { getErrorMessage } from "@/api/client";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { ConfigEnvsSection } from "./components/config-envs-section";

export default function PreviewConfigDetailPage() {
  const { configId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { defaultTeamName } = useResourceTeams();

  const [config, setConfig] = useState<StackPreviewConfig | null>(null);
  const [baseBranch, setBaseBranch] = useState("");
  const [stackfilePath, setStackfilePath] = useState("");
  const [maxActive, setMaxActive] = useState(10);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !configId) return;
    let cancelled = false;
    getPreviewConfig(orgId, defaultTeamName, configId)
      .then((cfg) => {
        if (cancelled) return;
        setConfig(cfg);
        setBaseBranch(cfg.git_repository?.base_branch ?? "");
        setStackfilePath(cfg.stackfile_path ?? "stackfile.yaml");
        setMaxActive(cfg.max_active_previews ?? 10);
      })
      .catch((e) => {
        if (!cancelled) setError(getErrorMessage(e));
      });
    return () => {
      cancelled = true;
    };
  }, [configId, defaultTeamName]);

  const save = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !configId || !config) return;
    setSaving(true);
    try {
      const updated = await updatePreviewConfig(orgId, defaultTeamName, configId, {
        git_repository: {
          repo_url: config.git_repository?.repo_url ?? "",
          base_branch: baseBranch,
        },
        stackfile_path: stackfilePath,
        max_active_previews: maxActive,
      });
      setConfig(updated);
      toast({ title: "Configuration saved" });
    } catch (e) {
      toast({ title: "Save failed", description: getErrorMessage(e), variant: "destructive" });
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !defaultTeamName || !configId) return;
    try {
      await deletePreviewConfig(orgId, defaultTeamName, configId);
      toast({ title: "Configuration deleted" });
      navigate("/previews");
    } catch (e) {
      toast({ title: "Delete failed", description: getErrorMessage(e), variant: "destructive" });
    }
  };

  if (error) return <p className="p-6 text-sm text-destructive">{error}</p>;
  if (!config) return <p className="p-6 text-sm text-muted-foreground">Loading…</p>;

  return (
    <div className="space-y-6 p-6">
      <div>
        <Button variant="ghost" size="sm" asChild className="mb-2 -ml-2">
          <Link to="/previews">
            <ArrowLeft className="h-4 w-4" />
            Preview Environments
          </Link>
        </Button>
        <h1 className="text-lg font-semibold">{config.name}</h1>
        <p className="font-mono text-xs text-muted-foreground">{config.git_repository?.repo_url}</p>
      </div>

      <div className="max-w-md space-y-4 rounded-lg border p-4">
        <div className="space-y-1.5">
          <Label htmlFor="cd-branch">Base branch</Label>
          <Input id="cd-branch" value={baseBranch} onChange={(e) => setBaseBranch(e.target.value)} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="cd-stackfile">Stackfile path</Label>
          <Input id="cd-stackfile" value={stackfilePath} onChange={(e) => setStackfilePath(e.target.value)} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="cd-max">Max active previews</Label>
          <Input
            id="cd-max"
            type="number"
            min={1}
            value={maxActive}
            onChange={(e) => {
              const n = e.target.valueAsNumber;
              setMaxActive(Number.isNaN(n) ? 1 : Math.max(1, Math.floor(n)));
            }}
            className="w-28"
          />
        </div>
        <Button onClick={() => void save()} disabled={saving}>Save</Button>
      </div>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold">Environments</h2>
        <ConfigEnvsSection config={config} />
      </section>

      <div className="max-w-md rounded-lg border border-destructive/40 p-4">
        <h2 className="text-sm font-semibold text-destructive">Danger zone</h2>
        <p className="mb-3 mt-1 text-sm text-muted-foreground">
          Deleting the configuration stops new previews for this repository.
        </p>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" size="sm">Delete configuration</Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete {config.name}?</AlertDialogTitle>
              <AlertDialogDescription>
                Existing preview environments must be deleted first; the backend
                rejects the delete otherwise.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={() => void remove()}>Delete</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  );
}
