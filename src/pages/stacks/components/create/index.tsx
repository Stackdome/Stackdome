import React, { useState, useCallback, Fragment } from "react";
import { useNavigate } from "react-router-dom";
import StackResourcesForm from "./stack-resources-form";
import StackVolumesForm from "./stack-volumes-form";
import { Button } from "@/components/ui/button";
import { Rocket, Tag as TagIcon, X, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label as UILabel } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  StackSchema,
  type StackData,
  type StackResourceData,
  type VolumeFormData,
  stripUIFieldsFromStackData,
} from "@/pages/stacks/schemas/stack-create-schema";
import { createStack } from '@/api/stacks';
import { getCurrentOrganizationId } from '@/helpers/common';
import { useToast } from '@/components/ui/use-toast';

type FormErrors = { [path: string]: string | undefined };

export default function StackCreatePage() {
  const [formData, setFormData] = useState<Partial<StackData>>({
    name: "",
    workspace_name: "default",
    labels: [],
    spec: {
      stack_resources: [],
      volumes: [],
    },
  });
  const [currentLabelInput, setCurrentLabelInput] = useState("");
  const [formErrors, setFormErrors] = useState<FormErrors>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const { toast } = useToast();

  const handleChange = (path: string, value: string | number | boolean | object | null) => {
    setFormData(prev => {
      const newState = JSON.parse(JSON.stringify(prev)) as Partial<StackData>;
      setNestedValue(newState, path, value);
      return newState;
    });
    if (formErrors[path]) {
      setFormErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[path];
        return newErrors;
      });
    }
  };

  const handleLabelInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentLabelInput(e.target.value);
  };

  const handleAddLabel = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && currentLabelInput.trim()) {
      e.preventDefault();
      const labelStr = currentLabelInput.trim();
      const parts = labelStr.split('=');
      const key = parts[0].trim();
      const value = parts.length > 1 ? parts.slice(1).join('=').trim() : "";

      if (key) {
        setFormData(prev => ({
          ...prev,
          labels: [...(prev.labels || []), { key, value }],
        }));
        setCurrentLabelInput("");
        if (formErrors["labels"]) {
          setFormErrors(prev => {
            const newErrors = { ...prev };
            delete newErrors["labels"];
            return newErrors;
          });
        }
      }
    }
  };

  const removeLabel = (indexToRemove: number) => {
    setFormData(prev => ({
      ...prev,
      labels: (prev.labels || []).filter((_, idx) => idx !== indexToRemove),
    }));
    const errorPathsToClear = [`labels.${indexToRemove}.key`, `labels.${indexToRemove}.value`, `labels.${indexToRemove}`];
    setFormErrors(prev => {
      const newErrors = { ...prev };
      errorPathsToClear.forEach(path => delete newErrors[path]);
      return newErrors;
    });
  };

  const handleResourcesChange = useCallback((updatedResources: Partial<StackResourceData>[]) => {
    setFormData(prev => ({
      ...prev,
      spec: {
        ...(prev.spec || {}),
        stack_resources: updatedResources as StackResourceData[],
      }
    }));
    if (formErrors["spec.stack_resources"]) {
      setFormErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors["spec.stack_resources"];
        Object.keys(newErrors).forEach(key => {
          if (key.startsWith("spec.stack_resources.")) {
            delete newErrors[key];
          }
        });
        return newErrors;
      });
    }
  }, [formErrors]);

  const handleVolumesChange = useCallback((updatedVolumes: Partial<VolumeFormData>[]) => {
    setFormData(prev => {
      const newFormData = { ...prev };

      if (!newFormData.spec) {
        newFormData.spec = {
          stack_resources: [],
          volumes: []
        };
      }

      newFormData.spec = {
        ...newFormData.spec,
        volumes: updatedVolumes.map(vol => ({
          name: vol.name || '',
          workspace_name: vol.workspace_name || newFormData.workspace_name || 'default',
          labels: vol.labels,
          annotations: vol.annotations,
          spec: {
            size: vol.spec?.size || '',
            storage_class: vol.spec?.storage_class,
            needs_sync_before_use: vol.spec?.needs_sync_before_use || false,
            access_mode: vol.spec?.access_mode || 'ReadWriteOnce',
            source: vol.spec?.source
          }
        }))
      };

      return newFormData;
    });

    if (formErrors["spec.volumes"]) {
      setFormErrors((prev: FormErrors) => {
        const newErrors: FormErrors = { ...prev };
        delete newErrors["spec.volumes"];
        Object.keys(newErrors).forEach((key) => {
          if (key.startsWith("spec.volumes.")) {
            delete newErrors[key];
          }
        });
        return newErrors;
      });
    }
  }, [formErrors]);

  const handleSubmit = async () => {
    setIsLoading(true);
    setApiError(null);

    const payloadToValidate: StackData = {
      // ...existing code
      name: formData.name || "",
      workspace_name: formData.workspace_name || "default",
      labels: formData.labels || [],
      spec: {
        stack_resources: (formData.spec?.stack_resources || []).map(sr => {
          const resource: StackResourceData = {
            name: sr.name || "",
            sourceType: sr.sourceType || 'image',
            labels: sr.labels?.length ? sr.labels : undefined,
            depends_on: sr.depends_on?.length ? sr.depends_on : undefined,
            ports: sr.ports?.length ? sr.ports : undefined,
            execution_config: sr.execution_config && (sr.execution_config.command?.length || sr.execution_config.args?.length || sr.execution_config.environment_variables?.length) ? {
              command: sr.execution_config.command?.length ? sr.execution_config.command : undefined,
              args: sr.execution_config.args?.length ? sr.execution_config.args : undefined,
              environment_variables: sr.execution_config.environment_variables?.length ? sr.execution_config.environment_variables : undefined,
            } : undefined,
            init_spec: sr.init_spec && (sr.init_spec.command?.length || sr.init_spec.args?.length || sr.init_spec.image_spec?.image) ? {
              command: sr.init_spec.command?.length ? sr.init_spec.command : undefined,
              args: sr.init_spec.args?.length ? sr.init_spec.args : undefined,
              image_spec: sr.init_spec.image_spec?.image ? { image: sr.init_spec.image_spec.image } : undefined,
            } : undefined,
          };

          if (sr.sourceType === 'image') {
            // Always include image_spec in validation payload, even if empty,
            // so validation catches required fields
            resource.image_spec = {
              image: sr.image_spec?.image || ""
            };
            resource.build_spec = undefined;
          } else if (sr.sourceType === 'git') {
            // Map git revision UI fields to OpenAPI structure
            let git_repo_revision: {commit?: string; branch?: {name: string}; tag?: string} | undefined = undefined;

            // Ensure we have gitRevisionType and gitRevisionValue for validation
            // Setting to empty strings if undefined to ensure validation catches them
            const revType: "commit" | "branch" | "tag" | undefined = sr.gitRevisionType || undefined;
            const revValue = sr.gitRevisionValue || "";

            // Set these properties directly on the resource object to ensure validation
            // can access and validate them
            resource.gitRevisionType = revType;
            resource.gitRevisionValue = revValue;

            if (sr.gitRevisionType && sr.gitRevisionValue) {
              if (sr.gitRevisionType === 'commit') {
                git_repo_revision = { commit: sr.gitRevisionValue };
              } else if (sr.gitRevisionType === 'branch') {
                git_repo_revision = { branch: { name: sr.gitRevisionValue } };
              } else if (sr.gitRevisionType === 'tag') {
                git_repo_revision = { tag: sr.gitRevisionValue };
              }
            }
            resource.build_spec = sr.build_spec ? {
              source_context: {
                git_repo: sr.build_spec.source_context?.git_repo?.repo_url
                  ? { repo_url: sr.build_spec.source_context.git_repo.repo_url }
                  : undefined,
              },
              context_path_within_source: sr.build_spec.context_path_within_source || "./",
              dockerfile_path: sr.build_spec.dockerfile_path || "Dockerfile",
              image_repository_url: sr.build_spec.image_repository_url?.url
                ? { url: sr.build_spec.image_repository_url.url, cluster_registry_id: sr.build_spec.image_repository_url.cluster_registry_id }
                : { url: sr.build_spec.image_repository_url?.url || "", cluster_registry_id: sr.build_spec.image_repository_url?.cluster_registry_id },
              insecure_registry: sr.build_spec.insecure_registry || false,
              source_revision: git_repo_revision ? { git_repo_revision } : undefined,
            } : undefined;
            resource.image_spec = undefined;
          }
          if (resource.sourceType !== 'git') resource.build_spec = undefined;
          if (resource.sourceType !== 'image') resource.image_spec = undefined;

          return resource;
        }),
        volumes: (formData.spec?.volumes || [])
          .map(vol => {
            const sourceConfig = (() => {
              const typedVol = vol as VolumeFormData;
              if (!typedVol.sourceType || typedVol.sourceType === 'None') {
                return {};
              }
              return {
                source: {
                  source_type: typedVol.sourceType === "GitRepo"
                    ? "GitRepo" as const
                    : typedVol.sourceType === "RemoteDir"
                      ? "RemoteDir" as const
                      : "BuildArtifact" as const,
                  git_repo_source: typedVol.sourceType === "GitRepo" ? vol.spec?.source?.git_repo_source : undefined,
                  remote_source: typedVol.sourceType === "RemoteDir" ? vol.spec?.source?.remote_source : undefined,
                  build_source: typedVol.sourceType === "BuildArtifact" ? vol.spec?.source?.build_source : undefined,
                }
              };
            })();
            return {
              name: vol.name || "",
              workspace_name: vol.workspace_name || formData.workspace_name || "default",
              labels: vol.labels?.length ? vol.labels : undefined,
              annotations: vol.annotations?.length ? vol.annotations : undefined,
              spec: {
                size: vol.spec?.size || "",
                storage_class: vol.spec?.storage_class,
                needs_sync_before_use: vol.spec?.needs_sync_before_use || false,
                access_mode: vol.spec?.access_mode || "ReadWriteOnce",
                ...sourceConfig
              }
            };
          })
      }
    };

    const validationResult = StackSchema.safeParse(payloadToValidate);

    if (!validationResult.success) {
      const newErrors: FormErrors = {};

      validationResult.error.issues.forEach(issue => {
        const pathKey = issue.path.join('.');
        if (!newErrors[pathKey]) {
          newErrors[pathKey] = issue.message;
        }
      });

      setFormErrors(newErrors);
      setIsLoading(false);
      toast({
        title: 'Validation Error',
        description: 'Please fix the highlighted errors before submitting the form.',
        variant: 'destructive',
      });
      if (Object.keys(newErrors).length > 0 && !apiError) {
        setApiError('Please fix the highlighted errors before submitting the form');
      }
      return;
    }

    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setIsLoading(false);
      toast({
        title: 'Organization Error',
        description: 'No organization selected. Please select an organization and try again.',
        variant: 'destructive',
      });
      setApiError('No organization selected.');
      return;
    }

    try {
      await createStack(orgId, stripUIFieldsFromStackData(validationResult.data));
      setIsLoading(false);
      toast({
        title: 'Stack Created',
        description: 'Your stack has been successfully created.',
        variant: 'success',
      });
      navigate('/stacks');
    } catch (error) {
      setIsLoading(false);
      let errorMsg = 'An unknown error occurred during stack creation.';
      // Robust error extraction for axios-like errors, without using 'any'
      if (
        error &&
        typeof error === 'object' &&
        error !== null &&
        'response' in error &&
        typeof (error as Record<string, unknown>).response === 'object' &&
        (error as { response?: { data?: { message?: unknown } } }).response?.data?.message
      ) {
        errorMsg = String((error as { response?: { data?: { message?: unknown } } }).response?.data?.message);
      } else if (error instanceof Error) {
        errorMsg = error.message;
      }
      setApiError(errorMsg);
      toast({
        title: 'Stack Creation Failed',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  type StackResourcesFormErrors = { [index: number]: { [field: string]: string | undefined } };
  type StackVolumesFormErrors = { [index: number]: { [field: string]: string | undefined } };

  const resourcesErrors: StackResourcesFormErrors = Object.entries(formErrors)
    .filter(([key]) => key.startsWith("spec.stack_resources."))
    .reduce((acc: StackResourcesFormErrors, [key, value]) => {
      const pathParts = key.split('.');
      if (pathParts.length >= 4) {
        const index = parseInt(pathParts[2], 10);
        const fieldName = pathParts.slice(3).join('.');

        // Initialize the resource's error object if not already done
        if (!acc[index]) acc[index] = {};

        // Map errors to field names, preserving the nested path structure
        // This works with the getError helper function in the StackResourceItem component
        acc[index][fieldName] = value;
      }
      return acc;
    }, {});

  const volumesErrors: StackVolumesFormErrors = Object.entries(formErrors)
    .filter(([key]) => key.startsWith("spec.volumes."))
    .reduce((acc: StackVolumesFormErrors, [key, value]) => {
      const pathParts = key.split('.');
      if (pathParts.length >= 4) {
        const index = parseInt(pathParts[2], 10);
        const fieldName = pathParts.slice(3).join('.');
        if (!acc[index]) acc[index] = {};
        acc[index][fieldName] = value;
      }
      return acc;
    }, {});

  return (
    <div className="px-4 pt-6 pb-10">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Create New Stack</h2>
          <p className="text-muted-foreground mt-1">
            Define your stack resources to provision infrastructure
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={() => {
            if (window.history.length > 2 && window.history.state && window.history.state.idx !== 0) {
              navigate(-1);
            } else {
              navigate("/stacks", { replace: true });
            }
          }}>Cancel</Button>
          <Button variant="default" onClick={handleSubmit} disabled={isLoading}>
            {isLoading ? "Deploying..." : <><Rocket className="mr-2 h-4 w-4" /> Deploy</>}
          </Button>
        </div>
      </div>
      <Separator className="my-6" />

      {Object.keys(formErrors).length > 0 && (
        <div className="bg-red-500/10 border border-red-500 rounded-lg px-4 py-3 mb-6">
          <h3 className="text-red-500 font-semibold flex items-center">
            <AlertTriangle className="w-4 mr-2" />
            Please fix errors on the form to deploy.
          </h3>
        </div>
      )}

      {formErrors[""] && (
        <Alert variant="destructive" className="mb-6">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Configuration Error</AlertTitle>
          <AlertDescription className="text-red-500">{formErrors[""]}</AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col">
        <Card className="mb-6 rounded-lg overflow-hidden">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl">Stack Information</CardTitle>
          </CardHeader>
          <Separator />
          <CardContent className="pt-6">
            <div className="grid gap-6 max-w-5xl">
              <div>
                <UILabel htmlFor="stack-name" className="text-sm font-medium flex items-center gap-1 mb-2">
                  Stack Name <span className="text-red-500">*</span>
                </UILabel>
                <Input
                  id="stack-name"
                  value={formData.name || ""}
                  onChange={(e) => handleChange("name", e.target.value)}
                  className={`max-w-md ${formErrors.name ? "border-red-500" : ""}`}
                  placeholder="my-application-stack"
                  required
                  aria-invalid={!!formErrors.name}
                />
                {formErrors.name && <p className="text-sm text-destructive">{formErrors.name}</p>}
              </div>

              <div>
                <UILabel htmlFor="stack-labels" className="text-sm font-medium flex items-center gap-1 mb-2">
                  Labels
                </UILabel>
                <div className="flex items-center">
                  <TagIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                  <Input
                    id="stack-labels"
                    value={currentLabelInput}
                    onChange={handleLabelInputChange}
                    onKeyDown={handleAddLabel}
                    className={`max-w-md ${formErrors.labels ? "border-red-500" : ""}`}
                    placeholder="e.g., environment=dev or just tag (press Enter to add)"
                    aria-invalid={!!formErrors.labels}
                  />
                </div>
                {formErrors.labels && <p className="text-sm text-destructive whitespace-pre-line">{formErrors.labels}</p>}
                {(formData.labels || []).map((_label, idx) => {
                  const keyError = formErrors[`labels.${idx}.key`];
                  const valueError = formErrors[`labels.${idx}.value`];
                  const itemError = formErrors[`labels.${idx}`];
                  return (
                    <Fragment key={`label-err-${idx}`}>
                      {itemError && <p className="text-sm text-destructive">{`Label ${idx + 1}: ${itemError}`}</p>}
                      {keyError && <p className="text-sm text-destructive">{`Label ${idx + 1} Key: ${keyError}`}</p>}
                      {valueError && <p className="text-sm text-destructive">{`Label ${idx + 1} Value: ${valueError}`}</p>}
                    </Fragment>
                  );
                })}
                {(formData.labels && formData.labels.length > 0) && (
                  <div className="flex flex-wrap gap-2 mt-3">
                    {(formData.labels).map((label, idx) => (
                      <Badge
                        key={idx}
                        variant="secondary"
                        className="flex items-center gap-1 px-2.5 py-1"
                      >
                        <span>{label.key}{label.value && label.value !== "" ? `=${label.value}` : ""}</span>
                        <button
                          onClick={() => removeLabel(idx)}
                          className="ml-1 rounded-full hover:bg-secondary-foreground/20 h-4 w-4 flex items-center justify-center"
                          type="button"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="mb-6 rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl">Define Stack Resources</CardTitle>
            <CardDescription className="mt-1">
              Configure the containerized services that make up your stack
            </CardDescription>
          </CardHeader>
          <Separator />
          <CardContent className="p-0">
            <StackResourcesForm
              resources={formData.spec?.stack_resources || []}
              onResourcesChange={handleResourcesChange}
              errors={resourcesErrors}
            />
          </CardContent>
        </Card>

        {formErrors["spec.stack_resources"] && (
          <Alert variant="destructive" className="mb-6">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Resource Error</AlertTitle>
            <AlertDescription>
              {formErrors["spec.stack_resources"]}
              {formData.spec?.stack_resources.length === 0 && (
                <div className="mt-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleResourcesChange([{ name: "", sourceType: "image" }])}>
                    Add a resource
                  </Button>
                </div>
              )}
            </AlertDescription>
          </Alert>
        )}

        <Card className="mb-6 rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl">Define Stack Volumes</CardTitle>
            <CardDescription className="mt-1">
              Configure persistent volumes that your stack resources can use
            </CardDescription>
          </CardHeader>
          <Separator />
          <CardContent className="p-0">
            <StackVolumesForm
              volumes={formData.spec?.volumes || []}
              onVolumesChange={handleVolumesChange}
              errors={volumesErrors}
              workspace={formData.workspace_name}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

const setNestedValue = <T extends Record<string, unknown>>(
  obj: T,
  path: string,
  value: unknown
): T => {
  const keys = path.split('.');
  let current = obj as Record<string, unknown>;
  keys.forEach((key, index) => {
    if (index === keys.length - 1) {
      current[key] = value;
    } else {
      if (!current[key] || typeof current[key] !== 'object') {
        current[key] = /^[0-9]+$/.test(keys[index + 1]) ? [] : {};
      }
      current = current[key] as Record<string, unknown>;
    }
  });
  return obj;
};
