import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from './table'

const meta = {
  title: 'Primitives/Table',
  component: Table,
  tags: ['ai-generated'],
} satisfies Meta<typeof Table>

export default meta
type Story = StoryObj<typeof meta>

const ROWS = [
  { name: 'web', replicas: 3, cpu: '120m', memory: '256Mi' },
  { name: 'worker', replicas: 2, cpu: '80m', memory: '512Mi' },
  { name: 'scheduler', replicas: 1, cpu: '40m', memory: '128Mi' },
]

export const Default: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Resource</TableHead>
          <TableHead>Replicas</TableHead>
          <TableHead>CPU</TableHead>
          <TableHead>Memory</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {ROWS.map((row) => (
          <TableRow key={row.name}>
            <TableCell className="font-medium">{row.name}</TableCell>
            <TableCell className="mono-num">{row.replicas}</TableCell>
            <TableCell className="mono-num">{row.cpu}</TableCell>
            <TableCell className="mono-num">{row.memory}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
}

export const WithCaption: Story = {
  render: () => (
    <Table>
      <TableCaption>Resource usage as of the last reconcile.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>Resource</TableHead>
          <TableHead>Replicas</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {ROWS.map((row) => (
          <TableRow key={row.name}>
            <TableCell>{row.name}</TableCell>
            <TableCell className="mono-num">{row.replicas}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
}

// Contract: hairline-ruled rows, no outer container box/shadow around the table.
export const NoContainerBox: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Resource</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>web</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
  play: async ({ canvas }) => {
    const table = canvas.getByRole('table')
    const container = table.parentElement as HTMLElement
    const style = getComputedStyle(container)
    await expect(style.boxShadow).toBe('none')
    await expect(style.borderWidth).toBe('0px')
  },
}
