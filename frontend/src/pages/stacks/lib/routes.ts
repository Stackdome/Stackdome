/**
 * The three stack routes, as constants rather than a string repeated at seven
 * call sites.
 *
 * `/stacks/new` used to be the canvas on an unsaved draft. It is now the **New
 * stack** journey — the chooser — and the canvas moved to `/stacks/draft`.
 * The two are genuinely different screens: one picks a starting point, the
 * other edits the thing you picked.
 *
 * A draft has **no id until it is saved** (`isNewStack = !id` in the editor),
 * which is why the unsaved canvas needs a literal path of its own rather than
 * `/stacks/:id`. Nothing is written to the database until the user saves.
 */
export const NEW_STACK_PATH = "/stacks/new";

/** The canvas, on a draft that does not exist server-side yet. */
export const STACK_DRAFT_PATH = "/stacks/draft";

/** The canvas, on a saved stack. */
export function stackPath(id: string) {
  return `/stacks/${id}`;
}
