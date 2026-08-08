import { chromium } from 'playwright'
const ids = ['a-empty', 'b-action', 'c-search', 'd-header-absorbed']
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
for (const id of ids) {
  await p.goto(`http://localhost:6006/iframe.html?id=shell-topnav-options--${id}&globals=theme:light&viewMode=story`, { waitUntil: 'networkidle' })
  await p.waitForTimeout(2000)
  const r = await p.evaluate(() => {
    const list = document.querySelector('.divide-y')
    const rows = list ? [...list.children] : []
    const first = rows[0]?.getBoundingClientRect()
    const visible = rows.filter(a => a.getBoundingClientRect().bottom <= window.innerHeight).length
    return { firstRowTop: first ? +first.top.toFixed(0) : null, rowsFullyVisible: visible, rowCount: rows.length }
  })
  console.log(id.padEnd(20), JSON.stringify(r))
}
await b.close()
