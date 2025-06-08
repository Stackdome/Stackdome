import { useParams, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Play, Square, Rocket, Pencil, Check, Loader2 } from "lucide-react";
import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import StackResourcesForm from "@/pages/stacks/components/shared/stack-resources-form";
import StackVolumesForm from "@/pages/stacks/components/shared/stack-volumes-form";
import StackResourcesDetail from "@/pages/stacks/components/detail/stack-resources-detail";
import StackVolumesDetail from "@/pages/stacks/components/detail/stack-volumes-detail";
import { StackLogsTab } from "@/pages/stacks/components/logging";
import type { FormStackResourceData  , FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import { getStackById } from "@/api/stacks";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
import type { z } from "zod";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { convertApiResourceToFormResource, convertApiVolumeToFormVolume } from "@/pages/stacks/schemas/form-schema";
import type { ApiStackResourceSchema, ApiVolumeSchema } from "@/pages/stacks/schemas/api-schema";

// Helper to map API build_spec to form schema shape

function mapStackResourceToFormData(resource: StackResource): FormStackResourceData {
  return convertApiResourceToFormResource(resource as z.infer<typeof ApiStackResourceSchema>);
}

function mapVolumeToFormData(volume: Volume): VolumeFormData {
  return convertApiVolumeToFormVolume(volume as z.infer<typeof ApiVolumeSchema> & { status?: unknown });
}

export default function StackDetailPage() {
  const { id } = useParams();
  const { stacks } = useStacks();
  const [isRunning, setIsRunning] = useState(true);
  const [fetchedStack, setFetchedStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editingResources, setEditingResources] = useState(false);
  const [editingVolumes, setEditingVolumes] = useState(false);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  // Find the current stack in context
  const currentStack = stacks.find((stack) => stack.id === id);

  // Update breadcrumb with stack name
  useEffect(() => {
    const path = `/stacks/${id}`;

    if (currentStack) {
      // If stack is already available in context, use its name for breadcrumb
      setCustomLabel(path, currentStack.name|| 'Stack Details');
    } else if (id) {
      // Set loading state while fetching
      setPathLoading(path, true);

      const orgId = getCurrentOrganizationId();
      if (!orgId) {
        setError("Organization ID not found.");
        setPathLoading(path, false);
        return;
      }

      setLoading(true);
      setError(null);
      getStackById(orgId, id)
        .then((data) => {
          setFetchedStack(data);
          setLoading(false);
          // Update breadcrumb with fetched stack name
          setCustomLabel(path, data.name || 'Stack Details');
          setPathLoading(path, false);
        })
        .catch(() => {
          setError("Failed to load stack. Please try again later.");
          setLoading(false);
          setPathLoading(path, false);
        });
    }
  }, [currentStack, id, setCustomLabel, setPathLoading]);

  const stackToShow = currentStack || fetchedStack;

  const [resourcesErrors] = useState({}); // No errors in read-only mode
  const [volumesErrors] = useState({});

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading stack...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Error</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button asChild>
          <Link to="/stacks">Return to Stacks</Link>
        </Button>
      </div>
    );
  }

  if (!stackToShow) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Stack not found</h2>
        <p className="text-muted-foreground mb-4">The stack you're looking for doesn't exist or has been deleted.</p>
        <Button asChild>
          <Link to="/stacks">Return to Stacks</Link>
        </Button>
      </div>
    );
  }

  const resourcesForForm: FormStackResourceData[] = (stackToShow?.spec.stack_resources || []).map(mapStackResourceToFormData);
  const volumesForForm = (stackToShow.spec?.volumes || []).map(mapVolumeToFormData);

  // Mock metrics data
  const metrics = {
    cpu: "0.5%",
    memory: "256MB / 1GB",
    storage: "10GB / 20GB",
    network: "1.2 Mbps",
  };

  const toggleRunning = () => {
    setIsRunning(!isRunning);
  };

  return (
    <div className="p-6">
      <header className="mb-6">
        <div className="flex justify-between items-center">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold">{stackToShow.name}</h1>
              {/* Status label */}
              {stackToShow.status?.state && (
                <span>
                  {stackToShow.status.state.toLowerCase() === 'ready' && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800 border border-green-300">Ready</span>
                  )}
                  {stackToShow.status.state.toLowerCase() === 'pending' && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-300">Pending</span>
                  )}
                  {stackToShow.status.state.toLowerCase() === 'failed' && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800 border border-red-300">Failed</span>
                  )}
                  {!['ready','pending','failed'].includes(stackToShow.status.state.toLowerCase()) && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 border border-gray-300">{stackToShow.status.state}</span>
                  )}
                </span>
              )}
            </div>
            <div className="flex items-center gap-4 text-muted-foreground text-sm mb-1">
              <span>Services: {stackToShow.spec?.stack_resources?.length || 0}</span>
              <span>Volumes: {stackToShow.spec?.volumes?.length || 0}</span>
            </div>
          </div>
          <div className="flex gap-3">
            <Button
              variant="outline"
              size="lg"
              onClick={toggleRunning}
              className={isRunning ? "border-red-200 text-red-600 hover:bg-red-50 hover:text-red-700" : "border-green-200 text-green-600 hover:bg-green-50 hover:text-green-700"}
            >
              {isRunning ? <Square className="mr-2 h-4 w-4" /> : <Play className="mr-2 h-4 w-4" />}
              {isRunning ? "Stop Stack" : "Start Stack"}
            </Button>
            <Button variant="default" size="lg">
              <Rocket className="mr-2 h-4 w-4" />
              <span className="font-semibold">Deploy</span>
            </Button>
          </div>
        </div>
        <Separator className="mt-4" />
      </header>

      <Tabs defaultValue="configuration" className="w-full">
        <TabsList className="mb-6 w-full justify-start">
          <TabsTrigger value="configuration" className="flex-1">Configuration</TabsTrigger>
          <TabsTrigger value="logs" className="flex-1">Logs</TabsTrigger>
          <TabsTrigger value="metrics" className="flex-1">Metrics</TabsTrigger>
        </TabsList>

        {/* Configuration Tab: Stack Resources and Volumes */}
        <TabsContent value="configuration" className="space-y-8">
          <Card className="mb-6 rounded-lg">
            <CardHeader className="pb-3 flex flex-row justify-between items-center">
              <CardTitle className="text-xl">Stack Resources</CardTitle>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setEditingResources(!editingResources)}
                className="ml-auto"
              >
                {editingResources ? (
                  <>
                    <Check className="mr-2 h-4 w-4" />
                    <span>Done</span>
                  </>
                ) : (
                  <>
                    <Pencil className="mr-2 h-4 w-4" />
                    <span>Edit</span>
                  </>
                )}
              </Button>
            </CardHeader>
            <CardContent className="p-0">
              {editingResources ? (
                <StackResourcesForm
                  resources={resourcesForForm}
                  onResourcesChange={() => {}}
                  errors={resourcesErrors}
                  volumes={volumesForForm}
                  accordionDefaultOpen={false}
                />
              ) : (
                <StackResourcesDetail
                  resources={resourcesForForm}
                  accordionDefaultOpen={false}
                />
              )}
            </CardContent>
          </Card>
          <Card className="mb-6 rounded-lg">
            <CardHeader className="pb-3 flex flex-row justify-between items-center">
              <CardTitle className="text-xl">Stack Volumes</CardTitle>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setEditingVolumes(!editingVolumes)}
                className="ml-auto"
              >
                {editingVolumes ? (
                  <>
                    <Check className="mr-2 h-4 w-4" />
                    <span>Done</span>
                  </>
                ) : (
                  <>
                    <Pencil className="mr-2 h-4 w-4" />
                    <span>Edit</span>
                  </>
                )}
              </Button>
            </CardHeader>
            <CardContent className="p-0">
              {editingVolumes ? (
                <StackVolumesForm
                  volumes={volumesForForm}
                  onVolumesChange={() => {}}
                  errors={volumesErrors}
                  stackResources={resourcesForForm}
                  accordionDefaultOpen={false}
                />
              ) : (
                <StackVolumesDetail
                  volumes={volumesForForm}
                  stackResources={resourcesForForm}
                  accordionDefaultOpen={false}
                />
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Logs Tab */}
        <TabsContent value="logs">
          {stackToShow.id ? (
            <StackLogsTab
              stackId={stackToShow.id}
              organizationId={stackToShow.organisation_id || getCurrentOrganizationId() || ''}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>

        {/* Metrics Tab */}
        <TabsContent value="metrics">
          <div className="grid grid-cols-3 gap-4 mb-6">
            {["frontend", "backend", "database"].map((service) => (
              <Card key={service} className="overflow-hidden">
                <CardHeader className="bg-gray-50 pb-3">
                  <div className="flex justify-between">
                    <CardTitle className="text-sm font-medium">{service}</CardTitle>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <div className="flex items-center">
                          <div className="h-2 w-2 rounded-full mr-1 bg-green-500"></div>
                          <span className="text-xs">Running</span>
                        </div>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="capitalize">Running</p>
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </CardHeader>
                <CardContent className="pt-3">
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">CPU:</span>
                      <span>{metrics.cpu}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Memory:</span>
                      <span>{metrics.memory}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Storage:</span>
                      <span>{metrics.storage}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
