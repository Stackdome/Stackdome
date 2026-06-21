import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PlusCircle, Loader2, AlertCircle, Puzzle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, EmptyState } from "@/components/branded";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useCurrentUser } from "@/hooks/use-current-user";
import { usePostgresAddons } from "./hooks/use-postgres-addons";
import { AddonList } from "./components/addon-list";
import { AddonTypePickerDialog, type AddonType } from "./components/addon-type-picker-dialog";

export default function AddonsPage() {
  const navigate = useNavigate();
  const { addons, loading, error, refetch } = usePostgresAddons();
  const { canWriteAnyTeam, canWrite } = useCurrentUser();
  const [pickerOpen, setPickerOpen] = useState(false);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  useEffect(() => {
    setCustomLabel("/addons", "Addons");
    setPathLoading("/addons", loading);
  }, [setCustomLabel, setPathLoading, loading]);

  function handlePickType(type: AddonType) {
    setPickerOpen(false);
    if (type === "postgres") {
      navigate("/addons/create/postgres");
    }
  }

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading addons...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading Addons</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => refetch()}>Try Again</Button>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title="Addons"
          subtitle="Manage hosted addon services for your stacks"
          actions={
            canWriteAnyTeam ? (
              <Button onClick={() => setPickerOpen(true)}>
                <PlusCircle className="h-4 w-4" />
                Add Addon
              </Button>
            ) : undefined
          }
        />

        {addons.length === 0 ? (
          <EmptyState
            icon={<Puzzle className="h-8 w-8" />}
            title="No addons yet"
            description="Add an addon to provision a managed Postgres for your stacks."
            action={
              canWriteAnyTeam ? (
                <Button onClick={() => setPickerOpen(true)}>
                  <PlusCircle className="h-4 w-4" />
                  Add Addon
                </Button>
              ) : undefined
            }
          />
        ) : (
          <AddonList
            addons={addons}
            canWrite={(teamId?: string) => canWrite(teamId ?? "")}
          />
        )}

        <AddonTypePickerDialog
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          onSelect={handlePickType}
        />
      </div>
    </TooltipProvider>
  );
}
