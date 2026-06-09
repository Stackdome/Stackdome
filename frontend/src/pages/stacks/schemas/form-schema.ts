/**
 * Form schemas for stack resources
 * These extend the API schemas with additional UI-specific fields and validations
 */
import { z } from "zod";
import {
  ApiStackResourceSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSpecSchema,
  ApiVolumeSchema,
  ApiStackSchema,
} from "./api-schema";
import type { StackUpdateRequest, StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import { ADDON_OUTPUT_FIELDS } from "@/pages/stacks/lib/addon-presets";
import { buildDesiredConnections } from "@/pages/stacks/lib/connection-mapping";
import type { FormEnvRow } from "@/pages/stacks/lib/connection-mapping";

/**
 * Form-specific UI schema additions
 */
const FormGitRevisionTypeSchema = z.enum(["commit", "branch", "tag"]);

const FormEnvVarSchema = z.union([
  z.object({
    from: z.literal("stack"),
    name: z.string().min(1, "Required"),
    value: z.string(),
  }),
  z.object({
    from: z.literal("secret"),
    name: z.string().min(1, "Required"),
    secretId: z.string().min(1, "Pick a secret"),
    secretKey: z.string().min(1, "Pick a key"),
  }),
  z
    .object({
      from: z.literal("addon"),
      name: z.string().min(1, "Required"),
      addonId: z.string().min(1, "Pick an addon"),
      database: z.string().optional(),
      superuser: z.boolean().default(false),
      credField: z.enum(ADDON_OUTPUT_FIELDS).optional(),
    })
    .refine((d) => d.superuser || (typeof d.database === "string" && d.database.length > 0), {
      message: "Pick a database",
      path: ["database"],
    })
    .refine((d) => typeof d.credField === "string" && d.credField.length > 0, {
      message: "Pick a field",
      path: ["credField"],
    }),
  z.object({
    from: z.literal("resource"),
    name: z.string().min(1, "Required"),
    resourceName: z.string().min(1, "Pick a resource"),
    output: z.string().min(1, "Pick an output"),
  }),
  z.object({
    from: z.literal("self"),
    name: z.string().min(1, "Required"),
    selfOutput: z.string().min(1, "Pick an output"),
  }),
]);

type FormEnvVarData = z.infer<typeof FormEnvVarSchema>;

// Compile-time assertion: the zod-inferred env-row type must stay structurally
// identical to FormEnvRow (the canonical union in connection-mapping.ts).
type _AssertEnvRowsMatch = FormEnvVarData extends FormEnvRow
  ? FormEnvRow extends FormEnvVarData
    ? true
    : ["FormEnvRow has arms FormEnvVarData lacks"]
  : ["FormEnvVarData has arms FormEnvRow lacks"];
const _envRowsMatch: _AssertEnvRowsMatch = true;
void _envRowsMatch;

const FormStackResourceSchema = ApiStackResourceSchema.extend({
  // UI helper, not part of API spec for StackResource
  sourceType: z.enum(["image", "git"]).optional().default("image"),
  // UI helper fields for git revision, not part of API spec StackResource
  gitRevisionType: FormGitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
  // UI helper fields for secrets, not part of API spec StackResource
  useImageSecret: z.boolean().optional().default(false),
  selectedImageSecretId: z.string().optional(),
  useGitSecret: z.boolean().optional().default(false),
  selectedGitSecretId: z.string().optional(),
  // Override execution_config to use our form env var schema. The
  // environment_variables list holds literal env-var rows (`from: "stack"`);
  // the API array is reconstructed in the converter on save.
  execution_config: z.object({
    command: z.array(z.string()).optional(),
    args: z.array(z.string()).optional(),
    environment_variables: z.array(FormEnvVarSchema).optional(),
  }).optional(),
}).superRefine((data, ctx) => {
  // Validate that git revision fields are required when sourceType is git
  if (data.sourceType === "git") {
    if (!data.gitRevisionType) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required when source is Git",
        path: ["gitRevisionType"],
      });
    }

    if (!data.gitRevisionValue) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required when source is Git",
        path: ["gitRevisionValue"],
      });
    }

    // Make sure there's a valid Git repo URL when using Git repository
    if (!data.build_spec?.source_context?.git_repo?.repo_url) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required when source is Git",
        path: ["build_spec", "source_context", "git_repo", "repo_url"],
      });
    }
  }

  // Validate that image URL is required when sourceType is image
  if (data.sourceType === "image") {
    if (!data.image_spec?.image) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required",
        path: ["image_spec", "image"],
      });
    }
  }
});

