import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
for (const id of ['d-header-absorbed','e-chatgpt-scale','f-product-scale']) {
  await p.goto(`http://localhost:6006/iframe.html?id=shell-topnav-options--${id}&globals=theme:light&viewMode=story`, { waitUntil: 'networkidle' })
  await p.waitForTimeout(1800)
  const r = await p.evaluate(() => {
    const g = (el) => { if (!el) return null; const c = getComputedStyle(el), r = el.getBoundingClientRect()
      return { h: +r.height.toFixed(0), font: `${c.fontSize}/${c.lineHeight} w${c.fontWeight}` } }
    const title = document.querySelector('[data-slot="breadcrumb-page"]')
    const trail = document.querySelector('[data-slot="breadcrumb-link"]')
    const trig = document.querySelector('[data-sidebar="trigger"]')
    const slot = document.getElementById('topnav-actions')
    const fact = slot?.firstElementChild
    const btn = slot?.querySelector('button')
    const gap = fact && btn ? +(btn.getBoundingClientRect().left - fact.getBoundingClientRect().right).toFixed(0) : null
    return { title: g(title), trail: g(trail), toggle: g(trig), fact: g(fact), button: g(btn), factToBtnGap: gap }
  })
  console.log(id.padEnd(20), JSON.stringify(r))
}
await b.close()
