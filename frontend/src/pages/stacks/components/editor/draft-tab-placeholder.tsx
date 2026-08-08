import { Rocket } from "lucide-react";

/** Shown for Deployments/Logs/Metrics while the stack is an unsaved draft. */
export function DraftTabPlaceholder({ label }: { label: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 py-24 text-center">
      <Rocket className="size-5 text-muted-foreground" aria-hidden />
      <p className="text-body font-medium text-foreground">{label} available after you deploy</p>
      <p className="text-body text-muted-foreground">Save this stack to create it, then deploy to see live data.</p>
    </div>
  );
}
