/**
 * Sources that hand a pre-built stack into the create form via `location.state`.
 * The create page prefills its form when the incoming `importSource` is one of these.
 */
export const ImportSource = {
  DockerCompose: "docker-compose",
  Template: "template",
} as const;

export type ImportSource = (typeof ImportSource)[keyof typeof ImportSource];

export const PREFILL_IMPORT_SOURCES: ImportSource[] = [
  ImportSource.DockerCompose,
  ImportSource.Template,
];

export function isPrefillSource(source: unknown): source is ImportSource {
  return PREFILL_IMPORT_SOURCES.includes(source as ImportSource);
}
