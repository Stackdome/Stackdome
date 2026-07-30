import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { within } from 'storybook/test'
import { EndpointPills } from './stack-card'

const meta = {
  title: 'Features/StackCard/EndpointPills',
  component: EndpointPills,
  tags: ['ai-generated'],
} satisfies Meta<typeof EndpointPills>

export default meta
type Story = StoryObj<typeof meta>

export const TwoPills: Story = {
  args: {
    urls: [
      { resource: 'web', url: 'https://web.example.com' },
      { resource: 'api', url: 'https://api.example.com' },
    ],
  },
}

// Past two pills the rest collapse into a "+N" popover.
export const Overflow: Story = {
  args: {
    urls: [
      { resource: 'web', url: 'https://web.example.com' },
      { resource: 'api', url: 'https://api.example.com' },
      { resource: 'admin', url: 'https://admin.example.com' },
      { resource: 'docs', url: 'https://docs.example.com' },
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: /2 more endpoints/i }))
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('admin')).toBeInTheDocument()
    await expect(await body.findByText('docs')).toBeInTheDocument()
  },
}

export const UrllessEntriesFiltered: Story = {
  args: {
    urls: [{ resource: 'web', url: 'https://web.example.com' }, { resource: 'worker' }],
  },
}
