import * as React from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { EyebrowLabel } from "@/components/branded/eyebrow-label";
import { StackdomeMark, StackdomeWordmark } from "@/components/branded/stackdome-mark";
import { cn } from "@/lib/utils";

interface MetaCell {
  label: string;
  value: React.ReactNode;
  tone?: "default" | "brand" | "success";
}

interface AuthShellProps {
  headlineSolid: string;
  headlineStroke?: string;
  tagline?: React.ReactNode;
  sub?: string;
  stageStatus?: string;
  meta?: MetaCell[];
  checklist?: React.ReactNode[];
  children: React.ReactNode;
}

export function AuthShell({
  headlineSolid,
  headlineStroke,
  tagline,
  sub,
  stageStatus,
  meta,
  checklist,
  children,
}: AuthShellProps) {
  return (
    <div className="relative min-h-svh bg-background text-foreground overflow-hidden">
      {/* Page-wide grid backdrop, masked to fade at edges (matches reference .grid-bg) */}
      <div
        aria-hidden="true"
        className="pointer-events-none fixed inset-0 z-0 opacity-[0.10] text-foreground [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:64px_64px] [mask-image:radial-gradient(ellipse_at_center,black_30%,transparent_80%)] [-webkit-mask-image:radial-gradient(ellipse_at_center,black_30%,transparent_80%)]"
      />

      <div className="absolute right-4 top-4 z-30 md:right-6 md:top-6">
        <ThemeToggle />
      </div>

      {/* Centered wrap — empty bg shows on either side past 1240px */}
      <div className="relative z-10 mx-auto grid min-h-svh max-w-[1240px] lg:grid-cols-[1fr_440px]">
        {/* Brand panel — transparent so the grid shows through */}
        <aside className="relative hidden flex-col justify-between p-10 lg:flex xl:p-14">
          <div className="relative z-10">
            <StackdomeWordmark size={20} />
          </div>

          <div className="relative z-10 flex flex-col gap-8">
            <h1 className="font-semibold text-5xl leading-[0.95] tracking-tight xl:text-6xl">
              {headlineSolid}
              {headlineStroke && (
                <>
                  <br />
                  <span
                    className="text-brand"
                    style={{
                      WebkitTextStroke: "1.5px currentColor",
                      WebkitTextFillColor: "transparent",
                    }}
                  >
                    {headlineStroke}
                  </span>
                </>
              )}
            </h1>

            {tagline && (
              <p className="max-w-md text-2xl font-medium leading-snug tracking-tight text-foreground/90 xl:text-3xl">
                {tagline}
              </p>
            )}

            {sub && (
              <p className="max-w-md text-sm leading-relaxed text-muted-foreground">{sub}</p>
            )}

            {stageStatus && (
              <div className="relative max-w-md rounded-md border border-border bg-background/50 p-6">
                <div className="absolute left-4 top-3 font-mono text-[10px] uppercase tracking-[1.5px] text-muted-foreground">
                  stack.render()
                </div>
                <div className="absolute right-4 top-3 flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-[1.5px] text-brand">
                  <span className="inline-block h-1.5 w-1.5 rounded-full bg-brand animate-pulse" />
                  {stageStatus}
                </div>
                <div className="flex items-center justify-center py-6">
                  <StackdomeMark size={96} />
                </div>
              </div>
            )}
          </div>

          <div className="relative z-10">
            {meta && (
              <div className="grid grid-cols-3 gap-px overflow-hidden rounded-md border border-border bg-border">
                {meta.map((m) => (
                  <div key={m.label} className="bg-card px-3 py-3">
                    <div className="font-mono text-[10px] uppercase tracking-[1.5px] text-muted-foreground">
                      {m.label}
                    </div>
                    <div
                      className={cn(
                        "mt-1 font-mono text-[13px]",
                        m.tone === "brand" && "text-brand",
                        m.tone === "success" && "text-[#22c55e]",
                      )}
                    >
                      {m.value}
                    </div>
                  </div>
                ))}
              </div>
            )}
            {checklist && (
              <ul className="space-y-2">
                {checklist.map((item) => (
                  <li
                    key={item}
                    className="flex items-center gap-2 font-mono text-[12px] text-muted-foreground"
                  >
                    <span className="flex h-4 w-4 items-center justify-center rounded-sm bg-brand-bg text-brand">
                      ✓
                    </span>
                    {item}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </aside>

        {/* Form band — solid darker bg covers the grid behind it */}
        <main className="relative flex items-center justify-center bg-secondary px-6 py-10 sm:px-10 lg:border-x lg:border-border">
          <div className="w-full max-w-[360px]">
            {/* Mobile-only brand mark */}
            <div className="mb-8 lg:hidden">
              <StackdomeWordmark size={20} />
            </div>
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}

export function FormHead({
  step,
  title,
  trailing,
}: {
  step: string;
  title: string;
  trailing?: React.ReactNode;
}) {
  return (
    <div className="mb-6 space-y-1.5">
      <EyebrowLabel tone="brand" className="font-bold">{step}</EyebrowLabel>
      <h2 className="text-[28px] font-semibold leading-tight tracking-tight">{title}</h2>
      {trailing && <p className="text-sm text-muted-foreground">{trailing}</p>}
    </div>
  );
}

export function FieldLabel({
  htmlFor,
  children,
  hint,
}: {
  htmlFor: string;
  children: React.ReactNode;
  hint?: React.ReactNode;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className="flex items-center justify-between font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground"
    >
      <span className="flex items-center gap-1.5">
        <span className="text-brand">→</span>
        {children}
      </span>
      {hint && <span className="text-muted-foreground/70 normal-case tracking-normal">{hint}</span>}
    </label>
  );
}

export function FootRow({
  left,
  right,
}: {
  left?: React.ReactNode;
  right?: React.ReactNode;
}) {
  return (
    <div className="mt-6 flex items-center justify-between font-mono text-[10px] uppercase tracking-[1.5px] text-muted-foreground">
      <span>{left}</span>
      <span>{right}</span>
    </div>
  );
}
