/**
 * Form schemas for stack resources
 * These extend the API schemas with additional UI-specific fields and validations
 */
import { z } from "zod";
import { join as shlexJoin, split as shlexSplit } from "shlex";
import {
  ApiStackResourceSchema,
  ApiPortSchema,
  ApiVolumeSourceSchema,
  ApiVolumeSpecSchema,
  ApiVolumeSchema,
  ApiStackSchema,
  ApiGitSourceSchema,
  ApiImageSourceSchema,
} from "./api-schema";
import type { StackUpdateRequest, StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import { ADDON_OUTPUT_FIELDS } from "@/pages/stacks/lib/addon-presets";
import { buildDesiredConnections, mountsToConnections } from "@/pages/stacks/lib/connection-mapping";
import type { FormEnvRow, FormMountRow } from "@/pages/stacks/lib/connection-mapping";
import { splitImageRef } from "@/pages/stacks/lib/image-ref";
import { DEFAULT_BUILD_CONTEXT, DEFAULT_DOCKERFILE_PATH } from "@/pages/stacks/lib/stack-model/policy";
import { withGitDefaults } from "@/pages/stacks/lib/stack-model/normalize";

/**
 * Form-specific UI schema additions
 */
const FormGitRevisionTypeSchema = z.enum(["branch", "tag"]);

/**
 * Command/args are argv arrays in the API (K8s exec form). The form holds them
 * as a single terminal-style string so typing is never re-parsed per keystroke;
 * conversion happens only here, at the load/save boundary. shlex keeps the
 * round-trip lossless for quoted arguments (e.g. `sh -c 'a && b'`).
 */
export function argvToText(argv: string[] | undefined): string | undefined {
  return argv && argv.length > 0 ? shlexJoin(argv) : undefined;
}

export function textToArgv(text: string | undefined): string[] | undefined {
  const trimmed = text?.trim();
  return trimmed ? shlexSplit(trimmed) : undefined;
}

/** Save-path variant: an emptied field must serialize as an EXPLICIT empty
 *  array — gorm's Updates() skips absent fields, so `undefined` would make a
 *  cleared command impossible to remove server-side. */
function textToArgvExplicit(text: string | undefined): string[] {
  return textToArgv(text) ?? [];
}

function refineArgvText(text: string | undefined, ctx: z.RefinementCtx, path: (string | number)[]) {
  if (!text?.trim()) return;
  try {
    shlexSplit(text);
  } catch {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Unclosed quote or trailing backslash",
      path,
    });
  }
}

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
    from: z.literal("resourceTemplate"),
    name: z.string().min(1, "Required"),
    resourceName: z.string().min(1, "Pick a resource"),
    template: z.string().min(1, "Required"),
    values: z.record(z.string()),
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
  // UI helper fields for git revision, not part of API spec StackResource.
  // The commit pin is separate: the API wants commit alongside a branch or
  // tag, never instead of one.
  gitRevisionType: FormGitRevisionTypeSchema.optional(),
  gitRevisionValue: z.string().optional(),
  gitCommitPin: z.string().optional(),
  // UI-only stash for the "Build from" toggle: the API rejects a source with
  // both `git` and `image` set (source_conflict), so we cannot keep both
  // subtrees live in `source`. Toggling away from a branch stashes it here so
  // toggling back restores it instead of resetting to blank defaults.
  stashedGitSource: ApiGitSourceSchema.optional(),
  stashedImageSource: ApiImageSourceSchema.optional(),
  // Port name is optional in the form — the API requires it, so we auto-derive
  // `port-<number>` on save rather than asking the user for it. Number is
  // optional in the form so the field can be cleared while typing; presence is
  // enforced in superRefine below.
  ports: z.array(ApiPortSchema.extend({
    name: z.string().optional(),
    number: z.number().int().optional(),
  })).optional(),
  // Command/args as terminal-style text (see argvToText/textToArgv).
  init_spec: z.object({
    command: z.string().optional(),
    args: z.string().optional(),
  }).optional(),
  // Override execution_config to use our form env var schema. The
  // environment_variables list holds literal env-var rows (`from: "stack"`);
  // the API array is reconstructed in the converter on save. Command/args as
  // terminal-style text, same as init_spec.
  execution_config: z.object({
    command: z.string().optional(),
    args: z.string().optional(),
    environment_variables: z.array(FormEnvVarSchema).optional(),
  }).optional(),
}).superRefine((data, ctx) => {
  refineArgvText(data.init_spec?.command, ctx, ["init_spec", "command"]);
  refineArgvText(data.init_spec?.args, ctx, ["init_spec", "args"]);
  refineArgvText(data.execution_config?.command, ctx, ["execution_config", "command"]);
  refineArgvText(data.execution_config?.args, ctx, ["execution_config", "args"]);

  (data.ports ?? []).forEach((port, i) => {
    if (port.number == null) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Port number is required",
        path: ["ports", i, "number"],
      });
    }
  });

  // Drive validation off the actual `source` union (same discriminant the canvas
  // node card uses) rather than the `sourceType` UI helper, which can desync.
  if (data.source?.git) {
    if (!data.source.git.repo_url?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Git repository URL is required",
        path: ["source", "git", "repo_url"],
      });
    }
    // Revision (branch/tag/commit) is optional — the builder defaults to the
    // repo's default branch when none is given. But if a revision *type* is
    // chosen, its value must be filled.
    if (data.gitRevisionType && !data.gitRevisionValue) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Enter a value for the selected revision, or clear the revision type",
        path: ["gitRevisionValue"],
      });
    }
    if (data.gitCommitPin && !data.gitRevisionType) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Pin requires a branch or tag",
        path: ["gitCommitPin"],
      });
    }
  } else if (data.source?.image) {
    // A ref like "ghcr.io/" (host picked, no repo typed) is non-empty but has
    // no actual image identity once the host segment is split off — reject
    // that the same as a fully empty ref. Whitespace-only refs/remainders are
    // rejected the same way (a ref of "  " round-trips through splitImageRef
    // untouched, so it must be checked explicitly, not just falsiness).
    const ref = data.source.image.ref ?? "";
    const { remainder } = splitImageRef(ref);
    if (!ref.trim() || !remainder.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Container image reference is required",
        path: ["source", "image", "ref"],
      });
    }
  } else {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Select a source — a container image or a Git repository",
      path: ["source"],
    });
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
  // Resources are addressed by name everywhere, so two sharing a name make
  // every reference to it ambiguous and a delete of either take both.
  const seenNames = new Set<string>();
  (stack.spec?.stack_resources ?? []).forEach((r, idx) => {
    if (r?.name) {
      if (seenNames.has(r.name)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["spec", "stack_resources", idx, "name"],
          message: `Name already in use: "${r.name}"`,
        });
      }
      seenNames.add(r.name);
    }
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
    // An env row sourcing another resource becomes a connection addressed by
    // name. A name no longer in the stack renders as "references unknown
    // resource" and fails the deploy, so reject it here instead.
    (r?.execution_config?.environment_variables ?? []).forEach((row, rowIdx) => {
      if (row?.from !== "resource" && row?.from !== "resourceTemplate") return;
      if (names.has(row.resourceName)) return;
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: [
          "spec",
          "stack_resources",
          idx,
          "execution_config",
          "environment_variables",
          rowIdx,
          "resourceName",
        ],
        message: `Unknown resource "${row.resourceName}"`,
      });
    });
  });
});

