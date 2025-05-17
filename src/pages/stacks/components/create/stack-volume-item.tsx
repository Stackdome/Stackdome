import { useState } from "react";
import {
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/components/ui/accordion";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Plus, X, GitBranch, Database, HardDrive, Trash2 } from "lucide-react";
import type { VolumeFormData } from "@/pages/stacks/schemas/stack-create-schema";
import { Badge } from "@/components/ui/badge";

interface BuildArtifact {
  resource_ref: string;
  source_path: string;
  destination_path: string;
}

interface StackVolumeItemProps {
  volume: Partial<VolumeFormData>;
  index: number;
  itemRef: (el: HTMLButtonElement | null) => void;
  isOnlyVolume: boolean;
  onChange: (index: number, updatedVolume: Partial<VolumeFormData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
}

export default function StackVolumeItem({
  volume,
  index,
  itemRef,
  onChange,
  onRemove,
  errors,
}: StackVolumeItemProps) {
  // Helper for updating volume fields
  const update = (patch: Partial<VolumeFormData>) => {
    onChange(index, { ...volume, ...patch });
  };

  // For labels management
  const [currentLabelInput, setCurrentLabelInput] = useState("");

  // For source tab selection based on volume.sourceType
  const getInitialSourceTab = (): string => {
    if (!volume.sourceType || volume.sourceType === "None") return "no-source";
    if (volume.sourceType === "GitRepo") return "git-repo";
    if (volume.sourceType === "RemoteDir") return "remote-dir";
    if (volume.sourceType === "BuildArtifact") return "build-artifact";
    return "no-source";
  };

  const handleSourceTypeChange = (value: string) => {
    let sourceType: "None" | "GitRepo" | "RemoteDir" | "BuildArtifact" = "None";

    if (value === "git-repo") sourceType = "GitRepo";
    else if (value === "remote-dir") sourceType = "RemoteDir";
    else if (value === "build-artifact") sourceType = "BuildArtifact";

    // Update the UI helper field and prepare the source structure based on the selected type
    const updatedVolume: Partial<VolumeFormData> = {
      ...volume,
      sourceType
    };

    // Set up the spec with required fields
    updatedVolume.spec = {
      size: volume.spec?.size || "",
      needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
      access_mode: volume.spec?.access_mode || "ReadWriteOnce",
      storage_class: volume.spec?.storage_class
    };

    // Add source configuration based on source type
    if (sourceType !== "None") {
      updatedVolume.spec.source = {
        source_type: sourceType as "GitRepo" | "RemoteDir" | "BuildArtifact",
      };

      if (sourceType === "GitRepo") {
        updatedVolume.spec.source.git_repo_source = {
          repo_url: "",
          revision: { branch: { name: "" } }
        };
      } else if (sourceType === "RemoteDir") {
        updatedVolume.spec.source.remote_source = { path: "" };
      } else if (sourceType === "BuildArtifact") {
        updatedVolume.spec.source.build_source = [{
          resource_ref: "",
          source_path: "",
          destination_path: ""
        }];
      }
    }

    update(updatedVolume);
  };

  // Add a new label
  const addLabel = () => {
    if (!currentLabelInput.trim()) return;

    const labelParts = currentLabelInput.trim().split('=');
    const key = labelParts[0].trim();
    const value = labelParts.length > 1 ? labelParts.slice(1).join('=').trim() : "";

    if (key) {
      update({
        ...volume,
        labels: [...(volume.labels || []), { key, value }]
      });
      setCurrentLabelInput("");
    }
  };

  // Remove a label
  const removeLabel = (indexToRemove: number) => {
    update({
      ...volume,
      labels: (volume.labels || []).filter((_, idx: number) => idx !== indexToRemove)
    });
  };

  // Handle adding a build artifact (for BuildArtifact source type)
  const addBuildArtifact = () => {
    if (!volume.spec?.source?.build_source) return;

    update({
      ...volume,
      spec: {
        ...volume.spec,
        source: {
          ...volume.spec.source,
          build_source: [
            ...volume.spec.source.build_source,
            { resource_ref: "", source_path: "", destination_path: "" }
          ]
        }
      }
    });
  };

  // Handle removing a build artifact
  const removeBuildArtifact = (artifactIndex: number) => {
    if (!volume.spec?.source?.build_source) return;

    update({
      ...volume,
      spec: {
        ...volume.spec,
        source: {
          ...volume.spec.source,
          build_source: volume.spec.source.build_source.filter((_: BuildArtifact, idx: number) => idx !== artifactIndex)
        }
      }
    });
  };

  // Update a build artifact field
  const updateBuildArtifact = (artifactIndex: number, field: string, value: string) => {
    if (!volume.spec?.source?.build_source) return;

    update({
      ...volume,
      spec: {
        ...volume.spec,
        source: {
          ...volume.spec.source,
          build_source: volume.spec.source.build_source.map((artifact: BuildArtifact, idx: number) =>
            idx === artifactIndex ? { ...artifact, [field]: value } : artifact
          )
        }
      }
    });
  };

  return (
    <AccordionItem value={String(index)} className="border-0">
      <AccordionTrigger
        ref={itemRef}
        className="px-4 py-3 hover:bg-accent hover:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground rounded-t-md [&[data-state=open]]:rounded-b-none"
      >
        <div className="flex items-center gap-2 text-left">
          <Database className="h-5 w-5 text-muted-foreground shrink-0" />
          <div>
            <span className="font-medium">{volume.name || `Volume ${index + 1}`}</span>
            {volume.spec?.size && (
              <span className="ml-2 text-sm text-muted-foreground">({volume.spec.size})</span>
            )}
            {errors._form && (
              <div className="text-sm text-destructive mt-1">{errors._form}</div>
            )}
          </div>
        </div>
      </AccordionTrigger>

      <AccordionContent className="pb-4 pt-2">
        <div className="px-4 space-y-4">
          {/* Basic info section */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor={`volume-name-${index}`}>
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id={`volume-name-${index}`}
                placeholder="Volume name"
                value={volume.name || ""}
                onChange={(e) => update({ name: e.target.value })}
                className={errors.name ? "border-destructive" : ""}
                aria-invalid={!!errors.name}
              />
              {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor={`volume-size-${index}`}>
                Size <span className="text-destructive">*</span>
              </Label>
              <Input
                id={`volume-size-${index}`}
                placeholder="e.g., 1Gi, 500Mi"
                value={volume.spec?.size || ""}
                onChange={(e) => update({
                  spec: {
                    size: e.target.value,
                    needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                    access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                    storage_class: volume.spec?.storage_class,
                    source: volume.spec?.source
                  }
                })}
                className={errors["spec.size"] ? "border-destructive" : ""}
                aria-invalid={!!errors["spec.size"]}
              />
              {errors["spec.size"] && <p className="text-sm text-destructive">{errors["spec.size"]}</p>}
              <p className="text-xs text-muted-foreground">Volume size (e.g., 1Gi, 500Mi)</p>
            </div>
          </div>

          {/* Advanced settings */}
          <div className="space-y-2">
            <h3 className="font-medium text-sm">Access Mode</h3>
            <Select
              value={volume.spec?.access_mode || "ReadWriteOnce"}
              onValueChange={(value) => update({
                spec: {
                  size: volume.spec?.size || "",
                  needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                  access_mode: value as "ReadWriteOnce" | "ReadWriteMany" | "ReadOnlyMany",
                  storage_class: volume.spec?.storage_class,
                  source: volume.spec?.source
                }
              })}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select access mode" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ReadWriteOnce">ReadWriteOnce (RWO)</SelectItem>
                <SelectItem value="ReadWriteMany">ReadWriteMany (RWX)</SelectItem>
                <SelectItem value="ReadOnlyMany">ReadOnlyMany (ROX)</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              ReadWriteOnce: Can be mounted by a single node for read/write.<br />
              ReadWriteMany: Can be mounted by multiple nodes for read/write.<br />
              ReadOnlyMany: Can be mounted by multiple nodes for read only.
            </p>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id={`needs-sync-${index}`}
              checked={volume.spec?.needs_sync_before_use || false}
              onCheckedChange={(checked) => update({
                spec: {
                  size: volume.spec?.size || "",
                  needs_sync_before_use: checked,
                  access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                  storage_class: volume.spec?.storage_class,
                  source: volume.spec?.source
                }
              })}
            />
            <Label htmlFor={`needs-sync-${index}`} className="cursor-pointer">
              Needs sync before use
            </Label>
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="ml-1 cursor-help text-muted-foreground">(?)</span>
              </TooltipTrigger>
              <TooltipContent>
                If enabled, the volume will be synced before the stack can use it
              </TooltipContent>
            </Tooltip>
          </div>

          {/* Optional storage class */}
          <div className="space-y-2">
            <Label htmlFor={`storage-class-${index}`}>Storage Class (optional)</Label>
            <Input
              id={`storage-class-${index}`}
              placeholder="Storage class name"
              value={volume.spec?.storage_class || ""}
              onChange={(e) => update({
                spec: {
                  size: volume.spec?.size || "",
                  needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                  access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                  storage_class: e.target.value,
                  source: volume.spec?.source
                }
              })}
            />
            <p className="text-xs text-muted-foreground">
              Leave empty to use the cluster's default storage class
            </p>
          </div>

          <Separator />

          {/* Source Settings */}
          <div className="space-y-2">
            <h3 className="font-medium">Volume Source</h3>
            <Tabs
              defaultValue={getInitialSourceTab()}
              onValueChange={handleSourceTypeChange}
              className="w-full"
            >
              <TabsList className="grid grid-cols-4 w-full">
                <TabsTrigger value="no-source">No Source</TabsTrigger>
                <TabsTrigger value="git-repo">Git Repository</TabsTrigger>
                <TabsTrigger value="remote-dir">Remote Directory</TabsTrigger>
                <TabsTrigger value="build-artifact">Build Artifact</TabsTrigger>
              </TabsList>

              {/* No Source Content */}
              <TabsContent value="no-source" className="pt-4">
                <div className="text-center text-muted-foreground">
                  <HardDrive className="mx-auto h-8 w-8 mb-2" />
                  <p>Empty volume without an initial source.</p>
                </div>
              </TabsContent>

              {/* Git Repository Source */}
              <TabsContent value="git-repo" className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label htmlFor={`git-repo-url-${index}`}>
                    Git Repository URL <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id={`git-repo-url-${index}`}
                    placeholder="https://github.com/username/repository.git"
                    value={volume.spec?.source?.git_repo_source?.repo_url || ""}
                    onChange={(e) => update({
                      spec: {
                        size: volume.spec?.size || "",
                        needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                        access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                        storage_class: volume.spec?.storage_class,
                        source: {
                          source_type: "GitRepo" as const,
                          git_repo_source: {
                            repo_url: e.target.value,
                            revision: volume.spec?.source?.git_repo_source?.revision || { branch: { name: "main" } }
                          }
                        }
                      }
                    })}
                    className={errors["spec.source.git_repo_source.repo_url"] ? "border-destructive" : ""}
                  />
                  {errors["spec.source.git_repo_source.repo_url"] &&
                    <p className="text-sm text-destructive">{errors["spec.source.git_repo_source.repo_url"]}</p>}
                </div>

                {/* Git Revision Section */}
                <div className="space-y-4 border rounded-lg p-3 bg-muted/20">
                  <div className="flex items-center">
                    <GitBranch className="h-4 w-4 mr-2" />
                    <h4 className="font-medium">Git Revision</h4>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor={`git-branch-${index}`}>Branch Name</Label>
                    <Input
                      id={`git-branch-${index}`}
                      placeholder="main"
                      value={volume.spec?.source?.git_repo_source?.revision?.branch?.name || ""}
                      onChange={(e) => update({
                        spec: {
                          size: volume.spec?.size || "",
                          needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                          access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                          storage_class: volume.spec?.storage_class,
                          source: {
                            source_type: "GitRepo" as const,
                            git_repo_source: {
                              repo_url: volume.spec?.source?.git_repo_source?.repo_url || "",
                              revision: {
                                branch: { name: e.target.value },
                                commit: undefined,
                                tag: undefined
                              }
                            }
                          }
                        }
                      })}
                    />
                    <p className="text-xs text-muted-foreground">Default branch is "main"</p>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor={`git-commit-${index}`}>Commit SHA (optional)</Label>
                    <Input
                      id={`git-commit-${index}`}
                      placeholder="Specific commit SHA"
                      value={volume.spec?.source?.git_repo_source?.revision?.commit || ""}
                      onChange={(e) => update({
                        spec: {
                          size: volume.spec?.size || "",
                          needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                          access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                          storage_class: volume.spec?.storage_class,
                          source: {
                            source_type: "GitRepo" as const,
                            git_repo_source: {
                              repo_url: volume.spec?.source?.git_repo_source?.repo_url || "",
                              revision: {
                                commit: e.target.value,
                                branch: undefined,
                                tag: undefined
                              }
                            }
                          }
                        }
                      })}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor={`git-tag-${index}`}>Tag (optional)</Label>
                    <Input
                      id={`git-tag-${index}`}
                      placeholder="v1.0.0"
                      value={volume.spec?.source?.git_repo_source?.revision?.tag || ""}
                      onChange={(e) => update({
                        spec: {
                          size: volume.spec?.size || "",
                          needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                          access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                          storage_class: volume.spec?.storage_class,
                          source: {
                            source_type: "GitRepo" as const,
                            git_repo_source: {
                              repo_url: volume.spec?.source?.git_repo_source?.repo_url || "",
                              revision: {
                                tag: e.target.value,
                                branch: undefined,
                                commit: undefined
                              }
                            }
                          }
                        }
                      })}
                    />
                  </div>
                </div>
              </TabsContent>

              {/* Remote Directory Source */}
              <TabsContent value="remote-dir" className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label htmlFor={`remote-dir-path-${index}`}>
                    Remote Directory Path <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id={`remote-dir-path-${index}`}
                    placeholder="/path/to/directory"
                    value={volume.spec?.source?.remote_source?.path || ""}
                    onChange={(e) => update({
                      spec: {
                        size: volume.spec?.size || "",
                        needs_sync_before_use: volume.spec?.needs_sync_before_use || false,
                        access_mode: volume.spec?.access_mode || "ReadWriteOnce",
                        storage_class: volume.spec?.storage_class,
                        source: {
                          source_type: "RemoteDir" as const,
                          remote_source: {
                            path: e.target.value
                          },
                          git_repo_source: undefined,
                          build_source: undefined
                        }
                      }
                    })}
                    className={errors["spec.source.remote_source.path"] ? "border-destructive" : ""}
                  />
                  {errors["spec.source.remote_source.path"] &&
                    <p className="text-sm text-destructive">{errors["spec.source.remote_source.path"]}</p>}
                  <p className="text-xs text-muted-foreground">
                    Provide the path to the remote directory that should be mounted
                  </p>
                </div>
              </TabsContent>

              {/* Build Artifact Source */}
              <TabsContent value="build-artifact" className="space-y-4 pt-4">
                {(volume.spec?.source?.build_source || []).map((artifact, artifactIndex) => (
                  <div key={artifactIndex} className="border rounded-lg p-3 space-y-3 bg-muted/20">
                    <div className="flex justify-between items-center">
                      <h4 className="font-medium">Build Artifact {artifactIndex + 1}</h4>
                      {(volume.spec?.source?.build_source?.length || 0) > 1 && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeBuildArtifact(artifactIndex)}
                          className="h-8 w-8 p-0 text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor={`resource-ref-${index}-${artifactIndex}`}>
                        Resource Reference <span className="text-destructive">*</span>
                      </Label>
                      <Input
                        id={`resource-ref-${index}-${artifactIndex}`}
                        placeholder="Resource name"
                        value={artifact.resource_ref || ""}
                        onChange={(e) => updateBuildArtifact(artifactIndex, "resource_ref", e.target.value)}
                        className={errors[`spec.source.build_source.${artifactIndex}.resource_ref`] ? "border-destructive" : ""}
                      />
                      {errors[`spec.source.build_source.${artifactIndex}.resource_ref`] &&
                        <p className="text-sm text-destructive">{errors[`spec.source.build_source.${artifactIndex}.resource_ref`]}</p>}
                      <p className="text-xs text-muted-foreground">
                        Reference to a stack resource that produces build artifacts
                      </p>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor={`source-path-${index}-${artifactIndex}`}>
                        Source Path <span className="text-destructive">*</span>
                      </Label>
                      <Input
                        id={`source-path-${index}-${artifactIndex}`}
                        placeholder="/path/in/build/output"
                        value={artifact.source_path || ""}
                        onChange={(e) => updateBuildArtifact(artifactIndex, "source_path", e.target.value)}
                        className={errors[`spec.source.build_source.${artifactIndex}.source_path`] ? "border-destructive" : ""}
                      />
                      {errors[`spec.source.build_source.${artifactIndex}.source_path`] &&
                        <p className="text-sm text-destructive">{errors[`spec.source.build_source.${artifactIndex}.source_path`]}</p>}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor={`dest-path-${index}-${artifactIndex}`}>
                        Destination Path <span className="text-destructive">*</span>
                      </Label>
                      <Input
                        id={`dest-path-${index}-${artifactIndex}`}
                        placeholder="/path/in/volume"
                        value={artifact.destination_path || ""}
                        onChange={(e) => updateBuildArtifact(artifactIndex, "destination_path", e.target.value)}
                        className={errors[`spec.source.build_source.${artifactIndex}.destination_path`] ? "border-destructive" : ""}
                      />
                      {errors[`spec.source.build_source.${artifactIndex}.destination_path`] &&
                        <p className="text-sm text-destructive">{errors[`spec.source.build_source.${artifactIndex}.destination_path`]}</p>}
                    </div>
                  </div>
                ))}

                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={addBuildArtifact}
                  className="flex w-full items-center justify-center"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Build Artifact
                </Button>
              </TabsContent>
            </Tabs>
          </div>

          <Separator />

          {/* Labels Section */}
          <div className="space-y-3">
            <h3 className="font-medium">Labels (Optional)</h3>
            <div className="flex items-center space-x-2">
              <Input
                placeholder="key=value format"
                value={currentLabelInput}
                onChange={(e) => setCurrentLabelInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && addLabel()}
              />
              <Button
                type="button"
                onClick={addLabel}
                variant="outline"
                className="shrink-0"
              >
                Add
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Enter labels in key=value format and press Enter or click Add
            </p>

            {(volume.labels && volume.labels.length > 0) && (
              <div className="flex flex-wrap gap-2 mt-2">
                {volume.labels.map((label, labelIndex) => (
                  <Badge key={labelIndex} variant="secondary" className="pl-2 pr-1 py-1">
                    {label.key}={label.value}
                    <button
                      type="button"
                      className="ml-1 hover:bg-destructive/20 rounded focus:outline-none"
                      onClick={() => removeLabel(labelIndex)}
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            )}
          </div>

          {/* Delete button always visible at the bottom */}
          <div className="pt-4 border-t">
            <Button
              type="button"
              variant="ghost"
              className="text-destructive hover:text-destructive hover:bg-destructive/10"
              onClick={() => onRemove(index)}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              Remove Volume
            </Button>
          </div>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}
