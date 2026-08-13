import { HttpResponse } from 'msw'

export interface SseEvent {
  event?: string
  data: unknown
  delay?: number
}

// Streams scripted SSE frames to a browser EventSource. keepOpen (default)
// leaves the stream open after the last frame — closing it triggers the
// EventSource auto-reconnect loop, which would re-invoke the handler forever.
export function sseResponse(events: SseEvent[], opts: { keepOpen?: boolean } = {}) {
  const { keepOpen = true } = opts
  const enc = new TextEncoder()
  const stream = new ReadableStream({
    async start(controller) {
      for (const e of events) {
        if (e.delay) await new Promise((r) => setTimeout(r, e.delay))
        const data = typeof e.data === 'string' ? e.data : JSON.stringify(e.data)
        controller.enqueue(enc.encode(`${e.event ? `event: ${e.event}\n` : ''}data: ${data}\n\n`))
      }
      if (!keepOpen) controller.close()
    },
  })
  return new HttpResponse(stream, {
    headers: { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' },
  })
}