/**
 * Type definitions for form data
 */
type FormStackData = z.infer<typeof FormStackSchema>;
type FormStackResourceData = z.infer<typeof FormStackResourceSchema> & {
  // Server-computed, read-only in the API schema; carried in form data for the
  // Self/Resource output pickers but stripped before any PUT.
  outputs?: unknown;
};
type FormVolumeData = z.infer<typeof FormVolumeSchema>;
type FormVolumeExtendedData = z.infer<typeof FormVolumeExtendedSchema>;

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
    gitCommitPin,
    stashedGitSource,
    stashedImageSource,
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
      command: textToArgvExplicit(rest.execution_config.command),
      args: textToArgvExplicit(rest.execution_config.args),
      environment_variables: apiEnvVars,
    }
    : undefined;

  // A form init_spec exists once loaded from the server or touched in the
  // drawer — keep sending it (with explicit empty arrays) so clearing the
  // fields actually clears them server-side. Only a resource that never had
  // an init_spec omits it.
  const processedInitSpec = rest.init_spec
    ? {
      command: textToArgvExplicit(rest.init_spec.command),
      args: textToArgvExplicit(rest.init_spec.args),
    }
    : undefined;

  // The API requires a port name; derive `port-<number>` when the form left it
  // blank (k8s port names must contain a letter, so a bare number won't do).
  const apiPorts = rest.ports?.map((port) => ({
    ...port,
    // Null number → no auto-name ("port-undefined"); presence is validated before save.
    name: port.name && port.name.trim() !== "" ? port.name : port.number != null ? `port-${port.number}` : undefined,
  }));

  return {
    ...rest,
    ports: apiPorts,
    volume_mounts: cleanedVolumeMounts,
    init_spec: processedInitSpec,
    execution_config: processedExecutionConfig,
  } as StackResourceUpdateRequest;
}

