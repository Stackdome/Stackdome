import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Database, FileCode, GitBranch, Globe, HardDrive, Package } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useJourney } from "@/hooks/use-journey"
import { usePostgresAddons } from "@/hooks/use-postgres-addons"
import { getBlockById, blockCatalog } from "@/pages/stacks/data/blocks/registry"
import { DATA_BLOCK_CATEGORIES } from "@/pages/stacks/data/blocks/types"
import { addBlockToStack, emptyStack } from "@/pages/stacks/lib/block-to-form"
import { emptyDraftSeed } from "@/pages/stacks/lib/canvas/draft-seed"
import { buildGitSeed, defaultServiceName } from "@/pages/stacks/lib/git-source-seed"
import { templateToFormData } from "@/pages/stacks/data/templates/template-to-form"
import { convertDockerComposeToStackData } from "@/pages/stacks/lib/docker-compose-converter"
import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser"
import { DEFAULT_BUILD_CONTEXT, DEFAULT_DOCKERFILE_PATH } from "@/pages/stacks/lib/stack-model/policy"
import { STACK_DRAFT_PATH } from "@/pages/stacks/lib/routes"
import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema"
import type { DockerComposeFile } from "@/types/docker-compose"
import { BlockGlyph } from "@/pages/stacks/components/blocks/block-glyph"

import { InThisStack, type StackItem } from "./in-this-stack"
import { emptySelection, uniqueBlockName, type Selection } from "./selection"
import { StartingPointTabs } from "./starting-point-tabs"
import { SOURCE_LEDE, STARTING_POINTS, type Source } from "./starting-points"
import { templateServices } from "./template-services"
import { BlankTab } from "./tabs/blank-tab"
import { BlocksTab } from "./tabs/blocks-tab"
import { ComposeTab, parseCompose } from "./tabs/compose-tab"
import { RepositoryTab } from "./tabs/repository-tab"
import { TemplateDetail, TemplateTab } from "./tabs/template-tab"

/**
 * Where a service's port comes from when nobody was asked for one.
 *
 * The design moves branch, port, Dockerfile path and build context **off this
 * screen and onto the canvas** — they belong to a resource, not to a form step.
 * So the seed carries a default and the node inspector is where it is changed.
 */
const DEFAULT_SERVICE_PORT = 3000

/**
 * **New stack** — the journey (§12a). One page, five peer starting points, and
 * an ending that is always the same: a draft, open on the canvas.
 *
 * ### Why a page and not a dialog
 *
 * The flow it replaces was a modal wizard: pick a path, then commit to it. A
 * dialog is the wrong shape for a decision you might want to reverse — going
 * back cost a step, and the four paths could not be compared at all. Here they
 * are five tabs, all visible, and moving between them keeps what you already
 * chose.
 *
 * ### Cards on top, work below
 *
 * The starting points are **actions, not navigation**, so they sit as a
 * full-bleed strip across the top rather than in a left rail. A rail put 476px
 * — 37% of the width — of nav-shaped list beside a 240px sidebar that is
 * already navigation.
 *
 * ### The flow ends on the canvas
 *
 * `Create stack` makes a **draft** and opens it. Deploying happens on the
 * canvas, because a blocking error has to be seen before it is acted on, and
 * this screen has nowhere to show one.
 */
