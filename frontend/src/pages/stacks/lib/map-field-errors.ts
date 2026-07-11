// Maps backend validation field paths onto the resource-tab error map the stack
// editor already consumes. Backend paths use bracket indices and a different
// taxonomy for env vars; the editor keys are two-level (resource index → bare
// dot-notation leaf path). See use-resource-tab-props / getError for the target
// shape. Env keys must be exact because the env tab looks them up directly (no
// prefix-walk); everything else is matched by getError's prefix walk.
import type { ParsedFieldError } from "@/api/errors";
import type { EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";

export type MapDialect =
  | { dialect: "fat" }
  | { dialect: "thin"; resourceIndex: number };

export type MappedFieldErrors = {
  stackName?: string;
  settings: string[];
  connections: string[];
  resources: Record<number, Record<string, string>>;
  unmapped: ParsedFieldError[];
};

const RESOURCE_PREFIX = /^spec\.stack_resources\[(\d+)\]\.(.+)$/;
// The backend keys env errors under `execution_config.env[i]` (bare index, or
// with a `.name`/`.secret_key_ref`/`.self_output` leaf); the editor's env tab
// keys them under `execution_config.environment_variables.i.(name|value)`.
const ENV_PREFIX = /^execution_config\.env\.(\d+)/;
const ENV_TAB_KEY = "execution_config.environment_variables";

// Convert backend bracket indices (`ports[1]`) to the editor's dot segments
// (`ports.1`), then remap the env taxonomy onto the env-tab key with the
// correct .name / .value leaf.
function toEditorKey(barePath: string): string {
  const dotted = barePath.replace(/\[(\d+)\]/g, ".$1");
  const env = dotted.match(ENV_PREFIX);
  if (env) {
    const leaf = dotted.endsWith(".name") ? "name" : "value";
    return `${ENV_TAB_KEY}.${env[1]}.${leaf}`;
  }
  return dotted;
}

// Resource-drawer tab holding a given backend field path. Env fields land on
// the environment tab; everything else (including unrecognized paths) opens
// on configuration. Same prefix-walk as toEditorKey — reused so a jump target
// (e.g. the release-error banner) always agrees with the inline error map.
export function fieldTab(field: string): EditSessionTab {
  return toEditorKey(field).startsWith(ENV_TAB_KEY) ? "environment" : "configuration";
}

function put(resources: Record<number, Record<string, string>>, index: number, key: string, message: string) {
  const bucket = (resources[index] ??= {});
  if (bucket[key] === undefined) bucket[key] = message;
}

export function mapFieldErrors(errors: ParsedFieldError[], opts: MapDialect): MappedFieldErrors {
  const out: MappedFieldErrors = { settings: [], connections: [], resources: {}, unmapped: [] };

  for (const err of errors) {
    const { field, message } = err;

    if (opts.dialect === "thin") {
      if (field.startsWith("spec.")) {
        out.unmapped.push(err);
      } else {
        put(out.resources, opts.resourceIndex, toEditorKey(field), message);
      }
      continue;
    }

    // fat dialect
    const resourceMatch = field.match(RESOURCE_PREFIX);
    if (resourceMatch) {
      put(out.resources, Number(resourceMatch[1]), toEditorKey(resourceMatch[2]), message);
    } else if (field === "name") {
      out.stackName ??= message;
    } else if (field.startsWith("spec.settings")) {
      out.settings.push(message);
    } else if (field.startsWith("spec.connections")) {
      out.connections.push(message);
    } else {
      out.unmapped.push(err);
    }
  }

  return out;
}
