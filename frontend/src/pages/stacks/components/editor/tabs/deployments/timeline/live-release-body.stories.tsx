import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import {
  DEFAULT_PROJECT,
  ORG_ID,
  STACK_ID,
  makeRelease,
  makeReleaseDetail,
  makeReleaseEvent,
  makeStack,
} from '../../../../../../../../.storybook/fixtures'
import { baselineHandlers, releaseHandlers, RELEASES_PATH } from '../../../../../../../../.storybook/msw-handlers'
import { sseResponse } from '../../../../../../../../.storybook/sse'
import { ReleaseEventScope, ReleaseEventType } from '@/api/releases'
import { ReleaseState } from '../release-states'
import type { DetailState, ReleaseDetail } from '../use-release-detail'
import { LiveReleaseBody } from './live-release-body'

const RELEASE_ID = 'rel-14'

const inFlight = makeRelease({ id: RELEASE_ID, sequence: 14, state: ReleaseState.InProgress, completed_at: undefined })

const inFlightDetail = makeReleaseDetail({
  id: RELEASE_ID,
  sequence: 14,
  state: ReleaseState.InProgress,
  completed_at: undefined,
  live_status: {
    health: 'progressing',
    resources: {
      web: { state: 'Progressing', replicas: 2, available_replicas: 1 },
      worker: { state: 'Ready', replicas: 1, available_replicas: 1 },
    },
  },
  outcome: undefined,
})

// Prop-level ReleaseDetail stub — ensure/refresh inert, peek serves canned states.
const stubDetail = (states: Record<string, DetailState> = {}): ReleaseDetail => ({
  ensure: fn(),
  refresh: fn(),
  peek: (id) => (id && states[id]) || { loading: false },
})

const streamEvent = (overrides: Parameters<typeof makeReleaseEvent>[0]) =>
  makeReleaseEvent({ release_id: RELEASE_ID, ...overrides })

const meta = {
  title: 'Features/Deployments/LiveReleaseBody',
  component: LiveReleaseBody,
  tags: ['ai-generated'],
  args: {
    release: inFlight,
    stack: makeStack(),
    logContext: { orgId: ORG_ID, projectName: DEFAULT_PROJECT, stackId: STACK_ID },
    detail: stubDetail({ [RELEASE_ID]: { loading: false, data: inFlightDetail } }),
  },
  decorators: [
    (Story) => (
      <div className="max-w-[900px] rounded-md border border-border bg-card p-4">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof LiveReleaseBody>

export default meta
type Story = StoryObj<typeof meta>

// SSE frames arrive staged 400–800ms apart; the activity feed fills in live.
export const Progressing: Story = {
  parameters: {
    msw: [
      http.get(`${RELEASES_PATH}/:releaseId/events/stream`, () =>
        sseResponse([
          { data: streamEvent({ sequence: 1, message: 'web: rolling out revision 5' }), delay: 400 },
          {
            data: streamEvent({
              sequence: 2,
              type: ReleaseEventType.ResourceWaiting,
              message: 'web: waiting for readiness probe',
              occurred_at: '2026-07-30T11:58:20Z',
            }),
            delay: 700,
          },
          {
            data: streamEvent({
              sequence: 3,
              resource_name: 'worker',
              type: ReleaseEventType.ResourceReady,
              level: 'success',
              message: 'worker: Ready',
              occurred_at: '2026-07-30T11:58:31Z',
            }),
            delay: 800,
          },
        ]),
      ),
      ...releaseHandlers({}, []),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('live', undefined, { timeout: 4000 })).toBeInTheDocument()
    await expect(
      await canvas.findByText('Deploying — rolling out revision 5', undefined, { timeout: 4000 }),
    ).toBeInTheDocument()
    await expect(
      await canvas.findByText('Waiting — waiting for readiness probe', undefined, { timeout: 4000 }),
    ).toBeInTheDocument()
  },
}

// Progress frames, then a resource failure, then the stream errors out
// (event: error → EventSource onerror → reconnect); dedupe keeps the feed stable.
export const Failing: Story = {
  parameters: {
    msw: [
      http.get(`${RELEASES_PATH}/:releaseId/events/stream`, () =>
        sseResponse([
          { data: streamEvent({ sequence: 1, message: 'web: rolling out revision 5' }), delay: 400 },
          {
            data: streamEvent({
              sequence: 2,
              type: ReleaseEventType.ResourceFailed,
              level: 'error',
              scope: ReleaseEventScope.Resource,
              message: 'web: back-off restarting failed container',
              occurred_at: '2026-07-30T11:58:44Z',
            }),
            delay: 600,
          },
          { event: 'error', data: 'boom' },
        ]),
      ),
      ...releaseHandlers({}, []),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(
      await canvas.findByText('Failed to start — back-off restarting failed container', undefined, { timeout: 4000 }),
    ).toBeInTheDocument()
  },
}

// Stream endpoint unreachable: no "live" badge, feed stays empty while the
// hook cycles reconnect attempts before its poll fallback.
export const Disconnected: Story = {
  parameters: {
    msw: [
      http.get(`${RELEASES_PATH}/:releaseId/events/stream`, () => HttpResponse.error()),
      ...releaseHandlers({}, []),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('No activity yet')).toBeInTheDocument()
    await expect(canvas.queryByText('live')).toBeNull()
  },
}
