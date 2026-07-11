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

  useEffect(() => {
    setEvents([]);
    lastSeq.current = 0;
    seen.current = new Set();
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

    if (terminal) {
      setStatus("polling");
      void listReleaseEvents(orgId, teamName, stackId, releaseId)
        .then((page) => {
          if (disposed) return;
          ingest(page.items ?? []);
          setStatus("closed");
        })
        .catch(() => { if (!disposed) setStatus("error"); });
      return () => { disposed = true; };
    }

    let es: EventSource | null = null;
    let reconnects = 0;
    let pollTimer: ReturnType<typeof setInterval> | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

    const startPolling = () => {
      setStatus("polling");
      pollTimer = setInterval(() => {
        void listReleaseEvents(orgId, teamName, stackId, releaseId, lastSeq.current)
          .then((page) => ingest(page.items ?? []))
          .catch(() => setStatus("error"));
      }, POLL_MS);
    };

    const connect = () => {
      setStatus("connecting");
      es = new EventSource(buildReleaseEventStreamUrl(orgId, teamName, stackId, releaseId, lastSeq.current || undefined));
      es.onopen = () => setStatus("streaming");
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
