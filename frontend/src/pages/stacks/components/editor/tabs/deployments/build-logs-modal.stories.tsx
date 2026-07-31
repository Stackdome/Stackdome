import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, waitFor, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { sseResponse } from '../../../../../../../.storybook/sse'
import { baselineHandlers } from '../../../../../../../.storybook/msw-handlers'
import {
  DEFAULT_PROJECT,
  ORG_ID,
  STACK_ID,
  makeImageBuild,
} from '../../../../../../../.storybook/fixtures'
import { BuildPhase } from '@/api/image-builds'
import { BuildLogsModal } from './build-logs-modal'

const BUILD_URL = '/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/builds/:buildId'
const BUILD_LOGS_URL = `${BUILD_URL}/logs`

const buildLines = [
  { data: 'Step 1/7 : FROM node:22-alpine' },
  { data: 'Step 2/7 : WORKDIR /app', delay: 400 },
  { data: 'Step 3/7 : COPY package.json pnpm-lock.yaml ./', delay: 400 },
  { data: 'Step 4/7 : RUN pnpm install --frozen-lockfile', delay: 500 },
  { data: 'Successfully tagged ghcr.io/acme/orders-api:a1b2c3d', delay: 500 },
]

const buildHandler = (build = makeImageBuild()) =>
  http.get(BUILD_URL, () => HttpResponse.json(build))

const meta = {
  title: 'Features/Deployments/BuildLogsModal',
  component: BuildLogsModal,
  tags: ['ai-generated'],
  args: {
    open: true,
    onClose: fn(),
    orgId: ORG_ID,
    projectName: DEFAULT_PROJECT,
    stackId: STACK_ID,
    buildId: 'build-1',
    resourceName: 'web',
  },
} satisfies Meta<typeof BuildLogsModal>

export default meta
type Story = StoryObj<typeof meta>

export const Streaming: Story = {
  parameters: {
    msw: [
      buildHandler(makeImageBuild({ status: { state: BuildPhase.Pending } })),
      http.get(BUILD_LOGS_URL, () => sseResponse(buildLines)),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await waitFor(async () => {
      await expect(await body.findByText(/FROM node:22-alpine/)).toBeInTheDocument()
    }, { timeout: 8000 })
  },
}

// Terminal build whose stream closes: the outcome banner replaces the live tail.
export const Succeeded: Story = {
  parameters: {
    msw: [
      buildHandler(),
      http.get(BUILD_LOGS_URL, () => sseResponse(buildLines, { keepOpen: false })),
      ...baselineHandlers,
    ],
  },
}

// The build's job TTL pruned the logs. A 404 (not a network error) is what
// closes the EventSource for good — a network error just makes the browser
// retry, which is how a live build's first seconds look.
export const Expired: Story = {
  parameters: {
    msw: [
      buildHandler(),
      http.get(BUILD_LOGS_URL, () => new HttpResponse(null, { status: 404 })),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(
      await body.findByText(/no longer available/i, undefined, { timeout: 8000 }),
    ).toBeInTheDocument()
  },
}

// No BuildJobCreated condition yet — the modal waits and re-polls.
export const WaitingToStart: Story = {
  parameters: {
    msw: [
      buildHandler(
        makeImageBuild({ status: { state: BuildPhase.Pending, conditions: [] } }),
      ),
      http.get(BUILD_LOGS_URL, () => sseResponse([])),
      ...baselineHandlers,
    ],
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText(/Waiting for the build to start/i)).toBeInTheDocument()
  },
}

export const BuildFailed: Story = {
  parameters: {
    msw: [
      buildHandler(makeImageBuild({ status: { state: BuildPhase.Failed } })),
      http.get(BUILD_LOGS_URL, () =>
        sseResponse(
          [
            { data: 'Step 4/7 : RUN pnpm install --frozen-lockfile' },
            { data: 'ERR_PNPM_OUTDATED_LOCKFILE  Cannot install with frozen-lockfile', delay: 400 },
          ],
          { keepOpen: false },
        ),
      ),
      ...baselineHandlers,
    ],
  },
}
