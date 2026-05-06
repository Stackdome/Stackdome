import React, { useState, useCallback, Fragment, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import StackResourcesForm from "../shared/stack-resources-form";
import StackVolumesForm from "../shared/stack-volumes-form";
import { Button } from "@/components/ui/button";
import { X, AlertTriangle, Rocket } from "lucide-react";
import { Panel, FieldError } from "@/components/branded";
import AddonsInStackPanel from "@/pages/stacks/components/detail/addons-in-stack-panel";
import StickyActionBar, { type StickyActionBarSegment } from "@/pages/stacks/components/shared/sticky-action-bar";
import { getAddonLinkCount } from "@/pages/stacks/lib/stack-diff";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label as UILabel } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  FormStackSchema,
  type FormStackData,
  type FormStackResourceData,
  type FormVolumeExtendedData,
  convertFormStackToApiStack,
} from "@/pages/stacks/schemas/form-schema";
import { createStack } from '@/api/stacks';
import { getCurrentOrganizationId } from '@/helpers/common';
import { getErrorMessage } from '@/api/client';
import { useToast } from '@/components/ui/use-toast';

type FormErrors = { [path: string]: string | undefined };

export default function StackCreatePage() {
  const [formData, setFormData] = useState<Partial<FormStackData>>({
    name: "",
    labels: [],
    spec: {
      stack_resources: [],
      volumes: [],
    },
  });
  const [currentLabelInput, setCurrentLabelInput] = useState("");
  const [linkedAddonIds, setLinkedAddonIds] = useState<Set<string>>(new Set());
  const [formErrors, setFormErrors] = useState<FormErrors>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { toast } = useToast();

  // Handle imported data from navigation state
  useEffect(() => {
    const importedData = location.state?.importedData;
    const importSource = location.state?.importSource;

    if (importedData && importSource === 'docker-compose') {
      setFormData(importedData);

      // Clear the navigation state to prevent re-importing on refresh
      navigate(location.pathname, { replace: true });
    }
  }, [location.state, navigate, location.pathname]);

  const handleChange = (path: string, value: string | number | boolean | object | null) => {
    setFormData(prev => {
      const newState = JSON.parse(JSON.stringify(prev)) as Partial<FormStackData>;
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
      const value = currentLabelInput.trim();
      if (value) {
        setFormData(prev => ({
          ...prev,
          labels: [...(prev.labels || []), { key: "stackdome.io/user-defined-value", value }],
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

  const handleResourcesChange = useCallback((updatedResources: Partial<FormStackResourceData>[]) => {
    setFormData(prev => ({
      ...prev,
      spec: {
        ...(prev.spec || {}),
        stack_resources: updatedResources as FormStackResourceData[],
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

  const handleVolumesChange = useCallback((updatedVolumes: Partial<FormVolumeExtendedData>[]) => {
    setFormData(prev => {
      const newFormData = { ...prev };

      if (!newFormData.spec) {
        newFormData.spec = {
          stack_resources: [],
          volumes: []
        };
      }

      // Get current volume names for later comparison
      const previousVolumeNames = new Set(
        prev.spec?.volumes?.map(vol => vol.name) || []
      );

      // Get new volume names after update
      const updatedVolumeNames = new Set(
        updatedVolumes.map(vol => vol.name) || []
      );

      // Find volumes that were removed
      const removedVolumeNames = Array.from(previousVolumeNames)
        .filter(name => name && !updatedVolumeNames.has(name)) as string[];

      // If any volumes were removed, we need to unlink them from resources
      if (removedVolumeNames.length > 0 && newFormData.spec?.stack_resources?.length) {
        newFormData.spec.stack_resources = newFormData.spec.stack_resources.map(resource => {
          if (!resource.volume_mounts || resource.volume_mounts.length === 0) {
            return resource;
          }

          // Filter out volume mounts that reference the removed volumes
          const updatedVolumeMounts = resource.volume_mounts.filter(
            mount => !removedVolumeNames.includes(mount.source_volume_name)
          );

          return {
            ...resource,
            volume_mounts: updatedVolumeMounts
          };
        });
      }

      newFormData.spec = {
        ...newFormData.spec,
        volumes: updatedVolumes.map(vol => ({
          name: vol.name || '',
          labels: vol.labels,
          annotations: vol.annotations,
          spec: {
            size: vol.spec?.size || '',
            storage_class: vol.spec?.storage_class,
            needs_sync_before_use: vol.spec?.needs_sync_before_use ?? false,
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

    const payloadToValidate: FormStackData = {
      name: formData.name || "",
      labels: formData.labels || [],
      spec: {
        stack_resources: (formData.spec?.stack_resources || []).map(sr => {
          const resource: FormStackResourceData = {
            name: sr.name || "",
            sourceType: sr.sourceType || 'image',
            useImageSecret: sr.useImageSecret || false,
            useGitSecret: sr.useGitSecret || false,
            labels: sr.labels?.length ? sr.labels : undefined,
            depends_on: sr.depends_on?.length ? sr.depends_on : undefined,
            ports: sr.ports?.length ? sr.ports : undefined,
            volume_mounts: sr.volume_mounts?.length ? sr.volume_mounts : undefined,
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
              image_repository: sr.build_spec.image_repository
                ? {
                  external_image_repo_url: sr.build_spec.image_repository.external_image_repo_url || "",
                  use_internal_registry: sr.build_spec.image_repository.use_internal_registry,
                  cluster_registry_id: sr.build_spec.image_repository.cluster_registry_id,
                }
                : { external_image_repo_url: "" },
              insecure_registry: sr.build_spec.insecure_registry || false,
              source_revision: { volume_source_revision: undefined, git_repo_revision },
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
              const typedVol = vol as FormVolumeExtendedData;
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
              labels: vol.labels?.length ? vol.labels : undefined,
              annotations: vol.annotations?.length ? vol.annotations : undefined,
              spec: {
                size: vol.spec?.size || "",
                storage_class: vol.spec?.storage_class,
                needs_sync_before_use: vol.spec?.needs_sync_before_use ?? false,
                access_mode: vol.spec?.access_mode || "ReadWriteOnce",
                ...sourceConfig
              }
            };
          })
      }
    };

    const validationResult = FormStackSchema.safeParse(payloadToValidate);

    if (!validationResult.success) {
      const newErrors: FormErrors = {};

      validationResult.error.issues.forEach(issue => {
        const pathKey = issue.path.join('.');
        if (!newErrors[pathKey]) {
          newErrors[pathKey] = issue.message;
        }
      });

      console.error('Form validation failed:', newErrors);

      setFormErrors(newErrors);
      setIsLoading(false);
      toast({
        title: 'Validation Error',
        description: 'Please fix the highlighted errors before submitting the form.',
        variant: 'destructive',
      });
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
      await createStack(orgId, convertFormStackToApiStack(validationResult.data));
      setIsLoading(false);
      toast({
        title: 'Stack Created',
        description: 'Your stack has been successfully created.',
        variant: 'success',
      });
      navigate('/stacks');
    } catch (error) {
      setIsLoading(false);

      console.error('Stack creation API failed:', error);

      const errorMsg = getErrorMessage(error);
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

  const resourceCount = formData.spec?.stack_resources?.length ?? 0;
  const volumeCount = formData.spec?.volumes?.length ?? 0;
  const addonLinkCount = getAddonLinkCount(linkedAddonIds, formData.spec?.stack_resources || []);
  const segments: StickyActionBarSegment[] = [];
  if (resourceCount > 0) {
    segments.push({ num: resourceCount, label: resourceCount === 1 ? "RESOURCE" : "RESOURCES" });
  }
  if (volumeCount > 0) {
    segments.push({ num: volumeCount, label: volumeCount === 1 ? "VOLUME" : "VOLUMES" });
  }
  if (addonLinkCount > 0) {
    segments.push({ num: addonLinkCount, label: addonLinkCount === 1 ? "ADDON" : "ADDONS" });
  }
  const handleCancel = () => {
    if (window.history.length > 2 && window.history.state && window.history.state.idx !== 0) {
      navigate(-1);
    } else {
      navigate("/stacks", { replace: true });
    }
  };

  return (
    <div className="p-6">
      <StickyActionBar
        leadLabel="Draft"
        segments={segments}
        primary={{
          label: "Deploy",
          loadingLabel: "Deploying",
          icon: <Rocket className="h-3.5 w-3.5" />,
          isLoading: isLoading,
          onClick: handleSubmit,
        }}
        secondary={{
          label: "Cancel",
          onClick: handleCancel,
        }}
      />
      <header className="mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-2xl font-bold">Create New Stack</h1>
          </div>
          <div className="flex items-center gap-4 text-muted-foreground text-sm mb-1">
            <span>Define your stack to provision infrastructure</span>
          </div>
        </div>
        <Separator className="mt-4" />
      </header>

      {(apiError || formErrors[""]) && (
        <Alert variant="destructive" className="mb-6">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>{formErrors[""] ? "Configuration Error" : "Could not deploy"}</AlertTitle>
          <AlertDescription>{formErrors[""] || apiError}</AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col gap-8">
        <Panel title="Stack Information">
            <div className="grid gap-6 max-w-5xl">
              <div>
                <UILabel htmlFor="stack-name" className="text-sm font-medium flex items-center gap-1 mb-2">
                  Stack Name <span className="text-[15px] font-semibold text-brand/80 leading-none" aria-hidden>*</span>
                </UILabel>
                <Input
                  id="stack-name"
                  value={formData.name || ""}
                  onChange={(e) => handleChange("name", e.target.value)}
                  className={`max-w-md ${formErrors.name ? "border-danger" : ""}`}
                  placeholder="my-application-stack"
                  required
                  aria-invalid={!!formErrors.name}
                />
                <FieldError>{formErrors.name}</FieldError>
              </div>

              <div>
                <UILabel htmlFor="stack-labels" className="text-sm font-medium flex items-center gap-1 mb-2">
                  Labels
                </UILabel>
                <div className="flex items-center">
                  <Input
                    id="stack-labels"
                    value={currentLabelInput}
                    onChange={handleLabelInputChange}
                    onKeyDown={handleAddLabel}
                    className={`max-w-md ${formErrors.labels ? "border-danger" : ""}`}
                    placeholder="e.g., dev, prod, mytag (press Enter to add)"
                    aria-invalid={!!formErrors.labels}
                  />
                </div>
                <FieldError className="whitespace-pre-line">{formErrors.labels}</FieldError>
                {(formData.labels || []).map((_label, idx) => {
                  const keyError = formErrors[`labels.${idx}.key`];
                  const valueError = formErrors[`labels.${idx}.value`];
                  const itemError = formErrors[`labels.${idx}`];
                  return (
                    <Fragment key={`label-err-${idx}`}>
                      <FieldError>{itemError && `Label ${idx + 1}: ${itemError}`}</FieldError>
                      <FieldError>{keyError && `Label ${idx + 1} Key: ${keyError}`}</FieldError>
                      <FieldError>{valueError && `Label ${idx + 1} Value: ${valueError}`}</FieldError>
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
                        <span>{label.value}</span>
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
        </Panel>
        <Panel
          title="Stack Resources"
          count={formData.spec?.stack_resources?.length ?? 0}
          bodyClassName="p-0"
          invalid={!!formErrors["spec.stack_resources"]}
        >
            <StackResourcesForm
              resources={formData.spec?.stack_resources || []}
              onResourcesChange={handleResourcesChange}
              errors={resourcesErrors}
              emptyError={formErrors["spec.stack_resources"]}
              volumes={formData.spec?.volumes || []}
              availableAddonIds={(() => {
                const ids = new Set(linkedAddonIds);
                for (const r of formData.spec?.stack_resources || []) {
                  const envs = (r?.execution_config?.environment_variables || []) as Array<{ from?: string; addonId?: string }>;
                  for (const e of envs) {
                    if (e.from === "addon" && e.addonId) ids.add(e.addonId);
                  }
                }
                return ids;
              })()}
            />
        </Panel>

        <Panel
          title="Stack Volumes"
          count={formData.spec?.volumes?.length ?? 0}
          bodyClassName="p-0"
        >
            <StackVolumesForm
              volumes={formData.spec?.volumes || []}
              onVolumesChange={handleVolumesChange}
              errors={volumesErrors}
            />
        </Panel>

        <AddonsInStackPanel
          resources={formData.spec?.stack_resources || []}
          linkedAddonIds={linkedAddonIds}
          onLinkAddon={(addonId) =>
            setLinkedAddonIds((prev) => {
              const next = new Set(prev);
              next.add(addonId);
              return next;
            })
          }
          onRemoveLinkedAddon={(addonId) =>
            setLinkedAddonIds((prev) => {
              const next = new Set(prev);
              next.delete(addonId);
              return next;
            })
          }
        />
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
