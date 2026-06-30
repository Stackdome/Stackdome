// frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx
import {
  Grid3x3,
  LayoutTemplate,
  GitBranch,
  Container,
  Code,
  ArrowRight,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface WizardChooserProps {
  onPickBlocks: () => void;
  onPickTemplate: () => void;
  onPickCompose: () => void;
  onPickBlank: () => void;
}

interface AltStart {
  icon: LucideIcon;
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
      icon: Container,
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
              <a.icon className="h-[18px] w-[18px]" />
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
