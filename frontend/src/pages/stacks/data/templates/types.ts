/**
 * A curated stack template surfaced in the Templates Browser.
 *
 * Display fields drive the browser UI; `stackYaml` is a real docker-compose
 * document that is converted (via the existing docker-compose import pipeline)
 * into the create-form payload when the user picks the template.
 */
export interface Template {
  /** Stable slug; also the default stack name. */
  id: string;
  /** Display name, e.g. "ToolJet". */
  name: string;
  /** Two-letter mono badge, used as a fallback when the icon is unavailable. */
  initials: string;
  /** Imported logo asset URL. */
  icon: string;
  /** Single category label, e.g. "Dev Tools". */
  category: string;
  /** One-line subtitle shown in the list row. */
  shortDescription: string;
  /** Longer blurb shown in the detail pane. */
  longDescription: string;
  website: string;
  docs: string;
  /** App version shown in the detail pane; matches the app image tag in `stackYaml`. */
  version: string;
  /** Real docker-compose document for the deployable preset. */
  stackYaml: string;
}
