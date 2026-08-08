import { HardDrive } from "lucide-react";
import { NodeGlyph } from "@/pages/stacks/components/editor/tabs/architecture/nodes/node-glyph";
import { nodePresentation, type GlyphKind } from "@/pages/stacks/lib/canvas/node-presentation";
import type { Stack } from "@/api/stack-types";

export interface StackComponent {
  name: string;
  /** Absent for a volume, which has no image to infer a kind from. */
  glyph?: GlyphKind;
  brandSlug?: string;
  isVolume?: boolean;
}

/**
 * Everything the stack is made of, **by name and by kind** — services first,
 * then volumes.
 *
 * The icon is not decoration: it reports what sort of thing each component is,
 * which is the only reason it is allowed on screen (§7). It reuses the canvas
 * editor's own `NodeGlyph`, so a Postgres component shows the same elephant on
 * the list as it does on the canvas — one visual language, learned on the first
 * screen of the session instead of three clicks in. `HardDrive` for volumes is
 * the same glyph the canvas's attachment node uses.
 *
 * The kind is inferred from the image reference and ports, exactly as the canvas
 * infers it. There is no `kind` field on the API; both call sites share one
 * heuristic rather than growing a second.
 */
export function stackComponents(stack: Stack): StackComponent[] {
  const services = (stack.spec?.stack_resources ?? [])
    .filter((r) => r.name)
    .map((r) => {
      const p = nodePresentation({
        isAddon: false,
        image: r.source?.image?.ref,
        hasBuild: !!r.source?.git,
        ports: r.ports?.map((port) => ({
          number: port.number,
          protocol: port.protocol,
          exposedToPublic: port.exposed_to_public,
        })),
      });
      return { name: r.name, glyph: p.glyph, brandSlug: p.brandSlug };
    });

  const volumes = (stack.spec?.volumes ?? [])
    .filter((v) => v.name)
    .map((v) => ({ name: v.name as string, isVolume: true }));

  return [...services, ...volumes];
}

/**
 * The card is a FIXED height, so the chip row cannot wrap — it has to fit on one
 * line and say how many it dropped.
 *
 * Width is estimated rather than measured because the names are **mono**, where
 * every character is the same 7.2px at `mono/meta`. That makes the estimate
 * exact for the text and the only approximation is the icon, which is a constant.
 * Budgeted against the NARROWEST card (the three-up column at 1440), so a wider
 * grid never overflows — it only leaves a little room unused.
 */
const CHIP_ROW = 344; // 376 card − 16px padding each side
const CHIP_GAP = 5;
const MONO_ADVANCE = 7.2; // JetBrains Mono at 12px
const CHIP_CHROME = 32; // icon 12 + gap 4 + padding 14 + border 2
const MORE_CHIP = 44; // room for "+N"

const chipWidth = (name: string) => name.length * MONO_ADVANCE + CHIP_CHROME;

export function fitChips(parts: StackComponent[]): { shown: StackComponent[]; hidden: number } {
  const total = parts.reduce((w, p, i) => w + chipWidth(p.name) + (i ? CHIP_GAP : 0), 0);
  if (total <= CHIP_ROW) return { shown: parts, hidden: 0 };

  // They do not all fit, so the last slot has to pay for the "+N" that says so.
  const shown: StackComponent[] = [];
  let used = 0;
  for (const p of parts) {
    const next = used + (shown.length ? CHIP_GAP : 0) + chipWidth(p.name);
    if (next > CHIP_ROW - MORE_CHIP - CHIP_GAP) break;
    used = next;
    shown.push(p);
  }
  return { shown, hidden: parts.length - shown.length };
}

/** One component, named and typed. */
export function ComponentChip({ part }: { part: StackComponent }) {
  return (
    <span className="inline-flex h-5 shrink-0 items-center gap-1 rounded-sm border border-border px-[7px] font-mono text-meta text-fg-2">
      {part.isVolume ? (
        <HardDrive className="size-3 shrink-0 text-fg-muted" aria-hidden />
      ) : (
        <NodeGlyph
          glyph={part.glyph ?? "service"}
          brandSlug={part.brandSlug}
          size={12}
          className="size-3 shrink-0 text-fg-muted"
        />
      )}
      {part.name}
    </span>
  );
}
