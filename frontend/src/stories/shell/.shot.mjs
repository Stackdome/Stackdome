import { chromium } from 'playwright'
const OUT = '/Users/jaseem/Projects/audit'
const shots = [
  ['light', 'p2-stacks-light.png', null],
  ['dark', 'p2-stacks-dark.png', null],
  ['light', 'p2-zoom-light.png', { x: 240, y: 60, width: 760, height: 360 }],
]
const browser = await chromium.launch()
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
for (const [theme, file, clip] of shots) {
  await page.goto(`http://localhost:6006/iframe.html?id=shell-stacks-pass-2--proposal&globals=theme:${theme}&viewMode=story`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(2000)
  await page.screenshot({ path: `${OUT}/${file}`, ...(clip ? { clip } : {}) })
  console.log('wrote', file)
}
await browser.close()
