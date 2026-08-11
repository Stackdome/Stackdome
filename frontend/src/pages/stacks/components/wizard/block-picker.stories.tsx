import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { blockCatalog, BLOCK_CATEGORY_META, BlockId } from '@/pages/stacks/data/blocks/registry'
import { BlockPicker } from './block-picker'

const meta = {
  title: 'Features/Wizard/BlockPicker',
  component: BlockPicker,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[640px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof BlockPicker>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    catalog: blockCatalog,
    categories: BLOCK_CATEGORY_META,
    addedIds: [BlockId.Postgres],
    onAdd: fn(),
    query: '',
  },
}
