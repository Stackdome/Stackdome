import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TooltipProvider } from "@/components/ui/tooltip";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { getPostgresAddon, type PostgresAddon } from "@/api/addons";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { PostgresDetailHeader } from "./components/postgres-detail-header";
import { BackupsTab } from "./components/backups-tab";

export default function PostgresDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [addon, setAddon] = useState<PostgresAddon | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  const refetch = useCallback(async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId || !id) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getPostgresAddon(orgId, id);
      setAddon(data);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void refetch();
  }, [refetch]);

  useEffect(() => {
    if (!id) return;
    const path = `/addons/postgres/${id}`;
    setCustomLabel(path, addon?.name ?? "Postgres add-on");
    setPathLoading(path, loading);
  }, [id, addon?.name, loading, setCustomLabel, setPathLoading]);

  if (loading && !addon) {
    return (
      <div className="flex items-center justify-center p-12 text-sm text-muted-foreground">
        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Loading…
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <p className="text-sm text-danger">{error}</p>
      </div>
    );
  }

  if (!addon) return null;

  return (
    <TooltipProvider>
      <div className="flex flex-col gap-6 p-6">
        <PostgresDetailHeader addon={addon} />

        <Tabs defaultValue="backups">
          <TabsList className="w-full justify-start bg-transparent border-b border-border rounded-none p-0 h-auto gap-1 px-2">
            <TabsTrigger value="backups">Backups</TabsTrigger>
          </TabsList>
          <TabsContent value="backups" className="pt-4">
            <BackupsTab addon={addon} onAddonChanged={refetch} />
          </TabsContent>
        </Tabs>
      </div>
    </TooltipProvider>
  );
}
