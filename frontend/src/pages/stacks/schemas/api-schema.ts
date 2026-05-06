/**
 * API schemas - directly reflect API models from OpenAPI definition
 * These are the "raw" schemas that represent what the API expects to receive and returns
 */
import { z } from 'zod';


/**
 * Basic API schema definitions - bare minimum validation reflecting API models
 * These schemas provide minimal validation just to ensure type safety
 */
const ApiLabelSchema = z.object({
  key: z.string().min(1, "Required"),
  value: z.string().min(1, "Required"),
});

const ApiAnnotationSchema = ApiLabelSchema;

const ApiPortSchema = z.object({
  number: z.number().int().min(1, "Required"),
  protocol: z.enum(['tcp', 'http']).optional().default("tcp"),
  exposed_to_public: z.boolean().optional().default(false),
  subdomain_prefix: z.string().optional(),
});

const ApiSecretRefSchema = z.object({
  secret_id: z.string(),
});

const ApiImageSpecSchema = z.object({
  image: z.string().min(1, "Required"),
  pull_secret: ApiSecretRefSchema.optional(),
});

const ApiEnvVarSchema = z.object({
  name: z.string().min(1, "Required"),
  value: z.string(),
});

const ApiEnvVarFromSecretSchema = z.object({
  name: z.string().min(1, "Required"),
  secret_ref: ApiSecretRefSchema,
  key: z.string().min(1, "Required"),
});

const ApiPostgresAddonEnvSourceSchema = z.object({
  addon_id: z.string().min(1, "Required"),
  database: z.string().optional(),
  superuser: z.boolean(),
  env_mapping: z.record(z.string(), z.string()),
});

const ApiAddonEnvSourceSchema = z.object({
  postgres: ApiPostgresAddonEnvSourceSchema.optional(),
});

const ApiExecutionConfigSchema = z.object({
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
  environment_variables: z.array(ApiEnvVarSchema).optional(),
  environment_variables_from_secret: z.array(ApiEnvVarFromSecretSchema).optional(),
  env_from_addons: z.array(ApiAddonEnvSourceSchema).optional(),
});

const ApiInitSpecSchema = z.object({
  image_spec: ApiImageSpecSchema.optional(),
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
});

const ApiGitRepoRevisionSchema = z
  .object({
    branch: z.object({ name: z.string().optional() }).optional(),
    commit: z.string().optional(),
    tag: z.string().optional(),
  })
  .refine((data) => data.branch?.name || data.commit || data.tag, {
    message:
      "At least one of branch, commit, or tag must be provided for Git revision.",
  });

const ApiBuildSourceContextSchema = z.object({
  git_repo: z
    .object({
      repo_url: z.string().url("Invalid Git repository URL"),
      git_secret: ApiSecretRefSchema.optional(),
    })
    .optional(),
});

const ApiStackResourceBuildSpecSchema = z.object({
  source_context: ApiBuildSourceContextSchema,
  context_path_within_source: z.string().optional().default("./"),
  dockerfile_path: z.string().optional().default("Dockerfile"),
  // Make source_revision always present and always with both keys
  source_revision: z.object({
    volume_source_revision: z.object({ current_volume_hash: z.string() }).optional(),
    git_repo_revision: ApiGitRepoRevisionSchema.optional(),
  }),
  image_repository: z.object({
    external_image_repo_url: z
      .string()
      .min(1, "Required"),
    use_internal_registry: z.boolean().optional(),
    cluster_registry_id: z.string().optional(),
  }),
  insecure_registry: z.boolean().optional().default(false),
});

const ApiVolumeMountSchema = z.object({
  source_volume_name: z.string().min(1, "Required"),
  source_sub_path: z.string().optional(),
  target_path: z.string().min(1, "Required"),
});

const ApiStackResourceSchema = z.object({
  name: z.string().min(1, "Required"),
  labels: z.array(ApiLabelSchema).optional(),
  build_spec: ApiStackResourceBuildSpecSchema.optional(),
  image_spec: ApiImageSpecSchema.optional(),
  init_spec: ApiInitSpecSchema.optional(),
  execution_config: ApiExecutionConfigSchema.optional(),
  depends_on: z.array(z.string()).optional(),
  ports: z.array(ApiPortSchema).optional(),
  volume_mounts: z.array(ApiVolumeMountSchema).optional(),
});

const ApiVolumeSourceTypeSchema = z.enum([
  'RemoteDir',
  'BuildArtifact',
  'GitRepo'
]);

const ApiRemoteSourceSchema = z.object({
  path: z.string().min(1, "Required"),
  current_directory_hash: z.string().default(""),
});

const ApiBuildArtifactSchema = z.object({
  resource_ref: z.string().min(1, "Required"),
  source_path: z.string().min(1, "Required"),
  destination_path: z.string().min(1, "Required"),
});

const ApiGitRepoSourceSchema = z.object({
  repo_url: z.string().url("Invalid Git repository URL"),
  revision: ApiGitRepoRevisionSchema,
});

const ApiVolumeSourceSchema = z.object({
  git_repo_source: ApiGitRepoSourceSchema.optional(),
  source_type: ApiVolumeSourceTypeSchema,
  remote_source: ApiRemoteSourceSchema.optional(),
  build_source: z.array(ApiBuildArtifactSchema).optional(),
});

const ApiVolumeAccessModeSchema = z.enum([
  'ReadWriteOnce',
  'ReadWriteMany',
  'ReadOnlyMany'
]);

const ApiVolumeSpecSchema = z.object({
  size: z.string().min(1, "Required"),
  storage_class: z.string().optional(),
  // needs_sync_before_use is required and must always be a boolean
  needs_sync_before_use: z.boolean().default(false),
  access_mode: ApiVolumeAccessModeSchema.default("ReadWriteOnce"),
  source: ApiVolumeSourceSchema.optional(),
});

const ApiVolumeSchema = z.object({
  name: z.string().min(1, "Required"),
  labels: z.array(ApiLabelSchema).optional(),
  annotations: z.array(ApiAnnotationSchema).optional(),
  spec: ApiVolumeSpecSchema,
});

const ApiStackSpecSchema = z.object({
  stack_resources: z.array(ApiStackResourceSchema).min(1, "Add at least one resource"),
  volumes: z.array(ApiVolumeSchema).optional(),
});

const ApiStackSchema = z.object({
  name: z.string().min(1, "Required"),
  labels: z.array(ApiLabelSchema).optional(),
  spec: ApiStackSpecSchema,
});

// Status schemas for StackResource and Volume (for use in detail components)
const ApiStackResourceStatusSchema = z.object({
  public_ingress: z.array(z.any()).optional(),
  internal_service_name: z.string().optional(),
  last_restart_request_processed_at: z.string().optional(),
  state: z.string().optional(),
  observed_version: z.number().optional(),
  conditions: z.array(z.any()).optional(),
});

const ApiVolumeStatusSchema = z.object({
  conditions: z.array(z.any()).optional(),
  phase: z.string().optional(),
  build_artifact_syncs: z.array(z.any()).optional(),
  last_synced_git_revision: z.string().optional(),
  last_remote_sync_hash: z.string().optional(),
});

export {
  ApiStackResourceSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSpecSchema,
  ApiVolumeSchema,
  ApiStackSchema,
  ApiStackResourceStatusSchema,
  ApiVolumeStatusSchema,
  ApiEnvVarSchema,
  ApiEnvVarFromSecretSchema,
  ApiExecutionConfigSchema,
};