const FormVolumeSourceSchema = ApiVolumeSourceSchema.superRefine(
  (data, ctx) => {
    // Validate that the appropriate source is provided based on source_type
    if (data.source_type === "GitRepo" && !data.git_repo_source) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required for GitRepo source",
        path: ["git_repo_source"],
      });
    }

    if (data.source_type === "RemoteDir" && !data.remote_source) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required for RemoteDir source",
        path: ["remote_source"],
      });
    }

    if (
      data.source_type === "BuildArtifact" &&
      (!data.build_source || data.build_source.length === 0)
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Required for BuildArtifact source",
        path: ["build_source"],
      });
    }
  }
);

const FormVolumeSpecSchema = ApiVolumeSpecSchema.extend({
  source: FormVolumeSourceSchema.optional(),
});

const FormVolumeSchema = ApiVolumeSchema.extend({
  spec: FormVolumeSpecSchema,
});

// Helper for UI to manage different source types
const FormVolumeExtendedSchema = FormVolumeSchema.extend({
  // UI helper field, not part of the API spec
  sourceType: z
    .enum(["None", "GitRepo", "RemoteDir", "BuildArtifact"])
    .default("None"),
});

const FormStackSchema = ApiStackSchema.extend({
  spec: ApiStackSchema.shape.spec.extend({
    stack_resources: z
      .array(FormStackResourceSchema)
      .min(1, "Add at least one resource"),
    volumes: z.array(FormVolumeSchema).optional(),
  }),
}).superRefine((stack, ctx) => {
  const names = new Set(
    (stack.spec?.stack_resources ?? [])
      .map(r => r?.name)
      .filter((n): n is string => !!n),
  );
  (stack.spec?.stack_resources ?? []).forEach((r, idx) => {
    (r?.depends_on ?? []).forEach((dep, depIdx) => {
      if (!dep || !names.has(dep)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["spec", "stack_resources", idx, "depends_on", depIdx],
          message: dep
            ? `Unknown resource "${dep}"`
            : "Required",
        });
      }
    });
  });
});

/**
 * Type definitions for form data
 */
type FormStackData = z.infer<typeof FormStackSchema>;
type FormStackResourceData = z.infer<typeof FormStackResourceSchema> & {
  status?: unknown;
  // Server-computed, read-only in the API schema; carried in form data for the
  // Self/Resource output pickers but stripped before any PUT.
  outputs?: unknown;
};
type FormVolumeData = z.infer<typeof FormVolumeSchema> & {
  status?: unknown;
};
type FormVolumeExtendedData = z.infer<typeof FormVolumeExtendedSchema> & {
  status?: unknown;
};

/**
 * Conversion utilities between API and Form schemas
 */

