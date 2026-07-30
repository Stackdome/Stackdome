import { useEffect, useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http } from 'msw'
import { sseResponse } from '../../../.storybook/sse'

// Proves the MSW service worker intercepts EventSource requests and can stream
// scripted text/event-stream frames — the mechanism every Features-tier
// streaming story (release events, logs, metrics) depends on.
function SseSmoke() {
  const [messages, setMessages] = useState<string[]>([])
  const [status, setStatus] = useState('connecting')
  useEffect(() => {
    const es = new EventSource('/api/v1/__sse-smoke')
    es.onopen = () => setStatus('connected')
    es.onmessage = (e) => setMessages((m) => [...m, e.data])
    es.onerror = () => setStatus('error')
    return () => es.close()
  }, [])
  return (
    <div className="font-mono text-sm">
      <div data-testid="status">{status}</div>
      <ul data-testid="messages">
        {messages.map((m, i) => (
          <li key={i}>{m}</li>
        ))}
      </ul>
    </div>
  )
}

const meta = {
  title: 'Foundations/SSE Smoke',
  component: SseSmoke,
  tags: ['ai-generated', 'dev'],
  parameters: {
    msw: {
      handlers: [
        http.get('/api/v1/__sse-smoke', () =>
          sseResponse([
            { data: 'first' },
            { data: 'second', delay: 100 },
          ]),
        ),
      ],
    },
  },
} satisfies Meta<typeof SseSmoke>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  play: async ({ canvas }) => {
    await waitFor(
      async () => {
        await expect(canvas.getByTestId('status')).toHaveTextContent('connected')
        await expect(canvas.getByTestId('messages')).toHaveTextContent('second')
      },
      { timeout: 5000 },
    )
  },
}
