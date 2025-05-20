import { z } from 'zod';
import type { Stack } from '../types';

export const LabelSchema = z.object({
  key: z.string().min(1, "Key is required"),
  value: z.string().min(1, "Value is required"),
});

export const AnnotationSchema = LabelSchema;

export const PortSchema = z.object({
  number: z.number().int().min(1, "Port number is required"),
  protocol: z.enum(['tcp', 'http']).optional().default('tcp'),
  exposed_to_public: z.boolean().optional().default(false),
  subdomain_prefix: z.string().optional(),
});

export const ImageSpecSchema = z.object({
  image: z.string().min(1, "Image URL is required"), // Assuming it's a URL or image name
});

export const EnvVarSchema = z.object({
  name: z.string().min(1, "Environment variable name is required"),
  value: z.string(),
});

export const ExecutionConfigSchema = z.object({
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
  environment_variables: z.array(EnvVarSchema).optional(),
});

export const InitSpecSchema = z.object({
  image_spec: ImageSpecSchema.optional(), // If init uses a different image
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
});

// BuildSpec related schemas - focusing on user inputs

export const GitRevisionTypeSchema = z.enum(["commit", "branch", "tag"]);

export const StackResourceFormUISchema = z.object({
  gitRevisionType: GitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
});

export const GitRepoRevisionSchema = z.object({ // User input for BuildSourceRevision.git_repo_revision
  branch: z.object({ name: z.string().optional() }).optional(),
  commit: z.string().optional(),
  tag: z.string().optional(),
}).refine(data => data.branch?.name || data.commit || data.tag, {
  message: "At least one of branch, commit, or tag must be provided for Git revision.",
});

export const BuildSourceContextSchema = z.object({ // User input for StackResourceBuildSpec.source_context
  git_repo: z.object({
    repo_url: z.string().url("Invalid Git repository URL"),
  }).optional(),
  // volume selection might be handled differently in UI, not a direct string input here
});

export const StackResourceBuildSpecSchema = z.object({
  source_context: BuildSourceContextSchema,
  context_path_within_source: z.string().optional().default('./'),
  dockerfile_path: z.string().optional().default('Dockerfile'),
  source_revision: z.object({
    git_repo_revision: GitRepoRevisionSchema.optional(),
  }).optional(),
  image_repository: z.object({
    external_image_repo_url: z.string().min(1, "Image repository URL is required"),
    use_internal_registry: z.boolean().optional(),
    cluster_registry_id: z.string().optional(),
  }),
  insecure_registry: z.boolean().optional().default(false),
});


// Main Schemas

export const VolumeMountSchema = z.object({
  source_volume_name: z.string().min(1, "Volume name is required"),
  source_sub_path: z.string().optional(),
  target_path: z.string().min(1, "Target path is required"),
});

export const StackResourceSchema = z.object({
  name: z.string().min(1, "Resource name is required"),
  labels: z.array(LabelSchema).optional(),
  build_spec: StackResourceBuildSpecSchema.optional(),
  image_spec: ImageSpecSchema.optional(),
  init_spec: InitSpecSchema.optional(),
  execution_config: ExecutionConfigSchema.optional(),
  depends_on: z.array(z.string()).optional(),
  ports: z.array(PortSchema).optional(),
  volume_mounts: z.array(VolumeMountSchema).optional(),
  // UI helper, not part of API spec for StackResource
  sourceType: z.enum(["image", "git"]).optional().default("image"),
  // UI helper fields for git revision, not part of API spec StackResource
  gitRevisionType: GitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
}).superRefine((data, ctx) => {
  // Validate that git revision fields are required when sourceType is git
  if (data.sourceType === 'git') {
    if (!data.gitRevisionType) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Git revision type is required when using a Git repository',
        path: ['gitRevisionType'],
      });
    }

    if (!data.gitRevisionValue) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Git revision value is required when using a Git repository',
        path: ['gitRevisionValue'],
      });
    }

    // Make sure there's a valid Git repo URL when using Git repository
    if (!data.build_spec?.source_context?.git_repo?.repo_url) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Git repository URL is required when using a Git repository',
        path: ['build_spec', 'source_context', 'git_repo', 'repo_url'],
      });
    }
  }

  // Validate that image URL is required when sourceType is image
  if (data.sourceType === 'image') {
    if (!data.image_spec?.image) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Container image URL is required',
        path: ['image_spec', 'image'],
      });
    }
  }
});

export const VolumeAccessModeSchema = z.enum([
  'ReadWriteOnce',
  'ReadWriteMany',
  'ReadOnlyMany'
]).default('ReadWriteOnce');

export const VolumeSourceTypeSchema = z.enum([
  'RemoteDir',
  'BuildArtifact',
  'GitRepo'
]);

