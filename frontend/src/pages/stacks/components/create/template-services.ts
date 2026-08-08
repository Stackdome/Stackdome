import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser"
import type { Template } from "@/pages/stacks/data/templates/types"

/**
 * The service names a template puts on the canvas.
 *
 * A template's `stackYaml` **is** a real compose document, so this reads it with
 * the same parser the compose tab uses rather than keeping a second, hand-kept
 * list of services that would drift the first time a template changed.
 */
export function templateServices(template: Template): string[] {
  try {
    const parsed = parseAndValidateDockerCompose(template.stackYaml) as {
      services?: Record<string, unknown>
    }
    return Object.keys(parsed.services ?? {})
  } catch {
    // A template whose YAML no longer parses is a broken template, and the
    // import will say so far more usefully than a count would.
    return []
  }
}
