import type { PickedRepo } from "@/components/git-source-picker/types"
import type { Template } from "@/pages/stacks/data/templates/types"

/**
 * One block put into the stack. The same block can go in repeatedly, so an
 * instance carries its own de-duplicated `name` — `postgres`, `postgres-2` —
 * alongside the catalogue id it came from.
 */
export interface BlockInstance {
  blockId: string
  /** The resource name. Machine-set, and unique within the stack. */
  name: string
  /** The catalogue name, for the "In this stack" subtitle. */
  label: string
}

/** What the compose tab read out of the file it was given. */
export interface ComposePreview {
  services: string[]
  volumes: string[]
  warnings: string[]
  error: string | null
}

/**
 * Everything the five tabs have gathered. It is one object rather than five
 * because the footer, the "In this stack" panel and `Create stack` all have to
 * read whichever tab is live without knowing which one that is.
 *
 * **Switching tabs does not discard the others.** Picking a repository and then
 * looking at the templates must not silently throw the repository away — the
 * strip is a set of peers you move between, not a wizard step you commit to.
 */
export interface Selection {
  git: { mode: "provider" | "url"; repo: PickedRepo | null; url: string }
  template: Template | null
  compose: { yaml: string; preview: ComposePreview | null }
  blocks: { instances: BlockInstance[]; addonIds: string[] }
}

export function emptySelection(): Selection {
  return {
    git: { mode: "provider", repo: null, url: "" },
    template: null,
    compose: { yaml: "", preview: null },
    blocks: { instances: [], addonIds: [] },
  }
}

/**
 * The next free name for a block. Mirrors `uniqueName()` inside
 * `addBlockToStack`, which is what actually renames on the way to the canvas —
 * this exists so the panel can show the name *before* you get there.
 */
export function uniqueBlockName(blockId: string, taken: BlockInstance[]): string {
  const names = new Set(taken.map((i) => i.name))
  if (!names.has(blockId)) return blockId
  let n = 2
  while (names.has(`${blockId}-${n}`)) n += 1
  return `${blockId}-${n}`
}
