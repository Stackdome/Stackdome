import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor } from 'storybook/test'
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableRowActions,
} from './table'
import { Button } from './button'
import { MoreHorizontal, Trash2 } from 'lucide-react'

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

/** §7 — headers are sentence case, `text-label` (11px), `fg-muted`. Uppercase is
 *  a third signal doing a job that size and colour already did, and it costs the
 *  word-shape the eye reads by. */
export const HeadersAreNotUppercase: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Needs attention</TableHead>
          <TableHead>CPU</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>web</TableCell>
          <TableCell className="mono-num">120m</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
  play: async ({ canvas }) => {
    const head = canvas.getByRole('columnheader', { name: 'Needs attention' })
    const style = getComputedStyle(head)
    await expect(style.textTransform).toBe('none')
    await expect(parseFloat(style.fontSize)).toBe(11)
    // The rendered text keeps its sentence case — no CSS is faking it back.
    await expect(head.textContent).toBe('Needs attention')
    // CPU stays uppercase because the word is: §7's initialism exception.
    await expect(canvas.getByRole('columnheader', { name: 'CPU' }).textContent).toBe('CPU')
  },
}

/** §7 — row actions appear on hover. A kebab on every row at rest is chrome
 *  competing with content. Hidden by opacity, so the control keeps its tab stop
 *  and the row does not reflow when the pointer arrives. */
export const RowActionsRevealOnHover: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Resource</TableHead>
          <TableHead className="w-8" />
        </TableRow>
      </TableHeader>
      <TableBody>
        {ROWS.map((row) => (
          <TableRow key={row.name}>
            <TableCell className="font-medium">{row.name}</TableCell>
            <TableCell>
              <TableRowActions>
                <Button variant="ghost" size="icon-sm" shape="flat" aria-label={`Delete ${row.name}`}>
                  <Trash2 />
                </Button>
                <Button variant="ghost" size="icon-sm" shape="flat" aria-label={`More for ${row.name}`}>
                  <MoreHorizontal />
                </Button>
              </TableRowActions>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
  play: async ({ canvas }) => {
    const trigger = canvas.getByRole('button', { name: 'More for web' })
    const actions = trigger.closest('[data-slot="table-row-actions"]') as HTMLElement

    // At rest the row is content only.
    await expect(getComputedStyle(actions).opacity).toBe('0')

    // Hidden must not mean unreachable: the control keeps its tab stop, and
    // focus reveals it through the same declaration hover uses.
    //
    // The hover half is not asserted here on purpose. `:hover` is a real
    // pointer state — synthetic events do not enter it, so a `userEvent.hover`
    // assertion would read 0 and prove only that the harness cannot hover.
    // Review that half in Storybook; this guards the two states a test can
    // actually observe.
    trigger.focus()
    await waitFor(async () => {
      await expect(getComputedStyle(actions).opacity).toBe('1')
    })

    trigger.blur()
    await waitFor(async () => {
      await expect(getComputedStyle(actions).opacity).toBe('0')
    })
  },
}
