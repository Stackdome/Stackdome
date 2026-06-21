import { Rocket } from "lucide-react";
import { AlertBanner } from "@/components/branded";

export function DriftBanner({ onDeploy, busy }: { onDeploy: () => void; busy: boolean }) {
  return (
    <AlertBanner action={{ label: busy ? "Deploying…" : "Deploy changes", onClick: onDeploy, disabled: busy }}>
      <span className="flex items-center gap-2">
        <Rocket className="h-3.5 w-3.5" />
        Unreleased changes — your saved configuration differs from the active deployment (approximate).
      </span>
    </AlertBanner>
  );
}

export function ReleaseErrorBanner({ lead, text, onView }: { lead: string; text: string; onView?: () => void }) {
  return (
    <AlertBanner action={onView ? { label: "View details", onClick: onView } : undefined}>
      <span className="flex items-center gap-2">
        <span className="font-semibold text-foreground">{lead}</span>
        <span className="truncate">{text}</span>
      </span>
    </AlertBanner>
  );
}
