import { Layers, PlusCircle, Loader2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { getStacksByOrg } from "@/api/stacks";
import type { Stack } from "@/pages/stacks/types";
import {
  Card,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { formatDistanceToNow } from 'date-fns';
import { getCurrentOrganizationId } from "@/helpers/common";

export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

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
          setError("Failed to load stacks. Please try again later.");
        }
        setIsLoading(false);
      };

      fetchStacks();
    } else {
      setError("Organization ID not found. Unable to load stacks.");
      setIsLoading(false);
    }
  }, []);

  const handleCreateNewStack = () => {
    navigate("/stacks/create");
  };

  if (isLoading) {
    return (
      <div className="flex flex-1 flex-col p-4 pt-0 h-full items-center justify-center">
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
      <div className="flex flex-1 flex-col p-4 pt-0 h-full">
        {stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-[80vh] text-center">
            <div className="flex flex-col items-center max-w-md">
              <div className="rounded-full bg-primary/10 p-4 mb-4">
                <Layers className="h-10 w-10 text-primary" />
              </div>
              <h2 className="text-2xl font-bold mb-2">No stacks deployed yet</h2>
              <p className="text-muted-foreground mb-6">
                Deploy your first stack to get started.
              </p>
              <Button onClick={handleCreateNewStack}>
                <PlusCircle className="mr-2 h-4 w-4" />
                Create New Stack
              </Button>
            </div>
          </div>
        ) : (
          <>
            <div className="flex justify-between items-center py-4">
              <h1 className="text-2xl font-semibold">Stacks</h1>
              <Button onClick={handleCreateNewStack} variant="outline">
                <PlusCircle className="mr-2 h-4 w-4" />
                Add New Stack
              </Button>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5 gap-4 mt-0">
              {stacks.map((stack) => {
                const status = stack.status?.state?.toLowerCase() || 'unknown';
                let circleColor = 'bg-gray-500'; // Default for unknown
                if (status === 'ready') {
                  circleColor = 'bg-green-500';
                } else if (status === 'pending') {
                  circleColor = 'bg-yellow-500';
                } else if (status === 'failed') {
                  circleColor = 'bg-red-500';
                }

                return (
                  <Card key={stack.id || stack.name} className="flex flex-col w-full min-h-[130px] hover:shadow-lg dark:hover:shadow-[0_10px_15px_-3px_rgba(200,200,200,0.15),0_4px_6px_-4px_rgba(200,200,200,0.12)] transition-shadow duration-200">
                    <CardHeader className="pb-2">
                      <CardTitle className="flex items-center justify-between">
                        <Link to={`/stacks/${stack.id}`} className="hover:underline truncate pr-2" title={stack.name}>
                          {stack.name}
                        </Link>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className={`h-3 w-3 rounded-full ${circleColor}`} />
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="capitalize">{status}</p>
                          </TooltipContent>
                        </Tooltip>
                      </CardTitle>
                    </CardHeader>
                    <CardFooter className="flex justify-between items-baseline text-xs text-muted-foreground mt-auto pt-2 pb-3">
                      <div className="flex flex-col">
                        <span>
                          Resources: {stack.spec.stack_resources?.length || 0}
                        </span>
                        <span>
                          Volumes: {stack.spec.volumes?.length || 0}
                        </span>
                      </div>
                      <span className="text-right">
                        {stack.created_at ? formatDistanceToNow(new Date(stack.created_at), { addSuffix: true }).replace(/^about\s/, '') : 'N/A'}
                      </span>
                    </CardFooter>
                  </Card>
                );
              })}
            </div>
          </>
        )}
      </div>
    </TooltipProvider>
  );
}
