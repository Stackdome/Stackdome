import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { sseResponse } from '../../../../../../../.storybook/sse'
import { baselineHandlers } from '../../../../../../../.storybook/msw-handlers'
import { ORG_ID, STACK_ID } from '../../../../../../../.storybook/fixtures'
import { withHeight } from '../../../../../../../.storybook/decorators'
import { LogViewer } from './log-viewer'

const STACK_LOGS_URL = '/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/logs'
const RESOURCE_LOGS_URL =
  '/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/resources/:name/logs'

// Raw SSE line format the parser expects: optional "[source]:" prefix, then an
// ISO timestamp, then the message (see parseLogEntry in ./utils.ts).
const logLines = [
  { data: '[web]: 2026-07-31T10:15:01Z Server listening on :8080' },
  { data: '[web]: 2026-07-31T10:15:02Z GET /healthz 200 1ms', delay: 400 },
  { data: '[worker]: 2026-07-31T10:15:03Z Picked up job orders.sync', delay: 400 },
  { data: '[web]: 2026-07-31T10:15:04Z GET /orders/42 200 12ms', delay: 400 },
]

const meta = {
  title: 'Features/Logs/LogViewer',
  component: LogViewer,
  tags: ['ai-generated'],
  decorators: [withHeight(500)],
  args: {
    stackId: STACK_ID,
    organizationId: ORG_ID,
    resources: [{ name: 'web' }, { name: 'worker' }],
  },
} satisfies Meta<typeof LogViewer>

export default meta
type Story = StoryObj<typeof meta>

export const Streaming: Story = {
  parameters: {
    msw: [
      http.get(STACK_LOGS_URL, () => sseResponse(logLines)),
      http.get(RESOURCE_LOGS_URL, ({ params }) =>
        sseResponse(logLines.filter((l) => l.data.startsWith(`[${params.name as string}]`))),
      ),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvasElement }) => {
    // Last scripted line lands ~1.2s in — proves lines keep streaming in.
    await waitFor(
      () => expect(canvasElement).toHaveTextContent('GET /orders/42 200 12ms'),
      { timeout: 8000 },
    )
  },
}

// Stream connects but never emits: "No logs yet" empty state.
export const EmptyLogs: Story = {
  parameters: {
    msw: [
      http.get(STACK_LOGS_URL, () => sseResponse([])),
      http.get(RESOURCE_LOGS_URL, () => sseResponse([])),
      ...baselineHandlers,
    ],
  },
}

export const Disconnected: Story = {
  parameters: {
    msw: [
      http.get(STACK_LOGS_URL, () => HttpResponse.error()),
      http.get(RESOURCE_LOGS_URL, () => HttpResponse.error()),
      ...baselineHandlers,
    ],
  },
}
