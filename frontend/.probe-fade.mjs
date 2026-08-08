import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
await p.goto('http://localhost:6006/iframe.html?id=shell-topnav-options--d-header-absorbed&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(2000)
for (const y of [0, 180]) {
  const r = await p.evaluate((y) => {
    const sc = document.querySelector('.scrollbar-hide')
    sc.scrollTop = y
    const bar = sc.children[0].getBoundingClientRect()
    const fade = sc.children[1].getBoundingClientRect()
    return { scrollTop: sc.scrollTop, barBottom: +bar.bottom.toFixed(1), fadeTop: +fade.top.toFixed(1), fadeBottom: +fade.bottom.toFixed(1) }
  }, y)
  console.log(JSON.stringify(r))
}
await b.close()