// Remove UI-only fields from a resource before sending to API
function convertFormResourceToApiResource(
  resource: FormStackResourceData
): StackResourceUpdateRequest {

  const {
    sourceType,
    gitRevisionType,
    gitRevisionValue,
    useImageSecret,
    selectedImageSecretId,
    useGitSecret,
    selectedGitSecretId,
    status,
    outputs,
    ...rest
  } = resource;

  // Clean volume_mounts to remove read-only fields (stack_resource_id and source_volume_type)
  const cleanedVolumeMounts = rest.volume_mounts?.map((volumeMount) => {
    // Remove read-only fields from volume mount
    const {

      stack_resource_id,

      source_volume_type,
      ...cleanVolumeMount
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } = volumeMount as any;

    return cleanVolumeMount;
  });

  // Only literal (stack) and self rows persist as env vars. Secret/addon/resource
  // rows persist as stack connections, handled in the save orchestration.
  const envVars = (rest.execution_config?.environment_variables ?? []) as FormEnvVarData[];

  const apiEnvVars = envVars.flatMap((r): { name: string; value?: string; self_output?: string }[] => {
    if (r.from === "stack") return [{ name: r.name, value: r.value }];
    if (r.from === "self") return [{ name: r.name, self_output: r.selfOutput }];
    return [];
  });

  const processedExecutionConfig = rest.execution_config
    ? {
      ...rest.execution_config,
      environment_variables: apiEnvVars,
    }
    : undefined;

  return {
    ...rest,
    volume_mounts: cleanedVolumeMounts,
    execution_config: processedExecutionConfig,
  } as StackResourceUpdateRequest;
}

// Remove UI-only fields from a volume before sending to API
function convertFormVolumeToApiVolume(
  volume: FormVolumeExtendedData | FormVolumeData
): VolumeUpdateRequest {

  const { sourceType, status, ...rest } = volume as FormVolumeExtendedData;
  // Ensure needs_sync_before_use is always present and boolean
  // Ensure remote_source.current_directory_hash and path are always present and strings if remote_source exists
  let fixedSource = rest.spec?.source;
  if (fixedSource && fixedSource.source_type === "RemoteDir") {
    if (fixedSource.remote_source) {
      fixedSource = {
        ...fixedSource,
        remote_source: {
          path: fixedSource.remote_source.path || "",
          current_directory_hash: fixedSource.remote_source.current_directory_hash || "",
        },
      };
    } else {
      fixedSource = {
        ...fixedSource,
        remote_source: { path: "", current_directory_hash: "" },
      };
    }
  }
  return {
    ...rest,
    spec: {
      ...rest.spec,
      needs_sync_before_use: rest.spec?.needs_sync_before_use ?? false,
      source: fixedSource,
    },
  };
}

// Add UI-only fields to a resource from API data for use in forms
function convertApiResourceToFormResource(
  resource: z.infer<typeof ApiStackResourceSchema> & { status?: unknown } // API resource may have status
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

  // Env vars deserialize into literal + self rows. Secret/addon/resource rows
  // are merged in by the caller from the stack's connections.
  const processedEnvVars: FormEnvVarData[] = (
    resource.execution_config?.environment_variables ?? []
  ).map((v) =>
    v.self_output
      ? { from: "self" as const, name: v.name, selfOutput: v.self_output }
      : { from: "stack" as const, name: v.name, value: v.value ?? "" },
  );

  // Detect if secrets are being used
  const useImageSecret = Boolean(resource.image_spec?.pull_secret?.secret_id);
  const selectedImageSecretId = resource.image_spec?.pull_secret?.secret_id;

  const useGitSecret = Boolean(resource.build_spec?.source_context?.git_repo?.git_secret?.secret_id);
  const selectedGitSecretId = resource.build_spec?.source_context?.git_repo?.git_secret?.secret_id;

  // Ensure all required fields are present, defaulting as needed
  return {
    ...resource,
    name: resource.name ?? "",
    sourceType,
    gitRevisionType,
    gitRevisionValue,
    useImageSecret,
    selectedImageSecretId,
    useGitSecret,
    selectedGitSecretId,
    execution_config: resource.execution_config ? {
      ...resource.execution_config,
      environment_variables: processedEnvVars,
    } : undefined,
    status: resource.status ?? {},
  };
}

