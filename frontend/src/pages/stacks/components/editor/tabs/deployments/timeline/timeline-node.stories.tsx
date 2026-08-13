import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { http } from 'msw'
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
import { ReleaseEventType } from '@/api/releases'
import { ReleaseState } from '../release-states'
import type { DetailState, ReleaseDetail } from '../use-release-detail'
import { TimelineNode } from './timeline-node'

// Prop-level ReleaseDetail stub: ensure/refresh are inert, peek serves canned
// states — bodies render without the provider or a detail endpoint.
const stubDetail = (states: Record<string, DetailState> = {}): ReleaseDetail => ({
  ensure: fn(),
  refresh: fn(),
  peek: (id) => (id && states[id]) || { loading: false },
})

const logContext = { orgId: ORG_ID, projectName: DEFAULT_PROJECT, stackId: STACK_ID }

const meta = {
  title: 'Features/Deployments/TimelineNode',
  component: TimelineNode,
  tags: ['ai-generated'],
  args: {
    stack: makeStack(),
    logContext,
    isActive: false,
    isLive: false,
    isOpen: true,
    onToggle: fn(),
    onRollback: fn(),
    onCancel: fn(),
    onCopyId: fn(),
  },
  decorators: [
    (Story) => (
      <div className="max-w-[900px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof TimelineNode>

export default meta
type Story = StoryObj<typeof meta>

// Historical released node, open: post-mortem body with a one-shot event fetch.
export const Succeeded: Story = {
  args: {
    release: makeRelease(),
    detail: stubDetail({ 'rel-12': { loading: false, data: makeReleaseDetail() } }),
    isLive: true,
  },
  parameters: {
    msw: [
      ...releaseHandlers({}, [
        makeReleaseEvent({ sequence: 1 }),
        makeReleaseEvent({
          sequence: 2,
          type: ReleaseEventType.ResourceReady,
          level: 'success',
          message: 'web: Ready',
          occurred_at: '2026-07-30T11:59:20Z',
        }),
      ]),
      ...baselineHandlers,
    ],
  },
  // Event feed arrives async via the terminal /events fetch.
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('Deploying — rolling out revision 4')).toBeInTheDocument()
  },
}

export const Failed: Story = {
  args: {
    release: makeRelease({
      id: 'rel-13',
      sequence: 13,
      state: ReleaseState.Failed,
      message: 'timed out waiting for resource "web" to become ready',
      completed_at: '2026-07-30T12:09:25Z',
    }),
    detail: stubDetail({
      'rel-13': {
        loading: false,
        data: makeReleaseDetail({
          id: 'rel-13',
          sequence: 13,
          state: ReleaseState.Failed,
          live_status: undefined,
          outcome: {
            resources: {
              web: { phase: 'Failed', replicas: 2, ready_replicas: 0, message: 'CrashLoopBackOff' },
              worker: { phase: 'Ready', replicas: 1, ready_replicas: 1 },
            },
          },
        }),
      },
    }),
  },
  parameters: {
    msw: [
      ...releaseHandlers({}, [
        makeReleaseEvent({
          release_id: 'rel-13',
          sequence: 1,
          type: ReleaseEventType.ResourceFailed,
          level: 'error',
          message: 'web: back-off restarting failed container',
        }),
      ]),
      ...baselineHandlers,
    ],
  },
}

// Newest node mid-deploy: live body streaming events over SSE.
export const InProgress: Story = {
  args: {
    release: makeRelease({ id: 'rel-14', sequence: 14, state: ReleaseState.InProgress, completed_at: undefined }),
    detail: stubDetail({
      'rel-14': {
        loading: false,
        data: makeReleaseDetail({
          id: 'rel-14',
          sequence: 14,
          state: ReleaseState.InProgress,
          completed_at: undefined,
          live_status: {
            health: 'progressing',
            resources: { web: { state: 'Progressing', replicas: 2, available_replicas: 1 } },
          },
          outcome: undefined,
        }),
      },
    }),
    isActive: true,
  },
  parameters: {
    msw: [
      http.get(`${RELEASES_PATH}/:releaseId/events/stream`, () =>
        sseResponse([
          { data: makeReleaseEvent({ release_id: 'rel-14', sequence: 1, message: 'web: rolling out revision 5' }), delay: 400 },
        ]),
      ),
      ...releaseHandlers({}, []),
      ...baselineHandlers,
    ],
  },
  // The "live" badge only appears once the EventSource actually connects.
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('live', undefined, { timeout: 4000 })).toBeInTheDocument()
    await expect(
      await canvas.findByText('Deploying — rolling out revision 5', undefined, { timeout: 4000 }),
    ).toBeInTheDocument()
  },
}

// Rollback release renders its cause as "Rollback to #N"; collapsed row only.
export const RolledBack: Story = {
  args: {
    release: makeRelease({
      id: 'rel-15',
      sequence: 15,
      cause: { kind: 'rollback', detail: 'rollback to release #8' },
    }),
    detail: stubDetail(),
    isOpen: false,
  },
}
