import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { http, HttpResponse } from 'msw'
import { baselineHandlers } from '../../../../../.storybook/msw-handlers'
import { makeAddon } from '../../../../../.storybook/fixtures'
import { BlockComposer } from './block-composer'

const meta = {
  title: 'Features/Wizard/BlockComposer',
  component: BlockComposer,
  tags: ['ai-generated'],
  args: { onBack: fn(), onClose: fn() },
  decorators: [
    (Story) => (
      <div style={{ height: 640 }}>
        <Story />
      </div>
    ),
  ],
  parameters: {
    msw: [
      ...baselineHandlers,
      http.get('/api/v1/organizations/:orgId/projects/:projectName/addons/postgres', () =>
        HttpResponse.json({
          items: [makeAddon(), makeAddon({ id: 'pg-2', name: 'analytics-db' })],
          total: 2,
        }),
      ),
    ],
  },
} satisfies Meta<typeof BlockComposer>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