export default function CreateStackPage() {
  useJourney("/stacks", "New stack")
  const navigate = useNavigate()
  const { addons } = usePostgresAddons()

  const [source, setSource] = useState<Source>("git")
  const [selection, setSelection] = useState<Selection>(emptySelection)

  function update(patch: (previous: Selection) => Selection) {
    setSelection(patch)
  }

  const removeBlock = (index: number) =>
    update((s) => ({
      ...s,
      blocks: { ...s.blocks, instances: s.blocks.instances.filter((_, i) => i !== index) },
    }))

  const toggleAddon = (addonId: string) =>
    update((s) => ({
      ...s,
      blocks: {
        ...s.blocks,
        addonIds: s.blocks.addonIds.includes(addonId)
          ? s.blocks.addonIds.filter((id) => id !== addonId)
          : [...s.blocks.addonIds, addonId],
      },
    }))

  // What the LIVE tab has gathered. The other tabs keep their state, but the
  // panel and the footer only ever speak for the one you are looking at —
  // otherwise "In this stack" would describe a stack you are not making.
  const items = stackItems(source, selection, addons, removeBlock, toggleAddon)
  const ready = source === "blank" || items.length > 0

  function createStack() {
    navigate(STACK_DRAFT_PATH, { state: { seed: buildSeed(source, selection) } })
  }

  const side =
    source === "template" && selection.template ? (
      <TemplateDetail template={selection.template} />
    ) : items.length > 0 ? (
      <InThisStack items={items} />
    ) : null

  return (
    <div className="flex h-full min-h-0 flex-col">
      <StartingPointTabs
        options={STARTING_POINTS}
        value={source}
        onValueChange={setSource}
        aria-label="Start from"
      />

      {/* The body scrolls; the strip above and the footer below do not. */}
      <div className="min-h-0 flex-1 overflow-y-auto px-4 pb-4 pt-[18px]">
        <div className="flex items-start gap-7">
          <div className="w-[620px] max-w-full flex-none">
            <p className="text-body text-fg-2 mb-3.5">{SOURCE_LEDE[source]}</p>

            {source === "git" && (
              <RepositoryTab
                mode={selection.git.mode}
                onModeChange={(mode) => update((s) => ({ ...s, git: { ...s.git, mode } }))}
                repo={selection.git.repo}
                onRepoChange={(repo) => update((s) => ({ ...s, git: { ...s.git, repo } }))}
                url={selection.git.url}
                onUrlChange={(url) => update((s) => ({ ...s, git: { ...s.git, url } }))}
              />
            )}

            {source === "template" && (
              <TemplateTab
                picked={selection.template}
                onPick={(template) => update((s) => ({ ...s, template }))}
              />
            )}

            {source === "compose" && (
              <ComposeTab
                yaml={selection.compose.yaml}
                onChange={(yaml) =>
                  update((s) => ({ ...s, compose: { yaml, preview: parseCompose(yaml) } }))
                }
              />
            )}

            {source === "blocks" && (
              <BlocksTab
                instances={selection.blocks.instances}
                onAddBlock={(blockId) =>
                  update((s) => {
                    const block = getBlockById(blockId)
                    if (!block) return s
                    return {
                      ...s,
                      blocks: {
                        ...s.blocks,
                        instances: [
                          ...s.blocks.instances,
                          {
                            blockId,
                            name: uniqueBlockName(blockId, s.blocks.instances),
                            label: block.name,
                          },
                        ],
                      },
                    }
                  })
                }
                addonIds={selection.blocks.addonIds}
                onToggleAddon={toggleAddon}
              />
            )}

            {source === "blank" && <BlankTab />}
          </div>

          {side}
        </div>
      </div>

      {/* The one filled button on the screen (§9/§11), and the only thing in the
          footer. `flat`, not a pill: this does not ship anything — it opens the
          canvas, where Deploy is the commitment. */}
      <footer className="sheet-edge-t flex h-14 flex-none items-center justify-end px-4">
        <Button onClick={createStack} disabled={!ready}>
          Create stack
        </Button>
      </footer>
    </div>
  )
}

type Addon = ReturnType<typeof usePostgresAddons>["addons"][number]

/** Whether a block is a data store, so the panel shows the right glyph. */
function isDataBlock(blockId: string) {
  const block = blockCatalog.find((b) => b.id === blockId)
  return block ? DATA_BLOCK_CATEGORIES.has(block.category) : false
}

/**
 * What the live tab would put on the canvas, as rows.
 *
 * Only the blocks tab can remove one: everywhere else the items are derived
 * from a single pick, and taking one out would leave a selection that no longer
 * matches what is ticked in the list.
 */
