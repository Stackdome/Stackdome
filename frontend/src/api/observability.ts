import api from "./client";
import { API_BASE_URL } from "./base-url";

// Logs/metrics (incl. SSE) served from project-scoped stack endpoints; UI scopes to the org's default project.

/** Status of a single SSE connection, from the browser's perspective. */
export type SseStreamStatus = 'connecting' | 'connected' | 'reconnecting' | 'disconnected';

/** Key for the whole-stack stream in the hooks' per-source status maps. */
export const STACK_STREAM_KEY = '__stack__';

/**
 * Wires an EventSource, separating the backend's `event: error` stream errors
 * (the only real API error signal — see pkg/handlers/handler_helper.go) from
 * the browser's native connection-lifecycle errors. Backend errors arrive as a
 * MessageEvent with a payload; native errors are plain Events without one.
 * Native errors with readyState CONNECTING mean the browser is auto-retrying.
 */
export function attachSseHandlers(
  es: EventSource,
  handlers: {
    onData: (data: string) => void;
    onStreamError: (message: string) => void;
    onStatusChange: (status: SseStreamStatus) => void;
  },
): void {
  es.onopen = () => handlers.onStatusChange('connected');
  es.onmessage = (event) => handlers.onData(event.data);
  es.addEventListener('error', (event) => {
    if (event instanceof MessageEvent) {
      // The server closes the stream after a terminal error event — close our
      // side too, or the browser retry-loops and re-fires the error forever.
      es.close();
      handlers.onStreamError(String(event.data));
      handlers.onStatusChange('disconnected');
      return;
    }
    handlers.onStatusChange(es.readyState === EventSource.CLOSED ? 'disconnected' : 'reconnecting');
  });
}

/** Rolls up many per-connection statuses into one pill-worthy status. A single
 *  permanently-dead stream (e.g. the backend refused it pre-stream) must not
 *  mask the streams that ARE delivering, so connected wins over disconnected;
 *  in-flight states win over both so degradation stays visible. */
export function aggregateStreamStatus(statuses: SseStreamStatus[]): SseStreamStatus {
  if (statuses.some((s) => s === 'reconnecting')) return 'reconnecting';
  if (statuses.some((s) => s === 'connecting')) return 'connecting';
  if (statuses.some((s) => s === 'connected')) return 'connected';
  return 'disconnected';
}

interface LogStreamParams {
  follow?: boolean;
  since?: string;
  tail?: number;
}

function withLogParams(path: string, params?: LogStreamParams): string {
  const searchParams = new URLSearchParams();

  if (params?.follow !== undefined) {
    searchParams.set('follow', params.follow.toString());
  }

  if (params?.since) {
    searchParams.set('since', params.since);
  }

  if (params?.tail !== undefined) {
    searchParams.set('tail', params.tail.toString());
  }

  const queryString = searchParams.toString();
  return queryString ? `${path}?${queryString}` : path;
}

export function buildStackLogStreamUrl(
  organizationId: string,
  projectName: string,
  stackId: string,
  params?: LogStreamParams
): string {
  const baseUrl = API_BASE_URL;
  const path = `${baseUrl}/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/logs`;
  return withLogParams(path, params);
}

export function buildStackResourceLogStreamUrl(
  organizationId: string,
  projectName: string,
  stackId: string,
  resourceName: string,
  params?: LogStreamParams
): string {
  const baseUrl = API_BASE_URL;
  const path = `${baseUrl}/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/resources/${resourceName}/logs`;
  return withLogParams(path, params);
}

export async function getStackLogs(organizationId: string, projectName: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/logs`);
  return res.data;
}

export async function getStackMetrics(organizationId: string, projectName: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/metrics`);
  return res.data;
}

export function buildStackMetricsStreamUrl(
  organizationId: string,
  projectName: string,
  stackId: string
): string {
  const baseUrl = API_BASE_URL;
  return `${baseUrl}/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/metrics?stream=true`;
}

export async function getStackResourceMetrics(
  organizationId: string,
  projectName: string,
  stackId: string,
  resourceId: string
) {
  const res = await api.get(`/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/resources/${resourceId}/metrics`);
  return res.data;
}

export function buildStackResourceMetricsStreamUrl(
  organizationId: string,
  projectName: string,
  stackId: string,
  resourceName: string
): string {
  const baseUrl = API_BASE_URL;
  return `${baseUrl}/organizations/${organizationId}/projects/${projectName}/stacks/${stackId}/resources/${resourceName}/metrics?stream=true`;
}

/** One-shot read of a resource log endpoint with follow=false — returns plain lines.
 *  Best-effort: returns [] on any error (pod may be unreachable, issue #98). */
export async function fetchLogSnapshot(
  organizationId: string,
  projectName: string,
  stackId: string,
  resourceName: string,
  tail = 50,
): Promise<string[]> {
  const url = buildStackResourceLogStreamUrl(organizationId, projectName, stackId, resourceName, { follow: false, tail });
  try {
    const res = await fetch(url, { credentials: "include" });
    if (!res.ok) return [];
    const text = await res.text();
    return text.split("\n").map((l) => l.replace(/^data:\s?/, "").trim()).filter((l) => l.length > 0);
  } catch { return []; }
}
