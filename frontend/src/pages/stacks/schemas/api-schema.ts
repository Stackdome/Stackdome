/**
 * API schemas - directly reflect API models from OpenAPI definition
 * These are the "raw" schemas that represent what the API expects to receive and returns
 */
import { z } from 'zod';
import type { components } from '@/api/types/openapi';

/**
 * API Schema types - direct aliases to OpenAPI component schemas
 * These represent the raw API data structures
 */
export type ApiLabel = components["schemas"]["Label"];
export type ApiAnnotation = components["schemas"]["Annotation"];
export type ApiPort = components["schemas"]["Port"];
export type ApiImageSpec = components["schemas"]["ImageSpec"];
export type ApiEnvVar = components["schemas"]["EnvVar"];
export type ApiExecutionConfig = components["schemas"]["ExecutionConfig"];
export type ApiInitSpec = components["schemas"]["InitSpec"];
export type ApiStackResourceBuildSpec = components["schemas"]["StackResourceBuildSpec"];
export type ApiGitRepoRevision = components["schemas"]["GitRepoRevision"];
export type ApiVolumeMount = components["schemas"]["VolumeMount"];
export type ApiStackResource = components["schemas"]["StackResource"];
export type ApiVolumeSource = components["schemas"]["VolumeSource"];
export type ApiVolumeSpec = components["schemas"]["VolumeSpec"];
export type ApiVolume = components["schemas"]["Volume"];
export type ApiStackSpec = components["schemas"]["StackSpec"];
export type ApiStack = components["schemas"]["Stack"];

/**
 * Basic API schema definitions - bare minimum validation reflecting API models
 * These schemas provide minimal validation just to ensure type safety
 */
export const ApiLabelSchema = z.object({
  key: z.string(),
  value: z.string(),
});

export const ApiAnnotationSchema = ApiLabelSchema;

export const ApiPortSchema = z.object({
  number: z.number().int(),
  protocol: z.enum(['tcp', 'http']).optional(),
  exposed_to_public: z.boolean().optional(),
  subdomain_prefix: z.string().optional(),
});

export const ApiImageSpecSchema = z.object({
  image: z.string(),
});

export const ApiEnvVarSchema = z.object({
  name: z.string(),
  value: z.string(),
});

export const ApiExecutionConfigSchema = z.object({
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
  environment_variables: z.array(ApiEnvVarSchema).optional(),
});

export const ApiInitSpecSchema = z.object({
  image_spec: ApiImageSpecSchema.optional(),
  command: z.array(z.string()).optional(),
  args: z.array(z.string()).optional(),
});

export const ApiGitRepoRevisionSchema = z.object({
  branch: z.object({ name: z.string().optional() }).optional(),
  commit: z.string().optional(),
  tag: z.string().optional(),
});

export const ApiBuildSourceContextSchema = z.object({
  git_repo: z.object({
    repo_url: z.string(),
  }).optional(),
});

export const ApiStackResourceBuildSpecSchema = z.object({
  source_context: ApiBuildSourceContextSchema,
  context_path_within_source: z.string().optional(),
  dockerfile_path: z.string().optional(),
  source_revision: z.object({
    git_repo_revision: ApiGitRepoRevisionSchema.optional(),
  }).optional(),
  image_repository: z.object({
    external_image_repo_url: z.string(),
    use_internal_registry: z.boolean().optional(),
    cluster_registry_id: z.string().optional(),
  }),
  insecure_registry: z.boolean().optional(),
});

export const ApiVolumeMountSchema = z.object({
  source_volume_name: z.string(),
  source_sub_path: z.string().optional(),
  target_path: z.string(),
});

export const ApiStackResourceSchema = z.object({
  name: z.string(),
  labels: z.array(ApiLabelSchema).optional(),
  build_spec: ApiStackResourceBuildSpecSchema.optional(),
  image_spec: ApiImageSpecSchema.optional(),
  init_spec: ApiInitSpecSchema.optional(),
  execution_config: ApiExecutionConfigSchema.optional(),
  depends_on: z.array(z.string()).optional(),
  ports: z.array(ApiPortSchema).optional(),
  volume_mounts: z.array(ApiVolumeMountSchema).optional(),
});

export const ApiVolumeSourceTypeSchema = z.enum([
  'RemoteDir',
  'BuildArtifact',
  'GitRepo'
]);

export const ApiRemoteSourceSchema = z.object({
  path: z.string(),
  current_directory_hash: z.string().optional(),
});

export const ApiBuildArtifactSchema = z.object({
  resource_ref: z.string(),
  source_path: z.string(),
  destination_path: z.string(),
});

export const ApiGitRepoSourceSchema = z.object({
  repo_url: z.string(),
  revision: ApiGitRepoRevisionSchema,
});

export const ApiVolumeSourceSchema = z.object({
  git_repo_source: ApiGitRepoSourceSchema.optional(),
  source_type: ApiVolumeSourceTypeSchema,
  remote_source: ApiRemoteSourceSchema.optional(),
  build_source: z.array(ApiBuildArtifactSchema).optional(),
});

export const ApiVolumeAccessModeSchema = z.enum([
  'ReadWriteOnce',
  'ReadWriteMany',
  'ReadOnlyMany'
]);

export const ApiVolumeSpecSchema = z.object({
  size: z.string(),
  storage_class: z.string().optional(),
  needs_sync_before_use: z.boolean().optional(),
  access_mode: ApiVolumeAccessModeSchema,
  source: ApiVolumeSourceSchema.optional(),
});

export const ApiVolumeSchema = z.object({
  name: z.string(),
  labels: z.array(ApiLabelSchema).optional(),
  annotations: z.array(ApiAnnotationSchema).optional(),
  spec: ApiVolumeSpecSchema,
});

export const ApiStackSpecSchema = z.object({
  stack_resources: z.array(ApiStackResourceSchema),
  volumes: z.array(ApiVolumeSchema).optional(),
});

export const ApiStackSchema = z.object({
  name: z.string(),
  workspace_name: z.string(),
  labels: z.array(ApiLabelSchema).optional(),
  spec: ApiStackSpecSchema,
});
