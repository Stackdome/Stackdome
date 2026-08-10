/**
 * API schemas — derived from generated zod-schemas.ts (single source of truth via OpenAPI).
 * Form-schema.ts and other consumers import the Api*Schema names from here.
 */
import { schemas } from "@/api/zod-schemas";

const ApiLabelSchema = schemas.Label;
const ApiAnnotationSchema = schemas.Annotation;
const ApiPortSchema = schemas.Port;
const ApiEnvVarSchema = schemas.EnvVar;
const ApiExecutionConfigSchema = schemas.ExecutionConfig;
const ApiInitSpecSchema = schemas.InitSpec;
const ApiGitRepoRevisionSchema = schemas.GitRepoRevision;
const ApiBuildSourceContextSchema = schemas.BuildSourceContext;

// Source union (replaces the old image_spec / build_spec split).
const ApiSourceSpecSchema = schemas.SourceSpec;
const ApiGitSourceSchema = schemas.GitSource;
const ApiImageSourceSchema = schemas.ImageSource;

const ApiVolumeMountSchema = schemas.VolumeMount;

const ApiStackResourceSchema = schemas.StackResource;

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
  ApiBuildSourceContextSchema,
  ApiEnvVarSchema,
  ApiExecutionConfigSchema,
  ApiGitRepoRevisionSchema,
  ApiGitRepoSourceSchema,
  ApiGitSourceSchema,
  ApiImageSourceSchema,
  ApiInitSpecSchema,
  ApiLabelSchema,
  ApiPortSchema,
  ApiSourceSpecSchema,
  ApiStackResourceSchema,
  ApiStackResourceStatusSchema,
  ApiStackSchema,
  ApiStackSpecSchema,
  ApiVolumeAccessModeSchema,
  ApiVolumeMountSchema,
  ApiVolumeSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSpecSchema,
  ApiVolumeStatusSchema,
};
