/**
 * Form schemas for stack resources
 * These extend the API schemas with additional UI-specific fields and validations
 */
import { z } from "zod";
import {
  ApiVolumeAccessModeSchema,
  ApiVolumeSourceTypeSchema,
  type ApiStack,
} from "./api-schema";

/**
 * Form-specific schemas that extend API schemas with UI validations and fields
 */
export const FormLabelSchema = z.object({
  key: z.string().min(1, "Key is required"),
  value: z.string().min(1, "Value is required"),
});

export const FormAnnotationSchema = FormLabelSchema;

export const FormPortSchema = z.object({
  number: z.number().int().min(1, "Port number is required"),
  protocol: z.enum(["tcp", "http"]).optional().default("tcp"),
  exposed_to_public: z.boolean().optional().default(false),
  subdomain_prefix: z.string().optional(),
});

export const FormImageSpecSchema = z.object({
  image: z.string().min(1, "Image URL is required"),
});

export const FormEnvVarSchema = z.object({
  name: z.string().min(1, "Environment variable name is required"),
  value: z.string(),
});

export const FormExecutionConfigSchema = z.object({
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
  environment_variables: z.array(FormEnvVarSchema).optional(),
});

export const FormInitSpecSchema = z.object({
  image_spec: FormImageSpecSchema.optional(),
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
});

/**
 * Form-specific UI schema additions
 */
export const FormGitRevisionTypeSchema = z.enum(["commit", "branch", "tag"]);

export const FormStackResourceUISchema = z.object({
  gitRevisionType: FormGitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
});

export const FormGitRepoRevisionSchema = z
  .object({
    branch: z.object({ name: z.string().optional() }).optional(),
    commit: z.string().optional(),
    tag: z.string().optional(),
  })
  .refine((data) => data.branch?.name || data.commit || data.tag, {
    message:
      "At least one of branch, commit, or tag must be provided for Git revision.",
  });

export const FormBuildSourceContextSchema = z.object({
  git_repo: z
    .object({
      repo_url: z.string().url("Invalid Git repository URL"),
    })
    .optional(),
});

export const FormStackResourceBuildSpecSchema = z.object({
  source_context: FormBuildSourceContextSchema,
  context_path_within_source: z.string().optional().default("./"),
  dockerfile_path: z.string().optional().default("Dockerfile"),
  source_revision: z
    .object({
      git_repo_revision: FormGitRepoRevisionSchema.optional(),
    })
    .optional(),
  image_repository: z.object({
    external_image_repo_url: z
      .string()
      .min(1, "Image repository URL is required"),
    use_internal_registry: z.boolean().optional(),
    cluster_registry_id: z.string().optional(),
  }),
  insecure_registry: z.boolean().optional().default(false),
});

export const FormVolumeMountSchema = z.object({
  source_volume_name: z.string().min(1, "Volume name is required"),
  source_sub_path: z.string().optional(),
  target_path: z.string().min(1, "Target path is required"),
});

