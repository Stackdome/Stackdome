/**
 * Backward compatibility layer for stack schemas
 * This file re-exports schemas from api-schema.ts and form-schema.ts for backward compatibility
 * New code should import directly from api-schema.ts or form-schema.ts based on needs
 */

import { type ApiStack } from './api-schema';
import {
  FormLabelSchema as LabelSchema,
  FormAnnotationSchema as AnnotationSchema,
  FormPortSchema as PortSchema,
  FormImageSpecSchema as ImageSpecSchema,
  FormEnvVarSchema as EnvVarSchema,
  FormExecutionConfigSchema as ExecutionConfigSchema,
  FormInitSpecSchema as InitSpecSchema,
  FormGitRevisionTypeSchema as GitRevisionTypeSchema,
  FormStackResourceUISchema as StackResourceFormUISchema,
  FormGitRepoRevisionSchema as GitRepoRevisionSchema,
  FormBuildSourceContextSchema as BuildSourceContextSchema,
  FormStackResourceBuildSpecSchema as StackResourceBuildSpecSchema,
  FormVolumeMountSchema as VolumeMountSchema,
  FormStackResourceSchema as StackResourceSchema,
  ApiVolumeAccessModeSchema as VolumeAccessModeSchema,
  FormVolumeSourceTypeSchema as VolumeSourceTypeSchema,
  FormRemoteSourceSchema as RemoteSourceSchema,
  FormBuildArtifactSchema as BuildArtifactSchema,
  FormGitRepoSourceSchema as GitRepoSourceSchema,
  FormVolumeSourceSchema as VolumeSourceSchema,
  FormVolumeSpecSchema as VolumeSpecSchema,
  FormVolumeSchema as VolumeSchema,
  FormVolumeExtendedSchema as VolumeFormSchema,
  FormStackSchema as StackSchema,
  type FormStackData as StackData,
  type FormStackResourceData as StackResourceData,
  type FormVolumeData as VolumeData,
  type FormVolumeExtendedData as VolumeFormData,
  convertFormResourceToApiResource as removeUIFieldsFromResource,
  convertFormVolumeToApiVolume as removeUIFieldsFromVolume,
  convertApiResourceToFormResource as addUIFieldsToResource,
  convertApiVolumeToFormVolume as addUIFieldsToVolume,
  convertFormStackToApiStack
} from './form-schema';

// For backward compatibility - rename to a function name that matches what was previously used
export function stripUIFieldsFromStackData(stackData: StackData): Omit<ApiStack, 'id' | 'created_at' | 'status'> {
  return convertFormStackToApiStack(stackData);
}

// Export everything that was previously exported
export {
  LabelSchema,
  AnnotationSchema,
  PortSchema,
  ImageSpecSchema,
  EnvVarSchema,
  ExecutionConfigSchema,
  InitSpecSchema,
  GitRevisionTypeSchema,
  StackResourceFormUISchema,
  GitRepoRevisionSchema,
  BuildSourceContextSchema,
  StackResourceBuildSpecSchema,
  VolumeMountSchema,
  StackResourceSchema,
  VolumeAccessModeSchema,
  VolumeSourceTypeSchema,
  RemoteSourceSchema,
  BuildArtifactSchema,
  GitRepoSourceSchema,
  VolumeSourceSchema,
  VolumeSpecSchema,
  VolumeSchema,
  VolumeFormSchema,
  StackSchema,
  // Export types
  type StackData,
  type StackResourceData,
  type VolumeData,
  type VolumeFormData,
  // Export conversion functions
  removeUIFieldsFromResource,
  removeUIFieldsFromVolume,
  addUIFieldsToResource,
  addUIFieldsToVolume,
};

