import { useEffect, useRef, useState } from "react";
import { buildReleaseEventStreamUrl, listReleaseEvents, type ReleaseEvent } from "@/api/releases";

const RECONNECT_MS = 2000;
const MAX_RECONNECTS = 10;
const POLL_MS = 5000;

export type UseReleaseEventsStatus = "idle" | "connecting" | "streaming" | "polling" | "closed" | "error";

export interface UseReleaseEventsResult {
  events: ReleaseEvent[];
  status: UseReleaseEventsStatus;
}

export interface UseReleaseEventsArgs {
  orgId: string;
  teamName: string;
  stackId: string;
  releaseId?: string;
  terminal: boolean;
  onEvent?: (e: ReleaseEvent) => void;
}

export function useReleaseEvents({ orgId, teamName, stackId, releaseId, terminal, onEvent }: UseReleaseEventsArgs): UseReleaseEventsResult {
  const [events, setEvents] = useState<ReleaseEvent[]>([]);
  const [status, setStatus] = useState<UseReleaseEventsStatus>("idle");
  const lastSeq = useRef<number>(0);
  const seen = useRef<Set<number>>(new Set());
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;
  const prevReleaseId = useRef<string | undefined>(undefined);

  useEffect(() => {
    // Only clear when the release itself changes. `terminal` flipping false→true for the
    // SAME release (deploy just completed) must not wipe the feed — the one-shot fetch below
    // re-fetches from sequence 0 and `ingest` dedupes against what's already shown, so keeping
    // the prior list avoids a flash-of-empty while that fetch is in flight.
    const isNewRelease = prevReleaseId.current !== releaseId;
    prevReleaseId.current = releaseId;
    if (isNewRelease) {
      setEvents([]);
      lastSeq.current = 0;
      seen.current = new Set();
    }
    if (!releaseId) {
      setStatus("idle");
      return;
    }

    let disposed = false;
    const ingest = (incoming: ReleaseEvent[]) => {
      const fresh = incoming.filter((e) => e.sequence !== undefined && !seen.current.has(e.sequence));
      if (fresh.length === 0) return;
      fresh.forEach((e) => {
        seen.current.add(e.sequence!);
        lastSeq.current = Math.max(lastSeq.current, e.sequence!);
      });
      setEvents((prev) => [...prev, ...fresh].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0)));
      fresh.forEach((e) => onEventRef.current?.(e));
    };

    // The events endpoint is paginated (default 100/page). Follow next_after_sequence
    // so a release with many events keeps its tail — the release-completed / failure
    // events land last and are exactly what a post-mortem or catch-up poll needs.
    const fetchAllSince = async (after?: number): Promise<ReleaseEvent[]> => {
      const all: ReleaseEvent[] = [];
      let cursor = after;
      for (;;) {
        const page = await listReleaseEvents(orgId, teamName, stackId, releaseId, cursor);
        if (disposed) break;
        const items = page.items ?? [];
        all.push(...items);
        if (items.length === 0 || page.next_after_sequence === undefined) break;
        cursor = page.next_after_sequence;
      }
      return all;
    };

    if (terminal) {
      setStatus("polling");
      void fetchAllSince()
        .then((items) => {
          if (disposed) return;
          ingest(items);
          setStatus("closed");
        })
        .catch(() => { if (!disposed) setStatus("error"); });
      return () => { disposed = true; };
    }

    let es: EventSource | null = null;
    let reconnects = 0;
    let pollTimer: ReturnType<typeof setInterval> | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

    const poll = () => {
      void fetchAllSince(lastSeq.current)
        .then((items) => {
          if (disposed) return;
          ingest(items);
          setStatus("polling");
        })
        .catch(() => { if (!disposed) setStatus("error"); });
    };
    const startPolling = () => {
      setStatus("polling");
      poll(); // fire once immediately — don't make the user wait a full POLL_MS after giving up on SSE
      pollTimer = setInterval(poll, POLL_MS);
    };

    const connect = () => {
      setStatus("connecting");
      es = new EventSource(buildReleaseEventStreamUrl(orgId, teamName, stackId, releaseId, lastSeq.current || undefined));
      es.onopen = () => { reconnects = 0; setStatus("streaming"); }; // a clean reopen clears the budget — MAX_RECONNECTS means N failures in a row, not over the stream's life
      es.onmessage = (msg) => {
        try {
          ingest([JSON.parse(msg.data) as ReleaseEvent]);
        } catch {
          // skip malformed frame
        }
      };
      es.onerror = () => {
        es?.close();
        if (disposed) return;
        reconnects += 1;
        if (reconnects > MAX_RECONNECTS) {
          startPolling();
          return;
        }
        reconnectTimer = setTimeout(connect, RECONNECT_MS);
      };
    };
    connect();

    return () => {
      disposed = true;
      es?.close();
      if (pollTimer) clearInterval(pollTimer);
      if (reconnectTimer) clearTimeout(reconnectTimer);
    };
  }, [orgId, teamName, stackId, releaseId, terminal]);

  return { events, status };
}
