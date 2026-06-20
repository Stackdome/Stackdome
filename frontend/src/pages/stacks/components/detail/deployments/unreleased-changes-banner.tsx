import { Rocket } from "lucide-react";
import { AlertBanner } from "@/components/branded";

export interface UnreleasedChangesBannerProps {
  hasDrift: boolean;
  onDeploy: () => void;
  busy: boolean;
}

export function UnreleasedChangesBanner({ hasDrift, onDeploy, busy }: UnreleasedChangesBannerProps) {
  if (!hasDrift) return null;
  return (
    <AlertBanner action={{ label: busy ? "Deploying…" : "Deploy", onClick: onDeploy }}>
      <span className="flex items-center gap-2">
        <Rocket className="h-3.5 w-3.5" />
        Unreleased changes — your saved configuration differs from the active deployment (approximate).
      </span>
    </AlertBanner>
  );
}
