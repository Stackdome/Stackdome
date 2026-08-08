import { Search } from "lucide-react";
import { cn } from "@/lib/utils";

interface EmptyStateProps {
  /** A glyph or illustration. See `SearchGlyph` / `StackArchitectureGlyph`. */
  icon?: React.ReactNode;
  title: React.ReactNode;
  description?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
}

/**
 * A page or list with nothing in it (§11).
 *
 * **There is no box.** It used to draw a dashed bordered panel, which
 * re-introduced exactly the frame §11 removes from lists — the state sits ON
 * the sheet, like the rows it replaces.
 *
 * The type is deliberately quiet: `text-body` medium over `text-meta` muted.
 * An empty state is not a headline, and a page that says nothing does not get
 * to shout about it. Settled on the Shape + Hierarchy board (nodes `199:2386`,
 * `199:2399`).
 */
export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-5 px-10 py-18 text-center",
        className,
      )}
    >
      {icon && <div className="text-fg-2">{icon}</div>}
      {/* 4px between the two lines — they are one statement, not two. The
          description is capped at 292px (the board's measure): unbounded, it
          ran the full 1000px of the sheet as a single line, which reads as a
          system message rather than as copy. */}
      <div className="flex flex-col items-center gap-1">
        <div className="text-body font-medium text-foreground">{title}</div>
        {description && (
          <p className="text-meta max-w-[292px] text-balance text-fg-muted">{description}</p>
        )}
      </div>
      {action}
    </div>
  );
}

/**
 * The **no-results** glyph — a lens on the sheet, and nothing else.
 *
 * A filter that matched nothing is a small, recoverable moment, so it gets a
 * small mark. The decoration is spent on `StackArchitectureGlyph` instead,
 * because that one is the first screen a new user ever sees.
 */
export function SearchGlyph() {
  return (
    <div
      aria-hidden
      className="flex size-[34px] items-center justify-center rounded-full border border-border"
    >
      <Search className="size-4 text-fg-2" strokeWidth={1.5} />
    </div>
  );
}

/**
 * The **first-run** glyph — one stack on the dot-grid canvas, with its
 * connections sweeping off the edges.
 *
 * Taken from the marketing site's architecture view rather than invented: it is
 * the one image this product owns, and it shows a stack instead of describing
 * one. An earlier version used translucent shapes and read as a smudge —
 * **nothing here is transparent**, because transparency was doing the work that
 * shape should do.
 *
 * **One card, not two.** Two cards and a wire between them drew a closed
 * system. One card with wires running off the canvas says the stack connects to
 * more than fits — which is the truer picture and the calmer image.
 *
 * Geometry is the board's (node `201:1177`), authored at **148×88** and drawn
 * at `SCALE`. Every number below is the board's own, so the two can still be
 * compared line by line — resizing is one constant, not a re-measure.
 */
const BOARD_W = 148;
const BOARD_H = 88;
/** Drawn at 200px wide. Bumped from 1× — at 148 the card read as a favicon. */
const SCALE = 200 / BOARD_W;
/** Board units → rendered px. */
const u = (n: number) => Math.round(n * SCALE * 100) / 100;

export function StackArchitectureGlyph() {
  return (
    <div
      aria-hidden
      className="relative overflow-hidden"
      style={{ width: u(BOARD_W), height: u(BOARD_H) }}
    >
      {/* The canvas — 1.5px dots on a 16px pitch. A dot needs far more alpha
          than a line to read at all (§4), so these run at the bold rung. */}
      <div
        className="absolute inset-0"
        style={{
          backgroundImage: `radial-gradient(circle at ${u(0.75)}px ${u(0.75)}px, var(--grid-bold) ${u(0.75)}px, transparent ${u(0.75)}px)`,
          backgroundSize: `${u(16)}px ${u(16)}px`,
          backgroundPosition: `${u(4)}px ${u(4)}px`,
        }}
      />

      {/* The wires. Curves, so they need real paths — and each fades out along
          its length, which is what stops them reading as two stray hairs.
          The board fades brand → white; this fades brand → transparent, because
          white is only invisible in the light theme.

          The viewBox stays in BOARD units, so the paths are the board's numbers
          verbatim and the stroke scales with the drawing. */}
      <svg
        className="absolute inset-0"
        width={u(BOARD_W)}
        height={u(BOARD_H)}
        viewBox={`0 0 ${BOARD_W} ${BOARD_H}`}
        fill="none"
        aria-hidden
      >
        <defs>
          <linearGradient id="sd-wire-l" x1="43.3" y1="44" x2="15" y2="60.5" gradientUnits="userSpaceOnUse">
            <stop stopColor="var(--brand)" />
            <stop offset="1" stopColor="var(--brand)" stopOpacity="0" />
          </linearGradient>
          <linearGradient id="sd-wire-r" x1="104.7" y1="44" x2="133" y2="60.5" gradientUnits="userSpaceOnUse">
            <stop stopColor="var(--brand)" />
            <stop offset="1" stopColor="var(--brand)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d="M15 60.5C31 60.5 31.8 44 43.3 44" stroke="url(#sd-wire-l)" />
        <path d="M133 60.5C117 60.5 116.2 44 104.7 44" stroke="url(#sd-wire-r)" />
      </svg>

      {/* The stack itself: a name, a detail line, three service chips, and the
          status said ONCE as a dot. The card's own anatomy in miniature.
          The border stays 1px — a hairline is a hairline at any drawing size,
          and scaling it to 1.35 would put it off the §4 ladder. */}
      <div
        className="absolute rounded-xl border border-border bg-card shadow-md"
        style={{ left: u(42), top: u(24), width: u(64), height: u(40) }}
      >
        <Bar left={6} top={8} width={20} />
        <Bar left={6} top={16} width={30} />
        <Bar left={6} top={31} width={13} />
        <Bar left={23} top={31} width={10} />
        <Bar left={37} top={31} width={13} />
        <div
          className="absolute rounded-full bg-brand"
          style={{ left: u(54), top: u(6), width: u(5), height: u(5) }}
        />
      </div>
    </div>
  );
}

/** A single ghosted line inside the glyph's card, in board units. */
function Bar({ left, top, width }: { left: number; top: number; width: number }) {
  return (
    <div
      className="absolute rounded-full"
      style={{
        left: u(left),
        top: u(top),
        width: u(width),
        height: u(4),
        background: "var(--wash-selected-hover)",
      }}
    />
  );
}
