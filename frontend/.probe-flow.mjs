import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
await p.goto('http://localhost:6006/iframe.html?id=shell-topnav-options--d-header-absorbed&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(2000)
console.log(JSON.stringify(await p.evaluate(() => {
  const sc = document.querySelector('.scrollbar-hide')
  const kids = [...sc.children]
  return {
    scroll: { clientH: sc.clientHeight, scrollH: sc.scrollHeight },
    children: kids.map(k => {
      const c = getComputedStyle(k)
      return {
        cls: k.className.toString().slice(0, 60),
        id: k.id || null,
        offsetH: k.offsetHeight,
        marginTop: c.marginTop,
        flowCost: k.offsetHeight + parseFloat(c.marginTop),
        z: c.zIndex, pos: c.position,
      }
    }),
  }
}), null, 2))
await b.close()
