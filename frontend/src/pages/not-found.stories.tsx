import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, within } from 'storybook/test'
import NotFoundPage from './not-found'

const meta = {
  title: 'Pages/NotFound',
  component: NotFoundPage,
  tags: ['ai-generated'],
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof NotFoundPage>

export default meta
type Story = StoryObj<typeof meta>

// Flat empty-state voice: sentence case, no illustration shadow, the theme
// toggle reads as a filled icon control in the corner.
export const Default: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText('Page not found')).toBeInTheDocument()
    await expect(
      canvas.getByRole('button', { name: /back to stacks/i }),
    ).toBeInTheDocument()
    await expect(canvas.getByRole('button', { name: /toggle theme/i })).toBeInTheDocument()
  },
}
