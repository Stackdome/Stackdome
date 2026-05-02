import { Layers, PlusCircle, Loader2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { getStacksByOrg } from "@/api/stacks";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import {
  Card,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  TooltipProvider,
} from "@/components/ui/tooltip";
import { PageHeader, EmptyState, StatusPill, variantFromState } from "@/components/branded";
import { formatDistanceToNow } from 'date-fns';
import { DockerComposeImportDropdown } from "@/pages/stacks/components/shared/import-dropdown";
import DockerComposeImportDialog from "@/pages/stacks/components/shared/docker-compose-import-dialog";
import { useDockerComposeImport } from "@/pages/stacks/hooks/use-docker-compose-import";

export default function StacksPage() {
  const { stacks, setStacks } = useStacks();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  // Import functionality
  const {
    isLoading: isImportLoading,
    error: importError,
    isDialogOpen: isImportDialogOpen,
    openDialog: openImportDialog,
    closeDialog: closeImportDialog,
    handleImport,
    clearError: clearImportError,
  } = useDockerComposeImport();

  useEffect(() => {
    const currentOrgId = getCurrentOrganizationId();

    if (currentOrgId) {
      const fetchStacks = async () => {
        setIsLoading(true);
        setError(null);
        try {
          const data = await getStacksByOrg(currentOrgId);
          setStacks(data.items || []);
        } catch (err) {
          console.error("Failed to fetch stacks:", err);
          setError(getErrorMessage(err));
        }
        setIsLoading(false);
      };

      fetchStacks();
    } else {
      setError("Organization ID not found. Unable to load stacks.");
      setIsLoading(false);
    }
  }, [setStacks]);

  const handleCreateNewStack = () => {
    navigate("/stacks/create");
  };

  if (isLoading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading stacks...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-1 flex-col p-4 pt-0 h-full items-center justify-center text-center">
        <AlertTriangle className="h-10 w-10 text-destructive mb-4" />
        <h2 className="text-2xl font-bold mb-2">Error</h2>
        <p className="text-muted-foreground mb-6">{error}</p>
        <Button onClick={() => window.location.reload()}>Try Again</Button>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="flex flex-1 flex-col p-8 space-y-8 h-full">
        <PageHeader
          eyebrow="Platform"
          title="Stacks"
          subtitle={`${stacks.length} ${stacks.length === 1 ? "stack" : "stacks"} deployed`}
          actions={
            <>
              <DockerComposeImportDropdown
                onDockerComposeImport={openImportDialog}
                variant="outline"
              />
              <Button onClick={handleCreateNewStack} className="bg-brand text-white hover:bg-brand-darker">
                <PlusCircle className="h-4 w-4" />
                New Stack
              </Button>
            </>
          }
        />

        {stacks.length === 0 ? (
          <EmptyState
            icon={<Layers className="h-8 w-8" />}
            title="No stacks deployed yet"
            description="Deploy your first stack to get started."
            action={
              <Button onClick={handleCreateNewStack} variant="outline">
                <PlusCircle className="h-4 w-4" />
                Create New Stack
              </Button>
            }
          />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5 gap-4">
            {stacks.map((stack) => {
              const variant = variantFromState(stack.status?.state);
              return (
                <Link
                  key={stack.id || stack.name}
                  to={`/stacks/${stack.id}`}
                  className="block group"
                >
                  <Card className="flex flex-col w-full hover:border-brand-border hover:bg-muted/30 transition-colors duration-150 py-4 gap-3">
                    <CardHeader className="px-4 pb-0">
                      <CardTitle className="flex items-start justify-between gap-2 text-base font-medium">
                        <span className="truncate group-hover:text-brand transition-colors" title={stack.name}>
                          {stack.name}
                        </span>
                        {stack.status?.state && (
                          <StatusPill variant={variant} className="shrink-0">
                            {stack.status.state}
                          </StatusPill>
                        )}
                      </CardTitle>
                    </CardHeader>
                    <CardFooter className="px-4 pb-0 flex justify-between items-end gap-2 font-mono text-[11px] text-muted-foreground whitespace-nowrap">
                      <div className="flex flex-col gap-0.5 tabular-nums">
                        <span>{stack.spec.stack_resources?.length || 0} resources</span>
                        <span>{stack.spec.volumes?.length || 0} volumes</span>
                      </div>
                      <span className="uppercase tracking-[0.5px] text-right">
                        {stack.created_at ? formatDistanceToNow(new Date(stack.created_at), { addSuffix: true }).replace(/^about\s/, '') : 'N/A'}
                      </span>
                    </CardFooter>
                  </Card>
                </Link>
              );
            })}
          </div>
        )}

        {/* Docker Compose Import Dialog */}
        <DockerComposeImportDialog
          open={isImportDialogOpen}
          onOpenChange={closeImportDialog}
          onImport={handleImport}
          isLoading={isImportLoading}
          error={importError}
          onClearError={clearImportError}
        />
      </div>
    </TooltipProvider>
  );
}
