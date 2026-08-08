import { chromium } from 'playwright'
const OUT = '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
await p.goto('http://localhost:6006/iframe.html?id=shell-platform--stacks-list&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(2500)
await p.screenshot({ path: `${OUT}/platform-now.png`, clip: { x: 224, y: 0, width: 1216, height: 300 } })
console.log('ok')
await b.close()
