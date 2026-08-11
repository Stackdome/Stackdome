import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { WizardChooser } from './wizard-chooser'

const meta = {
  title: 'Features/Wizard/WizardChooser',
  component: WizardChooser,
  tags: ['ai-generated'],
} satisfies Meta<typeof WizardChooser>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    onPickBlocks: fn(),
    onPickTemplate: fn(),
    onPickCompose: fn(),
    onPickBlank: fn(),
    onPickGit: fn(),
  },
  play: async ({ canvasElement }) => {
    // Option tiles hover to an ink wash, never brand orange (D8: assert the
    // parsed stylesheet, not a synthetic :hover).
    const canvas = within(canvasElement)
    const tiles = canvas.getAllByRole('button')
    for (const tile of tiles) {
      expect(tile.className).not.toContain('hover:border-brand')
    }
  },
}
