import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Check, ExternalLink, Plus, Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { blockCatalog, BLOCK_CATEGORY_META, getBlockById } from "@/pages/stacks/data/blocks/registry";
import { addBlockToStack, emptyStack } from "@/pages/stacks/lib/block-to-form";
import { ImportSource } from "@/pages/stacks/lib/import-source";
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import { createStack } from "@/api/stacks";
import { convertFormStackToApiStack, FormStackSchema, type FormStackData } from "@/pages/stacks/schemas/form-schema";
import { useResourceTeams } from "@/hooks/use-resource-teams";
import { getCurrentOrganizationId } from "@/helpers/common";
import { useToast } from "@/components/ui/use-toast";
import { BlockPicker } from "./block-picker";
import { BlockGlyph } from "./block-glyph";
import { WizardFooter } from "./wizard-footer";

interface BlockComposerProps {
  onBack: () => void;
  onClose: () => void;
}

export function BlockComposer({ onBack, onClose }: BlockComposerProps) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { defaultTeamName } = useResourceTeams();
  const [query, setQuery] = useState("");
  const [stack, setStack] = useState(emptyStack);
  const [selectedAddonIds, setSelectedAddonIds] = useState<Set<string>>(new Set());
  const [creating, setCreating] = useState(false);
  const { addons } = usePostgresAddons();

  // addedIds = block ids whose resource name (or a -N variant) is present
  const addedIds = useMemo(() => {
    const names = new Set(stack.spec.stack_resources.map((r) => r.name));
    return blockCatalog.filter((b) => names.has(b.id) || [...names].some((n) => n.startsWith(`${b.id}-`))).map((b) => b.id);
  }, [stack]);

  const addBlock = (id: string) => {
    const block = getBlockById(id);
    if (block) setStack((s) => addBlockToStack(s, block));
  };

  // Match a resource back to its block so the stack-so-far list shows the right glyph.
  const iconForResource = (name: string) =>
    blockCatalog.find((b) => b.id === name || name.startsWith(`${b.id}-`))?.icon ?? "box";

  const removeResource = (index: number) =>
    setStack((s) => ({
      ...s,
      spec: { ...s.spec, stack_resources: s.spec.stack_resources.filter((_, i) => i !== index) },
    }));

  const toggleAddon = (id: string) =>
    setSelectedAddonIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  // Managed addons already provisioned in the workspace, filtered by the search box.
  const q = query.trim().toLowerCase();
  const availableAddons = addons.filter((a) => a.id && (!q || a.name.toLowerCase().includes(q)));

  // Hand off to the full create form (naming, configuring, addon binding) when
  // the composed stack can't be created directly.
  const openFormEditor = () => {
    navigate("/stacks/create", {
      state: {
        importedData: stack,
        importSource: ImportSource.Blocks,
        linkedAddonIds: Array.from(selectedAddonIds),
      },
    });
    onClose();
  };

  // Prefer canvas-first: if the composed stack is named + fully configured,
  // create it now and drop the user straight into the canvas editor (carrying
  // any selected addons so they can be bound there). Otherwise fall back to the
  // create form, which also lands on the canvas after save.
  const createAndOpen = async () => {
    const orgId = getCurrentOrganizationId();
    const formData: FormStackData = {
      name: stack.name.trim(),
      labels: stack.labels,
      spec: { stack_resources: stack.spec.stack_resources, volumes: stack.spec.volumes },
    } as FormStackData;
    const parsed = FormStackSchema.safeParse(formData);
    if (!parsed.success || !orgId || !defaultTeamName) {
      openFormEditor();
      return;
    }
    setCreating(true);
    try {
      const created = await createStack(orgId, defaultTeamName, convertFormStackToApiStack(parsed.data));
      onClose();
      navigate(`/stacks/${created.id}`, { state: { linkedAddonIds: Array.from(selectedAddonIds) } });
    } catch (error) {
      setCreating(false);
      toast({
        title: "Couldn't create stack",
        description: error instanceof Error ? error.message : "Review the details and try again.",
        variant: "destructive",
      });
      openFormEditor();
    }
  };

  const count = stack.spec.stack_resources.length + selectedAddonIds.size;

  return (
    <div className="flex h-full flex-col">
      <div className="grid min-h-0 flex-1 grid-cols-[1fr_360px] gap-0 overflow-hidden">
        {/* LEFT: palette */}
        <div className="scrollbar-hide overflow-y-auto p-6">
          <div className="mb-1 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">COMPOSE</div>
          <h2 className="mb-1 text-2xl font-medium tracking-tight">What's in your stack?</h2>
          <p className="mb-5 text-sm text-muted-foreground">
            Add blocks below. Known software lands fully configured.
          </p>
          <div className="relative mb-5">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search services, data stores, addons…"
              className="pl-9"
            />
          </div>
          <BlockPicker catalog={blockCatalog} categories={BLOCK_CATEGORY_META} addedIds={addedIds} onAdd={addBlock} query={query} />

          {(availableAddons.length > 0 || (addons.length === 0 && !q)) && (
            <div className="mt-6">
              <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
                MANAGED ADD-ONS
              </div>
              {availableAddons.length > 0 ? (
                <div className="grid grid-cols-2 gap-2.5">
                  {availableAddons.map((a) => {
                    const added = selectedAddonIds.has(a.id!);
                    return (
                      <button
                        type="button"
                        key={a.id}
                        onClick={() => toggleAddon(a.id!)}
                        className={cn(
                          "flex min-h-[60px] items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary",
                          added && "border-primary/60",
                        )}
                      >
                        <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                          <AddonTypeIcon type="postgres" size={18} />
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="block text-sm font-medium text-foreground">{a.name}</span>
                          <span className="block truncate font-mono text-[11px] text-muted-foreground">managed postgres</span>
                        </span>
                        {added ? <Check className="h-[17px] w-[17px] text-success" /> : <Plus className="h-[17px] w-[17px] text-primary" />}
                      </button>
                    );
                  })}
                </div>
              ) : (
                // No managed addons in the workspace yet — offer a way to create one.
                <a
                  href="/addons"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex min-h-[60px] items-center gap-3 rounded-md border border-dashed bg-card/40 px-3 py-3 text-left transition-colors hover:border-primary hover:bg-card"
                >
                  <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                    <Plus className="h-[18px] w-[18px]" />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-sm font-medium text-foreground">Create an add-on</span>
                    <span className="block text-[12px] text-muted-foreground">
                      No managed add-ons yet. Set one up in the Addons page.
                    </span>
                  </span>
                  <ExternalLink className="h-4 w-4 flex-none text-muted-foreground" />
                </a>
              )}
            </div>
          )}
        </div>

        {/* RIGHT: your stack so far */}
        <div className="flex min-h-0 flex-col border-l bg-card/40">
          <div className="border-b px-4 py-3">
            <label htmlFor="wizard-stack-name" className="mb-1.5 block font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
              Stack name
            </label>
            <Input
              id="wizard-stack-name"
              value={stack.name}
              onChange={(e) => setStack((s) => ({ ...s, name: e.target.value }))}
              placeholder="e.g. my-app"
            />
          </div>
          <div className="border-b px-4 py-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Your stack so far · {count}
          </div>
          <div data-testid="stack-so-far" className="scrollbar-hide flex-1 space-y-1.5 overflow-y-auto p-4">
            {count === 0 ? (
              <p className="px-1 py-6 text-sm text-muted-foreground">Pick blocks on the left to start.</p>
            ) : (
              <>
                {stack.spec.stack_resources.map((r, i) => (
                  <div key={`${r.name}-${i}`} className="flex items-center gap-3 rounded border bg-card px-3 py-2">
                    <span className="h-2 w-2 flex-none rounded-full bg-success" />
                    <BlockGlyph icon={iconForResource(r.name)} size={16} />
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm text-foreground">{r.name}</span>
                      <span className="block truncate font-mono text-[11px] text-muted-foreground">
                        {r.image_spec?.image || "configure source"}
                      </span>
                    </span>
                    <button type="button" aria-label={`Remove ${r.name}`} onClick={() => removeResource(i)} className="text-muted-foreground hover:text-foreground">
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                ))}
                {Array.from(selectedAddonIds).map((id) => {
                  const addon = addons.find((a) => a.id === id);
                  const name = addon?.name ?? id;
                  return (
                    <div key={id} className="flex items-center gap-3 rounded border bg-card px-3 py-2">
                      <span className="h-2 w-2 flex-none rounded-full bg-success" />
                      <AddonTypeIcon type="postgres" size={16} />
                      <span className="min-w-0 flex-1">
                        <span className="block text-sm text-foreground">{name}</span>
                        <span className="block truncate font-mono text-[11px] text-muted-foreground">managed postgres · addon</span>
                      </span>
                      <button type="button" aria-label={`Remove ${name}`} onClick={() => toggleAddon(id)} className="text-muted-foreground hover:text-foreground">
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                  );
                })}
              </>
            )}
          </div>
        </div>
      </div>
      <WizardFooter
        onBack={onBack}
        onContinue={createAndOpen}
        continueDisabled={count === 0 || creating}
        hint={creating ? "Creating your stack…" : "Open in the canvas editor"}
      />
    </div>
  );
}