export const FormStackResourceSchema = z
  .object({
    name: z.string().min(1, "Resource name is required"),
    labels: z.array(FormLabelSchema).optional(),
    build_spec: FormStackResourceBuildSpecSchema.optional(),
    image_spec: FormImageSpecSchema.optional(),
    init_spec: FormInitSpecSchema.optional(),
    execution_config: FormExecutionConfigSchema.optional(),
    depends_on: z.array(z.string()).optional(),
    ports: z.array(FormPortSchema).optional(),
    volume_mounts: z.array(FormVolumeMountSchema).optional(),
    // UI helper, not part of API spec for StackResource
    sourceType: z.enum(["image", "git"]).optional().default("image"),
    // UI helper fields for git revision, not part of API spec StackResource
    gitRevisionType: FormGitRevisionTypeSchema.optional(),
    gitRevisionValue: z.string().optional(),
  })
  .superRefine((data, ctx) => {
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

export const FormVolumeSourceTypeSchema = ApiVolumeSourceTypeSchema;

export const FormRemoteSourceSchema = z.object({
  path: z.string().min(1, "Path is required"),
  current_directory_hash: z.string().optional(),
});

export const FormBuildArtifactSchema = z.object({
  resource_ref: z.string().min(1, "Resource reference is required"),
  source_path: z.string().min(1, "Source path is required"),
  destination_path: z.string().min(1, "Destination path is required"),
});

export const FormGitRepoSourceSchema = z.object({
  repo_url: z.string().url("Invalid Git repository URL"),
  revision: FormGitRepoRevisionSchema,
});

export const FormVolumeSourceSchema = z
  .object({
    git_repo_source: FormGitRepoSourceSchema.optional(),
    source_type: FormVolumeSourceTypeSchema,
    remote_source: FormRemoteSourceSchema.optional(),
    build_source: z.array(FormBuildArtifactSchema).optional(),
  })
  .superRefine((data, ctx) => {
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
  });

export const FormVolumeSpecSchema = z.object({
  size: z.string().min(1, "Volume size is required"),
  storage_class: z.string().optional(),
  needs_sync_before_use: z.boolean().default(false),
  access_mode: ApiVolumeAccessModeSchema.default("ReadWriteOnce"),
  source: FormVolumeSourceSchema.optional(),
});

export const FormVolumeSchema = z.object({
  name: z.string().min(1, "Volume name is required"),
  labels: z.array(FormLabelSchema).optional(),
  annotations: z.array(FormAnnotationSchema).optional(),
  spec: FormVolumeSpecSchema,
});

// Helper for UI to manage different source types
export const FormVolumeExtendedSchema = FormVolumeSchema.extend({
  // UI helper field, not part of the API spec
  sourceType: z
    .enum(["None", "GitRepo", "RemoteDir", "BuildArtifact"])
    .default("None"),
});

export const FormStackSchema = z.object({
  name: z.string().min(1, "Stack name is required"),
  workspace_name: z.string().min(1, "Workspace name is required"),
  labels: z.array(FormLabelSchema).optional(),
  spec: z.object({
    stack_resources: z
      .array(FormStackResourceSchema)
      .min(1, "At least one stack resource is required"),
    volumes: z.array(FormVolumeSchema).optional(),
  }),
});

/**
 * Type definitions for form data
 */
export type FormStackData = z.infer<typeof FormStackSchema>;
export type FormStackResourceData = z.infer<typeof FormStackResourceSchema>;
export type FormVolumeData = z.infer<typeof FormVolumeSchema>;
export type FormVolumeExtendedData = z.infer<typeof FormVolumeExtendedSchema>;

/**
 * Conversion utilities between API and Form schemas
 */

// Remove UI-only fields from a resource before sending to API
export function convertFormResourceToApiResource(
  resource: FormStackResourceData
): Omit<
  FormStackResourceData,
  "sourceType" | "gitRevisionType" | "gitRevisionValue"
> {
  const rest = { ...resource };
  delete (rest as Partial<FormStackResourceData>).sourceType;
  delete (rest as Partial<FormStackResourceData>).gitRevisionType;
  delete (rest as Partial<FormStackResourceData>).gitRevisionValue;
  return rest as Omit<
    FormStackResourceData,
    "sourceType" | "gitRevisionType" | "gitRevisionValue"
  >;
}

// Remove UI-only fields from a volume before sending to API
export function convertFormVolumeToApiVolume(
  volume: FormVolumeExtendedData | FormVolumeData
):
  | Omit<FormVolumeExtendedData, "sourceType">
  | Omit<FormVolumeData, "sourceType"> {
  const rest = { ...volume };
  delete (rest as Partial<FormVolumeExtendedData>).sourceType;
  return rest;
}

// Add UI-only fields to a resource from API data for use in forms
export function convertApiResourceToFormResource(
  resource: Omit<
    FormStackResourceData,
    "sourceType" | "gitRevisionType" | "gitRevisionValue"
  >
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

  return {
    ...resource,
    sourceType,
    gitRevisionType,
    gitRevisionValue,
  };
}

// Add UI-only fields to a volume from API data for use in forms
export function convertApiVolumeToFormVolume(
  volume:
    | Omit<FormVolumeExtendedData, "sourceType">
    | Omit<FormVolumeData, "sourceType">
): FormVolumeExtendedData {
  // Determine sourceType based on the volume's source specification
  let sourceType: "None" | "GitRepo" | "RemoteDir" | "BuildArtifact" = "None";

  if (volume.spec?.source) {
    sourceType = volume.spec.source.source_type;
  }

  return {
    ...volume,
    sourceType,
  };
}

// Convert a form stack to API stack for submission
export function convertFormStackToApiStack(
  stackData: FormStackData
): Omit<ApiStack, "id" | "created_at" | "status"> {
  // Process all stack resources by removing UI-only fields
  const apiStackResources = stackData.spec.stack_resources.map(
    convertFormResourceToApiResource
  );

  // Process volumes data if present
  const apiVolumes = stackData.spec.volumes
    ? stackData.spec.volumes.map(convertFormVolumeToApiVolume)
    : undefined;

  // Create a new clean spec object that will only include API-expected fields
  const apiSpec = {
    stack_resources: apiStackResources as ApiStack["spec"]["stack_resources"],
    volumes: apiVolumes as ApiStack["spec"]["volumes"] | undefined,
  };

  // Combine everything into the final API-compliant object
  return {
    name: stackData.name,
    workspace_name: stackData.workspace_name,
    labels: stackData.labels,
    spec: apiSpec,
  };
}
