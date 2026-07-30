import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { LazyLog } from "react-lazylog";
import { Loader2, Hourglass, WifiOff, RefreshCw } from "lucide-react";
import { StatusPill } from "@/components/branded";
import { statusVariant } from "@/components/branded/status-variant";
import type { SseStreamStatus } from "@/api/observability";
import { useBuildLogStream, type BuildLogPhase } from "./use-build-log-stream";

/** cluster-agent api/builds/v1alpha1/imagebuild_types.go:41 — flows through as a raw string. */
export const BuildPhase = {
  Pending: "Pending",
  Success: "Success",
  Failed: "Failed",
  Cancelled: "Cancelled",
} as const;

const TERMINAL_BUILD_STATES: ReadonlySet<string> = new Set<string>([
  BuildPhase.Success,
  BuildPhase.Failed,
  BuildPhase.Cancelled,
]);

const REVISION_LENGTH = 7;

const LOG_STYLE = {
  backgroundColor: "#070a0f",
  color: "#94a3b8",
  fontSize: "12px",
  lineHeight: "1.7",
  fontFamily: "var(--font-mono), ui-monospace, SFMono-Regular, Menlo, monospace",
} as const;

export type BuildLogsView =
  | "expired"
  | "waiting"
  | "endedWithoutLogs"
  | "interrupted"
  | "log"
  | "noOutput"
  | "connecting";

interface ViewInput {
  phase: BuildLogPhase;
  connectionStatus: SseStreamStatus;
  buildDone: boolean;
  hasLines: boolean;
}

/**
 * A native EventSource failure (stream 404 after the job's TTL pruned it) never
 * moves `phase` — only `connectionStatus` drops. On an already-terminal build
 * that is expiry; on a live one it is a real drop, which Retry can recover.
 * A build that died before its job was
 * ever created leaves `phase` at "waiting" forever, so a terminal state there
 * ends the wait too.
 */
export function deriveBuildLogsView({
  phase,
  connectionStatus,
  buildDone,
  hasLines,
}: ViewInput): BuildLogsView {
  if (phase === "unavailable") return "expired";

  const streamDead =
    phase === "error" || (phase === "streaming" && connectionStatus === "disconnected");
  if (streamDead) return buildDone ? "expired" : "interrupted";

  if (phase === "waiting") return buildDone ? "endedWithoutLogs" : "waiting";
  if (hasLines) return "log";
  if (phase === "ended") return "noOutput";
  return "connecting";
}

export interface BuildLogsModalProps {
  open: boolean;
  onClose: () => void;
  orgId: string;
  projectName: string;
  stackId: string;
  buildId: string;
  resourceName: string;
}

export function BuildLogsModal({ open, onClose, ...rest }: BuildLogsModalProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) onClose();
      }}
    >
      <DialogContent className="flex h-[85vh] max-w-4xl flex-col gap-0 overflow-hidden p-0 sm:max-w-4xl">
        <BuildLogsBody key={rest.buildId} open={open} {...rest} />
      </DialogContent>
    </Dialog>
  );
}

type BuildLogsBodyProps = Omit<BuildLogsModalProps, "onClose">;

