export interface DeployFailedBannerProps {
  message: string;
}

/**
 * The red "Deploy failed" box shown in a failed node's detail card. Shared by
 * the live and historical bodies so failure reads the same everywhere (and
 * replaces the old plain "WHY IT FAILED" marker on historical nodes).
 */
export function DeployFailedBanner({ message }: DeployFailedBannerProps) {
  return (
    <div className="mt-4 rounded-md border border-danger-border bg-danger-bg p-3.5">
      <div className="mb-1.5 flex items-center gap-2 font-sans text-[13px] font-semibold text-danger">
        <span>⊘</span> Deploy failed
      </div>
      <div className="font-mono text-[11.5px] leading-relaxed text-foreground">{message}</div>
    </div>
  );
}
