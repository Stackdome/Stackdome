import * as React from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { EyebrowLabel, StackdomeMark } from "@/components/branded";
import { cn } from "@/lib/utils";

interface MetaCell {
  label: string;
  value: React.ReactNode;
  tone?: "default" | "brand" | "success";
}

interface AuthShellProps {
  marker: { code: string; expr: string };
  headlineSolid: string;
  headlineStroke: string;
  sub: string;
  stageStatus: string;
  meta?: MetaCell[];
  checklist?: string[];
  children: React.ReactNode;
}

export function AuthShell({
  marker,
  headlineSolid,
  headlineStroke,
  sub,
  stageStatus,
  meta,
  checklist,
  children,
}: AuthShellProps) {
  return (
    <div className="relative min-h-svh bg-background text-foreground">
      <div className="absolute right-4 top-4 z-20 md:right-6 md:top-6">
        <ThemeToggle />
      </div>

      <div className="grid min-h-svh lg:grid-cols-2">
        {/* Brand panel */}
        <aside className="relative hidden flex-col justify-between border-r border-border bg-card p-10 lg:flex xl:p-14">
          <div className="absolute inset-0 pointer-events-none opacity-[0.06] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:32px_32px] text-foreground" />

          <div className="relative z-10 flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-sm bg-brand-bg border border-brand-border">
              <StackdomeMark size={18} variant="tinted" />
            </span>
            <span className="font-mono text-[12px] uppercase tracking-[1.5px] text-foreground">stackdome</span>
          </div>

          <div className="relative z-10 flex flex-col gap-8">
            <div className="flex items-center gap-3">
              <EyebrowLabel className="rounded-sm border border-brand-border bg-brand-bg px-2 py-1 text-brand">
                {marker.code}
              </EyebrowLabel>
              <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
                {marker.expr}
              </span>
            </div>

            <h1 className="font-semibold text-5xl leading-[0.95] tracking-tight xl:text-6xl">
              {headlineSolid}
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
            </h1>

            <p className="max-w-md text-sm leading-relaxed text-muted-foreground">{sub}</p>

            {/* Stage card */}
            <div className="relative max-w-md rounded-md border border-border bg-background/50 p-6">
              <div className="absolute left-4 top-3 font-mono text-[10px] uppercase tracking-[1.5px] text-muted-foreground">
                stack.render()
              </div>
              <div className="absolute right-4 top-3 flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-[1.5px] text-brand">
                <span className="inline-block h-1.5 w-1.5 rounded-full bg-brand animate-pulse" />
                {stageStatus}
              </div>
              <div className="flex items-center justify-center py-6">
                <StackdomeMark size={88} variant="tinted" />
              </div>
            </div>
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

        {/* Form panel */}
        <main className="flex items-center justify-center p-6 sm:p-10">
          <div className="w-full max-w-[400px]">
            {/* Mobile-only brand mark */}
            <div className="mb-8 flex items-center gap-2 lg:hidden">
              <span className="flex h-7 w-7 items-center justify-center rounded-sm bg-brand-bg border border-brand-border">
                <StackdomeMark size={18} variant="tinted" />
              </span>
              <span className="font-mono text-[12px] uppercase tracking-[1.5px]">stackdome</span>
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
    <div className="mb-6 space-y-2">
      <EyebrowLabel tone="muted">{step}</EyebrowLabel>
      <h2 className="text-3xl font-semibold tracking-tight">{title}</h2>
      {trailing && <p className="text-sm text-muted-foreground">{trailing}</p>}
    </div>
  );
}

export function FieldLabel({
  htmlFor,
  number,
  children,
  hint,
}: {
  htmlFor: string;
  number: string;
  children: React.ReactNode;
  hint?: React.ReactNode;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className="flex items-center justify-between font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground"
    >
      <span className="flex items-center gap-1.5">
        <span className="text-brand">{number}</span>
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
