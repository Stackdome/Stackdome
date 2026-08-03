import { ArrowLeft, ArrowRight, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";

interface WizardFooterProps {
  onBack: () => void;
  onContinue: () => void;
  /** Primary action label. Defaults to "Continue". */
  continueLabel?: string;
  continueDisabled?: boolean;
  /** Shows a spinner and disables both buttons while an action is in flight. */
  loading?: boolean;
  /** Optional muted context shown left of the Continue button (hidden on small widths). */
  hint?: string;
}

/**
 * Consistent wizard navigation footer: Back on the left, the primary
 * "Continue" action on the right. Used by every path panel so navigation is
 * uniform across the new-stack wizard.
 */
export function WizardFooter({
  onBack,
  onContinue,
  continueLabel = "Continue",
  continueDisabled = false,
  loading = false,
  hint,
}: WizardFooterProps) {
  return (
    <div className="flex items-center justify-between gap-4 border-t px-5 py-3.5">
      <Button type="button" variant="ghost" onClick={onBack} disabled={loading}>
        <ArrowLeft className="h-4 w-4" />
        Back
      </Button>
      <div className="flex items-center gap-3">
        {hint && (
          <span className="hidden text-xs text-muted-foreground sm:inline">{hint}</span>
        )}
        <Button type="button" onClick={onContinue} disabled={continueDisabled || loading}>
          {loading && <Loader2 className="h-4 w-4 animate-spin" />}
          {continueLabel}
          {!loading && <ArrowRight className="h-4 w-4" />}
        </Button>
      </div>
    </div>
  );
}
