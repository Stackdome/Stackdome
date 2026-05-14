/**
 * API schemas — derived from generated zod-schemas.ts (single source of truth via OpenAPI).
 * Form-schema.ts and other consumers import the Api*Schema names from here.
 */
import { schemas } from "@/api/zod-schemas";

const ApiLabelSchema = schemas.Label;
const ApiAnnotationSchema = schemas.Annotation;
const ApiPortSchema = schemas.Port;
const ApiSecretRefSchema = schemas.SecretRef;
const ApiImageSpecSchema = schemas.ImageSpec;
const ApiEnvVarSchema = schemas.EnvVar;
const ApiEnvVarFromSecretSchema = schemas.EnvVarFromSecret;
const ApiPostgresAddonEnvSourceSchema = schemas.PostgresAddonEnvSource;
const ApiAddonEnvSourceSchema = schemas.AddonEnvSource;
const ApiExecutionConfigSchema = schemas.ExecutionConfig;
const ApiInitSpecSchema = schemas.InitSpec;
const ApiGitRepoRevisionSchema = schemas.GitRepoRevision;
const ApiBuildSourceContextSchema = schemas.BuildSourceContext;
const ApiStackResourceBuildSpecSchema = schemas.StackResourceBuildSpec;
const ApiVolumeMountSchema = schemas.VolumeMount;
const ApiStackResourceSchema = schemas.StackResource;
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
  ApiAddonEnvSourceSchema,
  ApiAnnotationSchema,
  ApiBuildArtifactSchema,
  ApiBuildSourceContextSchema,
  ApiEnvVarFromSecretSchema,
  ApiEnvVarSchema,
  ApiExecutionConfigSchema,
  ApiGitRepoRevisionSchema,
  ApiGitRepoSourceSchema,
  ApiImageSpecSchema,
  ApiInitSpecSchema,
  ApiLabelSchema,
  ApiPortSchema,
  ApiPostgresAddonEnvSourceSchema,
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