function stackItems(
  source: Source,
  selection: Selection,
  addons: Addon[],
  removeBlock: (index: number) => void,
  toggleAddon: (addonId: string) => void,
): StackItem[] {
  if (source === "git") {
    const { mode, repo, url } = selection.git
    if (mode === "url") {
      if (!/^https?:\/\/\S+\/\S+/.test(url.trim())) return []
      const tail = url.trim().replace(/\/+$/, "").split("/").pop()!.replace(/\.git$/, "")
      return [{ key: "url", name: tail, sub: "from a public URL", icon: <GitBranch /> }]
    }
    if (!repo) return []
    return [
      {
        key: repo.fullName,
        name: defaultServiceName(repo),
        sub: `${repo.fullName} · ${repo.defaultBranch || "main"}`,
        icon: <GitBranch />,
      },
    ]
  }

  if (source === "template") {
    if (!selection.template) return []
    return templateServices(selection.template).map((name) => ({
      key: name,
      name,
      sub: "service",
      icon: <Package />,
    }))
  }

  if (source === "compose") {
    const preview = selection.compose.preview
    if (!preview || preview.error) return []
    return [
      ...preview.services.map((name) => ({
        key: `s:${name}`,
        name,
        sub: "service",
        icon: <FileCode />,
      })),
      ...preview.volumes.map((name) => ({
        key: `v:${name}`,
        name,
        sub: "volume",
        icon: <HardDrive />,
      })),
    ]
  }

  if (source === "blocks") {
    return [
      ...selection.blocks.instances.map((instance, index) => ({
        key: `b:${instance.name}`,
        name: instance.name,
        sub: instance.label,
        icon: isDataBlock(instance.blockId) ? (
          <BlockGlyph icon={blockCatalog.find((b) => b.id === instance.blockId)!.icon} size={14} />
        ) : (
          <Globe />
        ),
        // By position, not by block — removing "postgres-2" must not take
        // "postgres" with it.
        onRemove: () => removeBlock(index),
      })),
      ...selection.blocks.addonIds.map((id) => ({
        key: `a:${id}`,
        name: addons.find((a) => a.id === id)?.name ?? id,
        sub: "linked add-on",
        icon: <Database />,
        onRemove: () => toggleAddon(id),
      })),
    ]
  }

  return []
}

/** The navigation-state seed the canvas opens on. */
function buildSeed(source: Source, selection: Selection) {
  if (source === "git") {
    const { mode, repo, url } = selection.git
    const picked =
      mode === "url"
        ? { fullName: repoTailOf(url), cloneUrl: url.trim(), defaultBranch: "", integrationId: null }
        : repo
    if (!picked) return emptyDraftSeed()
    return buildGitSeed(picked, {
      serviceName: defaultServiceName(picked),
      branch: picked.defaultBranch || "main",
      dockerfilePath: DEFAULT_DOCKERFILE_PATH,
      buildContext: DEFAULT_BUILD_CONTEXT,
      port: DEFAULT_SERVICE_PORT,
      exposePublic: true,
    })
  }

  if (source === "template" && selection.template) {
    const { data } = templateToFormData(selection.template)
    return {
      ...emptyDraftSeed(),
      name: data.name ?? "",
      labels: data.labels ?? [],
      resources: data.spec?.stack_resources ?? [],
      volumes: data.spec?.volumes ?? [],
    }
  }

  if (source === "compose" && selection.compose.yaml.trim()) {
    const parsed = parseAndValidateDockerCompose(selection.compose.yaml)
    const converted = convertDockerComposeToStackData(parsed as DockerComposeFile)
    return {
      ...emptyDraftSeed(),
      name: converted.data?.name ?? "",
      labels: converted.data?.labels ?? [],
      resources: converted.data?.spec?.stack_resources ?? [],
      volumes: converted.data?.spec?.volumes ?? [],
    }
  }

  if (source === "blocks") {
    // `addBlockToStack` owns the de-duplication, so the instances are replayed
    // through it in order rather than the panel's names being trusted.
    let stack = emptyStack()
    for (const instance of selection.blocks.instances) {
      const block = getBlockById(instance.blockId)
      if (block) stack = addBlockToStack(stack, block)
    }
    return {
      ...emptyDraftSeed(),
      resources: stack.spec.stack_resources as unknown as FormStackResourceData[],
      volumes: (stack.spec.volumes ?? []) as unknown as FormVolumeExtendedData[],
      labels: stack.labels ?? [],
      linkedAddonIds: selection.blocks.addonIds,
    }
  }

  return emptyDraftSeed()
}

/** "https://github.com/acme/api.git" → "acme/api" */
function repoTailOf(url: string) {
  return url
    .trim()
    .replace(/\.git$/, "")
    .replace(/\/+$/, "")
    .split("/")
    .slice(-2)
    .join("/")
}
