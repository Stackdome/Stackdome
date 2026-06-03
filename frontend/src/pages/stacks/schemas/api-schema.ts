/**
 * API schemas — derived from generated zod-schemas.ts (single source of truth via OpenAPI).
 * Form-schema.ts and other consumers import the Api*Schema names from here.
 */
import { z } from "zod";
import { schemas } from "@/api/zod-schemas";

const ApiLabelSchema = schemas.Label;
const ApiAnnotationSchema = schemas.Annotation;
const ApiPortSchema = schemas.Port;
const ApiSecretRefSchema = schemas.SecretRef;
const ApiImageSpecSchema = schemas.ImageSpec;
const ApiEnvVarSchema = schemas.EnvVar;
const ApiExecutionConfigSchema = schemas.ExecutionConfig;
const ApiInitSpecSchema = schemas.InitSpec;
const ApiGitRepoRevisionSchema = schemas.GitRepoRevision;
const ApiBuildSourceContextSchema = schemas.BuildSourceContext;

// ImageRepository — generated schema + `cluster_registry_id` which the frontend uses
// when the user picks an internal cluster registry. Not yet in OpenAPI.
// TODO(openapi): add `cluster_registry_id?: string` to schemas.ImageRepository.
const ApiImageRepositorySchema = schemas.ImageRepository.extend({
  cluster_registry_id: z.string().optional(),
});

// StackResourceBuildSpec — generated schema + `insecure_registry` flag the frontend
// surfaces in the build form, and the extended ImageRepository above.
// TODO(openapi): add `insecure_registry?: boolean` to schemas.StackResourceBuildSpec.
const ApiStackResourceBuildSpecSchema = schemas.StackResourceBuildSpec.extend({
  image_repository: ApiImageRepositorySchema,
  insecure_registry: z.boolean().optional(),
});

const ApiVolumeMountSchema = schemas.VolumeMount;

// StackResource — rebuilt with the extended build_spec so consumers (form-schema,
// converters, page components) see the frontend-only fields above.
const ApiStackResourceSchema = schemas.StackResource.extend({
  build_spec: ApiStackResourceBuildSpecSchema.optional(),
});
// Generated emits this enum as `VolumeSourceTypes` (plural) — alias for our singular name.
const ApiVolumeSourceTypeSchema = schemas.VolumeSourceTypes;
const ApiRemoteSourceSchema = schemas.RemoteSource;
const ApiBuildArtifactSchema = schemas.BuildArtifact;
const ApiGitRepoSourceSchema = schemas.GitRepoSource;
const ApiVolumeSourceSchema = schemas.VolumeSource;
const ApiVolumeAccessModeSchema = schemas.VolumeAccessMode;
const ApiVolumeSpecSchema = schemas.VolumeSpec;
const ApiVolumeSchema = schemas.Volume;
const ApiStackSpecSchema = schemas.StackSpec;
const ApiStackSchema = schemas.Stack;
const ApiStackResourceStatusSchema = schemas.StackResourceStatus;
const ApiVolumeStatusSchema = schemas.VolumeStatus;

export {
  ApiAnnotationSchema,
  ApiBuildArtifactSchema,
  ApiBuildSourceContextSchema,
  ApiEnvVarSchema,
  ApiExecutionConfigSchema,
  ApiGitRepoRevisionSchema,
  ApiGitRepoSourceSchema,
  ApiImageRepositorySchema,
  ApiImageSpecSchema,
  ApiInitSpecSchema,
  ApiLabelSchema,
  ApiPortSchema,
  ApiRemoteSourceSchema,
  ApiSecretRefSchema,
  ApiStackResourceBuildSpecSchema,
  ApiStackResourceSchema,
  ApiStackResourceStatusSchema,
  ApiStackSchema,
  ApiStackSpecSchema,
  ApiVolumeAccessModeSchema,
  ApiVolumeMountSchema,
  ApiVolumeSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSourceTypeSchema,
  ApiVolumeSpecSchema,
  ApiVolumeStatusSchema,
};
