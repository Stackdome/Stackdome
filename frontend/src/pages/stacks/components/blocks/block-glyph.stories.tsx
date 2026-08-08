import type { Meta, StoryObj } from '@storybook/react-vite'
import { BRAND_ICONS } from '@/components/branded/brand-icon-registry'
import { BlockGlyph } from './block-glyph'

// Lucide fallbacks known to BlockGlyph, plus an unknown key that falls back to Box.
const LUCIDE_KEYS = ['globe', 'database', 'zap', 'box', 'unknown-key']

const meta = {
  title: 'Features/Wizard/BlockGlyph',
  component: BlockGlyph,
  tags: ['ai-generated'],
} satisfies Meta<typeof BlockGlyph>

export default meta
type Story = StoryObj<typeof meta>

export const AllGlyphs: Story = {
  args: { icon: 'box' },
  render: () => (
    <div className="grid max-w-[640px] grid-cols-4 gap-3">
      {[...Object.keys(BRAND_ICONS), ...LUCIDE_KEYS].map((icon) => (
        <div key={icon} className="flex items-center gap-2 rounded-md border bg-card px-3 py-2">
          <BlockGlyph icon={icon} size={20} />
          <span className="font-mono text-label text-muted-foreground">{icon}</span>
        </div>
      ))}
    </div>
  ),
}
