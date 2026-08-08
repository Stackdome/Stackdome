import * as React from "react";
import { Link } from "react-router-dom";
import { ThemeToggle } from "@/components/theme-toggle";
import { StackdomeWordmark } from "@/components/branded/stackdome-mark";

interface AuthShellProps {
  title: string;
  sub?: React.ReactNode;
  below?: React.ReactNode;
  children: React.ReactNode;
}

// "The room and the plate": the page around the form is the open-sky room
// (.auth-shell) and the form sits in an inset plate (.auth-plate). The header
// is absolute chrome so the column centers on the viewport, not on the
// leftover space under a flow header.
export function AuthShell({ title, sub, below, children }: AuthShellProps) {
  return (
    <div className="auth-shell relative flex min-h-svh flex-col overflow-hidden text-foreground">
      {/* Clouds sit on the horizon; day/night cuts are separate art, swapped by CSS. */}
      <div className="auth-sky" aria-hidden="true">
        <img className="auth-cloud ac-1 dark:hidden" src="/clouds/cloud-big-day.webp" alt="" draggable={false} />
        <img className="auth-cloud ac-2 dark:hidden" src="/clouds/cloud-streak-day.webp" alt="" draggable={false} />
        <img className="auth-cloud ac-1 hidden dark:block" src="/clouds/cloud-big-night.webp" alt="" draggable={false} />
        <img className="auth-cloud ac-2 hidden dark:block" src="/clouds/cloud-streak-night.webp" alt="" draggable={false} />
      </div>

      {/* Same 1280px track as the marketing nav, so brand and toggle don't drift apart on wide screens. */}
      <header className="auth-header absolute inset-x-0 top-0 z-10 mx-auto flex w-full max-w-[1280px] items-center justify-between">
        <Link to="/" aria-label="Stackdome — home" className="transition-opacity hover:opacity-70">
          <StackdomeWordmark size={15} />
        </Link>
        <ThemeToggle variant="ghost" />
      </header>

      <main className="relative z-[1] mx-auto flex w-full max-w-[428px] flex-1 flex-col px-6 py-20">
        <div className="mt-auto text-center">
          <h1 className="auth-title">{title}</h1>
          {sub && <p className="mx-auto mt-2.5 text-sm leading-relaxed text-muted-foreground">{sub}</p>}
        </div>

        {/* Pill controls inside a stadium plate — the plate is rounded like its contents.
            Deliberate exception above the rounded-lg (22px) ceiling: the stadium plate. */}
        <div className="auth-plate mt-8 [&_button]:rounded-full [&_input]:rounded-full [&_input]:pl-4">
          {children}
        </div>

        {below && <div className="mt-6 text-center text-sm text-muted-foreground">{below}</div>}
        <div className="mb-auto" aria-hidden="true" />
      </main>
    </div>
  );
}

export function SwapLink({ lead, to, label }: { lead: string; to: string; label: string }) {
  return (
    <p>
      {lead}{" "}
      <Link to={to} className="font-medium text-foreground transition-opacity hover:opacity-70">
        {label}
      </Link>
    </p>
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
      className="flex items-center justify-between text-[13px] font-medium text-muted-foreground"
    >
      <span>{children}</span>
      {hint && <span className="text-xs text-muted-foreground/70">{hint}</span>}
    </label>
  );
}