// Remove UI-only fields from a volume before sending to API
function convertFormVolumeToApiVolume(
  volume: FormVolumeExtendedData | FormVolumeData
): VolumeUpdateRequest {

  const { sourceType, ...rest } = volume as FormVolumeExtendedData;
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
  resource: z.infer<typeof ApiStackResourceSchema>
): FormStackResourceData {
  const git = resource.source?.git;
  const sourceType: "image" | "git" = git ? "git" : "image";
  let gitRevisionType: "branch" | "tag" | undefined = undefined;
  let gitRevisionValue: string | undefined = undefined;
  let gitCommitPin: string | undefined = undefined;

  if (git) {
    if (git.branch) {
      gitRevisionType = "branch";
      gitRevisionValue = git.branch;
    } else if (git.tag) {
      gitRevisionType = "tag";
      gitRevisionValue = git.tag;
    }
    // Independent of branch/tag — never gate this on gitRevisionType, or an
    // invalid commit-only spec silently drops the pin.
    gitCommitPin = git.commit;
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

  // Ensure all required fields are present, defaulting as needed
  return {
    ...resource,
    name: resource.name ?? "",
    sourceType,
    // The form spells the git build defaults out in its fields, so the read
    // path must too — otherwise every unset one reads as a user edit.
    ...(git ? { source: { ...resource.source, git: withGitDefaults(git) } } : {}),
    gitRevisionType,
    gitRevisionValue,
    gitCommitPin,
    init_spec: resource.init_spec ? {
      command: argvToText(resource.init_spec.command),
      args: argvToText(resource.init_spec.args),
    } : undefined,
    execution_config: resource.execution_config ? {
      ...resource.execution_config,
      command: argvToText(resource.execution_config.command),
      args: argvToText(resource.execution_config.args),
      environment_variables: processedEnvVars,
    } : undefined,
  };
}

// Add UI-only fields to a volume from API data for use in forms
function convertApiVolumeToFormVolume(
  volume: z.infer<typeof ApiVolumeSchema>
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
  };
}

// Convert a form stack to API stack for submission
// Prepare one form resource for the API: normalize git source_revision, attach
// selected secrets, then strip UI-only fields.
function prepareFormResourceForApi(resource: FormStackResourceData): StackResourceUpdateRequest {
  const prepared = { ...resource };

  if (resource.sourceType === 'git') {
    const existingGit = resource.source?.git;
    // Flatten the UI revision helpers into one of branch/tag; the commit pin
    // rides alongside.
    // push is optional (omit => internal cluster registry). integration_id
    // carries the clone-auth credential reference (e.g. seeded by the "From
    // git provider" wizard for private repos) and must pass through as-is.
    prepared.source = {
      git: {
        repo_url: existingGit?.repo_url ?? '',
        dockerfile_path: existingGit?.dockerfile_path ?? DEFAULT_DOCKERFILE_PATH,
        build_context: existingGit?.build_context ?? DEFAULT_BUILD_CONTEXT,
        branch: resource.gitRevisionType === 'branch' ? resource.gitRevisionValue : undefined,
        tag: resource.gitRevisionType === 'tag' ? resource.gitRevisionValue : undefined,
        commit: resource.gitCommitPin || undefined,
        push: existingGit?.push?.repository ? { repository: existingGit.push.repository } : undefined,
        integration_id: existingGit?.integration_id,
      },
    };
  } else if (resource.sourceType === 'image') {
    prepared.source = {
      image: {
        ref: resource.source?.image?.ref ?? '',
        registry_credentials_id: resource.source?.image?.registry_credentials_id,
      },
    };
  }

  return convertFormResourceToApiResource(prepared);
}

function convertFormStackToApiStack(
  stackData: FormStackData
): StackUpdateRequest {
  // Filter out empty or invalid resources (resources with empty names or no image)
  const validResources = stackData.spec.stack_resources.filter(resource => {
    // A resource is valid if it has a name and either an image or git source
    const hasName = resource.name && resource.name.trim() !== '';
    const hasImage = resource.source?.image?.ref && resource.source.image.ref.trim() !== '';
    const hasGit = resource.source?.git?.repo_url && resource.source.git.repo_url.trim() !== '';

    return hasName && (hasImage || hasGit);
  });

  // Process all valid stack resources by removing UI-only fields. Mounts are
  // fully represented as volume_mount connections (built below) — strip
  // volume_mounts so the payload has one source of truth, matching the
  // draft-sync save path (the server returns volume_mounts: [] and stores
  // mounts in connections).
  const apiStackResources = validResources.map((resource) => ({
    ...prepareFormResourceForApi(resource),
    volume_mounts: undefined,
  }));

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

  // Volume mounts persist as volume_mount connections too (compose blocks and
  // the canvas record mounts as form rows on the resource). Mounts referencing
  // volumes absent from the payload are dangling — never emit them.
  const volumeNames = new Set((validVolumes ?? []).map((v) => v.name));
  for (const r of validResources) {
    const liveMounts = ((r.volume_mounts ?? []) as FormMountRow[]).filter(
      (m) => volumeNames.has(m.source_volume_name ?? ""),
    );
    connections.push(...mountsToConnections(r.name ?? "", liveMounts));
  }

  // Create a new clean spec object that will only include API-expected fields
  const apiSpec = {
    stack_resources: apiStackResources,
    volumes: apiVolumes,
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
  FormStackResourceSchema,
  convertApiResourceToFormResource,
  convertFormResourceToApiResource,
  convertApiVolumeToFormVolume,
  convertFormVolumeToApiVolume,
  convertFormStackToApiStack,
  prepareFormResourceForApi,
};