// Add UI-only fields to a volume from API data for use in forms
function convertApiVolumeToFormVolume(
  volume: z.infer<typeof ApiVolumeSchema> & { status?: unknown } // API volume may have status
): FormVolumeExtendedData {
  // Determine sourceType based on the volume's source specification
  let sourceType: "None" | "GitRepo" | "RemoteDir" | "BuildArtifact" = "None";

  if (volume.spec?.source) {
    sourceType = volume.spec.source.source_type;
  }

  // Ensure all required fields are present, defaulting as needed
  return {
    ...volume,
    sourceType,
    status: volume.status ?? {},
  };
}

// Convert a form stack to API stack for submission
function convertFormStackToApiStack(
  stackData: FormStackData
): StackUpdateRequest {
  // Filter out empty or invalid resources (resources with empty names or no image)
  const validResources = stackData.spec.stack_resources.filter(resource => {
    // A resource is valid if it has a name and either an image or build_spec
    const hasName = resource.name && resource.name.trim() !== '';
    const hasImage = resource.image_spec?.image && resource.image_spec.image.trim() !== '';
    const hasBuildSpec = resource.build_spec?.source_revision;

    return hasName && (hasImage || hasBuildSpec);
  });

  // Process all valid stack resources by removing UI-only fields
  const apiStackResources = validResources.map(resource => {
    // Always provide both keys for source_revision if build_spec is present
    if (resource.build_spec) {
      const gitRepoRev = resource.build_spec.source_revision?.git_repo_revision;
      resource.build_spec.source_revision = {
        volume_source_revision: undefined,
        git_repo_revision: gitRepoRev,
      };
    }

    // Add secret references before conversion
    const resourceWithSecrets = { ...resource };

    // Add image pull secret if selected
    if (resource.sourceType === 'image' && resource.useImageSecret && resource.selectedImageSecretId) {
      resourceWithSecrets.image_spec = {
        image: resourceWithSecrets.image_spec?.image || '',
        ...resourceWithSecrets.image_spec,
        pull_secret: {
          secret_id: resource.selectedImageSecretId
        }
      };
    }

    // Add git secret if selected
    if (resource.sourceType === 'git' && resource.useGitSecret && resource.selectedGitSecretId) {
      if (resourceWithSecrets.build_spec?.source_context?.git_repo) {
        resourceWithSecrets.build_spec.source_context.git_repo.git_secret = {
          secret_id: resource.selectedGitSecretId
        };
      }
    }

    return convertFormResourceToApiResource(resourceWithSecrets);
  });

  // Filter out empty or invalid volumes (volumes with empty names)
  const validVolumes = stackData.spec.volumes?.filter(volume => {
    return volume.name && volume.name.trim() !== '';
  });

  // Process volumes data if present
  const apiVolumes = validVolumes && validVolumes.length > 0
    ? validVolumes.map(convertFormVolumeToApiVolume)
    : undefined;

  // Secret/addon/resource env rows persist as stack connections, not env vars.
  // The per-resource serializer already drops them from environment_variables;
  // rebuild the full desired connection set here so the stack PUT carries it.
  const connections = buildDesiredConnections(
    validResources.map((r) => ({
      name: r.name ?? "",
      rows: (r.execution_config?.environment_variables ?? []) as FormEnvRow[],
    })),
  );

  // Create a new clean spec object that will only include API-expected fields
  const apiSpec = {
    stack_resources: apiStackResources as z.infer<typeof ApiStackSchema>["spec"]["stack_resources"],
    volumes: apiVolumes as z.infer<typeof ApiStackSchema>["spec"]["volumes"] | undefined,
    ...(connections.length > 0 ? { connections } : {}),
  };

  // Combine everything into the final API-compliant object
  return {
    name: stackData.name,
    labels: stackData.labels,
    spec: apiSpec,
  };
}

export type {
  FormStackData,
  FormStackResourceData,
  FormVolumeData,
  FormVolumeExtendedData,
  FormEnvVarData,
};

export {
  FormEnvVarSchema,
  FormVolumeExtendedSchema,
  FormStackSchema,
  convertApiResourceToFormResource,
  convertFormResourceToApiResource,
  convertApiVolumeToFormVolume,
  convertFormStackToApiStack,
};