function BuildLogsBody({
  open,
  orgId,
  projectName,
  stackId,
  buildId,
  resourceName,
}: BuildLogsBodyProps) {
  const { lines, phase, connectionStatus, build, error, retry } = useBuildLogStream({
    orgId,
    projectName,
    stackId,
    buildId,
    enabled: open,
  });

  const state = build?.status?.state ?? "";
  const buildDone = TERMINAL_BUILD_STATES.has(state);
  const view = deriveBuildLogsView({
    phase,
    connectionStatus,
    buildDone,
    hasLines: lines.length > 0,
  });
  const revision = build?.status?.build_source_revision?.slice(0, REVISION_LENGTH);

  return (
    <>
      <DialogHeader className="flex-row items-center gap-3 space-y-0 border-b border-border px-4 py-3">
        <DialogTitle className="text-sm font-semibold">Build logs — {resourceName}</DialogTitle>
        <DialogDescription className="sr-only">
          Streamed output from the image build for {resourceName}.
        </DialogDescription>
        {revision && (
          <span className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[11px] text-fg-muted">
            {revision}
          </span>
        )}
        {state && <StatusPill variant={statusVariant("build", state)}>{state}</StatusPill>}
      </DialogHeader>

      <div className="relative min-h-0 flex-1 bg-[#070a0f]">
        {view === "log" ? (
          <div className="h-full">
            <LazyLog
              text={lines.join("\n")}
              follow={phase === "streaming"}
              height="auto"
              extraLines={1}
              enableSearch
              caseInsensitive
              selectableLines
              style={LOG_STYLE}
            />
          </div>
        ) : (
          <CenterState {...centerStateFor(view, error, retry)} />
        )}
      </div>

      {phase === "ended" && buildDone && <OutcomeBanner state={state} />}
    </>
  );
}

interface CenterStateProps {
  icon: React.ReactNode;
  title: string;
  body: string;
  action?: React.ReactNode;
}

const EXPIRED_BODY =
  "Logs for completed builds are kept only for a short time. The build's outcome and failure details remain on this build.";

function centerStateFor(
  view: Exclude<BuildLogsView, "log">,
  error: string | null,
  retry: () => void,
): CenterStateProps {
  switch (view) {
    case "expired":
      return {
        icon: <Hourglass className="h-6 w-6 text-fg-muted" />,
        title: "Build logs are no longer available",
        body: EXPIRED_BODY,
      };
    case "endedWithoutLogs":
      return {
        icon: <Hourglass className="h-6 w-6 text-fg-muted" />,
        title: "The build ended before it produced logs",
        body: "The build never produced any output. Its outcome and failure details remain on this build.",
      };
    case "waiting":
      return {
        icon: <Loader2 className="h-6 w-6 animate-spin text-info" />,
        title: "Waiting for the build to start…",
        body: "The build hasn't started yet. Checking again every few seconds.",
      };
    case "interrupted":
      return {
        icon: <WifiOff className="h-6 w-6 text-warn" />,
        title: "Log stream interrupted",
        body: error ?? "The log stream dropped before the build finished.",
        action: (
          <Button variant="outline" size="sm" onClick={retry}>
            <RefreshCw className="mr-1.5 h-3.5 w-3.5" />
            Retry
          </Button>
        ),
      };
    case "noOutput":
      return {
        icon: <Hourglass className="h-6 w-6 text-fg-muted" />,
        title: "The build produced no log output",
        body: "The stream completed without emitting a single line.",
      };
    case "connecting":
      return {
        icon: <Loader2 className="h-6 w-6 animate-spin text-info" />,
        title: "Connecting to the build…",
        body: "Log lines will appear as soon as the build produces output.",
      };
  }
}

function OutcomeBanner({ state }: { state: string }) {
  const ok = state === BuildPhase.Success;
  return (
    <div
      className={
        ok
          ? "border-t border-border bg-success/10 px-4 py-2 font-mono text-xs text-success"
          : "border-t border-border bg-danger/10 px-4 py-2 font-mono text-xs text-danger"
      }
    >
      {ok ? "✓ Build succeeded — log stream complete" : "✕ Build failed — log stream complete"}
    </div>
  );
}

function CenterState({ icon, title, body, action }: CenterStateProps) {
  return (
    <div className="absolute inset-0 grid place-items-center p-6 text-center">
      <div className="max-w-sm">
        <div className="flex justify-center">{icon}</div>
        <h3 className="mt-3 text-sm font-semibold text-foreground">{title}</h3>
        <p className="mt-1 text-xs text-fg-muted">{body}</p>
        {action && <div className="mt-3 flex justify-center">{action}</div>}
      </div>
    </div>
  );
}
