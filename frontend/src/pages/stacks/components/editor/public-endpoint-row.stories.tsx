import type { Meta, StoryObj } from '@storybook/react-vite'
import { EndpointInlineList, PublicEndpointRow, type PublicEndpoint } from './public-endpoint-row'

const endpoints: PublicEndpoint[] = [
  { service: 'web', url: 'https://web.example.com', port: 443, variant: 'ready' },
  {
    service: 'api',
    url: 'https://api.example.com',
    port: 8080,
    variant: 'pending',
    urls: [
      { url: 'https://api.example.com', target_port: 8080 },
      { url: 'https://admin.api.example.com', target_port: 9090 },
      { url: 'https://metrics.api.example.com', target_port: 9100 },
    ],
  },
]

const meta = {
  title: 'Features/EditorChrome/PublicEndpointRow',
  component: PublicEndpointRow,
  tags: ['ai-generated'],
  args: { endpoints },
} satisfies Meta<typeof PublicEndpointRow>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const LongURL: Story = {
  args: {
    endpoints: [
      {
        service: 'web',
        url: 'https://extremely-long-subdomain-that-tests-truncation.us-east-1.staging.platform.example.com',
        port: 443,
        variant: 'ready',
      },
    ],
  },
}

/** Zen/collapsed header bar: label-less tooltip chips. */
export const Compact: Story = {
  args: { compact: true },
}

/** Drawer-header block: best URL inline, the rest behind a "+N" popover. */
export const InlineList: Story = {
  render: () => (
    <EndpointInlineList
      service="api"
      urls={[
        { url: 'https://api.example.com', target_port: 8080 },
        { url: 'https://admin.api.example.com', target_port: 9090 },
      ]}
    />
  ),
}
