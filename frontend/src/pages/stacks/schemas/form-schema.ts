/**
 * Form schemas for stack resources
 * These extend the API schemas with additional UI-specific fields and validations
 */
import { z } from "zod";
import {
  ApiStackResourceSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSpecSchema,
  ApiVolumeSchema,
  ApiStackSchema,
} from "./api-schema";
import type { StackUpdateRequest, StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";

/**
 * Form-specific UI schema additions
 */
const FormGitRevisionTypeSchema = z.enum(["commit", "branch", "tag"]);

const FormStackResourceSchema = ApiStackResourceSchema.extend({
  // UI helper, not part of API spec for StackResource
  sourceType: z.enum(["image", "git"]).optional().default("image"),
  // UI helper fields for git revision, not part of API spec StackResource
  gitRevisionType: FormGitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
}).superRefine((data, ctx) => {
  // Validate that git revision fields are required when sourceType is git
  if (data.sourceType === "git") {
    if (!data.gitRevisionType) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Git revision type is required when using a Git repository",
        path: ["gitRevisionType"],
      });
    }

    if (!data.gitRevisionValue) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Git revision value is required when using a Git repository",
        path: ["gitRevisionValue"],
      });
    }

    // Make sure there's a valid Git repo URL when using Git repository
    if (!data.build_spec?.source_context?.git_repo?.repo_url) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Git repository URL is required when using a Git repository",
        path: ["build_spec", "source_context", "git_repo", "repo_url"],
      });
    }
  }

  // Validate that image URL is required when sourceType is image
  if (data.sourceType === "image") {
    if (!data.image_spec?.image) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Container image URL is required",
        path: ["image_spec", "image"],
      });
    }
  }
});

const FormVolumeSourceSchema = ApiVolumeSourceSchema.superRefine(
  (data, ctx) => {
    // Validate that the appropriate source is provided based on source_type
    if (data.source_type === "GitRepo" && !data.git_repo_source) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message:
          "Git repository source is required when source type is GitRepo",
        path: ["git_repo_source"],
      });
    }

    if (data.source_type === "RemoteDir" && !data.remote_source) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Remote source is required when source type is RemoteDir",
        path: ["remote_source"],
      });
    }

    if (
      data.source_type === "BuildArtifact" &&
      (!data.build_source || data.build_source.length === 0)
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message:
          "Build artifacts are required when source type is BuildArtifact",
        path: ["build_source"],
      });
    }
  }
);

const FormVolumeSpecSchema = ApiVolumeSpecSchema.extend({
  source: FormVolumeSourceSchema.optional(),
});

const FormVolumeSchema = ApiVolumeSchema.extend({
  spec: FormVolumeSpecSchema,
});

// Helper for UI to manage different source types
const FormVolumeExtendedSchema = FormVolumeSchema.extend({
  // UI helper field, not part of the API spec
  sourceType: z
    .enum(["None", "GitRepo", "RemoteDir", "BuildArtifact"])
    .default("None"),
});

const FormStackSchema = ApiStackSchema.extend({
  spec: ApiStackSchema.shape.spec.extend({
    stack_resources: z
      .array(FormStackResourceSchema)
      .min(1, "At least one stack resource is required"),
    volumes: z.array(FormVolumeSchema).optional(),
  }),
});

/**
 * Type definitions for form data
 */
type FormStackData = z.infer<typeof FormStackSchema>;
type FormStackResourceData = z.infer<typeof FormStackResourceSchema> & {
  status?: unknown;
};
type FormVolumeData = z.infer<typeof FormVolumeSchema> & {
  status?: unknown;
};
type FormVolumeExtendedData = z.infer<typeof FormVolumeExtendedSchema> & {
  status?: unknown;
};

/**
 * Conversion utilities between API and Form schemas
 */

// Remove UI-only fields from a resource before sending to API
function convertFormResourceToApiResource(
  resource: FormStackResourceData
): StackResourceUpdateRequest {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { sourceType, gitRevisionType, gitRevisionValue, status, ...rest } = resource;
  
  // Clean volume_mounts to remove read-only fields (stack_resource_id and source_volume_type)
  const cleanedVolumeMounts = rest.volume_mounts?.map((volumeMount) => {
    // Remove read-only fields from volume mount
    const {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      stack_resource_id,
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      source_volume_type,
      ...cleanVolumeMount
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } = volumeMount as any;
    
    return cleanVolumeMount;
  });
  
  return {
    ...rest,
    volume_mounts: cleanedVolumeMounts
  } as StackResourceUpdateRequest;
}

