import { useState, useEffect, useRef, useCallback } from "react";
import {
  getImageBuild,
  isBuildJobCreated,
  isBuildTerminal,
  buildImageBuildLogStreamUrl,
  type ImageBuild,
} from "@/api/image-builds";
import { attachSseHandlers, type SseStreamStatus } from "@/api/observability";

const PREFLIGHT_POLL_MS = 3000;
const BUILD_LOG_TAIL = 200;
const MAX_STREAM_RETRIES = 10;

export type BuildLogPhase = "waiting" | "streaming" | "ended" | "unavailable" | "error";

interface UseBuildLogStreamProps {
  orgId: string;
  projectName: string;
  stackId: string;
  buildId: string;
  /** false while the modal is closed — nothing fetches or connects. */
  enabled: boolean;
}

interface UseBuildLogStreamReturn {
  lines: string[];
  phase: BuildLogPhase;
  connectionStatus: SseStreamStatus;
  build: ImageBuild | null;
  error: string | null;
  retry: () => void;
}

export function useBuildLogStream({
  orgId,
  projectName,
  stackId,
  buildId,
  enabled,
}: UseBuildLogStreamProps): UseBuildLogStreamReturn {
  const [lines, setLines] = useState<string[]>([]);
  const [phase, setPhase] = useState<BuildLogPhase>("waiting");
  const [connectionStatus, setConnectionStatus] = useState<SseStreamStatus>("disconnected");
  const [build, setBuild] = useState<ImageBuild | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const esRef = useRef<EventSource | null>(null);

  const retry = useCallback(() => setRetryNonce((n) => n + 1), []);

  useEffect(() => {
    if (!enabled || !orgId || !projectName || !stackId || !buildId) return;

    let alive = true;
    let pollTimer: ReturnType<typeof setTimeout> | null = null;
    let latestBuild: ImageBuild | null = null;
    let streamAttempts = 0;

    setLines([]);
    setPhase("waiting");
    setConnectionStatus("disconnected");
    setError(null);

    // BuildJobCreated goes True before the job's pod is running, so the stream
    // request can still 409 and the browser gives the connection up for dead.
    // That is the normal first seconds of a build, not an interruption: drop
    // back into the poll loop instead of showing the interrupted view.
    const resumeWaiting = (es: EventSource) => {
      es.close();
      if (esRef.current === es) esRef.current = null;
      if (isBuildTerminal(latestBuild)) return;
      if (streamAttempts >= MAX_STREAM_RETRIES) {
        setPhase("error");
        return;
      }
      setPhase("waiting");
      pollTimer = setTimeout(() => void preflight(), PREFLIGHT_POLL_MS);
    };

    const openStream = () => {
      streamAttempts += 1;
      let sawLine = false;
      let settled = false;
      const url = buildImageBuildLogStreamUrl(orgId, projectName, stackId, buildId, {
        follow: true,
        tail: BUILD_LOG_TAIL,
      });
      const es = new EventSource(url);
      esRef.current = es;
      setPhase("streaming");
      setConnectionStatus("connecting");

      attachSseHandlers(es, {
        onData: (data) => {
          if (!alive) return;
          sawLine = true;
          setLines((prev) => [...prev, data]);
        },
        onStreamError: (message) => {
          if (!alive) return;
          settled = true;
          setError(message);
          setPhase("error");
        },
        onStatusChange: (status) => {
          if (!alive) return;
          setConnectionStatus(status);
          if (status !== "disconnected" || settled || sawLine) return;
          resumeWaiting(es);
        },
      });

      es.addEventListener("end", () => {
        if (!alive) return;
        settled = true;
        es.close();
        setConnectionStatus("disconnected");
        setPhase("ended");
        void getImageBuild(orgId, projectName, stackId, buildId)
          .then((b) => {
            if (alive) {
              latestBuild = b;
              setBuild(b);
            }
          })
          .catch((e: unknown) => {
            if (alive) setError(e instanceof Error ? e.message : String(e));
          });
      });
    };

    const preflight = async () => {
      let b: ImageBuild;
      try {
        b = await getImageBuild(orgId, projectName, stackId, buildId);
      } catch (e) {
        if (!alive) return;
        setPhase("unavailable");
        setError(e instanceof Error ? e.message : String(e));
        return;
      }
      if (!alive) return;
      latestBuild = b;
      setBuild(b);
      if (isBuildJobCreated(b)) {
        openStream();
      } else {
        setPhase("waiting");
        pollTimer = setTimeout(() => void preflight(), PREFLIGHT_POLL_MS);
      }
    };

    void preflight();

    return () => {
      alive = false;
      if (pollTimer) clearTimeout(pollTimer);
      esRef.current?.close();
      esRef.current = null;
    };
  }, [enabled, orgId, projectName, stackId, buildId, retryNonce]);

  return { lines, phase, connectionStatus, build, error, retry };
}
