import { chromium } from 'playwright'
const OUT = '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
await p.goto('http://localhost:6006/iframe.html?id=shell-topnav-options--d-header-absorbed&globals=theme:dark&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(2200)
await p.evaluate(() => { const el = document.querySelector('.scrollbar-hide'); if (el) el.scrollTop = 180 })
await p.waitForTimeout(400)
await p.screenshot({ path: `${OUT}/tn-d2-dark.png`, clip: { x: 0, y: 0, width: 1440, height: 340 } })
console.log('ok')
await b.close()
