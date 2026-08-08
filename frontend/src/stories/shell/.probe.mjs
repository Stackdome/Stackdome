import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
await p.goto('http://localhost:6006/iframe.html?id=shell-stacks-pass-2--proposal&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(1500)
console.log(await p.evaluate(() => {
  const row = document.querySelectorAll('main a')[0]
  const dot = row.firstElementChild
  const cs = getComputedStyle(dot)
  const main = document.querySelector('main > div')
  return {
    rowBox: row.getBoundingClientRect().toJSON(),
    dotBox: dot.getBoundingClientRect().toJSON(),
    dotBg: cs.backgroundColor, dotW: cs.width, dotDisplay: cs.display,
    containerBox: main.getBoundingClientRect().toJSON(),
    containerMaxW: getComputedStyle(main).maxWidth,
  }
}))
await b.close()
