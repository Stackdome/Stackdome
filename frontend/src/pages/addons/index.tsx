import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PlusCircle, Loader2, AlertCircle, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useToast } from "@/components/ui/use-toast";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage, isErrorStatus } from "@/api/client";
import { deletePostgresAddon, type PostgresAddon } from "@/api/addons";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { usePostgresAddons } from "./hooks/use-postgres-addons";
import { AddonTable } from "./components/addon-table";
import { AddonTypePickerDialog, type AddonType } from "./components/addon-type-picker-dialog";
import { DeleteAddonDialog } from "./components/delete-addon-dialog";

export default function AddonsPage() {
  const navigate = useNavigate();
  const { addons, loading, error, refetch } = usePostgresAddons();
  const [pickerOpen, setPickerOpen] = useState(false);
  const [deletingAddon, setDeletingAddon] = useState<PostgresAddon | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const { toast } = useToast();
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

  async function handleDeleteConfirm() {
    if (!deletingAddon?.id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;

    setDeleteLoading(true);
    setDeleteError(null);
    try {
      await deletePostgresAddon(orgId, deletingAddon.id);
      toast({
        title: "Addon deleted",
        description: `"${deletingAddon.name}" is being torn down.`,
        variant: "destructive",
      });
      setDeletingAddon(null);
      refetch();
    } catch (e) {
      console.error("Failed to delete addon:", e);
      if (isErrorStatus(e, 409)) {
        setDeleteError(
          `${getErrorMessage(e)}\n\nRemove the stack references first, then try again.`,
        );
      } else {
        setDeleteError(getErrorMessage(e));
      }
    } finally {
      setDeleteLoading(false);
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
      <div className="p-6">
        <header className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3 mb-1">
                <h1 className="text-2xl font-bold">Addons Management</h1>
              </div>
              <p className="text-muted-foreground">
                Manage hosted Addon services for your stacks.
              </p>
            </div>
            <Button onClick={() => setPickerOpen(true)}>
              <PlusCircle className="mr-2 h-4 w-4" />
              Add Addon
            </Button>
          </div>
          <Separator className="mt-4" />
        </header>

        <Card className="rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl flex items-center gap-2">
              <Package className="h-5 w-5" />
              Addons
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {addons.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20">
                <Package className="h-12 w-12 mb-4 text-muted-foreground" />
                <h3 className="text-xl font-medium mb-2">No addons yet</h3>
                <p className="text-muted-foreground mb-6">
                  Add an addon to provision a managed Postgres for your stacks.
                </p>
                <Button onClick={() => setPickerOpen(true)}>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Add Addon
                </Button>
              </div>
            ) : (
              <AddonTable addons={addons} onDelete={(a) => setDeletingAddon(a)} />
            )}
          </CardContent>
        </Card>

        <AddonTypePickerDialog
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          onSelect={handlePickType}
        />

        <DeleteAddonDialog
          open={!!deletingAddon}
          addonName={deletingAddon?.name}
          loading={deleteLoading}
          error={deleteError}
          onConfirm={handleDeleteConfirm}
          onCancel={() => {
            if (!deleteLoading) {
              setDeletingAddon(null);
              setDeleteError(null);
            }
          }}
        />
      </div>
    </TooltipProvider>
  );
}
