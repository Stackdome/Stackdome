import { useParams, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Rocket, Pencil, Loader2, X } from "lucide-react";
import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import StackResourcesForm from "@/pages/stacks/components/shared/stack-resources-form";
import StackVolumesForm from "@/pages/stacks/components/shared/stack-volumes-form";
import StackResourcesDetail from "@/pages/stacks/components/detail/stack-resources-detail";
import StackVolumesDetail from "@/pages/stacks/components/detail/stack-volumes-detail";
import { StackLogsTab } from "@/pages/stacks/components/detail/logs/stack-logs-tab";
import { StackMetricsTab } from "@/pages/stacks/components/detail/metrics/stack-metrics-tab";
import type { FormStackResourceData, FormVolumeExtendedData as VolumeFormData, FormStackData } from "@/pages/stacks/schemas/form-schema";
import type { StackResource, Volume, Stack } from "@/pages/stacks/types";
import { getStackById, updateStack } from "@/api/stacks";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { getCurrentOrganizationId } from "@/helpers/common";
import type { z } from "zod";
import { convertApiResourceToFormResource, convertApiVolumeToFormVolume, convertFormStackToApiStack } from "@/pages/stacks/schemas/form-schema";
import { useToast } from "@/components/ui/use-toast";
import type { ApiStackResourceSchema, ApiVolumeSchema } from "@/pages/stacks/schemas/api-schema";

// Helper to map API build_spec to form schema shape

function mapStackResourceToFormData(resource: StackResource): FormStackResourceData {
  // Remove read-only fields before converting to form data
  const { id: _id, stack_id: _stackId, revision: _revision, ...writableResource } = resource;

  const cleanedVolumeMounts = writableResource.volume_mounts?.map((volumeMount) => {
    const { stack_resource_id: _stackResourceId, source_volume_type: _sourceVolumeType, ...cleanVolumeMount } = volumeMount;
    return cleanVolumeMount;
  });

  const resourceWithCleanedMounts = {
    ...writableResource,
    volume_mounts: cleanedVolumeMounts
  };

  return convertApiResourceToFormResource(resourceWithCleanedMounts as z.infer<typeof ApiStackResourceSchema> & { status?: unknown });
}

function mapVolumeToFormData(volume: Volume): VolumeFormData {
  // Remove read-only fields before converting to form data
  const { id: _id, ...writableVolume } = volume;
  return convertApiVolumeToFormVolume(writableVolume as z.infer<typeof ApiVolumeSchema> & { status?: unknown });
}

export default function StackDetailPage() {
  const { id } = useParams();
  const { stacks } = useStacks();
  const [fetchedStack, setFetchedStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [editFormData, setEditFormData] = useState<{
    resources: FormStackResourceData[];
    volumes: VolumeFormData[];
  } | null>(null);
  const [validationErrors, setValidationErrors] = useState<{
    resources: { [index: number]: { [field: string]: string | undefined } };
    volumes: { [index: number]: { [field: string]: string | undefined } };
  }>({ resources: {}, volumes: {} });

  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();

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

  const initializeEditForm = () => {
    if (!stackToShow) return;

    const resourcesForForm: FormStackResourceData[] = (stackToShow.spec.stack_resources || []).map(mapStackResourceToFormData);
    const volumesForForm = (stackToShow.spec?.volumes || []).map(mapVolumeToFormData);

    setEditFormData({
      resources: resourcesForForm,
      volumes: volumesForForm
    });
  };

  const handleEditToggle = () => {
    if (!isEditing) {
      initializeEditForm();
      setIsEditing(true);
    } else {
      // Cancel edit mode
      setIsEditing(false);
      setEditFormData(null);
      setValidationErrors({ resources: {}, volumes: {} });
    }
  };

  const handleSave = async () => {
    if (!stackToShow || !editFormData || !id) return;

    setIsSaving(true);
    setValidationErrors({ resources: {}, volumes: {} });

    try {
      const orgId = getCurrentOrganizationId();
      if (!orgId) {
        throw new Error("Organization ID not found");
      }

      // Convert form data to API format
      const formStackData: FormStackData = {
        name: stackToShow.name || '',
        labels: stackToShow.labels || [],
        spec: {
          stack_resources: editFormData.resources,
          volumes: editFormData.volumes
        }
      };

      const apiData = convertFormStackToApiStack(formStackData);
      const updatedStack = await updateStack(orgId, id, apiData);

      // Update local state
      setFetchedStack(updatedStack);
      setIsEditing(false);
      setEditFormData(null);

      toast({
        title: "Stack updated successfully",
        description: "Your stack configuration has been saved.",
        variant: "default"
      });

    } catch (error) {
      console.error('Failed to update stack:', error);
      toast({
        title: "Failed to update stack",
        description: error instanceof Error ? error.message : "An unexpected error occurred. Please try again.",
        variant: "destructive"
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleResourcesChange = (updatedResources: Partial<FormStackResourceData>[]) => {
    if (!editFormData) return;
    setEditFormData({
      ...editFormData,
      resources: updatedResources as FormStackResourceData[]
    });
  };

  const handleVolumesChange = (updatedVolumes: Partial<VolumeFormData>[]) => {
    if (!editFormData) return;
    setEditFormData({
      ...editFormData,
      volumes: updatedVolumes as VolumeFormData[]
    });
  };

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

  return (
    <div className="p-6">
      <header className="mb-6">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-4">
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
          </div>

          <div className="flex gap-3">
            {/* Edit Mode Toggle */}
            {isEditing ? (
              <>
                <Button
                  variant="outline"
                  size="lg"
                  onClick={handleEditToggle}
                  disabled={isSaving}
                >
                  <X className="mr-2 h-4 w-4" />
                  Cancel
                </Button>
                <Button
                  variant="default"
                  size="lg"
                  onClick={handleSave}
                  disabled={isSaving}
                >
                  {isSaving ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Deploying...
                    </>
                  ) : (
                    <>
                      <Rocket className="mr-2 h-4 w-4" />
                      Deploy
                    </>
                  )}
                </Button>
              </>
            ) : (
              <Button
                variant="outline"
                size="lg"
                onClick={handleEditToggle}
              >
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </Button>
            )}
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
            <CardHeader className="pb-3">
              <CardTitle className="text-xl">Stack Resources</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {isEditing ? (
                <StackResourcesForm
                  resources={editFormData?.resources || []}
                  onResourcesChange={handleResourcesChange}
                  errors={validationErrors.resources}
                  volumes={editFormData?.volumes || []}
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
            <CardHeader className="pb-3">
              <CardTitle className="text-xl">Stack Volumes</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {isEditing ? (
                <StackVolumesForm
                  volumes={editFormData?.volumes || []}
                  onVolumesChange={handleVolumesChange}
                  errors={validationErrors.volumes}
                  stackResources={editFormData?.resources || []}
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
              resources={stackToShow.spec.stack_resources?.map(r => ({ name: r.name || '', id: r.id || '' })) || []}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>

        {/* Metrics Tab */}
        <TabsContent value="metrics">
          {stackToShow.id ? (
            <StackMetricsTab
              stackId={stackToShow.id}
              organizationId={stackToShow.organisation_id || getCurrentOrganizationId() || ''}
              resources={stackToShow.spec.stack_resources || []}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
