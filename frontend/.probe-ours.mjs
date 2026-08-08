import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
await p.goto('http://localhost:6006/iframe.html?id=shell-topnav-options--d-header-absorbed&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(2200)
console.log(JSON.stringify(await p.evaluate(() => {
  const desc = (el, label) => {
    if (!el) return { label, missing: true }
    const r = el.getBoundingClientRect(), c = getComputedStyle(el)
    return {
      label, tag: el.tagName, text: el.textContent.trim().slice(0, 20),
      x: +r.x.toFixed(0), y: +r.y.toFixed(0), w: +r.width.toFixed(0), h: +r.height.toFixed(0),
      font: `${c.fontSize}/${c.lineHeight} w${c.fontWeight}`,
      color: c.color, bg: c.backgroundColor, radius: c.borderRadius, pad: c.padding,
      pos: c.position, z: c.zIndex,
    }
  }
  const bar = document.querySelector('.sticky.top-0')
  const trigger = document.querySelector('[data-sidebar="trigger"]')
  const crumbs = document.querySelector('[data-slot="breadcrumb-list"], nav[aria-label="breadcrumb"] ol')
  const last = crumbs?.lastElementChild
  const slot = document.getElementById('topnav-actions')
  const fact = slot?.firstElementChild
  const action = slot?.querySelector('button')
  return {
    bar: desc(bar, 'bar'),
    trigger: desc(trigger, 'sidebar toggle'),
    crumbFirst: desc(crumbs?.firstElementChild, 'crumb first'),
    crumbLast: desc(last, 'crumb last (title)'),
    fact: desc(fact, 'fact'),
    action: desc(action, 'action'),
    gapCrumbToSlot: slot && crumbs ? +(slot.getBoundingClientRect().left - crumbs.getBoundingClientRect().right).toFixed(0) : null,
    rightInset: slot ? +(document.querySelector('main,[data-slot="sidebar-inset"]').getBoundingClientRect().right - slot.getBoundingClientRect().right).toFixed(0) : null,
  }
}), null, 2))
await b.close()
