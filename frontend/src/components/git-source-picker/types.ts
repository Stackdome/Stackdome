/** A repository chosen in the git source picker. Canonical home of the type
    (the enable-repo wizard re-exports it for its existing consumers). */
export interface PickedRepo {
  /** e.g. "acme/webapp" */
  fullName: string;
  cloneUrl: string;
  /** empty string when unknown (manual URL / credentials host) */
  defaultBranch: string;
  /** null when the user typed a public URL (no integration involved) */
  integrationId: string | null;
}
