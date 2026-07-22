import { cloneElement } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface LimitedActionProps {
  /** When true, the child action is disabled and the limit message is shown on hover. */
  limitReached: boolean;
  limitMessage: string;
  children: React.ReactElement<{ disabled?: boolean }>;
}

export function LimitedAction({ limitReached, limitMessage, children }: LimitedActionProps) {
  if (!limitReached) return children;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        {/* Disabled buttons swallow pointer events, so the tooltip anchors to a focusable wrapper. */}
        <span tabIndex={0}>{cloneElement(children, { disabled: true })}</span>
      </TooltipTrigger>
      <TooltipContent>{limitMessage}</TooltipContent>
    </Tooltip>
  );
}
