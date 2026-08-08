import { chromium } from 'playwright'
const OUT = '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
for (const [id, file] of [['e-chatgpt-scale','sc-e'],['f-product-scale','sc-f'],['d-header-absorbed','sc-d']]) {
  await p.goto(`http://localhost:6006/iframe.html?id=shell-topnav-options--${id}&globals=theme:light&viewMode=story`, { waitUntil: 'networkidle' })
  await p.waitForTimeout(2200)
  await p.screenshot({ path: `${OUT}/${file}.png`, clip: { x: 224, y: 0, width: 1216, height: 260 } })
  console.log('wrote', file)
}
await b.close()
