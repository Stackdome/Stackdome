import type { Meta, StoryObj } from '@storybook/react-vite'
import { AddonTypeIcon } from './addon-type-icon'

// Raster brand icons, the inline Ollama glyph, and an unknown type that falls
// back to the Puzzle icon.
const TYPES = ['postgres', 'redis', 'ollama', 'unknown']

const meta = {
  title: 'Features/Addons/AddonTypeIcon',
  component: AddonTypeIcon,
  tags: ['ai-generated'],
} satisfies Meta<typeof AddonTypeIcon>

export default meta
type Story = StoryObj<typeof meta>

export const AllTypes: Story = {
  args: { type: 'postgres' },
  render: () => (
    <div className="grid max-w-[420px] grid-cols-2 gap-3">
      {TYPES.map((type) => (
        <div key={type} className="flex items-center gap-2 rounded-md border bg-card px-3 py-2">
          <AddonTypeIcon type={type} size={20} />
          <span className="font-mono text-label text-muted-foreground">{type}</span>
        </div>
      ))}
    </div>
  ),
}