// Remove UI-only fields from a volume before sending to API
function convertFormVolumeToApiVolume(
  volume: FormVolumeExtendedData | FormVolumeData
): VolumeUpdateRequest {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { sourceType, status, ...rest } = volume as FormVolumeExtendedData;
  // Ensure needs_sync_before_use is always present and boolean
  // Ensure remote_source.current_directory_hash and path are always present and strings if remote_source exists
  let fixedSource = rest.spec?.source;
  if (fixedSource && fixedSource.source_type === "RemoteDir") {
    if (fixedSource.remote_source) {
      fixedSource = {
        ...fixedSource,
        remote_source: {
          path: fixedSource.remote_source.path || "",
          current_directory_hash: fixedSource.remote_source.current_directory_hash || "",
        },
      };
    } else {
      fixedSource = {
        ...fixedSource,
        remote_source: { path: "", current_directory_hash: "" },
      };
    }
  }
  return {
    ...rest,
    spec: {
      ...rest.spec,
      needs_sync_before_use: rest.spec?.needs_sync_before_use ?? false,
      source: fixedSource,
    },
  };
}

// Add UI-only fields to a resource from API data for use in forms
function convertApiResourceToFormResource(
  resource: z.infer<typeof ApiStackResourceSchema> & { status?: unknown } // API resource may have status
): FormStackResourceData {
  const sourceType: "image" | "git" = resource.build_spec ? "git" : "image";
  let gitRevisionType: "commit" | "branch" | "tag" | undefined = undefined;
  let gitRevisionValue: string | undefined = undefined;

  if (resource.build_spec) {
    const rev = resource.build_spec.source_revision?.git_repo_revision;
    if (rev?.commit) {
      gitRevisionType = "commit";
      gitRevisionValue = rev.commit;
    } else if (rev?.branch?.name) {
      gitRevisionType = "branch";
      gitRevisionValue = rev.branch.name;
    } else if (rev?.tag) {
      gitRevisionType = "tag";
      gitRevisionValue = rev.tag;
    }
  }

  // Ensure all required fields are present, defaulting as needed
  return {
    ...resource,
    name: resource.name ?? "",
    sourceType,
    gitRevisionType,
    gitRevisionValue,
    status: resource.status ?? {},
  };
}

// Add UI-only fields to a volume from API data for use in forms
function convertApiVolumeToFormVolume(
  volume: z.infer<typeof ApiVolumeSchema> & { status?: unknown } // API volume may have status
): FormVolumeExtendedData {
  // Determine sourceType based on the volume's source specification
  let sourceType: "None" | "GitRepo" | "RemoteDir" | "BuildArtifact" = "None";

  if (volume.spec?.source) {
    sourceType = volume.spec.source.source_type;
  }

  // Ensure all required fields are present, defaulting as needed
  return {
    ...volume,
    sourceType,
    status: volume.status ?? {},
  };
}

// Convert a form stack to API stack for submission
function convertFormStackToApiStack(
  stackData: FormStackData
): StackUpdateRequest {
  // Filter out empty or invalid resources (resources with empty names or no image)
  const validResources = stackData.spec.stack_resources.filter(resource => {
    // A resource is valid if it has a name and either an image or build_spec
    const hasName = resource.name && resource.name.trim() !== '';
    const hasImage = resource.image_spec?.image && resource.image_spec.image.trim() !== '';
    const hasBuildSpec = resource.build_spec?.source_revision;
    
    return hasName && (hasImage || hasBuildSpec);
  });

  // Process all valid stack resources by removing UI-only fields
  const apiStackResources = validResources.map(resource => {
    // Always provide both keys for source_revision if build_spec is present
    if (resource.build_spec) {
      const gitRepoRev = resource.build_spec.source_revision?.git_repo_revision;
      resource.build_spec.source_revision = {
        volume_source_revision: undefined,
        git_repo_revision: gitRepoRev,
      };
    }
    return convertFormResourceToApiResource(resource);
  });

  // Filter out empty or invalid volumes (volumes with empty names)
  const validVolumes = stackData.spec.volumes?.filter(volume => {
    return volume.name && volume.name.trim() !== '';
  });

  // Process volumes data if present
  const apiVolumes = validVolumes && validVolumes.length > 0
    ? validVolumes.map(convertFormVolumeToApiVolume)
    : undefined;

  // Create a new clean spec object that will only include API-expected fields
  const apiSpec = {
    stack_resources: apiStackResources as z.infer<typeof ApiStackSchema>["spec"]["stack_resources"],
    volumes: apiVolumes as z.infer<typeof ApiStackSchema>["spec"]["volumes"] | undefined,
  };

  // Combine everything into the final API-compliant object
  return {
    name: stackData.name,
    labels: stackData.labels,
    spec: apiSpec,
  };
}

export type {
  FormStackData,
  FormStackResourceData,
  FormVolumeData,
  FormVolumeExtendedData,
};

export {
  FormVolumeExtendedSchema,
  FormStackSchema,
  convertApiResourceToFormResource,
  convertApiVolumeToFormVolume,
  convertFormStackToApiStack,
};
