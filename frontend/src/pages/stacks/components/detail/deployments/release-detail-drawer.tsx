import type { StackRelease } from "@/api/releases";

export interface ReleaseDetailDrawerProps {
  orgId: string;
  teamName: string;
  stackId: string;
  releaseId: string;
  previousRelease?: StackRelease;
  onClose: () => void;
}

// Temporary stub — replaced in the release-detail-drawer task (Task 14).
export function ReleaseDetailDrawer(_props: ReleaseDetailDrawerProps) {
  return null;
}
