import type { Meta, StoryObj } from '@storybook/react-vite'
import { http } from 'msw'
import { sseResponse } from '../../../../../../../.storybook/sse'
import { baselineHandlers } from '../../../../../../../.storybook/msw-handlers'
import { ORG_ID, STACK_ID } from '../../../../../../../.storybook/fixtures'
import { withHeight } from '../../../../../../../.storybook/decorators'
import { LogsTab } from './logs-tab'

const meta = {
  title: 'Features/Logs/LogsTab',
  component: LogsTab,
  tags: ['ai-generated'],
  decorators: [withHeight(500)],
} satisfies Meta<typeof LogsTab>

export default meta
type Story = StoryObj<typeof meta>

// Pure pass-through to LogViewer — one story proves the wiring.
export const Default: Story = {
  args: {
    stackId: STACK_ID,
    organizationId: ORG_ID,
    resources: [{ name: 'web' }, { name: 'worker' }],
  },
  parameters: {
    msw: [
      http.get('/api/v1/organizations/:orgId/projects/:proj/stacks/:stackId/logs', () =>
        sseResponse([
          { data: '[web]: 2026-07-31T10:15:01Z Server listening on :8080' },
          { data: '[web]: 2026-07-31T10:15:02Z GET /healthz 200 1ms', delay: 400 },
        ]),
      ),
      ...baselineHandlers,
    ],
  },
}
