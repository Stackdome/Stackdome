import { z } from 'zod';

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
  source_revision: z.object({ // Simplified, actual revision comes from git_repo_revision within
    git_repo_revision: GitRepoRevisionSchema.optional(),
    // volume_source_revision is not direct user input for this schema part
  }).optional(),
  image_repository_url: z.object({ // User input
    url: z.string().min(1, "Image repository URL is required"),
    cluster_registry_id: z.string().optional(), // Assuming this might be selected or auto-filled
  }),
  insecure_registry: z.boolean().optional().default(false),
});


// Main Schemas

export const StackResourceSchema = z.object({
  name: z.string().min(1, "Resource name is required"),
  labels: z.array(LabelSchema).optional(),
  build_spec: StackResourceBuildSpecSchema.optional(),
  image_spec: ImageSpecSchema.optional(),
  init_spec: InitSpecSchema.optional(),
  execution_config: ExecutionConfigSchema.optional(),
  depends_on: z.array(z.string()).optional(),
  ports: z.array(PortSchema).optional(),
  // UI helper, not part of API spec for StackResource itself
  sourceType: z.enum(["image", "git"]).optional().default("image"), 
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
  workspace_name: z.string().min(1, 'Workspace name is required'), 
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

