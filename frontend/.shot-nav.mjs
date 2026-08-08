import { chromium } from 'playwright'
const OUT = process.argv[2] || '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const shots = [
  ['light', 'nav-light.png', null],
  ['dark', 'nav-dark.png', null],
  ['light', 'nav-zoom-light.png', { x: 0, y: 0, width: 640, height: 900 }],
]
const browser = await chromium.launch()
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
for (const [theme, file, clip] of shots) {
  await page.goto(`http://localhost:6006/iframe.html?id=shell-platform--stacks-list&globals=theme:${theme}&viewMode=story`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(2000)
  await page.screenshot({ path: `${OUT}/${file}`, ...(clip ? { clip } : {}) })
  console.log('wrote', file)
}
await browser.close()
