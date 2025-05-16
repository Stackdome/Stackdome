import { z } from 'zod';

// Basic Schemas based on // user input fields from openapi.d.ts

export const LabelSchema = z.object({
  key: z.string().min(1, "Key is required"),
  value: z.string().min(1, "Value is required"),
});

export const PortSchema = z.object({
  number: z.number().int().min(1, "Port number is required"),
  protocol: z.enum(['tcp', 'http', 'udp']).optional().default('tcp'), // Assuming tcp/http/udp are common, default to tcp
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

export const StackCreateSchema = z.object({
  name: z.string().min(1, "Stack name is required"),
  workspace_name: z.string().min(1, "Workspace name is required"), // This might be pre-filled or selected
  labels: z.array(LabelSchema).optional(),
  spec: z.object({
    stack_resources: z.array(StackResourceSchema).min(1, "At least one stack resource is required"),
  }),
});

export type StackCreateData = z.infer<typeof StackCreateSchema>;
export type StackResourceData = z.infer<typeof StackResourceSchema>;
// For individual field errors, a simple map should suffice
// export type FieldErrors = Record<string, string | undefined>;