export const RemoteSourceSchema = z.object({
  path: z.string().min(1, 'Path is required'),
  current_directory_hash: z.string().optional(),
});

export const BuildArtifactSchema = z.object({
  resource_ref: z.string().min(1, 'Resource reference is required'),
  source_path: z.string().min(1, 'Source path is required'),
  destination_path: z.string().min(1, 'Destination path is required'),
});

export const GitRepoSourceSchema = z.object({
  repo_url: z.string().url('Invalid Git repository URL'),
  revision: GitRepoRevisionSchema,
});

export const VolumeSourceSchema = z.object({
  git_repo_source: GitRepoSourceSchema.optional(),
  source_type: VolumeSourceTypeSchema,
  remote_source: RemoteSourceSchema.optional(),
  build_source: z.array(BuildArtifactSchema).optional(),
}).superRefine((data, ctx) => {
  // Validate that the appropriate source is provided based on source_type
  if (data.source_type === 'GitRepo' && !data.git_repo_source) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Git repository source is required when source type is GitRepo',
      path: ['git_repo_source'],
    });
  }

  if (data.source_type === 'RemoteDir' && !data.remote_source) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Remote source is required when source type is RemoteDir',
      path: ['remote_source'],
    });
  }

  if (data.source_type === 'BuildArtifact' && (!data.build_source || data.build_source.length === 0)) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Build artifacts are required when source type is BuildArtifact',
      path: ['build_source'],
    });
  }
});

export const VolumeSpecSchema = z.object({
  size: z.string().min(1, 'Volume size is required'),
  storage_class: z.string().optional(),
  needs_sync_before_use: z.boolean().default(false),
  access_mode: VolumeAccessModeSchema,
  source: VolumeSourceSchema.optional(),
});

export const VolumeSchema = z.object({
  name: z.string().min(1, 'Volume name is required'),
  labels: z.array(LabelSchema).optional(),
  annotations: z.array(AnnotationSchema).optional(),
  spec: VolumeSpecSchema,
});

// Helper for UI to manage different source types
export const VolumeFormSchema = VolumeSchema.extend({
  // UI helper field, not part of the API spec
  sourceType: z.enum(['None', 'GitRepo', 'RemoteDir', 'BuildArtifact']).default('None'),
});

export const StackSchema = z.object({
  name: z.string().min(1, "Stack name is required"),
  workspace_name: z.string().min(1, "Workspace name is required"),
  labels: z.array(LabelSchema).optional(),
  spec: z.object({
    stack_resources: z.array(StackResourceSchema).min(1, "At least one stack resource is required"),
    volumes: z.array(VolumeSchema).optional(),
  }),
});

export type StackData = z.infer<typeof StackSchema>;
export type StackResourceData = z.infer<typeof StackResourceSchema>;
export type VolumeData = z.infer<typeof VolumeSchema>;
export type VolumeFormData = z.infer<typeof VolumeFormSchema>;

// FIXME: Remove this helper if OpenAPI spec supports conditional fields (e.g., oneOf) for UI-only fields

function omitUIFieldsFromResource(resource: StackResourceData): Omit<StackResourceData, "sourceType" | "gitRevisionType" | "gitRevisionValue"> {
  const rest = { ...resource };
  delete (rest as Partial<StackResourceData>).sourceType;
  delete (rest as Partial<StackResourceData>).gitRevisionType;
  delete (rest as Partial<StackResourceData>).gitRevisionValue;

  return rest as Omit<StackResourceData, "sourceType" | "gitRevisionType" | "gitRevisionValue">;
}

function omitUIFieldsFromVolume(volume: VolumeFormData | VolumeData): Omit<VolumeFormData, "sourceType"> | Omit<VolumeData, "sourceType"> {
  const rest = { ...volume };
  delete (rest as Partial<VolumeFormData>).sourceType;

  return rest;
}

export function stripUIFieldsFromStackData(stackData: StackData): Omit<Stack, 'workspace_name'> {
  // Process all stack resources by removing UI-only fields
  const cleanStackResources = stackData.spec.stack_resources.map(omitUIFieldsFromResource);

  // Process volumes data if present
  const cleanVolumes = stackData.spec.volumes
    ? stackData.spec.volumes.map(omitUIFieldsFromVolume)
    : undefined;

  // Create a new clean spec object that will only include API-expected fields
  const cleanSpec: Partial<Stack["spec"]> = {
    stack_resources: cleanStackResources as Stack["spec"]["stack_resources"],
  };

  if (cleanVolumes && cleanVolumes.length > 0) {
    cleanSpec.volumes = cleanVolumes as Stack["spec"]["volumes"];
  }

  // Combine everything into the final API-compliant object
  return {
    name: stackData.name,
    labels: stackData.labels,
    spec: cleanSpec as Stack["spec"],
  };
}

