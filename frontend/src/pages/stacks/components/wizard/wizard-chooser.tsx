// frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx
import type { ComponentType } from "react";
import {
  Grid3x3,
  LayoutTemplate,
  GitBranch,
  Code,
  ArrowRight,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Docker whale as a monochrome glyph (fill=currentColor) so it matches the
 * muted-foreground color of the lucide tiles around it.
 */
function DockerGlyph({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 512 512" fill="currentColor" aria-hidden className={className}>
      <path d="M501.4 212.3c-11.5-8-38-11-58.6-7-2.4-20-13.5-37.5-32.7-53l-11-8-7.7 11.5c-9.6 15-14.4 36-13 56 .5 7 2.9 19.5 10.1 30.5-6.7 4-20.7 9-38.9 9H2.3l-1 4c-3.4 20-3.4 82.5 36 130.5 29.8 36.5 74 55 132.1 55 125.9 0 219.1-60.5 262.8-170 17.3.5 54.3 0 73-37.5.5-1 1.4-3 4.8-10.5l1.9-4zM280 71.3h-52.8v50H280zm0 60h-52.8v50H280zm-62.5 0h-52.8v50h52.8zm-62.4 0h-52.8v50h52.8zm-62.5 60H39.8v50h52.8zm62.5 0h-52.8v50h52.8zm62.4 0h-52.8v50h52.8zm62.5 0h-52.8v50H280zm62.4 0h-52.8v50h52.8z" />
    </svg>
  );
}

interface WizardChooserProps {
  onPickBlocks: () => void;
  onPickTemplate: () => void;
  onPickCompose: () => void;
  onPickBlank: () => void;
}

interface AltStart {
  /** Lucide icon, or a custom glyph component (both render at muted-foreground). */
  icon?: LucideIcon;
  glyph?: ComponentType<{ className?: string }>;
  label: string;
  desc: string;
  onClick?: () => void;
  disabled?: boolean;
  soon?: boolean;
}

export function WizardChooser({
  onPickBlocks,
  onPickTemplate,
  onPickCompose,
  onPickBlank,
}: WizardChooserProps) {
  const alts: AltStart[] = [
    {
      icon: LayoutTemplate,
      label: "From template",
      desc: "Curated self-hosted apps, ready to deploy.",
      onClick: onPickTemplate,
    },
    {
      icon: GitBranch,
      label: "GitHub repo",
      desc: "Auto-detect build & start.",
      disabled: true,
      soon: true,
    },
    {
      glyph: DockerGlyph,
      label: "Docker compose",
      desc: "Import a compose.yml.",
      onClick: onPickCompose,
    },
    {
      icon: Code,
      label: "Blank slate",
      desc: "Build it up yourself.",
      onClick: onPickBlank,
    },
  ];

  return (
    <div className="p-8">
      <h2 className="mb-1 text-2xl font-medium tracking-tight">
        How do you want to start?
      </h2>
      <p className="mb-6 text-sm text-muted-foreground">
        Let&apos;s get something running. Pick a starting point. You can change
        anything later.
      </p>

      {/* Primary tile */}
      <button
        type="button"
        onClick={onPickBlocks}
        className="mb-6 flex w-full items-center gap-4 rounded-lg border border-primary bg-card p-5 text-left transition-colors hover:bg-card/80"
      >
        <span className="flex h-11 w-11 flex-none items-center justify-center rounded bg-primary/10 text-primary">
          <Grid3x3 className="h-5 w-5" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="mb-0.5 flex items-center gap-2">
            <span className="text-base font-medium text-foreground">
              Build from blocks
            </span>
            <span className="rounded-full bg-primary/10 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-primary">
              Recommended
            </span>
          </span>
          <span className="block text-sm text-muted-foreground">
            Assemble from recognizable building blocks like web, Postgres, Redis,
            and workers. Known software lands fully configured.
          </span>
        </span>
        {/*
          Rendered as a <span> (not <Button>) to avoid a nested <button> inside
          this <button> tile, which is invalid HTML. The outer button already
          has the accessible name "Compose blocks" via its text content, so the
          test's getByRole("button", { name: /Compose blocks/i }) still matches.
        */}
        <span className="inline-flex flex-none cursor-default items-center gap-1.5 rounded-sm bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">
          Compose blocks <ArrowRight className="h-4 w-4" />
        </span>
      </button>

      <div className="mb-4 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
        OR START FROM
      </div>

      <div className="grid grid-cols-2 gap-2.5">
        {alts.map((a) => (
          <button
            type="button"
            key={a.label}
            onClick={a.onClick}
            disabled={a.disabled}
            className={cn(
              "flex min-h-[76px] items-start gap-3 rounded-md border bg-card p-4 text-left transition-colors",
              a.disabled
                ? "cursor-not-allowed opacity-50"
                : "hover:border-primary",
            )}
          >
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded bg-muted text-muted-foreground">
              {a.glyph ? (
                <a.glyph className="h-[18px] w-[18px]" />
              ) : (
                a.icon && <a.icon className="h-[18px] w-[18px]" />
              )}
            </span>
            <span className="min-w-0 flex-1">
              <span className="mb-0.5 flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">
                  {a.label}
                </span>
                {a.soon && (
                  <span className="rounded-full border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wider text-muted-foreground">
                    soon
                  </span>
                )}
              </span>
              <span className="block text-xs text-muted-foreground">
                {a.desc}
              </span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
