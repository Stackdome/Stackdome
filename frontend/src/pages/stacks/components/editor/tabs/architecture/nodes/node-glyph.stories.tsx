import type { ReactNode } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import type { GlyphKind } from '@/pages/stacks/lib/canvas/node-presentation'
import { NodeGlyph } from './node-glyph'

const meta = {
  title: 'Features/Canvas/NodeGlyph',
  component: NodeGlyph,
  tags: ['ai-generated'],
} satisfies Meta<typeof NodeGlyph>

export default meta
type Story = StoryObj<typeof meta>

const GLYPHS: GlyphKind[] = ['web', 'postgres', 'redis', 'database', 'object', 'worker', 'service']
const BRAND_SLUGS = ['postgres', 'redis', 'mysql', 'mongo', 'minio', 'grafana', 'clickhouse']

function Cell({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex w-20 flex-col items-center gap-1.5 rounded-md border border-border p-3">
      {children}
      <span className="font-mono text-[10px] text-fg-muted">{label}</span>
    </div>
  )
}

/** Every generic Lucide glyph, then the brand-icon variants that override them. */
export const AllGlyphs: Story = {
  args: { glyph: 'web' },
  render: () => (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        {GLYPHS.map((glyph) => (
          <Cell key={glyph} label={glyph}>
            <NodeGlyph glyph={glyph} size={17} className="size-[17px] text-fg-2" />
          </Cell>
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        {BRAND_SLUGS.map((slug) => (
          <Cell key={slug} label={slug}>
            <NodeGlyph glyph="service" brandSlug={slug} size={17} className="size-[17px]" />
          </Cell>
        ))}
      </div>
    </div>
  ),
}
