import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { sseResponse } from '../../../../../../../.storybook/sse'
import { baselineHandlers } from '../../../../../../../.storybook/msw-handlers'
import { ORG_ID, STACK_ID } from '../../../../../../../.storybook/fixtures'
import { MetricsTab } from './metrics-tab'

const STACK_METRICS_URL = '/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/metrics'
const RESOURCE_METRICS_URL =
  '/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/resources/:name/metrics'

// ResourceMetrics samples: cpu_usage in millicores, memory_usage in Mi, both strings.
const stackSamples = [
  { cpu_usage: '120', memory_usage: '256', timestamp: '2026-07-31T10:00:00Z' },
  { cpu_usage: '155', memory_usage: '281', timestamp: '2026-07-31T10:00:05Z' },
  { cpu_usage: '190', memory_usage: '300', timestamp: '2026-07-31T10:00:10Z' },
  { cpu_usage: '165', memory_usage: '290', timestamp: '2026-07-31T10:00:15Z' },
  { cpu_usage: '240', memory_usage: '312', timestamp: '2026-07-31T10:00:20Z' },
].map((data, i) => ({ data, delay: i === 0 ? 0 : 800 }))

const resourceSamples: Record<string, { cpu_usage: string; memory_usage: string }> = {
  web: { cpu_usage: '180', memory_usage: '200' },
  worker: { cpu_usage: '60', memory_usage: '112' },
}

const meta = {
  title: 'Features/Metrics/MetricsTab',
  component: MetricsTab,
  tags: ['ai-generated'],
  args: {
    stackId: STACK_ID,
    organizationId: ORG_ID,
    resources: [
      { name: 'web', workload_type: 'Service' },
      { name: 'worker', workload_type: 'Service' },
    ],
  },
} satisfies Meta<typeof MetricsTab>

export default meta
type Story = StoryObj<typeof meta>

export const Live: Story = {
  parameters: {
    msw: [
      http.get(STACK_METRICS_URL, () => sseResponse(stackSamples)),
      http.get(RESOURCE_METRICS_URL, ({ params }) => {
        const sample = resourceSamples[params.name as string] ?? resourceSamples.worker
        return sseResponse([
          { data: { ...sample, timestamp: '2026-07-31T10:00:00Z' }, delay: 400 },
        ])
      }),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    // Last scripted stack sample lands ~3.2s in — proves samples keep streaming.
    await waitFor(() => expect(canvas.getByText('240m')).toBeVisible(), { timeout: 8000 })
    await expect(canvas.getByText('180m')).toBeVisible()
  },
}

// Streams connect but never emit: Live pill, dashes, per-resource placeholders.
export const Empty: Story = {
  parameters: {
    msw: [
      http.get(STACK_METRICS_URL, () => sseResponse([])),
      http.get(RESOURCE_METRICS_URL, () => sseResponse([])),
      ...baselineHandlers,
    ],
  },
}

export const Disconnected: Story = {
  parameters: {
    msw: [
      http.get(STACK_METRICS_URL, () => HttpResponse.error()),
      http.get(RESOURCE_METRICS_URL, () => HttpResponse.error()),
      ...baselineHandlers,
    ],
  },
}
