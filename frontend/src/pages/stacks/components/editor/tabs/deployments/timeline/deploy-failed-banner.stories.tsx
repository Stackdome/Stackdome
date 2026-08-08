import type { Meta, StoryObj } from '@storybook/react-vite'
import { DeployFailedBanner } from './deploy-failed-banner'

const meta = {
  title: 'Features/Deployments/DeployFailedBanner',
  component: DeployFailedBanner,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[720px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof DeployFailedBanner>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    message: 'render failed: secret "db-credentials" referenced by resource "web" was not found in project "default"',
  },
}
