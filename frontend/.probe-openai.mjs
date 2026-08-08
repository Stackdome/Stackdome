import { chromium } from 'playwright'
const OUT = '/private/tmp/claude-502/-Users-jaseem-Projects-Stackdome/cbfd4875-c83a-45f9-8873-8507a5609d98/scratchpad'
const browser = await chromium.launch({ headless: true })
const page = await browser.newPage({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 2,
  userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36',
})
await page.goto('https://chatgpt.com/', { waitUntil: 'domcontentloaded', timeout: 45000 })
await page.waitForTimeout(8000)

const data = await page.evaluate(() => {
  const desc = (el) => {
    if (!el) return null
    const b = el.getBoundingClientRect(), c = getComputedStyle(el)
    return {
      text: el.textContent.trim().slice(0, 24),
      tag: el.tagName,
      x: +b.x.toFixed(0), y: +b.y.toFixed(0), w: +b.width.toFixed(0), h: +b.height.toFixed(0),
      font: `${c.fontSize}/${c.lineHeight} w${c.fontWeight}`,
      color: c.color, bg: c.backgroundColor,
      pad: c.padding, gap: c.gap, radius: c.borderRadius,
      border: c.borderWidth === '0px' ? 'none' : `${c.borderWidth} ${c.borderColor}`,
      pos: c.position, z: c.zIndex,
    }
  }

  // Anything sitting on the top 60px band, wide enough to be real UI.
  const band = [...document.querySelectorAll('button, a, span, svg, h1, div')]
    .filter(e => {
      const b = e.getBoundingClientRect()
      return b.y >= 0 && b.y < 56 && b.height >= 14 && b.height <= 48 && b.width >= 14 && b.width < 400
    })
    .map(desc)

  // Sidebar: the nav column on the left
  const nav = document.querySelector('nav') || document.querySelector('[class*="sidebar"]')
  // The element that actually scrolls the conversation
  const scroller = [...document.querySelectorAll('div')].find(e => {
    const c = getComputedStyle(e)
    return (c.overflowY === 'auto' || c.overflowY === 'scroll') && e.getBoundingClientRect().width > 600
  })

  return {
    band,
    nav: desc(nav),
    scroller: desc(scroller),
    pageBg: getComputedStyle(document.body).backgroundColor,
    viewport: { w: innerWidth, h: innerHeight },
  }
})
console.log(JSON.stringify(data, null, 2))
await page.screenshot({ path: `${OUT}/oai-header.png`, clip: { x: 0, y: 0, width: 1440, height: 120 } })
await browser.close()
