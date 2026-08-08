import { chromium } from 'playwright'
const OUT = '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const browser = await chromium.launch()
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })

async function shot(id, file, { scroll = 0, clip } = {}) {
  await page.goto(`http://localhost:6006/iframe.html?id=shell-topnav-options--${id}&globals=theme:light&viewMode=story`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(2200)
  if (scroll) {
    await page.evaluate((y) => {
      const el = document.querySelector('.scrollbar-hide')
      if (el) el.scrollTop = y
    }, scroll)
    await page.waitForTimeout(400)
  }
  await page.screenshot({ path: `${OUT}/${file}.png`, clip: clip ?? { x: 224, y: 0, width: 1216, height: 330 } })
  console.log('wrote', file)
}

await shot('d-header-absorbed', 'tn-d2-top')
await shot('d-header-absorbed', 'tn-d2-scrolled', { scroll: 180 })
await shot('a-empty', 'tn-a2-top')
await browser.close()
